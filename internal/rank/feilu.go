package rank

import (
	"github.com/xingmang/xingmang-desktop/internal/feilu"
)

// feiluSource 是飞卢小说网榜单数据源。
//
// 走公开静态榜单页（GB2312 + HTML 解析），没有 JSON 接口。
// ⚠ 两个与其它源不同的点：需要 cookie jar 过 C3VK 挑战；数值口径按榜型各不相同
// （周/月/总收藏榜页面根本不给收藏数，只给字数）。
type feiluSource struct{}

func newFeiluSource() Source { return feiluSource{} }

func (feiluSource) ID() string   { return "feilu" }
func (feiluSource) Name() string { return "飞卢小说" }
func (feiluSource) Desc() string { return "数据来自飞卢榜单（公开榜单页，GB2312）" }

func (feiluSource) Groups() ([]Group, error) {
	groups := feilu.Groups()
	cats := feilu.Catalogs()

	// 题材是硬过滤（实测纯度 100%），作为每个榜单共用的第三层筛选维度
	filters := categoryFilter(cats)

	out := make([]Group, 0, len(groups))
	for _, g := range groups {
		items := make([]GroupItem, 0, len(g.Items))
		for _, rt := range g.Items {
			items = append(items, GroupItem{
				ID:      rt.Code,
				Name:    rt.Name,
				Filters: filters,
			})
		}
		out = append(out, Group{Title: g.Title, Items: items})
	}
	return out, nil
}

func (feiluSource) Books(q BookQuery) ([]Book, error) {
	books, err := feilu.Books(q.Tab, q.Filters, q.Offset, q.Limit)
	if err != nil {
		return nil, err
	}
	return toFeiluBooks(books), nil
}

// categoryFilter 把题材清单转成一层筛选维度。
// ⚠ 清单由接口动态给出：我先前从书库页只抓到 13 个，首页有 34 个，硬编码必漏。
// 全部题材（ID "0_0"）排在最前作为缺省值。
func categoryFilter(cats []feilu.Category) []FilterDim {
	if len(cats) == 0 {
		return nil
	}
	opts := make([]FilterOption, 0, len(cats)+1)
	opts = append(opts, FilterOption{ID: "", Name: "全部题材"})
	for _, c := range cats {
		opts = append(opts, FilterOption{ID: c.ID, Name: c.Name})
	}
	return []FilterDim{{
		Key:       "cat",
		Name:      "题材",
		Options:   opts,
		DefaultID: "",
	}}
}

func toFeiluBooks(in []feilu.Book) []Book {
	out := make([]Book, 0, len(in))
	for _, b := range in {
		out = append(out, Book{
			Rank:         b.Rank,
			Title:        b.Title,
			Author:       b.Author,
			Category:     b.Category,
			Status:       "", // 榜单页不公开完结状态，卡片标签回落显示「实时」
			ReadingCount: b.MetricValue,
			WordCount:    b.WordCount,
			CoverURL:     b.CoverURL,
			BookID:       b.BookID,
			Link:         b.Link,
			// 飞卢没有名次变动字段，连方向都没有：
			// 留空，等 internal/ranksnap 用快照算幅度；没快照就显示「—」
			RankDiffMode: "",
			Synopsis:     b.Synopsis,
			MetricLabel:  b.MetricName,
		})
	}
	return out
}
