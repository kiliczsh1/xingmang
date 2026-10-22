package rank

// 章节能力（可选接口）：AI 拆书弹窗的「关联章节」用。
// 数据来源与榜单同构：外部番茄桥优先（/chapters、/chapter_content），
// 桥不可用或未配置时回退 internal/fanqie 的 Go 原生实现。

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/xingmang/xingmang-desktop/internal/fanqie"
)

// RankChapter 是在线书籍的一章。Content 仅在显式拉取正文时填充。
type RankChapter struct {
	ID     string `json:"id"`
	Index  int    `json:"index"`
	Title  string `json:"title"`
	IsFree bool   `json:"isFree"`
}

// ChapterLister 是可选能力：能在线取书籍章节目录与单章正文。
// 榜单源逐个实现；不支持的源由调用方降级隐藏。
type ChapterLister interface {
	// ChapterList 返回该书前 limit 章目录（limit<=0 表示全部）。
	ChapterList(bookID string, limit int) ([]RankChapter, error)
	// ChapterContent 返回单章标题与正文。bookID 供需要书号+章号定位正文的源使用（番茄忽略它）。
	ChapterContent(bookID, chapterID string) (title, content string, err error)
}

// Chapters 实现 ChapterLister（番茄桥 / 内置双通道）。
func (c *BridgeClient) ChapterList(bookID string, limit int) ([]RankChapter, error) {
	url := fmt.Sprintf("%s/chapters?book_id=%s&limit=%d", c.baseURL, bookID, max(limit, 0))
	key := fmt.Sprintf("chapters:%s:%d", bookID, limit)

	var payload struct {
		Chapters []struct {
			ID     string `json:"id"`
			Index  int    `json:"index"`
			Title  string `json:"title"`
			IsFree bool   `json:"is_free"`
		} `json:"chapters"`
	}
	err := c.get(key, url, &payload, func() ([]byte, error) {
		chapters, err := fanqie.FetchChapters(bookID)
		if err != nil {
			return nil, err
		}
		if limit > 0 && len(chapters) > limit {
			chapters = chapters[:limit]
		}
		return json.Marshal(map[string]interface{}{"chapters": chapters})
	})
	if err != nil {
		return nil, err
	}
	out := make([]RankChapter, 0, len(payload.Chapters))
	for _, ch := range payload.Chapters {
		out = append(out, RankChapter{ID: ch.ID, Index: ch.Index, Title: ch.Title, IsFree: ch.IsFree})
	}
	return out, nil
}

// ChapterContent 实现 ChapterLister（番茄桥 / 内置双通道）。bookID 番茄用不到。
func (c *BridgeClient) ChapterContent(bookID, chapterID string) (string, string, error) {
	if c.baseURL != "" && !c.isDown() {
		url := fmt.Sprintf("%s/chapter_content?chapter_id=%s", c.baseURL, chapterID)
		if body, err := c.do(url); err == nil {
			var payload struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if json.Unmarshal(body, &payload) == nil {
				return payload.Title, payload.Content, nil
			}
		} else if c.markDown() {
			log.Printf("外部番茄桥不可用 (%v)，章节正文已切换为内置实现", err)
		}
	}
	return fanqie.FetchChapterContent(chapterID)
}

// ChapterListerOf 返回数据源的章节能力；不支持时返回 nil。
func ChapterListerOf(src Source) ChapterLister {
	if cl, ok := src.(ChapterLister); ok {
		return cl
	}
	return nil
}
