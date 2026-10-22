// Package qimao 提供七猫小说网榜单的抓取实现。
//
// ⚠ 用的是**分类接口**，不是榜单接口：
//
//	榜单接口  /api/rank/book-list  —— 完全不支持题材过滤
//	          （复验过 category1+category2 / category_id / catId / category2 / 带 channel 全套，
//	            六种写法返回结果完全一致），只有 100 本封顶，但会给热度数值。
//	分类接口  /api/classify/*      —— 支持题材（一级/二级）与四种排序，单题材约 220 本，
//	          但没有热度数值，也没有日榜/月榜。
//
// 方案 A 选了分类接口：用「热度数值 + 日榜月榜」换「真题材维度 + 更大的榜单深度」。
//
// 两个接口都是公开 JSON，无需签名、无字体混淆、无 Cookie，只做参数拼装与字段映射。
package qimao

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	optionURL = "https://www.qimao.com/api/classify/select-option"
	listURL   = "https://www.qimao.com/api/classify/book-list"

	// 分类接口每页固定 15 本
	pageSize = 15

	// optionsTTL 筛选选项缓存：分类树不常变
	optionsTTL = 6 * time.Hour
	// pageTTL 单页缓存
	pageTTL = 10 * time.Minute
	// fetchGap 翻页间隔
	fetchGap = 300 * time.Millisecond

	uaWeb = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	referer = "https://www.qimao.com/shuku/"
	// allValue 是接口里「全部」的取值（书库 URL 里也是 a）
	allValue = "a"
)

// ─── 对外模型 ─────────────────────────────────────────────────────────

// FilterOption 是一个筛选取值。
type FilterOption struct {
	Value string
	Label string
}

// Category 是一级分类及其二级子项。
// Type 是它归属的频道，取值与频道筛选项的 value 一致（"0" 男生原创 / "1" 女生原创 /
// "2" 出版图书）。分类树是跨频道的一份全集，但每个分类只属于其中一个频道 ——
// 题材清单必须按它过滤，否则会出现「男生原创 · 现代言情」这种恒为空的组合。
type Category struct {
	ID       string
	Name     string
	Type     string
	Children []Category
}

// Options 是分类接口给的全部筛选选项。
type Options struct {
	Channels   []FilterOption
	Words      []FilterOption
	UpdateTime []FilterOption
	IsOver     []FilterOption
	Orders     []FilterOption
	Categories []Category
}

// Channel 是一组榜单（一个频道）。
type Channel struct {
	Value string
	Label string
}

// Book 已经过格式化，字段可直接用于展示。
type Book struct {
	Rank      int
	Title     string
	Author    string
	Category  string
	Status    string
	WordCount string
	Updated   string
	CoverURL  string
	BookID    string
	Link      string
	Synopsis  string
	LastChap  string
}

// ─── 榜单类型与频道 ───────────────────────────────────────────────────

// rankOrder 是一种排序方式，也就是一个「榜」。
// 接口的 label 是「按点击量 / 按总字数 / 最近更新 / 按收藏数」，这里用更短的榜名。
type rankOrder struct {
	Code string
	Name string
}

var rankOrders = []rankOrder{
	{"click", "点击榜"},
	{"words_num", "字数榜"},
	{"update_time", "更新榜"},
	{"favorite_uv", "收藏榜"},
}

// RankOrders 返回可用的排序方式，也就是全部「榜」。
func RankOrders() []rankOrder {
	out := make([]rankOrder, 0, len(rankOrders))
	for _, o := range rankOrders {
		out = append(out, rankOrder{Code: o.Code, Name: o.Name})
	}
	return out
}

func rankOrderOf(code string) (rankOrder, bool) {
	for _, o := range rankOrders {
		if o.Code == code {
			return o, true
		}
	}
	return rankOrder{}, false
}

// TabID 拼榜单标识：{channel}_{order}，如 "0_click"。
func TabID(channel, order string) string { return channel + "_" + order }

// ParseTab 解析榜单标识，非法时返回 false。
func ParseTab(tab string) (channel, order string, ok bool) {
	parts := strings.SplitN(tab, "_", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	if _, exists := rankOrderOf(parts[1]); !exists {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// ─── HTTP ─────────────────────────────────────────────────────────────

var (
	httpMu   sync.Mutex
	lastCall time.Time
	client   = &http.Client{Timeout: 25 * time.Second}
)

func get(rawURL string) ([]byte, error) {
	httpMu.Lock()
	if wait := fetchGap - time.Since(lastCall); wait > 0 {
		time.Sleep(wait)
	}
	lastCall = time.Now()
	httpMu.Unlock()

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", uaWeb)
	req.Header.Set("Referer", referer)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("七猫请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("七猫 HTTP %d", resp.StatusCode)
	}
	return body, nil
}

// ─── 筛选选项 ─────────────────────────────────────────────────────────

type apiPair struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type apiKid struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type apiCategory struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Children []apiKid `json:"children"`
}

type optionResp struct {
	Data struct {
		Filters struct {
			Channel    []apiPair `json:"channel"`
			Words      []apiPair `json:"words"`
			UpdateTime []apiPair `json:"update_time"`
			IsOver     []apiPair `json:"is_over"`
			Order      []apiPair `json:"order"`
		} `json:"filters"`
		Category []apiCategory `json:"category"`
	} `json:"data"`
}

var (
	optMu      sync.Mutex
	optCache   Options
	optExpires time.Time
)

func pairs(in []apiPair) []FilterOption {
	out := make([]FilterOption, 0, len(in))
	for _, p := range in {
		out = append(out, FilterOption{Value: p.Value, Label: p.Label})
	}
	return out
}

// FetchOptions 返回全部筛选选项（含完整分类树）。
//
// ⚠ 一律从接口动态取，不硬编码——分类树有 30 多个二级题材，硬编码必漏。
// 取不到时会返回上一次的缓存（哪怕已过期），保证榜单还能用。
func FetchOptions() (Options, error) {
	optMu.Lock()
	cached, expires, haveCache := optCache, optExpires, len(optCache.Channels) > 0
	optMu.Unlock()
	if haveCache && time.Now().Before(expires) {
		return cached, nil
	}

	body, err := get(optionURL)
	if err != nil {
		if haveCache {
			return cached, nil
		}
		return Options{}, err
	}

	var r optionResp
	if err := json.Unmarshal(body, &r); err != nil {
		if haveCache {
			return cached, nil
		}
		return Options{}, fmt.Errorf("七猫筛选选项解析失败: %w", err)
	}

	cats := make([]Category, 0, len(r.Data.Category))
	for _, c := range r.Data.Category {
		// 「全部」不是一个真分类
		if c.ID == "" || c.ID == "0" || c.ID == allValue {
			continue
		}
		kids := make([]Category, 0, len(c.Children))
		for _, k := range c.Children {
			if k.ID == "" || k.ID == "0" {
				continue
			}
			kids = append(kids, Category{ID: k.ID, Name: k.Name, Type: k.Type})
		}
		cats = append(cats, Category{ID: c.ID, Name: c.Name, Type: c.Type, Children: kids})
	}

	out := Options{
		Channels:   pairs(r.Data.Filters.Channel),
		Words:      pairs(r.Data.Filters.Words),
		UpdateTime: pairs(r.Data.Filters.UpdateTime),
		IsOver:     pairs(r.Data.Filters.IsOver),
		Orders:     pairs(r.Data.Filters.Order),
		Categories: cats,
	}

	optMu.Lock()
	optCache = out
	optExpires = time.Now().Add(optionsTTL)
	optMu.Unlock()
	return out, nil
}

// Channels 返回可用的频道（第一层分组）。
// 拿不到接口时退回「全部 / 女生原创 / 男生原创 / 出版图书」这四项。
func Channels() []Channel {
	opts, err := FetchOptions()
	if err != nil || len(opts.Channels) == 0 {
		return []Channel{
			{allValue, "全部"},
			{"1", "女生原创"},
			{"0", "男生原创"},
			{"2", "出版图书"},
		}
	}
	out := make([]Channel, 0, len(opts.Channels))
	for _, c := range opts.Channels {
		out = append(out, Channel{Value: c.Value, Label: c.Label})
	}
	return out
}

// ─── 书目 ─────────────────────────────────────────────────────────────

type apiBook struct {
	Index              string `json:"index"`
	BookID             string `json:"book_id"`
	ReadURL            string `json:"read_url"`
	IsOverTxt          string `json:"is_over_txt"`
	WordsNum           string `json:"words_num"`
	ImageLink          string `json:"image_link"`
	Author             string `json:"author"`
	Title              string `json:"title"`
	Category2Name      string `json:"category2_name"`
	UpdateTime         string `json:"update_time"`
	UpdateTimeTxt      string `json:"update_time_txt"`
	Intro              string `json:"intro"`
	LatestChapterTitle string `json:"latest_chapter_title"`
}

type listResp struct {
	Data struct {
		Title    string    `json:"title"`
		BookList []apiBook `json:"book_list"`
		PageData struct {
			Count    string `json:"count"`
			Page     string `json:"page"`
			PageSize string `json:"page_size"`
		} `json:"page_data"`
	} `json:"data"`
	Errors *struct {
		Code  string `json:"code"`
		Title string `json:"title"`
	} `json:"errors"`
}

// ClassifyQuery 是一次分类查询。空字符串表示「全部」。
type ClassifyQuery struct {
	Channel    string
	Category1  string
	Category2  string
	Words      string
	UpdateTime string
	IsVip      string
	IsOver     string
	Order      string
	Page       int
}

// ClassifyBooks 取一页分类书目，返回书目与总数。
//
// ⚠ 接口要求参数齐全：缺任一都会返回 `errors.code=40001002 请求参数错误`。
// 「全部」用 "a"（书库 URL 里也是这个值）。
func ClassifyBooks(q ClassifyQuery) ([]Book, int, error) {
	fill := func(s string) string {
		if s == "" {
			return allValue
		}
		return s
	}
	vals := url.Values{}
	vals.Set("channel", fill(q.Channel))
	vals.Set("category1", fill(q.Category1))
	vals.Set("category2", fill(q.Category2))
	vals.Set("words", fill(q.Words))
	vals.Set("update_time", fill(q.UpdateTime))
	vals.Set("is_vip", fill(q.IsVip))
	vals.Set("is_over", fill(q.IsOver))
	vals.Set("order", q.Order)
	if q.Page < 1 {
		q.Page = 1
	}
	vals.Set("page", strconv.Itoa(q.Page))

	body, err := get(listURL + "?" + vals.Encode())
	if err != nil {
		return nil, 0, err
	}

	var r listResp
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, 0, fmt.Errorf("七猫分类书目解析失败: %w", err)
	}
	if r.Errors != nil {
		return nil, 0, fmt.Errorf("七猫分类书目接口报错(%s): %s", r.Errors.Code, r.Errors.Title)
	}

	total := 0
	if r.Data.PageData.Count != "" {
		total, _ = strconv.Atoi(r.Data.PageData.Count)
	}

	books := make([]Book, 0, len(r.Data.BookList))
	for i, b := range r.Data.BookList {
		if b.BookID == "" || b.Title == "" {
			continue
		}
		books = append(books, Book{
			Rank:      (q.Page-1)*pageSize + i + 1,
			Title:     strings.TrimSpace(b.Title),
			Author:    strings.TrimSpace(b.Author),
			Category:  strings.TrimSpace(b.Category2Name),
			Status:    strings.TrimSpace(b.IsOverTxt),
			WordCount: strings.TrimSpace(b.WordsNum),
			Updated:   strings.TrimSpace(b.UpdateTime),
			CoverURL:  b.ImageLink,
			BookID:    b.BookID,
			Link:      b.ReadURL,
			Synopsis:  cleanText(b.Intro),
			LastChap:  strings.TrimSpace(b.LatestChapterTitle),
		})
	}
	return books, total, nil
}

type pageEntry struct {
	books   []Book
	total   int
	expires time.Time
}

var (
	cacheMu   sync.Mutex
	pageCache = map[string]pageEntry{}
)

func classifyCached(q ClassifyQuery) ([]Book, int, error) {
	key := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%d",
		q.Channel, q.Category1, q.Category2, q.Words, q.UpdateTime, q.IsVip, q.IsOver, q.Order, q.Page)

	cacheMu.Lock()
	if e, ok := pageCache[key]; ok && time.Now().Before(e.expires) {
		books, total := e.books, e.total
		cacheMu.Unlock()
		return books, total, nil
	}
	cacheMu.Unlock()

	books, total, err := ClassifyBooks(q)
	if err != nil {
		return nil, 0, err
	}
	cacheMu.Lock()
	pageCache[key] = pageEntry{books: books, total: total, expires: time.Now().Add(pageTTL)}
	cacheMu.Unlock()
	return books, total, nil
}

// Books 返回某榜第 offset 条起的 limit 本书。
// tab 形如 "{channel}_{order}"；filters["cat"] 形如 "{一级}_{二级}"（二级可以是 "a" 表示该一级全部）。
func Books(tab string, filters map[string]string, offset, limit int) ([]Book, error) {
	channel, order, ok := ParseTab(tab)
	if !ok {
		return nil, fmt.Errorf("未知的七猫榜单: %s", tab)
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = pageSize
	}

	c1, c2 := "", ""
	if filters != nil {
		if v := filters["cat"]; v != "" {
			if parts := strings.SplitN(v, "_", 2); len(parts) == 2 {
				c1, c2 = parts[0], parts[1]
			}
		}
	}

	base := ClassifyQuery{Channel: channel, Category1: c1, Category2: c2, Order: order}
	first := offset/pageSize + 1
	drop := offset % pageSize
	out := make([]Book, 0, limit+drop)

	for page := first; len(out) < limit+drop; page++ {
		q := base
		q.Page = page
		books, _, err := classifyCached(q)
		if err != nil {
			return nil, err
		}
		if len(books) == 0 {
			break
		}
		out = append(out, books...)
		if len(books) < pageSize {
			break // 最后一页
		}
	}

	if drop >= len(out) {
		return []Book{}, nil
	}
	out = out[drop:]
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ─── 题材清单 ─────────────────────────────────────────────────────────

// GenreOption 是题材筛选的一个取值。
type GenreOption struct {
	ID   string
	Name string
}

// Genres 返回**某个频道下**可用的题材筛选取值（一级分类）。
//
// channel 传空或 "a"（全部）时给出全集；否则只保留 Category.Type 等于该频道的分类。
// 这个过滤不能省：分类树是跨频道的一份全集，而接口既不按频道过滤、
// 也不会对非法组合报错（返回 count=0、book_list 为空），
// 不过滤的话 UI 上会出现一大片点进去就是空态的组合。
//
// 分类树是两级的（23 个一级 + 99 个二级），这里只用一级，理由是 UI 装不下——
// 99 个 chip 会堆成五六行，把卡片区挤没。一级的粒度（都市 / 现代言情 / 玄幻 / 历史…）
// 与其它源的题材维度相当，够用。
//
// 选项 ID 形如 "{一级}_a"，末尾的 a 表示「该一级下的全部二级」。
// 真要细到二级，得先给前端加级联筛选（选完一级再出二级），那是另一件事。
func Genres(channel string) []GenreOption {
	opts, err := FetchOptions()
	if err != nil {
		return nil
	}
	scoped := channel != "" && channel != allValue

	out := make([]GenreOption, 0, len(opts.Categories))
	matched := 0
	for _, c := range opts.Categories {
		if scoped && c.Type != channel {
			continue
		}
		matched++
		out = append(out, GenreOption{ID: c.ID + "_" + allValue, Name: c.Name})
	}
	if matched == 0 {
		// 频道与分类的对应关系对不上（接口改了形状）时，退回全集而不是给出空清单
		for _, c := range opts.Categories {
			out = append(out, GenreOption{ID: c.ID + "_" + allValue, Name: c.Name})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
