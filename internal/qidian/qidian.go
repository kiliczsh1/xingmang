// Package qidian 提供起点中文网（阅文）榜单的抓取实现。
//
// 只走移动端 m.qidian.com。PC 站 www.qidian.com 被风控挡死（202 + 209 字节空壳，
// 里面是 probe.js 风控脚本），而移动端没有任何字体反爬，数字全是明文，
// 所以网上那些「起点字体解密」的坑（qidian.gtimg.com/qd_anti_spider/*.woff）这里不用碰。
//
// 接口（唯一正确姿势，SSR 只给第 1 页且忽略 query 参数）：
//
//	https://m.qidian.com/majax/rank/{type}list?_csrfToken=<token>&gender={male|female}&catId=&pageNum=
//
// ⚠ 必须带 _csrfToken，否则 code=1 / msg=失败。token 来自任意一次访问 m.qidian.com 的
// Set-Cookie，不需要登录；这里用 cookiejar 维护，遇到 code!=0 就强制重取一次。
//
// ⚠ 题材（catId）是性别作用域内的，跨性别传会 code=0 但 total=0（静默空列表），
// 所以筛选维度一律取自接口返回的 data.filters，绝不硬编码。
package qidian

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	baseHost = "https://m.qidian.com"
	rankPage = baseHost + "/rank/"
	rankTmpl = baseHost + "/majax/rank/%slist"

	bookTmpl  = "https://www.qidian.com/book/%s/"
	coverTmpl = "https://bookcover.yuewen.com/qdbimg/349573/%s/180"

	pageSize = 20

	// catalogTTL 榜单目录（含筛选维度）缓存：维度不常变，但 yearmonth 是按月的。
	catalogTTL = 6 * time.Hour
	// pageTTL 单页缓存，避免同一页被反复命中。
	pageTTL = 10 * time.Minute
	// fetchGap 翻页间隔，整榜抓取时用来限速。
	fetchGap = 300 * time.Millisecond
	// catalogWorkers 拉榜单目录时的并发度（12 个榜，顺序拉太慢）。
	catalogWorkers = 4

	uaMobile = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1"
)

// ─── 榜单类型 ─────────────────────────────────────────────────────────

// rankType 是起点的一个榜单。Metric 是 rankCnt 的数值口径；
// 为空表示该榜不公开数值（畅销榜、阅读指数榜、新书榜、新人作者榜都拿不到）。
type rankType struct {
	Code   string
	Name   string
	Metric string
}

var rankTypes = []rankType{
	{"hotsales", "畅销榜", ""},
	{"yuepiao", "月票榜", "月票"},
	{"readindex", "阅读指数榜", ""},
	{"newbook", "新书榜", ""},
	{"newauthor", "新人作者榜", ""},
	{"update", "更新榜", "月更字"},
}

func rankTypeOf(code string) (rankType, bool) {
	for _, rt := range rankTypes {
		if rt.Code == code {
			return rt, true
		}
	}
	return rankType{}, false
}

// gender 是频道。
type gender struct {
	Code string
	Name string
}

var genders = []gender{{"male", "男频"}, {"female", "女频"}}

// ─── 对外模型 ─────────────────────────────────────────────────────────

// FilterOption 是筛选器的一个取值。
type FilterOption struct {
	ID   string
	Name string
}

// FilterDim 是榜单内的一层筛选维度。
type FilterDim struct {
	Key       string
	Name      string
	Options   []FilterOption
	DefaultID string
}

// RankItem 是一个榜单及其筛选维度。
type RankItem struct {
	Code    string
	Name    string
	Filters []FilterDim
}

// Catalog 是一个频道下的全部榜单。
type Catalog struct {
	Gender string
	Name   string
	Items  []RankItem
}

// Book 已经过格式化，字段可直接用于展示。
type Book struct {
	Rank         int
	Title        string
	Author       string
	Category     string
	Status       string
	ReadingCount string
	WordCount    string
	CoverURL     string
	BookID       string
	Link         string
	MetricLabel  string
	Synopsis     string
}

// ─── HTTP 客户端与 token ─────────────────────────────────────────────

var (
	clientOnce sync.Once
	client     *http.Client

	tokenMu sync.Mutex
)

func httpClient() *http.Client {
	clientOnce.Do(func() {
		jar, _ := cookiejar.New(nil)
		client = &http.Client{Timeout: 20 * time.Second, Jar: jar}
	})
	return client
}

func get(rawURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", uaMobile)
	req.Header.Set("Referer", rankPage)

	// 不手动设 Accept-Encoding：交给 net/http 自动协商并透明解压
	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("起点请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("起点 HTTP %d", resp.StatusCode)
	}
	return body, nil
}

// tokenOf 从 cookiejar 里取 _csrfToken。
func tokenOf() string {
	u, err := url.Parse(baseHost)
	if err != nil {
		return ""
	}
	for _, ck := range httpClient().Jar.Cookies(u) {
		if ck.Name == "_csrfToken" {
			return ck.Value
		}
	}
	return ""
}

// ensureToken 保证手里有 _csrfToken。force 为真时忽略缓存重新取一次。
// 拿不到就直接返回错误——majax 没有 token 只会得到 code=1 / 「失败」。
func ensureToken(force bool) (string, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	if !force {
		if t := tokenOf(); t != "" {
			return t, nil
		}
	}
	// 访问一次榜单页即可从 Set-Cookie 拿到，不需要登录
	if _, err := get(rankPage); err != nil {
		return "", err
	}
	if t := tokenOf(); t != "" {
		return t, nil
	}
	return "", fmt.Errorf("起点未返回 _csrfToken")
}

// ─── 接口原始结构 ─────────────────────────────────────────────────────

// flexString 兼容接口里 value 有时是数字、有时是字符串的情况：
// catId / rankPeriod 给数字，yearmonth 给 "202608" 这样的字符串。
// 统一成字符串，前端也只当字符串用。
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexString(s)
		return nil
	}
	*f = flexString(b)
	return nil
}

type apiFilterItem struct {
	Text  string     `json:"text"`
	Value flexString `json:"value"`
}

type apiFilter struct {
	Items         []apiFilterItem `json:"items"`
	Key           string          `json:"key"`
	SelectedIndex int             `json:"selectedIndex"`
}

type apiRecord struct {
	BID     string `json:"bid"`
	BName   string `json:"bName"`
	BAuth   string `json:"bAuth"`
	Cat     string `json:"cat"`
	SubCat  string `json:"subCat"`
	Cnt     string `json:"cnt"`
	Desc    string `json:"desc"`
	RankNum int    `json:"rankNum"`
	RankCnt string `json:"rankCnt"`
}

type apiData struct {
	Total   int         `json:"total"`
	IsLast  int         `json:"isLast"`
	PageNum int         `json:"pageNum"`
	Records []apiRecord `json:"records"`
	Filters []apiFilter `json:"filters"`
}

type apiResp struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data apiData `json:"data"`
}

// filterKey 把筛选条件拼成稳定的缓存键。
func filterKey(filters map[string]string) string {
	if len(filters) == 0 {
		return ""
	}
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+filters[k])
	}
	return strings.Join(parts, "&")
}

// fetchPage 调 majax 取某榜某页。filters 里的空值会被丢掉（空筛选等于不筛）。
func fetchPage(g, typ string, filters map[string]string, page int) (*apiData, error) {
	q := url.Values{}
	q.Set("gender", g)
	for k, v := range filters {
		if k != "" && v != "" {
			q.Set(k, v)
		}
	}
	q.Set("pageNum", strconv.Itoa(page))

	var lastMsg string
	for attempt := 0; attempt < 2; attempt++ {
		token, err := ensureToken(attempt > 0)
		if err != nil {
			return nil, err
		}
		body, err := get(fmt.Sprintf(rankTmpl, typ) + "?_csrfToken=" + url.QueryEscape(token) + "&" + q.Encode())
		if err != nil {
			return nil, err
		}
		var resp apiResp
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("起点榜单响应解析失败: %w", err)
		}
		if resp.Code == 0 {
			return &resp.Data, nil
		}
		// code != 0 多半是 token 过期，第二轮强制重取
		lastMsg = resp.Msg
	}
	return nil, fmt.Errorf("起点榜单接口返回失败: %s", lastMsg)
}

// ─── 页级缓存 ─────────────────────────────────────────────────────────

type pageEntry struct {
	data    *apiData
	expires time.Time
}

var (
	cacheMu   sync.Mutex
	pageCache = map[string]pageEntry{}
)

// fetchPageCached 带缓存的取页。整榜抓取也走这里，重复抓同一页不会重复请求。
func fetchPageCached(g, typ string, filters map[string]string, page int) (*apiData, error) {
	key := g + "_" + typ + "|" + filterKey(filters) + "|" + strconv.Itoa(page)

	cacheMu.Lock()
	if e, ok := pageCache[key]; ok && time.Now().Before(e.expires) {
		data := e.data
		cacheMu.Unlock()
		return data, nil
	}
	cacheMu.Unlock()

	data, err := fetchPage(g, typ, filters, page)
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	pageCache[key] = pageEntry{data: data, expires: time.Now().Add(pageTTL)}
	cacheMu.Unlock()
	return data, nil
}

// ─── 榜单目录（含筛选维度）────────────────────────────────────────────

var (
	catalogMu      sync.Mutex
	catalogCache   []Catalog
	catalogExpires time.Time
)

// Catalogs 返回 2 个频道下的 6 个榜单，并带上各自的筛选维度。
//
// ⚠ 筛选维度一律取自接口返回的 filters：男频 6 个榜一致，女频各榜不同
// （女频畅销榜甚至没有筛选器），硬编码必然出错。首次调用会拉 12 个榜，之后走缓存。
func Catalogs() []Catalog {
	catalogMu.Lock()
	if len(catalogCache) > 0 && time.Now().Before(catalogExpires) {
		out := catalogCache
		catalogMu.Unlock()
		return out
	}
	catalogMu.Unlock()

	// token 先备好，避免并发时多个 goroutine 同时 bootstrap
	if _, err := ensureToken(false); err != nil {
		// 拿不到 token 也要返回目录（只是筛选维度会缺），由 Books 再报错
		return fallbackCatalogs()
	}

	out := make([]Catalog, 0, len(genders))
	for _, gd := range genders {
		items := make([]RankItem, len(rankTypes))
		for i, rt := range rankTypes {
			items[i] = RankItem{Code: rt.Code, Name: rt.Name}
		}

		sem := make(chan struct{}, catalogWorkers)
		var wg sync.WaitGroup
		for i := range items {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				// 轻微错峰，别 12 个请求同时打过去
				time.Sleep(time.Duration(idx) * 60 * time.Millisecond)

				data, err := fetchPage(gd.Code, items[idx].Code, nil, 1)
				if err != nil {
					return // 单个榜拉不到就留空筛选，不影响其它榜
				}
				items[idx].Filters = toFilterDims(data.Filters)
			}(i)
		}
		wg.Wait()

		out = append(out, Catalog{Gender: gd.Code, Name: gd.Name, Items: items})
	}

	catalogMu.Lock()
	catalogCache = out
	catalogExpires = time.Now().Add(catalogTTL)
	catalogMu.Unlock()
	return out
}

// fallbackCatalogs 拿不到接口时的兜底目录：只有榜单名，没有筛选维度。
// 前端会少一行筛选，但至少能切榜、能报错。
func fallbackCatalogs() []Catalog {
	out := make([]Catalog, 0, len(genders))
	for _, gd := range genders {
		items := make([]RankItem, 0, len(rankTypes))
		for _, rt := range rankTypes {
			items = append(items, RankItem{Code: rt.Code, Name: rt.Name})
		}
		out = append(out, Catalog{Gender: gd.Code, Name: gd.Name, Items: items})
	}
	return out
}

func toFilterDims(fs []apiFilter) []FilterDim {
	out := make([]FilterDim, 0, len(fs))
	for _, f := range fs {
		opts := make([]FilterOption, 0, len(f.Items))
		def := ""
		for i, it := range f.Items {
			id := string(it.Value)
			opts = append(opts, FilterOption{ID: id, Name: it.Text})
			if i == f.SelectedIndex {
				def = id
			}
		}
		if len(opts) == 0 {
			continue
		}
		out = append(out, FilterDim{
			Key:       f.Key,
			Name:      filterName(f.Key),
			Options:   opts,
			DefaultID: def,
		})
	}
	return out
}

func filterName(key string) string {
	switch key {
	case "catId":
		return "题材"
	case "yearmonth":
		return "月份"
	case "rankPeriod":
		return "周期"
	default:
		return key
	}
}

// ─── 榜单标识 ─────────────────────────────────────────────────────────

// TabID 拼榜单标识：{gender}_{type}，如 "male_hotsales"。
func TabID(g, typ string) string { return g + "_" + typ }

// ParseTab 解析榜单标识，非法时返回 false。
func ParseTab(tab string) (g, typ string, ok bool) {
	parts := strings.SplitN(tab, "_", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	if _, exists := rankTypeOf(parts[1]); !exists {
		return "", "", false
	}
	for _, gd := range genders {
		if gd.Code == parts[0] {
			return parts[0], parts[1], true
		}
	}
	return "", "", false
}

// ─── 书籍 ─────────────────────────────────────────────────────────────

// Books 返回某榜第 offset 条起的 limit 本书。
// 接口固定 20 条/页，这里按 offset/limit 折算成所需页再切片，对调用方表现为连续分页。
func Books(tab string, filters map[string]string, offset, limit int) ([]Book, error) {
	g, typ, ok := ParseTab(tab)
	if !ok {
		return nil, fmt.Errorf("未知的起点榜单: %s", tab)
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = pageSize
	}

	first := offset/pageSize + 1
	drop := offset % pageSize
	out := make([]Book, 0, limit+drop)

	for page := first; len(out) < limit+drop; page++ {
		data, err := fetchPageCached(g, typ, filters, page)
		if err != nil {
			return nil, err
		}
		if len(data.Records) == 0 {
			break
		}
		for _, r := range data.Records {
			out = append(out, r.toBook(typ))
		}
		if data.IsLast == 1 {
			break
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

// FullResult 是整榜抓取的结果。
type FullResult struct {
	Books []Book
	// Complete 表示真的翻到了榜尾（isLast）。
	// ⚠ 只有 Complete 为真才能拿去做「掉榜」判定：截断的部分会被误判成掉榜。
	Complete bool
}

// BooksFull 把整榜翻完，用于当日整榜快照。
// maxPages 是硬上限（超过就停，Complete=false）；页与页之间按 fetchGap 限速。
func BooksFull(tab string, filters map[string]string, maxPages int) (FullResult, error) {
	g, typ, ok := ParseTab(tab)
	if !ok {
		return FullResult{}, fmt.Errorf("未知的起点榜单: %s", tab)
	}
	if maxPages <= 0 {
		maxPages = 60
	}

	res := FullResult{Books: []Book{}}
	for page := 1; page <= maxPages; page++ {
		if page > 1 {
			time.Sleep(fetchGap)
		}
		data, err := fetchPageCached(g, typ, filters, page)
		if err != nil {
			return res, err
		}
		if len(data.Records) == 0 {
			res.Complete = true
			break
		}
		for _, r := range data.Records {
			res.Books = append(res.Books, r.toBook(typ))
		}
		if data.IsLast == 1 {
			res.Complete = true
			break
		}
	}
	return res, nil
}

func (r apiRecord) toBook(typ string) Book {
	category := r.SubCat
	if category == "" {
		category = r.Cat
	}
	return Book{
		Rank:         r.RankNum,
		Title:        r.BName,
		Author:       r.BAuth,
		Category:     category,
		Status:       "", // 榜单接口不公开完结状态，详情页才有
		ReadingCount: r.RankCnt,
		WordCount:    r.Cnt,
		CoverURL:     fmt.Sprintf(coverTmpl, r.BID),
		BookID:       r.BID,
		Link:         fmt.Sprintf(bookTmpl, r.BID),
		MetricLabel:  metricOf(typ),
		Synopsis:     cleanText(r.Desc),
	}
}

func metricOf(typ string) string {
	if rt, ok := rankTypeOf(typ); ok {
		return rt.Metric
	}
	return ""
}

// cleanText 把简介里的换行压成空格，避免展示与 AI 素材里出现硬换行。
func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
