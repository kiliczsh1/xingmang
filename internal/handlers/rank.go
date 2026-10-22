package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/rank"
	"github.com/xingmang/xingmang-desktop/internal/ranksnap"
)

// 整榜抓取（用于可信的「新上榜 / 掉榜」）的两道闸：
// 页数上限兜住单次成本，每日次数上限兜住总成本（12 个榜各一次）。
// 一次整榜要翻几十页，不能任由前端触发；是否值得抓由 FullCapturer 自己判断。
const (
	fullCaptureMaxPages = 60
	fullCaptureDailyCap = 12
)

// fullCaptureState 记录当天已投递过整榜抓取的榜，避免重复投递。
var fullCaptureState = struct {
	sync.Mutex
	day   string
	keys  map[string]bool
	count int
}{keys: map[string]bool{}}

// GetRankSources 返回可用榜单数据源（前端右上角切换器 + 页脚来源说明）。
// 数据源在 internal/rank 的注册表里维护，新增站点不需要改前端。
func (h *Handler) GetRankSources(w http.ResponseWriter, r *http.Request) {
	respondOK(w, map[string]interface{}{
		"sources": rank.List(),
		"default": rank.DefaultSourceID,
	})
}

// GetRankLists 返回指定数据源的榜单分组。
// source 缺省或非法时回退默认数据源；分组结构由各数据源给出，如
// 番茄 [{title:"男频阅读榜", items:[{id:"1_2_1141", name:"西方奇幻"}]}]，
// 七猫 [{title:"男生榜",   items:[{id:"0_1_1",    name:"大热榜·日榜"}]}]，
// 起点 [{title:"男频",     items:[{id:"male_hotsales", name:"畅销榜", filters:[题材/月份/周期]}]}]。
func (h *Handler) GetRankLists(w http.ResponseWriter, r *http.Request) {
	src := rank.Get(r.URL.Query().Get("source"))
	groups, err := src.Groups()
	if err != nil {
		respondFail(w, http.StatusBadGateway, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"source": src.ID(), "groups": groups})
}

// GetRankBooks 返回指定数据源某榜的书。
// tab 与 filters 的格式由各数据源自定（filters 形如 "catId:21,yearmonth:202608"）。
// 数据源不可用时返回 502，前端如实显示错误态（无演示数据回退）。
//
// 结果会过一遍 ranksnap：落一份当前页快照，并把缺幅度的源补成有幅度
// （七猫只给方向，起点连方向都没有），失败不影响本次返回。
// 若该源支持整榜抓取，会顺带在后台投递一次当日整榜快照。
func (h *Handler) GetRankBooks(w http.ResponseWriter, r *http.Request) {
	src := rank.Get(r.URL.Query().Get("source"))
	page := queryInt(r, "page", 0)
	if page < 0 {
		page = 0
	}
	size := queryInt(r, "size", 30)
	if size <= 0 || size > 30 {
		size = 30
	}

	q := rank.BookQuery{
		Tab:     r.URL.Query().Get("tab"),
		Filters: parseFilters(r.URL.Query().Get("filters")),
		Offset:  page * size,
		Limit:   size,
	}

	books, err := src.Books(q)
	if err != nil {
		respondFail(w, http.StatusBadGateway, err.Error())
		return
	}
	books = ranksnap.Enrich(h.db, src.ID(), q.ScopeKey(), ranksnap.ScopePage, books)
	maybeCaptureFull(h.db, src, q)

	respondOK(w, map[string]interface{}{"source": src.ID(), "books": books})
}

// GetRankDrops 返回某榜的「掉榜」名单：上一份整榜里有、今天整榜里没有的书。
// 只有可穷尽的数据源（当前是起点）才有意义，其它源会返回 ready=false。
func (h *Handler) GetRankDrops(w http.ResponseWriter, r *http.Request) {
	src := rank.Get(r.URL.Query().Get("source"))
	key := rank.BookQuery{
		Tab:     r.URL.Query().Get("tab"),
		Filters: parseFilters(r.URL.Query().Get("filters")),
	}.ScopeKey()

	res, err := ranksnap.Drops(h.db, src.ID(), key)
	if err != nil {
		respondFail(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w, res)
}

// maybeCaptureFull 对实现了 FullCapturer 的源，每天为每个榜投递一次整榜抓取。
// 抓完写进 scope=full 的快照，之后新上榜与掉榜判定才有可信依据。
// 投递是异步的：整榜要翻几十页，不能拖在一次 HTTP 请求里。
func maybeCaptureFull(db *sql.DB, src rank.Source, q rank.BookQuery) {
	fc, ok := src.(rank.FullCapturer)
	if !ok || db == nil {
		return
	}
	// 由数据源判断这次查询值不值得整榜抓取（起点只在全站视图返回 true），
	// 否则用户点几轮题材就把每日预算耗光，真正要看的全站榜反而抓不到。
	if !fc.WorthFullCapture(q) {
		return
	}

	key := q.ScopeKey()
	today := time.Now().Format("2006-01-02")

	fullCaptureState.Lock()
	if fullCaptureState.day != today {
		fullCaptureState.day = today
		fullCaptureState.keys = map[string]bool{}
		fullCaptureState.count = 0
	}
	if fullCaptureState.keys[key] || fullCaptureState.count >= fullCaptureDailyCap {
		fullCaptureState.Unlock()
		return
	}
	fullCaptureState.keys[key] = true
	fullCaptureState.count++
	fullCaptureState.Unlock()

	sourceID := src.ID()
	go func() {
		books, err := fc.BooksFull(q, fullCaptureMaxPages)
		if err != nil {
			log.Printf("整榜抓取失败 (%s/%s): %v", sourceID, key, err)
			return
		}
		ranksnap.Enrich(db, sourceID, key, ranksnap.ScopeFull, books)
		log.Printf("整榜快照完成 (%s/%s): %d 本", sourceID, key, len(books))
	}()
}

// parseFilters 解析 filters 查询参数，格式 "catId:21,yearmonth:202608"。
// 解析不出东西时返回 nil，让各数据源走自己的缺省值。
func parseFilters(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	out := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		k, v, ok := strings.Cut(pair, ":")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if k != "" && v != "" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// GetRankBookChapters 返回在线书籍的前 N 章目录（可选带正文）。
// GET /api/ai/rank/book-chapters?source=&bookId=&limit=10&content=1&maxChars=
//
// 仅实现了 ChapterLister 的数据源可用（目前只有番茄）；
// 其他源返回 501，前端据此隐藏「关联章节」区。
// 正文只对免费章抓取（付费章接口拿不到）。
// maxChars 仅在调用方显式传 >0 时截断（rune 计），缺省不截断——
// 番茄单章通常 2~4 千字，截断会让拆书素材残缺。
// 串行抓取避免触发站点限流；目录/正文结果由 BridgeClient 缓存 10 分钟。
func (h *Handler) GetRankBookChapters(w http.ResponseWriter, r *http.Request) {
	src := rank.Get(r.URL.Query().Get("source"))
	cl := rank.ChapterListerOf(src)
	if cl == nil {
		respondFail(w, http.StatusNotImplemented, "该数据源暂不支持在线章节")
		return
	}
	bookID := strings.TrimSpace(r.URL.Query().Get("bookId"))
	if bookID == "" {
		respondFail(w, http.StatusBadRequest, "缺少 bookId")
		return
	}
	limit := queryInt(r, "limit", 10)
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	withContent := queryInt(r, "content", 0) == 1
	maxChars := queryInt(r, "maxChars", 0) // 0 = 不截断

	// 本地缓存：当天抓过的书 24h 内直接回（正文抓取是弹窗打开时的耗时大头）。
	// 缓存按 source+bookId+limit+content 分文件；maxChars 只影响展示层截断，不参与 key。
	if data := loadChapterCache(src.ID(), bookID, limit, withContent); data != nil {
		respondOK(w, data)
		return
	}

	chapters, err := cl.ChapterList(bookID, limit)
	if err != nil {
		respondFail(w, http.StatusBadGateway, err.Error())
		return
	}

	type chapterOut struct {
		rank.RankChapter
		Content string `json:"content"`
	}
	out := make([]chapterOut, 0, len(chapters))
	for _, ch := range chapters {
		item := chapterOut{RankChapter: ch}
		if withContent && ch.IsFree && ch.ID != "" {
			if _, content, err := cl.ChapterContent(bookID, ch.ID); err == nil {
				if maxChars > 0 {
					content = truncateRunes(content, maxChars)
				}
				item.Content = content
			}
		}
		out = append(out, item)
	}
	data := map[string]interface{}{
		"source":   src.ID(),
		"bookId":   bookID,
		"chapters": out,
	}
	saveChapterCache(src.ID(), bookID, limit, withContent, data)
	respondOK(w, data)
}

// truncateRunes 按 rune 截断，避免把多字节字符截成乱码。
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
