package qimao

// 七猫章节抓取 —— 基于公开网页接口（同 qimao.com 书页使用的前端 API）：
//
//   - 章节目录：GET https://www.qimao.com/api/book/chapter-list?book_id=<id>
//     返回 data.chapters[] = {id, title, index, is_vip}，公开 JSON、无签名无混淆；
//   - 章节正文：GET https://www.qimao.com/shuku/<book_id>-<chapter_id>/
//     网页里 div.article 是正文段落，提取后去标签。
//
// VIP 章（is_vip=1）网页只给试读或整章加密，调用方按 IsFree 处理。

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

const (
	chapterListURL = "https://www.qimao.com/api/book/chapter-list?book_id="
	shukuURL       = "https://www.qimao.com/shuku/"
)

// Chapter 是目录里的一章。
type Chapter struct {
	ID     string `json:"id"`
	Index  int    `json:"index"`
	Title  string `json:"title"`
	IsFree bool   `json:"is_free"`
}

// FetchChapters 拉取全书章节目录。limit>0 时只返回前 limit 章。
func FetchChapters(bookID string, limit int) ([]Chapter, error) {
	body, err := get(chapterListURL + bookID)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			Chapters []struct {
				ID    string      `json:"id"`
				Title string      `json:"title"`
				Index interface{} `json:"index"`
				IsVip interface{} `json:"is_vip"`
			} `json:"chapters"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("七猫章节目录解析失败: %w", err)
	}
	list := resp.Data.Chapters
	if len(list) == 0 {
		return nil, fmt.Errorf("七猫章节目录为空")
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	chapters := make([]Chapter, 0, len(list))
	for i, c := range list {
		chapters = append(chapters, Chapter{
			ID:     strings.TrimSpace(c.ID),
			Index:  looseInt(c.Index, i+1),
			Title:  strings.TrimSpace(c.Title),
			IsFree: looseInt(c.IsVip, 0) == 0,
		})
	}
	return chapters, nil
}

// FetchChapterContent 拉取单章正文（shuku 网页解析）。
func FetchChapterContent(bookID, chapterID string) (title, content string, err error) {
	page, err := get(shukuURL + bookID + "-" + chapterID + "/")
	if err != nil {
		return "", "", err
	}
	htmlText := string(page)

	if t := reQmTitle.FindStringSubmatch(htmlText); t != nil {
		inner := t[1]
		if inner == "" {
			inner = t[2] // h1 分支的捕获组
		}
		title = strings.TrimSpace(html.UnescapeString(stripTags(inner)))
	}
	m := reQmArticle.FindStringSubmatch(htmlText)
	if m == nil {
		return "", "", fmt.Errorf("七猫正文区块未命中（页面结构可能变化或为 VIP 章）")
	}
	// 先把段落标签转换行、再去标签，否则段落会被拼成一段
	content = strings.TrimSpace(normalizeParagraphs(stripTags(reQmPara.ReplaceAllString(m[1], "\n"))))
	if title == "" {
		title = chapterTitleFallback(htmlText)
	}
	return title, content, nil
}

// ─── 页面解析辅助 ─────────────────────────────────────────────────────

var (
	// 正文区块：div.article（内含 <p> 段落，懒惰匹配到第一个 </div>）
	reQmArticle = regexp.MustCompile(`(?s)<div[^>]*class="[^"]*\barticle\b[^"]*"[^>]*>(.*?)(?:</div>)`)
	// 章节标题：shuku 页用 <h2 class="chapter-title">；h1 是别的站型的兜底
	reQmTitle = regexp.MustCompile(`(?s)<h2[^>]*class="[^"]*\bchapter-title\b[^"]*"[^>]*>(.*?)</h2>|<h1[^>]*>(.*?)</h1>`)
	reQmPage  = regexp.MustCompile(`<title>(.*?)</title>`)
	// 段落标签转换行
	reQmPara = regexp.MustCompile(`(?i)</p\s*>|<br\s*/?>`)
	reQmTag  = regexp.MustCompile(`<[^>]+>`)
)

func stripTags(s string) string { return reQmTag.ReplaceAllString(s, "") }

// normalizeParagraphs 把段落分隔收敛为单个换行并去多余空白行
func normalizeParagraphs(s string) string {
	s = reQmPara.ReplaceAllString(s, "\n")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}

func chapterTitleFallback(htmlText string) string {
	if m := reQmPage.FindStringSubmatch(htmlText); m != nil {
		return strings.TrimSpace(html.UnescapeString(m[1]))
	}
	return ""
}

// looseInt 宽松转 int（接口字段可能是字符串或数字），失败返回 def。
func looseInt(v interface{}, def int) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(x)); err == nil {
			return n
		}
	}
	return def
}
