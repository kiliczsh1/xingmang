package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/database"
	"github.com/xingmang/xingmang-desktop/internal/handlers"
)

// Start starts the HTTP server
func Start(addr string, devMode bool, webFS embed.FS) {
	db, err := database.Init("")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	h := handlers.New(db)
	mux := http.NewServeMux()

	// API routes - Go 1.22+ pattern matching
	// Note: some routes use manual dispatch to avoid wildcard conflicts

	// Books
	mux.HandleFunc("GET /api/books", h.GetBooks)
	mux.HandleFunc("GET /api/books/{id}", h.GetBook)
	mux.HandleFunc("POST /api/books", h.CreateBook)
	mux.HandleFunc("PUT /api/books/{id}", h.UpdateBook)
	mux.HandleFunc("DELETE /api/books/{id}", h.DeleteBook)

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

	// Memos (manual dispatch for search vs {id} conflict)
	mux.HandleFunc("GET /api/memos", h.GetMemos)
	mux.HandleFunc("POST /api/memos", h.CreateMemo)
	mux.HandleFunc("GET /api/memos/", h.DispatchMemos)
	mux.HandleFunc("PUT /api/memos/{id}", h.UpdateMemo)
	mux.HandleFunc("DELETE /api/memos/{id}", h.DeleteMemo)

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
	mux.HandleFunc("POST /api/ai/generate-description", h.GenerateDescription)
	mux.HandleFunc("POST /api/ai/generate-image", h.GenerateImage)
	mux.HandleFunc("POST /api/ai/recognize-characters", h.RecognizeCharacters)
	mux.HandleFunc("GET /api/ai/rank-reference/categories", h.GetRankReferenceCategories)
	mux.HandleFunc("GET /api/ai/rank-reference/books", h.GetRankReferenceBooks)

	// 番茄真实榜单（AI 扫榜）
	mux.HandleFunc("GET /api/ai/rank/lists", h.GetRankLists)
	mux.HandleFunc("GET /api/ai/rank/books", h.GetRankBooks)

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

	// Workflows（工作流工作台）
	// 注意路由顺序无关，Go 1.22 ServeMux 按「更具体的字面量优先」匹配，
	// 所以 /api/workflows/node-library/* 不会被 /api/workflows/{id} 抢走。
	mux.HandleFunc("GET /api/workflows", h.GetWorkflows)
	mux.HandleFunc("POST /api/workflows", h.CreateWorkflow)
	mux.HandleFunc("GET /api/workflows/node-library/all", h.GetWorkflowNodeLibrary)
	mux.HandleFunc("POST /api/workflows/node-library", h.SaveWorkflowNodeLibraryItem)
	mux.HandleFunc("DELETE /api/workflows/node-library/{id}", h.DeleteWorkflowNodeLibraryItem)
	mux.HandleFunc("GET /api/workflows/{id}", h.GetWorkflow)
	mux.HandleFunc("PUT /api/workflows/{id}", h.UpdateWorkflow)
	mux.HandleFunc("DELETE /api/workflows/{id}", h.DeleteWorkflow)

	// Entries (manual dispatch for categories vs {id} conflict)
	mux.HandleFunc("GET /api/entries", h.GetEntries)
	mux.HandleFunc("POST /api/entries", h.CreateEntry)
	mux.HandleFunc("GET /api/entries/", h.DispatchEntries)
	mux.HandleFunc("PUT /api/entries/{id}", h.UpdateEntry)
	mux.HandleFunc("DELETE /api/entries/{id}", h.DeleteEntry)

	// Health
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"服务运行正常"}`))
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
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			if _, err := fs.Stat(subFS, strings.TrimPrefix(path, "/")); err != nil {
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	log.Printf("服务已启动: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
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
