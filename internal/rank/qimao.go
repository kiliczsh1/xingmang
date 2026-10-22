package rank

import (
	"github.com/xingmang/xingmang-desktop/internal/qimao"
)

// qimaoSource 是七猫小说榜单数据源。
//
// 走的是**分类接口**（/api/classify/*）而不是榜单接口：
// 榜单接口完全不支持题材过滤，没有题材维度可言。
// 代价是拿不到热度数值、也没有日榜/月榜——见 internal/qimao 的包注释。
type qimaoSource struct{}

func newQimaoSource() Source { return qimaoSource{} }

func (qimaoSource) ID() string   { return "qimao" }
func (qimaoSource) Name() string { return "七猫小说" }
func (qimaoSource) Desc() string {
	return "数据来自七猫书库分类榜单（四种排序 + 题材）"
}

// Groups 返回「频道 → 榜」两层，每个榜带题材筛选维度。
func (qimaoSource) Groups() ([]Group, error) {
	channels := qimao.Channels()

	out := make([]Group, 0, len(channels))
	for _, ch := range channels {
		// 题材清单按频道过滤：分类树是跨频道的一份全集，
		// 不过滤就会出现「男生原创 · 现代言情」这种恒为空的组合
		filters := genreFilter(ch.Value)
		items := make([]GroupItem, 0, len(qimao.RankOrders()))
		for _, o := range qimao.RankOrders() {
			items = append(items, GroupItem{
				ID:      qimao.TabID(ch.Value, o.Code),
				Name:    o.Name,
				Filters: filters,
			})
		}
		out = append(out, Group{Title: ch.Label, Items: items})
	}
	return out, nil
}

func (qimaoSource) Books(q BookQuery) ([]Book, error) {
	books, err := qimao.Books(q.Tab, q.Filters, q.Offset, q.Limit)
	if err != nil {
		return nil, err
	}
	return toQimaoBooks(books), nil
}

// ─── 章节能力（ChapterLister） ────────────────────────────────────────
// 走 qimao.com 公开网页接口：目录 /api/book/chapter-list，正文 shuku 页面解析。

// ChapterList 实现 ChapterLister。
func (qimaoSource) ChapterList(bookID string, limit int) ([]RankChapter, error) {
	chapters, err := qimao.FetchChapters(bookID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]RankChapter, 0, len(chapters))
	for _, ch := range chapters {
		out = append(out, RankChapter{
			ID:     ch.ID,
			Index:  ch.Index,
			Title:  ch.Title,
			IsFree: ch.IsFree,
		})
	}
	return out, nil
}

// ChapterContent 实现 ChapterLister（七猫正文 URL 需要书号+章号）。
func (qimaoSource) ChapterContent(bookID, chapterID string) (string, string, error) {
	return qimao.FetchChapterContent(bookID, chapterID)
}

// genreFilter 把某个频道下可用的题材清单转成一层筛选维度。
// ⚠ 清单由接口动态给出（23 个一级分类），不硬编码。
func genreFilter(channel string) []FilterDim {
	genres := qimao.Genres(channel)
	if len(genres) == 0 {
		return nil
	}
	opts := make([]FilterOption, 0, len(genres)+1)
	opts = append(opts, FilterOption{ID: "", Name: "全部题材"})
	for _, g := range genres {
		opts = append(opts, FilterOption{ID: g.ID, Name: g.Name})
	}
	return []FilterDim{{
		Key:       "cat",
		Name:      "题材",
		Options:   opts,
		DefaultID: "",
	}}
}

func toQimaoBooks(in []qimao.Book) []Book {
	out := make([]Book, 0, len(in))
	for _, b := range in {
		out = append(out, Book{
			Rank:     b.Rank,
			Title:    b.Title,
			Author:   b.Author,
			Category: b.Category,
			Status:   b.Status,
			// 分类接口不给热度数值：留空，前端整列隐藏
			ReadingCount: "",
			WordCount:    b.WordCount,
			CoverURL:     b.CoverURL,
			BookID:       b.BookID,
			Link:         b.Link,
			// 七猫没有名次变动字段（榜单接口有个 index_change 方向，分类接口连它都没有）：
			// 留空，等 internal/ranksnap 用快照算幅度；没快照就显示「—」
			RankDiffMode: "",
			Synopsis:     b.Synopsis,
			MetricLabel:  "",
		})
	}
	return out
}
