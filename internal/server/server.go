package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/database"
	"github.com/xingmang/xingmang-desktop/internal/handlers"
	"github.com/xingmang/xingmang-desktop/internal/sharestore"
)

// BuildInfo 是构建期注入的版本信息（见 cmd/xingmang/main.go 的 ldflags）。
//
// 它通过 /api/health 暴露，供**新实例探测旧实例版本**使用：
// 旧进程没退出时用户双击新 exe，若直接复用旧实例，用户看到的会是旧界面 + 旧数据，
// 表现为「更新了但好像没生效」。带上版本号后，新 exe 才能识别并明确提示用户。
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// shutdownGracePeriod 是优雅关闭的等待上限。超过则强制断开，避免残留连接把退出卡住。
const shutdownGracePeriod = 5 * time.Second

// Start 启动 HTTP 服务并**阻塞直到退出**，负责全部收尾工作（HTTP 优雅关闭 + 数据库关闭）。
//
// 退出有两条路径，都会让 ctx 结束：
//   - 调用方收到 SIGINT/SIGTERM 后自行 cancel；
//   - 前端点「退出程序」→ POST /api/shutdown → 本函数调用 shutdown() 触发。
//
// shutdown 由调用方提供（通常是派生 ctx 的 cancel），因此是幂等的。
func Start(ctx context.Context, addr string, devMode bool, webFS embed.FS, build BuildInfo, shutdown context.CancelFunc) {
	db, err := database.Init("")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	// 不在此 defer 关闭：必须等 HTTP 服务停止、在途请求写完之后再关闭，
	// 否则正在处理的长请求会被抽掉数据库连接。收尾统一放在函数末尾。

	// 一次性种入系统引导帖（原来写死在模板里的两条假置顶，现在是真帖子）。
	// 幂等：靠 app_meta 标记，只做一次；用户删掉之后不会再复活。
	//
	// ⚠⚠ 必须排在「取 maxShareID / sharestore.New」**之前**：
	// seed 往库里插了行会抬高 max(id)，而文件帖的 id 水位是从这一刻的 max 起算的。
	// 顺序反了会让新帖复用 seed 帖的 id → 文件帖和库内帖同 id，
	// 表现为列表里同一篇出现两次、置顶计数漏算（已踩过一次，别再调回去）。
	if _, err := handlers.SeedSystemPosts(db); err != nil {
		log.Printf("系统引导帖种入失败（不影响使用）: %v", err)
	}

	// 帖子文件存储：水位取库里现有最大 id，保证文件帖 id 不与旧帖冲突
	var maxShareID int
	_ = db.QueryRow(`SELECT COALESCE(MAX(id), 0) FROM experience_shares`).Scan(&maxShareID)
	store, err := sharestore.New("", maxShareID)
	if err != nil {
		log.Printf("帖子文件存储初始化失败（新帖将退回数据库）: %v", err)
		store = nil
	}

	// 一次性迁移：早期文件帖的 JSON 里没有 board 键，补成 experience（幂等）。
	// 库里存量由 database.migrateExperienceSharesColumns 刷平，这里只管文件帖。
	if store != nil {
		if n, err := store.BackfillBoard(); err != nil {
			log.Printf("存量文件帖补 board 失败（已处理 %d 条）: %v", n, err)
		} else if n > 0 {
			log.Printf("已为 %d 条存量文件帖补上 board=experience", n)
		}
	}

	h := handlers.NewWithShareStore(db, store)
	mux := http.NewServeMux()

	// 一次性数据修复：早期"换封面"只认 data URL，AI 生成封面返回的站内 URL 被
	// 当成字符串写进了 books.cover —— 封面因此"住"在 data/generated-images 里，
	// 作品文件夹内的 gallery/ 永远是空的。这里在开始服务前把它们搬成真正的文件。
	// 幂等：搬完 cover 就空了，下次启动不会再命中。
	handlers.HealLegacyCovers(db)
	// 再把缓存目录里剩下的生成图按归属搬进 books/<书名>/gallery/（认不出的进 _orphan-images/）。
	// 必须在 HealLegacyCovers 之后：那一步会把"URL 形态的封面"落成文件，
	// 这一步才能靠内容指纹认出这些图属于谁。同样幂等。
	handlers.MigrateGeneratedImages(db)

	// 记忆落盘的后台合并写入（docs/books-memory-storage.md §8.1）。
	// 对话/图谱的写操作只标记脏，由它按 2 秒间隔合并写盘。
	//
	// 这里注册 Agent 会话的渲染函数：Agent 包只负责 MarkMemoryDirty（不引 bookstore 的反向依赖），
	// 真正"账本 → Markdown"的渲染放在 bookstore 里，保证内容与账本一致。
	svc := bookstore.NewService(db)
	bookstore.RegisterMemoryRenderer(bookstore.MemoryKindAgent, svc.MirrorAgentSession)
	// 图谱的 key 有两种：版本 id，或 "book:<书 id>"（未分版本的历史数据）
	bookstore.RegisterMemoryRenderer(bookstore.MemoryKindGraph, svc.MirrorGraphByKey)
	bookstore.StartMemoryFlusher()

	// API routes - Go 1.22+ pattern matching
	// Note: some routes use manual dispatch to avoid wildcard conflicts

	// Books
	mux.HandleFunc("GET /api/books", h.GetBooks)
	mux.HandleFunc("GET /api/books/{id}", h.GetBook)
	mux.HandleFunc("POST /api/books", h.CreateBook)
	mux.HandleFunc("PUT /api/books/{id}", h.UpdateBook)
	mux.HandleFunc("DELETE /api/books/{id}", h.DeleteBook)
	// 封面 / 画廊 / 在文件管理器中定位（作品文件夹化，见 docs/books-folder-storage.md）
	mux.HandleFunc("GET /api/books/{id}/cover", h.GetBookCover)
	mux.HandleFunc("GET /api/books/{id}/gallery", h.GetBookGallery)
	mux.HandleFunc("POST /api/books/{id}/gallery", h.AddGalleryImage)
	mux.HandleFunc("GET /api/books/{id}/gallery/{name}", h.GetGalleryImage)
	mux.HandleFunc("DELETE /api/books/{id}/gallery/{name}", h.DeleteGalleryImage)
	mux.HandleFunc("POST /api/books/{id}/gallery/activate", h.ActivateBookCover)
	mux.HandleFunc("POST /api/books/{id}/reveal", h.RevealBook)
	mux.HandleFunc("POST /api/books/{id}/rescan", h.RescanBook)
	mux.HandleFunc("POST /api/books/import-folder", h.ImportFolder)
	mux.HandleFunc("POST /api/books/import-json", h.ImportBookJSON)
	// 作品库目录（设置）
	mux.HandleFunc("GET /api/settings/books-root", h.GetBooksRoot)
	mux.HandleFunc("PUT /api/settings/books-root", h.SetBooksRoot)
	mux.HandleFunc("POST /api/settings/pick-folder", h.PickFolder)
	mux.HandleFunc("POST /api/settings/books-root/preview", h.PreviewBooksRoot)
	mux.HandleFunc("PUT /api/settings/books-root/link", h.LinkBooksRoot)
	mux.HandleFunc("POST /api/books/rebuild-index", h.RebuildBookIndex)
	mux.HandleFunc("GET /api/books/{id}/export", h.ExportBook)
	// 拆书 / 生成历史（memory_records）+ 记忆概览
	mux.HandleFunc("GET /api/books/{id}/memories", h.GetBookMemories)
	mux.HandleFunc("POST /api/books/{id}/memories", h.CreateMemoryRecord)
	mux.HandleFunc("GET /api/books/{id}/memory/summary", h.GetBookMemorySummary)
	mux.HandleFunc("GET /api/memories", h.GetAllMemories)
	mux.HandleFunc("GET /api/memories/{id}", h.GetMemoryRecord)
	mux.HandleFunc("PUT /api/memories/{id}", h.UpdateMemoryRecord)
	mux.HandleFunc("DELETE /api/memories/{id}", h.DeleteMemoryRecord)
	// 书内自由文档（大纲 / 灵感）
	mux.HandleFunc("GET /api/books/{id}/docs", h.GetBookDocs)
	mux.HandleFunc("POST /api/books/{id}/docs", h.CreateBookDoc)
	mux.HandleFunc("PUT /api/docs/{id}", h.UpdateBookDoc)
	mux.HandleFunc("DELETE /api/docs/{id}", h.DeleteBookDoc)

	// Chapters (manual dispatch for book/{bookId} vs {id} conflict)
	mux.HandleFunc("GET /api/chapters", h.NotFound)
	mux.HandleFunc("POST /api/chapters", h.CreateChapter)
	mux.HandleFunc("GET /api/chapters/", h.DispatchChapters)
	mux.HandleFunc("POST /api/chapters/", h.DispatchChapters)
	mux.HandleFunc("PUT /api/chapters/{id}", h.UpdateChapter)
	mux.HandleFunc("DELETE /api/chapters/{id}", h.DeleteChapter)

	// Prompts
	mux.HandleFunc("GET /api/prompts", h.GetPrompts)
	mux.HandleFunc("GET /api/prompts/{id}", h.GetPrompt)
	mux.HandleFunc("POST /api/prompts", h.CreatePrompt)
	mux.HandleFunc("PUT /api/prompts/{id}", h.UpdatePrompt)
	mux.HandleFunc("DELETE /api/prompts/{id}", h.DeletePrompt)

	// 卡片使用心得（一张卡片多条，时间线）
	// ⚠ 不能把更新/删除写成 /api/prompts/insights/{id}：它会和 /api/prompts/{id}/insights
	// 在 "prompts/insights/insights" 这类路径上重叠，Go 1.22 的 ServeMux 注册时直接 panic。
	mux.HandleFunc("GET /api/prompt-insights", h.GetAllPromptInsights)
	mux.HandleFunc("GET /api/prompts/{id}/insights", h.GetPromptInsights)
	mux.HandleFunc("POST /api/prompts/{id}/insights", h.CreatePromptInsight)
	mux.HandleFunc("PUT /api/prompt-insights/{id}", h.UpdatePromptInsight)
	mux.HandleFunc("DELETE /api/prompt-insights/{id}", h.DeletePromptInsight)

	// Memos (manual dispatch for search vs {id} conflict)
	mux.HandleFunc("GET /api/memos", h.GetMemos)
	mux.HandleFunc("POST /api/memos", h.CreateMemo)
	mux.HandleFunc("GET /api/memos/", h.DispatchMemos)
	mux.HandleFunc("PUT /api/memos/{id}", h.UpdateMemo)
	mux.HandleFunc("DELETE /api/memos/{id}", h.DeleteMemo)

	// 工作流画布（节点式工作流；执行在前端，Go 只做持久化 + 落盘）
	mux.HandleFunc("GET /api/workflows", h.GetWorkflows)
	mux.HandleFunc("POST /api/workflows", h.CreateWorkflow)
	mux.HandleFunc("GET /api/workflows/{id}", h.GetWorkflow)
	mux.HandleFunc("PUT /api/workflows/{id}", h.UpdateWorkflow)
	mux.HandleFunc("DELETE /api/workflows/{id}", h.DeleteWorkflow)

	mux.HandleFunc("GET /api/workflow-runs", h.ListWorkflowRuns)
	mux.HandleFunc("POST /api/workflow-runs", h.CreateWorkflowRun)
	mux.HandleFunc("PUT /api/workflow-runs/{id}", h.UpdateWorkflowRun)
	mux.HandleFunc("DELETE /api/workflow-runs/{id}", h.DeleteWorkflowRun)

	// 执行落盘：子路径与集合路径同前缀，均为字面量，注册期无冲突
	mux.HandleFunc("GET /api/workflow-outputs", h.ListWorkflowOutputs)
	mux.HandleFunc("POST /api/workflow-outputs", h.WriteWorkflowOutput)
	mux.HandleFunc("GET /api/workflow-outputs/read", h.ReadWorkflowOutput)
	mux.HandleFunc("GET /api/workflow-outputs/root", h.RunOutputDir)
	mux.HandleFunc("POST /api/workflow-outputs/mkdir", h.MkdirWorkflowOutput)

	// Providers
	mux.HandleFunc("GET /api/providers", h.GetProviders)
	mux.HandleFunc("GET /api/providers/{id}", h.GetProvider)
	mux.HandleFunc("GET /api/providers/{id}/models", h.DiscoverModels)
	mux.HandleFunc("POST /api/providers", h.CreateProvider)
	mux.HandleFunc("PUT /api/providers/{id}", h.UpdateProvider)
	mux.HandleFunc("DELETE /api/providers/{id}", h.DeleteProvider)

	// Config (manual dispatch for provider/{providerId} vs {id} conflict)
	mux.HandleFunc("GET /api/config", h.GetAllConfigs)
	mux.HandleFunc("POST /api/config", h.CreateConfig)
	mux.HandleFunc("GET /api/config/", h.DispatchConfig)
	mux.HandleFunc("PUT /api/config/", h.DispatchConfig)
	mux.HandleFunc("DELETE /api/config/{id}", h.DeleteConfig)
	mux.HandleFunc("POST /api/config/{id}/test", h.TestModel)

	// AI
	mux.HandleFunc("POST /api/ai/chat", h.AIChat)
	// 自适应 max_tokens 的探测结果（只读）。供界面展示「实际探测到的输出上限」——
	// 否则这个值只存在于日志里，用户没法判断某个模型到底探到了多少。
	mux.HandleFunc("GET /api/ai/max-tokens-cache", h.GetMaxTokensCache)
	// 单模型探测：导入弹窗的「测试选中」与模型行内测试共用。
	// 支持"尚未落库"的 provider_id + model，因此不挂在 /config/{id} 下。
	mux.HandleFunc("POST /api/ai/probe", h.ProbeModel)
	mux.HandleFunc("POST /api/ai/generate-description", h.GenerateDescription)
	mux.HandleFunc("POST /api/ai/generate-image", h.GenerateImage)
	mux.HandleFunc("POST /api/ai/recognize-characters", h.RecognizeCharacters)

	// 工作流媒体节点：生图 / 图片分析（视频类待接入具体服务商）
	mux.HandleFunc("POST /api/media/image", h.GenerateMediaImage)
	mux.HandleFunc("POST /api/media/analyze", h.AnalyzeMediaImage)
	mux.HandleFunc("POST /api/media/video", h.GenerateMediaVideo)
	mux.HandleFunc("POST /api/media/image-to-video", h.MediaImageToVideo)

	// 工作流文件输入节点：只读文本（UTF-8 / GBK 自动识别）
	mux.HandleFunc("POST /api/files/read-text", h.ReadTextFile)
	mux.HandleFunc("GET /api/ai/rank-reference/categories", h.GetRankReferenceCategories)
	mux.HandleFunc("GET /api/ai/rank-reference/books", h.GetRankReferenceBooks)

	// 榜单数据源（AI 扫榜）：sources 返回可用数据源，lists/books 按 ?source= 取数据，
	// drops 返回掉榜名单（仅可穷尽的数据源有值）
	mux.HandleFunc("GET /api/ai/rank/sources", h.GetRankSources)
	mux.HandleFunc("GET /api/ai/rank/lists", h.GetRankLists)
	mux.HandleFunc("GET /api/ai/rank/books", h.GetRankBooks)
	mux.HandleFunc("GET /api/ai/rank/drops", h.GetRankDrops)
	mux.HandleFunc("GET /api/ai/rank/book-chapters", h.GetRankBookChapters)

	// Conversations (manual dispatch for book/{bookId} vs {id}/messages conflict)
	mux.HandleFunc("GET /api/conversations", h.GetConversations)
	mux.HandleFunc("POST /api/conversations", h.CreateConversation)
	mux.HandleFunc("GET /api/conversations/", h.DispatchConversations)
	mux.HandleFunc("POST /api/conversations/", h.DispatchConversations)
	mux.HandleFunc("PUT /api/conversations/", h.DispatchConversations)
	mux.HandleFunc("DELETE /api/conversations/", h.DispatchConversations)

	// Generators
	mux.HandleFunc("GET /api/generators", h.GetGenerators)
	mux.HandleFunc("GET /api/generators/{id}", h.GetGenerator)
	mux.HandleFunc("POST /api/generators", h.CreateGenerator)
	mux.HandleFunc("PUT /api/generators/{id}", h.UpdateGenerator)
	mux.HandleFunc("DELETE /api/generators/{id}", h.DeleteGenerator)

	// Characters (manual dispatch for book/{bookId} vs {id} conflict)
	mux.HandleFunc("GET /api/characters", h.NotFound)
	mux.HandleFunc("POST /api/characters", h.CreateCharacter)
	mux.HandleFunc("GET /api/characters/", h.DispatchCharacters)
	mux.HandleFunc("PUT /api/characters/{id}", h.UpdateCharacter)
	mux.HandleFunc("DELETE /api/characters/{id}", h.DeleteCharacter)

	// Stats
	mux.HandleFunc("GET /api/stats/overview", h.GetOverview)
	mux.HandleFunc("GET /api/stats/daily", h.GetDaily)
	mux.HandleFunc("GET /api/stats/model-stats", h.GetModelStats)
	mux.HandleFunc("GET /api/stats/monthly", h.GetMonthly)
	mux.HandleFunc("POST /api/stats/record", h.RecordUsage)

	// Volumes (manual dispatch for book/{bookId} vs {id} conflict)
	mux.HandleFunc("GET /api/volumes", h.NotFound)
	mux.HandleFunc("POST /api/volumes", h.CreateVolume)
	mux.HandleFunc("GET /api/volumes/", h.DispatchVolumes)
	mux.HandleFunc("PUT /api/volumes/{id}", h.UpdateVolume)
	mux.HandleFunc("DELETE /api/volumes/{id}", h.DeleteVolume)

	// Experience Shares
	mux.HandleFunc("GET /api/experience-shares", h.GetAllExperienceShares)
	mux.HandleFunc("GET /api/experience-shares/export", h.ExportAll)
	mux.HandleFunc("GET /api/experience-shares/{id}", h.GetExperienceShare)
	mux.HandleFunc("POST /api/experience-shares", h.CreateExperienceShare)
	mux.HandleFunc("POST /api/experience-shares/import", h.ImportAll)
	mux.HandleFunc("POST /api/experience-shares/import-pdf", h.ImportPdf)
	mux.HandleFunc("PUT /api/experience-shares/{id}", h.UpdateExperienceShare)
	mux.HandleFunc("DELETE /api/experience-shares/{id}", h.DeleteExperienceShare)
	// 分区/精华/点赞：子资源用独立后缀，避免与 {id} 路由在注册期冲突
	mux.HandleFunc("POST /api/experience-shares/{id}/like", h.LikeExperienceShare)
	mux.HandleFunc("POST /api/experience-shares/{id}/feature", h.SetExperienceShareFeatured)
	mux.HandleFunc("POST /api/experience-shares/{id}/pin", h.SetExperienceSharePinned)
	mux.HandleFunc("POST /api/experience-shares/{id}/section", h.SetExperienceShareSection)
	mux.HandleFunc("POST /api/experience-shares/batch-section", h.BatchUpdateExperienceShareSection)

	// 经验分享封面图（本地落盘 + 静态访问 + 外链图片代理）
	mux.HandleFunc("GET /api/experience-covers/{file}", h.ServeExperienceCover)
	mux.HandleFunc("GET /api/image-proxy", h.ProxyImage)

	// 正文插图（本地落盘 + 静态访问）
	mux.HandleFunc("POST /api/images", h.UploadImage)
	mux.HandleFunc("GET /api/editor-images/{file}", h.ServeEditorImage)
	mux.HandleFunc("GET /api/generated-images/{file}", h.ServeGeneratedImage)

	// Experience Sections（贴吧式自定义分区，上限 15）
	mux.HandleFunc("GET /api/experience-sections", h.GetExperienceSections)
	mux.HandleFunc("POST /api/experience-sections", h.CreateExperienceSection)
	mux.HandleFunc("POST /api/experience-sections/reorder", h.ReorderExperienceSections)
	mux.HandleFunc("PUT /api/experience-sections/{id}", h.UpdateExperienceSection)
	mux.HandleFunc("DELETE /api/experience-sections/{id}", h.DeleteExperienceSection)

	// Knowledge Graph (manual dispatch for entities/book/{bookId} vs entities/{id})
	mux.HandleFunc("GET /api/knowledge-graph/entities", h.NotFound)
	mux.HandleFunc("POST /api/knowledge-graph/entities", h.CreateEntity)
	mux.HandleFunc("GET /api/knowledge-graph/entities/", h.DispatchKgEntities)
	mux.HandleFunc("PUT /api/knowledge-graph/entities/{id}", h.UpdateEntity)
	mux.HandleFunc("DELETE /api/knowledge-graph/entities/{id}", h.DeleteEntity)
	mux.HandleFunc("PUT /api/knowledge-graph/entities/{id}/chapters", h.SetEntityChapters)

	mux.HandleFunc("GET /api/knowledge-graph/relations", h.NotFound)
	mux.HandleFunc("POST /api/knowledge-graph/relations", h.CreateRelation)
	mux.HandleFunc("GET /api/knowledge-graph/relations/book/{bookId}", h.GetRelations)
	mux.HandleFunc("PUT /api/knowledge-graph/relations/{id}", h.UpdateRelation)
	mux.HandleFunc("DELETE /api/knowledge-graph/relations/{id}", h.DeleteRelation)
	mux.HandleFunc("PUT /api/knowledge-graph/relations/{id}/chapters", h.SetRelationChapters)

	mux.HandleFunc("GET /api/knowledge-graph/chapter-refs/{bookId}", h.GetChapterRefs)

	mux.HandleFunc("GET /api/knowledge-graph/graph/{bookId}", h.GetGraph)
	mux.HandleFunc("DELETE /api/knowledge-graph/graph/{bookId}", h.ClearGraph)
	mux.HandleFunc("POST /api/knowledge-graph/graph/import", h.ImportGraph)
	mux.HandleFunc("POST /api/knowledge-graph/analyze", h.AnalyzeGraph)
	mux.HandleFunc("POST /api/knowledge-graph/ai-supplement", h.AISupplement)
	mux.HandleFunc("GET /api/knowledge-graph/versions/{bookId}", h.GetVersions)
	mux.HandleFunc("POST /api/knowledge-graph/versions", h.CreateVersion)
	mux.HandleFunc("PUT /api/knowledge-graph/versions/{id}", h.RenameVersion)
	mux.HandleFunc("DELETE /api/knowledge-graph/versions/{bookId}/{versionId}", h.DeleteVersion)
	mux.HandleFunc("GET /api/knowledge-graph/version/{versionId}", h.GetVersionData)

	// 实体合并系统（主词条 / 副词条）
	mux.HandleFunc("POST /api/knowledge-graph/entities/merge", h.MergeEntities)
	mux.HandleFunc("POST /api/knowledge-graph/entities/unmerge", h.UnmergeEntities)
	mux.HandleFunc("GET /api/knowledge-graph/merge-candidates/{bookId}", h.GetMergeCandidates)

	// Entries (manual dispatch for categories vs {id} conflict)
	mux.HandleFunc("GET /api/entries", h.GetEntries)
	mux.HandleFunc("POST /api/entries", h.CreateEntry)
	mux.HandleFunc("GET /api/entries/", h.DispatchEntries)
	mux.HandleFunc("PUT /api/entries/{id}", h.UpdateEntry)
	mux.HandleFunc("DELETE /api/entries/{id}", h.DeleteEntry)

	// 星芒 Agent（独立板块，Go 原生内核）
	mux.HandleFunc("GET /api/agent/models", h.AgentModels)
	mux.HandleFunc("GET /api/agent/books", h.AgentBookChoices)
	mux.HandleFunc("GET /api/agent/sessions", h.AgentSessionsList)
	mux.HandleFunc("POST /api/agent/sessions", h.AgentSessionCreate)
	mux.HandleFunc("GET /api/agent/sessions/{id}/stream", h.AgentStream)
	mux.HandleFunc("GET /api/agent/sessions/{id}", h.AgentSessionGet)
	mux.HandleFunc("DELETE /api/agent/sessions/{id}", h.AgentSessionDelete)
	mux.HandleFunc("POST /api/agent/sessions/{id}/runs", h.AgentRunStart)
	mux.HandleFunc("POST /api/agent/runs/{id}/cancel", h.AgentRunCancel)
	mux.HandleFunc("DELETE /api/agent/runs/{id}", h.AgentRunDelete)
	mux.HandleFunc("PUT /api/agent/runs/{id}/answer", h.AgentRunEditAnswer)
	mux.HandleFunc("PATCH /api/agent/sessions/{id}", h.AgentSessionPatch)
	mux.HandleFunc("POST /api/agent/proposals/{id}/approve", h.AgentProposalApprove)
	mux.HandleFunc("POST /api/agent/proposals/{id}/reject", h.AgentProposalReject)
	mux.HandleFunc("POST /api/agent/proposals/{id}/rollback", h.AgentProposalRollback)

	// Health：同时作为「单实例探测 + 版本握手」接口。
	// 新增程序启动时会先请求这里，比对 version 决定是复用还是拒绝启动。
	// 禁止缓存，避免拿到别的实例的陈旧结果。
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"message":   "服务运行正常",
			"version":   build.Version,
			"commit":    build.Commit,
			"buildTime": build.BuildTime,
		})
	})

	// 退出程序。界面上的「退出程序」按钮会打到这里。
	//
	// ⚠ 必须校验来源：服务默认绑定所有网卡（:19825），局域网里任何一台设备都能访问。
	// 若不校验，别人可以直接把你正在使用的程序关掉、甚至反复关着玩。
	// 这里只接受回环地址或本机网卡地址，并要求 POST（GET 会被浏览器预取/图片标签误触发）。
	mux.HandleFunc("POST /api/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if !isLocalRequest(r) {
			log.Printf("拒绝来自 %s 的退出请求：非本机来源", r.RemoteAddr)
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"success":false,"message":"仅允许在本机退出程序"}`))
			return
		}

		log.Printf("收到退出请求（来自 %s），程序即将退出", r.RemoteAddr)
		_, _ = w.Write([]byte(`{"success":true,"message":"程序正在退出"}`))

		// 先把响应真正发给浏览器，再触发关闭：
		// 否则服务可能先停掉，前端只会看到请求失败，没法显示「已退出」提示。
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		go func() {
			time.Sleep(300 * time.Millisecond)
			shutdown()
		}()
	})

	// Static files (if not dev mode)
	if !devMode {
		subFS, err := fs.Sub(webFS, "web/dist")
		if err != nil {
			log.Fatalf("Failed to create sub filesystem: %v", err)
		}
		fileServer := http.FileServer(http.FS(subFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}

			reqPath := r.URL.Path
			if reqPath == "/" {
				reqPath = "/index.html"
			}
			rel := strings.TrimPrefix(reqPath, "/")

			if _, err := fs.Stat(subFS, rel); err != nil {
				// ⚠ SPA fallback 不能无差别生效，否则会「用 HTML 冒充 JS」：
				// 换版本后浏览器还拿着旧的 index.html，去请求已改名的 chunk，
				// 这里若回吐 index.html，浏览器把 HTML 当 ES module 解析 →
				// SyntaxError → 路由懒加载失败 → 整个界面白屏，且 Network 面板
				// 全是 200，极难排查（实测过：/assets/不存在.js → 200 text/html）。
				// 因此：只对「外观像静态资源」的路径返回真 404，其余才回退到首页。
				if isStaticAssetPath(reqPath) {
					http.NotFound(w, r)
					return
				}
				r.URL.Path = "/"
				// 首页禁止缓存：否则上述「旧 HTML 配新 chunk」的场景会因浏览器
				// 启发式缓存而放大（无 Cache-Control 时浏览器可自行决定缓存多久）。
				w.Header().Set("Cache-Control", "no-cache")
				fileServer.ServeHTTP(w, r)
				return
			}

			// 带内容 hash 的构建产物可长期强缓存；文件名一变 URL 就变，无需担心失效。
			if isHashedAssetPath(reqPath) {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if reqPath == "/index.html" {
				w.Header().Set("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	log.Printf("服务已启动: http://localhost%s", addr)

	srv := &http.Server{Addr: addr, Handler: mux}
	srvErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("正在关闭服务...")
	case err := <-srvErr:
		log.Printf("服务异常终止: %v", err)
	}

	// 优雅关闭：给在途请求留出完成时间（AI 对话这类长请求可能还在写库）
	shutCtx, cancelShut := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer cancelShut()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("HTTP 服务未能在 %s 内关闭（%v），强制断开", shutdownGracePeriod, err)
		_ = srv.Close()
	}

	// 先把记忆层没落盘的改动 flush 出去，再关数据库：顺序反了就会丢掉最后一次编辑。
	if n, err := bookstore.StopMemoryFlusher(2 * time.Second); err != nil {
		log.Printf("记忆文件收尾写入未完全成功（已写 %d 条）: %v", n, err)
	}

	// HTTP 已停止，此时关闭数据库不会打断任何写操作；Close 内部会做 WAL checkpoint
	database.Close()
	log.Println("服务已停止")
}

// isLocalRequest 判断请求是否来自本机。用于保护 /api/shutdown 这类会终止进程的接口。
//
// 放行的情形：
//   - 来源地址是回环地址（127.0.0.1 / ::1）；
//   - 来源地址是本机某块网卡的地址（用户用局域网 IP 打开自己的界面，源地址不是回环）；
//   - Origin 头存在时，其主机也必须是本机（挡住其他网站发起的跨站请求）。
func isLocalRequest(r *http.Request) bool {
	if !isLocalHost(r.RemoteAddr) {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || !isLocalHost(u.Host) {
			return false
		}
	}
	return true
}

// isLocalHost 判断 host/host:port 形式的主机是否属于本机。
func isLocalHost(hostport string) bool {
	if hostport == "" {
		return false
	}
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
	}
	host = strings.Trim(host, "[]")

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return true
		}
		return isLocalInterfaceIP(ip)
	}
	// 主机名形式（浏览器里的 localhost）
	return strings.EqualFold(host, "localhost")
}

// isLocalInterfaceIP 判断 ip 是否为本机某块网卡上配置的地址。
func isLocalInterfaceIP(ip net.IP) bool {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		switch v := a.(type) {
		case *net.IPNet:
			if v.IP.Equal(ip) {
				return true
			}
		case *net.IPAddr:
			if v.IP.Equal(ip) {
				return true
			}
		}
	}
	return false
}

// staticAssetExts 是按扩展名判定的「静态资源」白名单。
// 命中这些扩展名却在磁盘上找不到时，必须返回 404，绝不能回退到 index.html ——
// 否则浏览器会把 HTML 当成对应类型解析（JS 报 SyntaxError、CSS 静默失效），
// 表现为难以定位的白屏。
var staticAssetExts = map[string]bool{
	".js": true, ".mjs": true, ".cjs": true,
	".css": true, ".map": true,
	".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".svg": true, ".ico": true, ".bmp": true, ".avif": true,
	".json": true, ".wasm": true, ".mp4": true, ".webm": true, ".mp3": true,
}

// isStaticAssetPath 判断请求路径是否「看起来是一个静态资源文件」。
// 判定依据是扩展名，而不是目录名 —— 这样 /assets/xxx.js 与 /docs/a.png
// 都能覆盖，且不会误伤 /books 这类无扩展名的前端路由。
func isStaticAssetPath(p string) bool {
	return staticAssetExts[strings.ToLower(path.Ext(p))]
}

// isHashedAssetPath 判断是否为「带内容 hash 的构建产物」，可安全长期强缓存。
// Vite 产出的资源形如 /assets/index-LxxifY_P.js、/assets/Books-DyeXoVbv.css，
// 文件名里嵌了 8 位以上 hash，内容一变文件名就变，因此可以 immutable。
func isHashedAssetPath(p string) bool {
	if !strings.HasPrefix(p, "/assets/") {
		return false
	}
	base := path.Base(p)
	if !isStaticAssetPath(base) {
		return false
	}
	name := strings.TrimSuffix(base, path.Ext(base))
	// 从右往左取最后一个 '-' 之后的部分，判断是否够长且像 hash
	idx := strings.LastIndex(name, "-")
	if idx < 0 {
		return false
	}
	tail := name[idx+1:]
	if len(tail) < 8 {
		return false
	}
	for _, c := range tail {
		isAlnum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
		if !isAlnum {
			return false
		}
	}
	return true
}

// parsePathSegments extracts segments from the URL path after a prefix
func parsePathSegments(r *http.Request, prefix string) []string {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

// parseID parses an integer ID from a string
func parseID(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil
}
