// Package rank 提供 AI 扫榜的数据源。
//
// 目前注册两个数据源（见 source.go 的注册表）：
//   - fanqie 番茄小说：默认使用 Go 原生内置实现（internal/fanqie，a_bogus 签名 + PUA 字体解码），
//     无需 Python、零外部进程依赖，打包版开箱即用。
//     如需临时使用外部 Python 桥（scripts/fanqie_bridge.py），
//     显式设置 FANQIE_BRIDGE_URL 环境变量（如 http://127.0.0.1:19900）即可，
//     此时外部桥优先、内置实现兜底；桥不可用会切换为内置并只提示一次。
//   - qimao 七猫小说：榜单接口是公开 JSON，直接抓取（internal/qimao）。
package rank

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/fanqie"
)

// BridgeClient 是榜单数据源客户端（带内存缓存）。
//   - baseURL 为空（默认）：纯内置模式，直接走 Go 原生实现，零外部连接；
//   - baseURL 非空（FANQIE_BRIDGE_URL 显式指定）：外部桥优先，
//     一旦连接失败即标记不可用，后续请求直接走内置实现（仅首次提示一次）。
type BridgeClient struct {
	baseURL string
	http    *http.Client
	mu      sync.Mutex
	cache   map[string]cacheEntry
	down    bool // 外部桥不可用标记
}

type cacheEntry struct {
	body    []byte
	expires time.Time
}

// NewBridgeClient 创建数据源客户端。baseURL 为空时使用内置实现。
func NewBridgeClient(baseURL string) *BridgeClient {
	return &BridgeClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 25 * time.Second},
		cache:   make(map[string]cacheEntry),
	}
}

// ID/Name/Desc 实现 Source。BridgeClient 是番茄数据源。
func (c *BridgeClient) ID() string   { return "fanqie" }
func (c *BridgeClient) Name() string { return "番茄小说" }
func (c *BridgeClient) Desc() string {
	return "数据来自番茄榜单（a_bogus 签名 + PUA 字体解码）"
}

// rankDiffMode 的两个取值。番茄接口给真实名次差，七猫只给方向，
// 展示层据此决定「变动」是 ▲3 还是只有 ▲。
const (
	RankDiffModeStep = "step" // 有幅度
	RankDiffModeDir  = "dir"  // 仅方向，RankDiff 只有 ±1 的意义
)

// Group 是榜单分组（一个 tab），Items 是该分组下的题材列表。
type Group struct {
	Title string      `json:"title"`
	Items []GroupItem `json:"items"`
}

// GroupItem 是一个榜单。ID 格式由各数据源自定（番茄 "1_2_1141"；七猫 "0_1_1"）。
// Filters 是该榜单的进一步筛选维度，没有第三层的源留空。
type GroupItem struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Filters []FilterDim `json:"filters,omitempty"`
}

// Book 是榜单里的一本书。
type Book struct {
	Rank         int    `json:"rank"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	Category     string `json:"category"`     // 题材 categoryId（如 "1141"）
	Status       string `json:"status"`       // 连载中 / 已完结
	ReadingCount string `json:"readingCount"` // 如 "44.97万人"
	WordCount    string `json:"wordCount"`    // 如 "375.26万字"
	CoverURL     string `json:"coverUrl"`
	BookID       string `json:"bookId"`
	Link         string `json:"link"`
	RankDiff     int    `json:"rankDiff"` // 名次变动：正=上升 负=下降 0=持平
	// RankDiffMode 说明 RankDiff 的口径：step=真实名次差；dir=仅方向（此时只有 ±1 的意义）。
	RankDiffMode string `json:"rankDiffMode"`
	// IsNew 表示这本在上一份榜单快照里没有出现过（新上榜）。
	// 只由 internal/ranksnap 回填。
	IsNew bool `json:"isNew"`
	// DaysOnList 是这本在快照里出现过的天数（含今天）。只由 internal/ranksnap 回填。
	DaysOnList int    `json:"daysOnList"`
	Synopsis   string `json:"synopsis"`
	// MetricLabel 是 ReadingCount 的数值口径（在读 / 热度 / 收藏）。
	MetricLabel string `json:"metricLabel"`
}

// bridgeBook 是桥返回的单本书原始结构。
type bridgeBook struct {
	BookID         string `json:"bookId"`
	Rank           int    `json:"rank"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	Synopsis       string `json:"synopsis"`
	CoverURL       string `json:"coverUrl"`
	WordNumber     int    `json:"wordNumber"`
	ReadCount      int    `json:"readCount"`
	RankDiff       int    `json:"rankDiff"`
	CreationStatus int    `json:"creationStatus"`
	CategoryID     string `json:"categoryId"`
}

func (b bridgeBook) toBook() Book {
	status := "连载中"
	if b.CreationStatus == 0 {
		status = "已完结"
	}
	return Book{
		Rank:         b.Rank,
		Title:        b.Title,
		Author:       b.Author,
		Category:     b.CategoryID,
		Status:       status,
		ReadingCount: fmt.Sprintf("%.2f万人", float64(b.ReadCount)/10000),
		WordCount:    fmt.Sprintf("%.2f万字", float64(b.WordNumber)/10000),
		CoverURL:     b.CoverURL,
		BookID:       b.BookID,
		Link:         "https://fanqienovel.com/page/" + b.BookID,
		RankDiff:     b.RankDiff,
		RankDiffMode: RankDiffModeStep,
		Synopsis:     b.Synopsis,
		MetricLabel:  "在读",
	}
}

// Groups 返回榜单分组（tab + 题材），实现 Source。
func (c *BridgeClient) Groups() ([]Group, error) {
	var payload struct {
		Groups []Group `json:"groups"`
	}
	if err := c.get("rankings", c.baseURL+"/rankings", &payload, fanqie.RankingsJSON); err != nil {
		return nil, err
	}
	return payload.Groups, nil
}

// Books 实现 Source。Tab 格式：gender_rankMold_categoryId，如 "1_2_1141"。
// 番茄没有第三层筛选，q.Filters 被忽略。
func (c *BridgeClient) Books(q BookQuery) ([]Book, error) {
	gender, rankMold, categoryID := ParseFanqieTab(q.Tab)
	return c.booksByGender(gender, rankMold, categoryID, q.Offset, q.Limit)
}

// ParseFanqieTab 解析番茄榜单标识；缺失或不足三段时退回 男频·阅读榜·西方奇幻。
func ParseFanqieTab(tab string) (gender, rankMold, categoryID string) {
	gender, rankMold, categoryID = "1", "2", "1141"
	if parts := strings.Split(tab, "_"); len(parts) >= 3 {
		gender, rankMold, categoryID = parts[0], parts[1], parts[2]
	}
	return
}

// booksByGender 返回某榜（gender/rankMold/categoryID）的书列表。
func (c *BridgeClient) booksByGender(gender, rankMold, categoryID string, offset, limit int) ([]Book, error) {
	url := fmt.Sprintf("%s/ranking?gender=%s&rank_mold=%s&category_id=%s&offset=%d&limit=%d",
		c.baseURL, gender, rankMold, categoryID, offset, limit)
	key := fmt.Sprintf("ranking:%s:%s:%s:%d:%d", gender, rankMold, categoryID, offset, limit)

	var payload struct {
		Books []bridgeBook `json:"books"`
	}
	err := c.get(key, url, &payload, func() ([]byte, error) {
		return fanqie.BooksJSON(gender, rankMold, categoryID, offset, limit)
	})
	if err != nil {
		return nil, err
	}
	books := make([]Book, 0, len(payload.Books))
	for _, b := range payload.Books {
		books = append(books, b.toBook())
	}
	return books, nil
}

// get 带缓存地请求数据源并解析 JSON。
//   - 纯内置模式（baseURL 为空）：直接走 Go 原生实现，零外部连接、无日志；
//   - 外部桥模式（FANQIE_BRIDGE_URL 显式设置）：外部桥优先，
//     不可用时回退内置实现（仅首次提示一次），回退结果同样进缓存。
func (c *BridgeClient) get(key, url string, out interface{}, fallback func() ([]byte, error)) error {
	c.mu.Lock()
	if e, ok := c.cache[key]; ok && time.Now().Before(e.expires) {
		body := e.body
		c.mu.Unlock()
		return json.Unmarshal(body, out)
	}
	c.mu.Unlock()

	var body []byte
	var err error
	if c.baseURL == "" {
		// 默认：内置实现就是唯一数据源
		body, err = fallback()
	} else if c.isDown() {
		// 外部桥已确认不可用，直接走内置实现
		body, err = fallback()
	} else {
		body, err = c.do(url)
		if err != nil {
			if c.markDown() {
				log.Printf("外部番茄桥不可用 (%v)，已切换为内置实现", err)
			}
			body, err = fallback()
		}
	}
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.cache[key] = cacheEntry{body: body, expires: time.Now().Add(10 * time.Minute)}
	c.mu.Unlock()
	return json.Unmarshal(body, out)
}

// isDown 外部桥是否已被标记为不可用。
func (c *BridgeClient) isDown() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.down
}

// markDown 标记外部桥不可用，返回是否首次标记（仅首次需要打日志）。
func (c *BridgeClient) markDown() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.down {
		return false
	}
	c.down = true
	return true
}

func (c *BridgeClient) do(url string) ([]byte, error) {
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("番茄桥不可用: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("番茄桥错误(%d): %s", resp.StatusCode, string(body))
	}
	return body, nil
}
