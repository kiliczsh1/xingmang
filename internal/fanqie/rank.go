package fanqie

// 番茄小说榜单抓取 —— 从 scripts/fanqie_bridge.py + fanqie-rank-mcp/api.py 移植。
//
// 输出 JSON 与 Python 桥的 /rankings、/ranking 接口保持一致，
// 因此调用方（internal/rank）可以用与解析 HTTP 桥完全相同的方式解析。

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	rankAPI = "https://fanqienovel.com/api/rank/category/list"
	appID   = "2503"

	uaRank = "Dalvik/2.1.0 (Linux; U; Android 10; SM-G975F Build/QP1A.190711.020) " +
		"com.ss.android.article.news/831"
	uaWeb = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// ─── HTTP 客户端 ──────────────────────────────────────────────────────

func httpGet(url, ua string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second} // 默认跟随重定向
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return string(body), nil
}

// ─── 分组/书籍 JSON 结构（与 Python 桥输出一致） ────────────────────────

type groupItemJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type groupJSON struct {
	Title string          `json:"title"`
	Items []groupItemJSON `json:"items"`
}

type bookJSON struct {
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

// ─── 字段解析辅助（对齐 Python 的 `it.get(a) or it.get(b)` 真值语义） ──

// rawStr 按顺序取第一个"真值"字段并转为字符串（数字 0 视为假）。
func rawStr(m map[string]json.RawMessage, keys ...string) string {
	for _, k := range keys {
		raw, ok := m[k]
		if !ok || string(raw) == "null" {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if t != "" {
				return t
			}
		case float64:
			if t != 0 {
				return strconv.FormatInt(int64(t), 10)
			}
		}
	}
	return ""
}

// rawInt 按顺序取第一个数字字段（字符串也尝试解析，对齐 _to_int）。
func rawInt(m map[string]json.RawMessage, keys ...string) int {
	for _, k := range keys {
		raw, ok := m[k]
		if !ok || string(raw) == "null" {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			continue
		}
		switch t := v.(type) {
		case float64:
			return int(t)
		case string:
			if n, err := strconv.Atoi(strings.TrimSpace(t)); err == nil {
				return n
			}
		}
	}
	return 0
}

func normalizeCoverURL(url string) string {
	switch {
	case url == "":
		return ""
	case strings.HasPrefix(url, "//"):
		return "https:" + url
	case strings.HasPrefix(url, "http"):
		return url
	default:
		return "https://p3-novel.byteimg.com/" + strings.TrimLeft(url, "/")
	}
}

// ─── 榜单分组 ─────────────────────────────────────────────────────────

var (
	reArcoMenu  = regexp.MustCompile(`class="arco-menu-inner"([\s\S]*?)</div>\s*</div>\s*</div>\s*</div>`)
	reGroupName = regexp.MustCompile(`<span><span>([^<]+)</span></span>`)
	reRankLink  = regexp.MustCompile(`<a[^>]*href="/rank/([^"]+)"[^>]*>([^<]+)</a>`)
)

func fetchLeaderboards() ([]groupJSON, error) {
	html, err := httpGet("https://fanqienovel.com/rank", uaWeb)
	if err != nil {
		return nil, err
	}

	m := reArcoMenu.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("未找到榜单菜单结构")
	}
	menuHTML := m[1]
	blocks := strings.Split(menuHTML, `class="arco-menu-inline-header"`)

	groups := []groupJSON{}
	for _, block := range blocks[1:] {
		title := ""
		if tm := reGroupName.FindStringSubmatch(block); tm != nil {
			title = tm[1]
		}
		items := []groupItemJSON{}
		for _, lm := range reRankLink.FindAllStringSubmatch(block, -1) {
			items = append(items, groupItemJSON{ID: lm[1], Name: lm[2]})
		}
		if len(items) > 0 {
			groups = append(groups, groupJSON{Title: title, Items: items})
		}
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("榜单菜单解析结果为空")
	}
	return groups, nil
}

// fallbackLeaderboards 兜底榜单（官网解析失败时使用）。
// 注意：题材 ID 必须为官网当前真实 ID（1142/1143/1144 等旧 ID 接口已返回空榜）。
func fallbackLeaderboards() []groupJSON {
	return []groupJSON{
		{
			Title: "男频阅读榜",
			Items: []groupItemJSON{
				{ID: "1_2_1141", Name: "西方奇幻"},
				{ID: "1_2_1140", Name: "东方仙侠"},
				{ID: "1_2_8", Name: "科幻末世"},
				{ID: "1_2_261", Name: "都市日常"},
				{ID: "1_2_124", Name: "都市修真"},
			},
		},
		{
			Title: "女频阅读榜",
			Items: []groupItemJSON{
				{ID: "0_2_1139", Name: "古风世情"},
				{ID: "0_2_267", Name: "现言脑洞"},
				{ID: "0_2_248", Name: "玄幻言情"},
				{ID: "0_2_23", Name: "种田"},
			},
		},
	}
}

// ─── 榜单书籍 ─────────────────────────────────────────────────────────

// fetchBooks 请求番茄榜单接口并解析 book_list（增强字段，对齐 Python 桥）。
func fetchBooks(gender, rankMold, categoryID string, offset, limit int) ([]bookJSON, error) {
	query := fmt.Sprintf(
		"app_id=%s&rank_list_type=3&offset=%d&limit=%d&category_id=%s&rank_version=&gender=%s&rankMold=%s",
		appID, offset, limit, categoryID, gender, rankMold,
	)
	aBogus := generateABogus(query, uaRank)
	url := fmt.Sprintf("%s?%s&a_bogus=%s", rankAPI, query, aBogus)

	body, err := httpGet(url, uaRank)
	if err != nil {
		return nil, err
	}

	var data struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
		Data struct {
			BookList []map[string]json.RawMessage `json:"book_list"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, err
	}
	if data.Code != 0 {
		return nil, fmt.Errorf("番茄接口错误: %s", data.Msg)
	}

	books := make([]bookJSON, 0, len(data.Data.BookList))
	for idx, it := range data.Data.BookList {
		rank := rawInt(it, "currentPos")
		if rank == 0 {
			rank = idx + 1
		}
		books = append(books, bookJSON{
			BookID:         rawStr(it, "book_id", "bookId"),
			Rank:           rank,
			Title:          decryptPUA(rawStr(it, "bookName", "book_name")),
			Author:         decryptPUA(rawStr(it, "author", "author_name")),
			Synopsis:       decryptPUA(rawStr(it, "abstract", "description")),
			CoverURL:       normalizeCoverURL(rawStr(it, "thumbUri", "thumb_url")),
			WordNumber:     rawInt(it, "wordNumber", "word_number"),
			ReadCount:      rawInt(it, "read_count", "readCount"),
			RankDiff:       rawInt(it, "rankPosDiff"),
			CreationStatus: rawInt(it, "creationStatus"),
			CategoryID:     rawStr(it, "curent_category_id", "category_id"),
		})
	}
	return books, nil
}
