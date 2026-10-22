package fanqie

// 番茄小说章节抓取 —— 从 fanqie-rank-mcp/api.py 的
// fetch_chapters / fetch_chapter_content / parse_initial_state 移植。
//
//   - 章节目录走 fanqienovel.com/api/reader/directory/detail，无需 a_bogus 签名；
//   - 正文走 fanqienovel.com/reader/<chapterId> 页面，从 window.__INITIAL_STATE__
//     提取 reader.chapterData，清洗 HTML 后做 PUA 解密。
//   - 付费章（needPay=1 或 isPaidStory）正文拿不到，调用方按 IsFree 处理。

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const (
	directoryAPI = "https://fanqienovel.com/api/reader/directory/detail?bookId="
	readerPage   = "https://fanqienovel.com/reader/"
)

// Chapter 是目录里的一章。
type Chapter struct {
	ID     string `json:"id"`
	Index  int    `json:"index"`
	Title  string `json:"title"`
	IsFree bool   `json:"is_free"`
}

// ─── HTTP（带指数退避重试，对齐 api.py http_get_with_retry） ────────────

func httpGetWithRetry(url, ua string, maxRetries int) (string, error) {
	backoff := 600 * time.Millisecond
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		body, err := httpGet(url, ua)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if attempt >= maxRetries {
			break
		}
		sleep := backoff + time.Duration(rand.Intn(300))*time.Millisecond
		if sleep > 3*time.Second {
			sleep = 3 * time.Second
		}
		time.Sleep(sleep)
		backoff *= 2
		if backoff > 3*time.Second {
			backoff = 3 * time.Second
		}
	}
	return "", fmt.Errorf("重试%d次后失败: %w", maxRetries, lastErr)
}

// ─── 章节目录 ─────────────────────────────────────────────────────────

// FetchChapters 拉取全书章节目录（与 Python fetch_chapters 输出同构）。
func FetchChapters(bookID string) ([]Chapter, error) {
	body, err := httpGetWithRetry(directoryAPI+bookID, uaWeb, 3)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			ChapterListWithVolume json.RawMessage `json:"chapterListWithVolume"`
			List                  json.RawMessage `json:"list"`
			ItemList              json.RawMessage `json:"item_list"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, fmt.Errorf("章节目录解析失败: %w", err)
	}
	items := resp.Data.ChapterListWithVolume
	if len(items) == 0 || string(items) == "null" {
		items = resp.Data.List
	}
	if len(items) == 0 || string(items) == "null" {
		items = resp.Data.ItemList
	}
	if len(items) == 0 || string(items) == "null" {
		return nil, fmt.Errorf("章节目录为空")
	}

	// 可能是嵌套数组（按卷分组），展平一层
	var nested [][]map[string]json.RawMessage
	var flat []map[string]json.RawMessage
	if err := json.Unmarshal(items, &nested); err == nil {
		for _, vol := range nested {
			flat = append(flat, vol...)
		}
	} else if err := json.Unmarshal(items, &flat); err != nil {
		return nil, fmt.Errorf("章节列表结构不符合预期: %w", err)
	}

	chapters := make([]Chapter, 0, len(flat))
	for i, c := range flat {
		needPay := rawInt(c, "needPay", "need_pay")
		if needPay == 0 && !rawKeyExists(c, "needPay", "need_pay") {
			needPay = 1 // 字段缺失时按付费处理（对齐 Python 默认）
		}
		isPaidStory := true
		if v, ok := c["isPaidStory"]; ok {
			_ = json.Unmarshal(v, &isPaidStory)
		}
		chapters = append(chapters, Chapter{
			ID:     rawStr(c, "itemId", "item_id"),
			Index:  i + 1,
			Title:  decryptPUA(rawStr(c, "title", "chapterName")),
			IsFree: needPay == 0 && !isPaidStory,
		})
	}
	return chapters, nil
}

func rawKeyExists(m map[string]json.RawMessage, keys ...string) bool {
	for _, k := range keys {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

// ─── 章节正文 ─────────────────────────────────────────────────────────

// FetchChapterContent 拉取单章正文（与 Python fetch_chapter_content 同构）。
func FetchChapterContent(chapterID string) (title, content string, err error) {
	html, err := httpGetWithRetry(readerPage+chapterID, uaWeb, 1)
	if err != nil {
		return "", "", err
	}
	state := parseInitialState(html)
	if state == nil {
		return "", "", fmt.Errorf("页面缺少 __INITIAL_STATE__")
	}
	reader, _ := state["reader"].(map[string]interface{})
	if reader == nil {
		return "", "", nil
	}
	chData, _ := reader["chapterData"].(map[string]interface{})
	if chData == nil {
		return "", "", nil
	}
	if t, ok := chData["title"].(string); ok {
		title = strings.TrimSpace(t)
	}
	raw, _ := chData["content"].(string)
	content = decryptPUA(stripHTMLTags(cleanHTMLContent(raw)))
	return title, content, nil
}

// ─── __INITIAL_STATE__ 提取（对齐 api.py parse_initial_state） ─────────

var (
	reTrailingComma = regexp.MustCompile(`,\s*([}\]])`)
	reHTMLTags      = regexp.MustCompile(`<[^>]+>`)
)

// parseInitialState 从页面 HTML 中提取 window.__INITIAL_STATE__ JSON 对象。
// 手工做括号配平（跳过字符串与转义），再修掉 undefined 与尾逗号。
func parseInitialState(html string) map[string]interface{} {
	const marker = "window.__INITIAL_STATE__"
	pos := strings.Index(html, marker)
	if pos == -1 {
		return nil
	}
	rest := html[pos+len(marker):]
	eq := strings.Index(rest, "=")
	if eq == -1 {
		return nil
	}
	trimmed := strings.TrimSpace(rest[eq+1:])
	if !strings.HasPrefix(trimmed, "{") {
		return nil
	}

	depth, inString, escaped, end := 0, false, false, 0
	for i, ch := range trimmed {
		if escaped {
			escaped = false
			continue
		}
		switch {
		case ch == '\\' && inString:
			escaped = true
		case ch == '"':
			inString = !inString
		case !inString && ch == '{':
			depth++
		case !inString && ch == '}':
			depth--
			if depth == 0 {
				end = i + 1
			}
		}
		if end != 0 {
			break
		}
	}
	if depth != 0 || end == 0 {
		return nil
	}

	raw := strings.ReplaceAll(trimmed[:end], "undefined", "null")
	cleaned := reTrailingComma.ReplaceAllString(raw, "$1")
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &state); err != nil {
		return nil
	}
	return state
}

// ─── HTML 清洗（对齐 api.py clean_html_content / strip_html_tags） ─────

func cleanHTMLContent(raw string) string {
	r := strings.NewReplacer(
		"<p>", "\n", "</p>", "",
		"<br>", "\n", "<br/>", "\n", "<br />", "\n",
		"&nbsp;", " ", "&lt;", "<", "&gt;", ">",
		"&amp;", "&", "&quot;", `"`,
	)
	return strings.TrimSpace(r.Replace(raw))
}

func stripHTMLTags(s string) string {
	return reHTMLTags.ReplaceAllString(s, "")
}
