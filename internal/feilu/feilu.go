// Package feilu 提供飞卢小说网榜单的抓取实现。
//
// 榜单页是公开静态 HTML，GB2312 编码，条目明文（书名/作者/题材/数值/字数/简介/最新章节/封面）。
// 网上常说的「飞卢防盗」防的是 **VIP 章节正文**，不是榜单列表——扫榜不读正文，不受影响。
//
// ⚠ 两个必须先知道的坑：
//
//  1. **Cookie 挑战**：飞卢会返回 302 把请求重定向回自己，同时下发 `C3VK` cookie
//     （Max-Age=300）。不带 cookie jar 的客户端会陷入无限重定向。
//     解决办法跟起点的 _csrfToken 一样——用 net/http/cookiejar 就行。
//
//  2. **GB2312 编码**：Go 标准库没有 GBK 编解码器，这里用 golang.org/x/text。
//
// URL 结构（7 段，y_ 与 l_ 同一套布局）：
//
//	https://b.faloo.com/y_{一级题材}_{二级题材}_{保留}_{字数}_{状态}_{排序}_{页}.html
//
// 方案 B 只暴露「榜单类型（排序段）」和「题材」两个维度，字数与状态固定为 0（全部）。
package feilu

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	baseURL = "https://b.faloo.com"
	// homeURL 用来取题材导航（分类清单不硬编码，跟起点同一条原则）
	homeURL  = baseURL + "/"
	bookTmpl = baseURL + "/%s.html"

	pageSize = 30

	// catalogTTL 题材导航缓存：分类不常变
	catalogTTL = 6 * time.Hour
	// pageTTL 单页缓存，压低请求量（飞卢有 cookie 挑战，请求越省越好）
	pageTTL = 15 * time.Minute
	// fetchGap 两次请求之间的最小间隔，别把网站打急了
	fetchGap = 700 * time.Millisecond

	uaWeb = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// ─── 榜单类型（URL 排序段）─────────────────────────────────────────────

// rankType 是飞卢的一个榜单。Metric 是数值口径，
// 为空表示该榜页面上不展示数值（周/月/总收藏榜就只给字数）。
type rankType struct {
	Code   string
	Name   string
	Metric string
}

// rankGroups 把 18 个榜按用途分四组，让第一层 chip 也有意义。
var rankGroups = []struct {
	Title string
	Items []rankType
}{
	{"人气", []rankType{
		{"1", "周点击榜", "周点击"},
		{"2", "月点击榜", "月点击"},
		{"3", "总点击榜", "总点击"},
		{"19", "催更榜", "催更"},
	}},
	{"付费", []rankType{
		{"15", "月票榜", "月票"},
		{"13", "周打赏榜", "周打赏"},
		{"17", "月打赏榜", "月打赏"},
		{"18", "总打赏榜", "总打赏"},
	}},
	{"互动", []rankType{
		{"4", "周鲜花榜", "周鲜花"},
		{"5", "月鲜花榜", "月鲜花"},
		{"6", "总鲜花榜", "总鲜花"},
		{"7", "周收藏榜", ""},
		{"8", "月收藏榜", ""},
		{"9", "总收藏榜", ""},
	}},
	{"创作", []rankType{
		// 总字数榜的「字数」就是卡片上本来就有的字数列，再当数值列显示就重复了，故留空
		{"10", "总字数榜", ""},
		{"12", "月字数榜", "月字数"},
		{"66", "新书PK榜", "指数"},
		{"0", "最新更新榜", "总点击"},
	}},
}

func rankTypeOf(code string) (rankType, bool) {
	for _, g := range rankGroups {
		for _, rt := range g.Items {
			if rt.Code == code {
				return rt, true
			}
		}
	}
	return rankType{}, false
}

// ─── 对外模型 ─────────────────────────────────────────────────────────

// Category 是一个二级题材。ID 是 URL 里的 "{一级}_{二级}"。
type Category struct {
	ID   string
	Name string
}

// Group 是一组榜单（人气/付费/互动/创作）。
type Group struct {
	Title string
	Items []rankType
}

// Book 已经过格式化，字段可直接用于展示。
type Book struct {
	Rank        int
	Title       string
	Author      string
	Category    string
	MetricName  string // 数值口径名（月票/周点击/…）
	MetricValue string // 数值，如 "1248"；为空表示该榜不展示数值
	WordCount   string
	CoverURL    string
	BookID      string
	Link        string
	Synopsis    string
}

// ─── HTTP ─────────────────────────────────────────────────────────────

var (
	clientOnce sync.Once
	client     *http.Client

	// reqMu 串行化请求：既保证 fetchGap 生效，也避免并发把网站打急
	reqMu    sync.Mutex
	lastCall time.Time
)

func httpClient() *http.Client {
	clientOnce.Do(func() {
		jar, _ := cookiejar.New(nil)
		client = &http.Client{Timeout: 25 * time.Second, Jar: jar}
	})
	return client
}

// decodeGB 把 GB2312/GBK/GB18030 字节转成 UTF-8 字符串。
// Go 标准库没有这套编解码器，必须借 golang.org/x/text。
func decodeGB(raw []byte) (string, error) {
	out, _, err := transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), raw)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// fetchText 取一个页面并按 GB18030 解码。请求之间至少隔 fetchGap。
func fetchText(url string) (string, error) {
	reqMu.Lock()
	if wait := fetchGap - time.Since(lastCall); wait > 0 {
		time.Sleep(wait)
	}
	lastCall = time.Now()
	reqMu.Unlock()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", uaWeb)
	req.Header.Set("Referer", baseURL+"/")

	resp, err := httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("飞卢请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("飞卢 HTTP %d", resp.StatusCode)
	}
	return decodeGB(body)
}

// ─── 题材导航 ─────────────────────────────────────────────────────────

var catRe = regexp.MustCompile(`href="(?:https?:)?//b\.faloo\.com/l_(\d+)_(\d+)_1\.html"[^>]*title="([^"]*)"`)

var (
	catalogMu      sync.Mutex
	catalogCache   []Category
	catalogExpires time.Time
)

// Catalogs 返回全部二级题材。
// ⚠ 一律从首页导航动态取，不硬编码——我先前从书库页只抓到 13 个，首页有 34 个，
// 硬编码必漏。
func Catalogs() []Category {
	catalogMu.Lock()
	if len(catalogCache) > 0 && time.Now().Before(catalogExpires) {
		out := catalogCache
		catalogMu.Unlock()
		return out
	}
	catalogMu.Unlock()

	page, err := fetchText(homeURL)
	if err != nil {
		return nil
	}

	seen := map[string]bool{}
	out := make([]Category, 0, 34)
	for _, m := range catRe.FindAllStringSubmatch(page, -1) {
		id := m[1] + "_" + m[2]
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, Category{ID: id, Name: strings.TrimSpace(m[3])})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	catalogMu.Lock()
	catalogCache = out
	catalogExpires = time.Now().Add(catalogTTL)
	catalogMu.Unlock()
	return out
}

// Groups 返回分组后的榜单。
func Groups() []Group {
	out := make([]Group, 0, len(rankGroups))
	for _, g := range rankGroups {
		out = append(out, Group{Title: g.Title, Items: g.Items})
	}
	return out
}

// ─── 榜单页 ───────────────────────────────────────────────────────────

// entrySplitRe 用来切分条目块：条目以 TwoBox02_01 或 TwoBox02_02 开头。
var entrySplitRe = regexp.MustCompile(`<div class="TwoBox02_0[12]">`)

var (
	bidRe    = regexp.MustCompile(`<a href="(?:https?:)?//b\.faloo\.com/(\d+)\.html"`)
	titleRe  = regexp.MustCompile(`<h1 class="fontSize17andHei" title="([^"]*)"`)
	coverRe  = regexp.MustCompile(`src="(https?://img\.faloo\.com/[^"]+)"`)
	authorRe = regexp.MustCompile(`TwoBox02_09".*?title="([^"]*)"`)
	catlnkRe = regexp.MustCompile(`<a href="(?:https?:)?//b\.faloo\.com/l_\d+_\d+_1\.html"[^>]*title="([^"]*)"`)
	metricRe = regexp.MustCompile(`<span>([^<]{2,8}：[^\s<]{1,16})</span>`)
	synopRe  = regexp.MustCompile(`class="TwoBox02_06"><a[^>]*>([^<]*)<`)
)

// PageURL 拼榜单页 URL。cat 形如 "3_19"，为空表示全部题材。
func PageURL(cat, sortCode string, page int) string {
	c1, c2 := "0", "0"
	if parts := strings.SplitN(cat, "_", 2); len(parts) == 2 && parts[0] != "" {
		c1, c2 = parts[0], parts[1]
	}
	return fmt.Sprintf("%s/y_%s_%s_0_0_0_%s_%d.html", baseURL, c1, c2, sortCode, page)
}

// FetchPage 取一页榜单。cat 为题材 ID（""=全部），sortCode 为榜单类型。
func FetchPage(cat, sortCode string, page int) ([]Book, error) {
	rt, ok := rankTypeOf(sortCode)
	if !ok {
		return nil, fmt.Errorf("未知的飞卢榜单: %s", sortCode)
	}
	html, err := fetchText(PageURL(cat, sortCode, page))
	if err != nil {
		return nil, err
	}
	return parseEntries(html, rt, (page-1)*pageSize), nil
}

type pageEntry struct {
	books   []Book
	expires time.Time
}

var (
	cacheMu   sync.Mutex
	pageCache = map[string]pageEntry{}
)

// FetchPageCached 带缓存的取页。
func FetchPageCached(cat, sortCode string, page int) ([]Book, error) {
	key := cat + "|" + sortCode + "|" + strconv.Itoa(page)
	cacheMu.Lock()
	if e, ok := pageCache[key]; ok && time.Now().Before(e.expires) {
		books := e.books
		cacheMu.Unlock()
		return books, nil
	}
	cacheMu.Unlock()

	books, err := FetchPage(cat, sortCode, page)
	if err != nil {
		return nil, err
	}
	cacheMu.Lock()
	pageCache[key] = pageEntry{books: books, expires: time.Now().Add(pageTTL)}
	cacheMu.Unlock()
	return books, nil
}

// Books 返回某榜第 offset 条起的 limit 本书。
// tab 是榜单类型码；filters["cat"] 是题材 ID（如 "3_19"）。
func Books(tab string, filters map[string]string, offset, limit int) ([]Book, error) {
	if _, ok := rankTypeOf(tab); !ok {
		return nil, fmt.Errorf("未知的飞卢榜单: %s", tab)
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = pageSize
	}

	cat := ""
	if filters != nil {
		cat = filters["cat"]
	}

	first := offset/pageSize + 1
	drop := offset % pageSize
	out := make([]Book, 0, limit+drop)

	for page := first; len(out) < limit+drop; page++ {
		books, err := FetchPageCached(cat, tab, page)
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

// parseEntries 解析一页 HTML 里的全部条目。base 是上一页最后一本的名次。
//
// ⚠ 页面里混着广告块，切分出来的片段不一定都是书目，所以名次必须**只对解析成功的条目**
// 顺序递增，不能按片段下标算——否则名次会跳号（实测出现过 2、3、5…）。
func parseEntries(html string, rt rankType, base int) []Book {
	parts := entrySplitRe.Split(html, -1)
	out := make([]Book, 0, pageSize)
	rank := base
	for _, part := range parts[1:] {
		b := parseOne(part, rt)
		if b.BookID == "" || b.Title == "" {
			continue
		}
		rank++
		b.Rank = rank
		out = append(out, b)
	}
	return out
}

func parseOne(part string, rt rankType) Book {
	bid := ""
	if m := bidRe.FindStringSubmatch(part); m != nil {
		bid = m[1]
	}
	title := ""
	if m := titleRe.FindStringSubmatch(part); m != nil {
		title = strings.TrimSpace(m[1])
	}
	if bid == "" || title == "" {
		return Book{}
	}

	author := ""
	if m := authorRe.FindStringSubmatch(part); m != nil {
		author = strings.TrimSpace(m[1])
	}
	category := ""
	if m := catlnkRe.FindStringSubmatch(part); m != nil {
		category = strings.TrimSpace(m[1])
	}
	cover := ""
	if m := coverRe.FindStringSubmatch(part); m != nil {
		cover = m[1]
	}
	synopsis := ""
	if m := synopRe.FindStringSubmatch(part); m != nil {
		synopsis = strings.TrimSpace(m[1])
	}

	// 数值行形如 <span>月票：1248</span> / <span>字数：236万</span>
	// 先全收进 map 再按榜单类型挑，避免「字数」既是数值又是字数时被 switch 抢先吞掉
	metrics := map[string]string{}
	for _, m := range metricRe.FindAllStringSubmatch(part, -1) {
		if k, v, ok := strings.Cut(m[1], "："); ok {
			metrics[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	words := metrics["字数"]
	name, value := "", ""
	if rt.Metric != "" {
		if v := metrics[rt.Metric]; v != "" {
			name, value = rt.Metric, v
		}
	}
	// 该榜不展示数值（如周/月/总收藏榜）就留空，前端会整列隐藏——不用字数充数，
	// 否则卡片上会出现两遍字数

	return Book{
		Rank:        0, // 由 parseEntries 顺序赋名次
		Title:       title,
		Author:      author,
		Category:    category,
		MetricName:  name,
		MetricValue: value,
		WordCount:   words,
		CoverURL:    cover,
		BookID:      bid,
		Link:        fmt.Sprintf(bookTmpl, bid),
		Synopsis:    clean(synopsis),
	}
}

func clean(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
