package rank

import (
	"fmt"

	"github.com/xingmang/xingmang-desktop/internal/qidian"
)

// qidianSource 是起点中文网榜单数据源。
//
// 抓取逻辑与限速都在 internal/qidian 里（走移动端 majax 接口，无字体反爬），
// 本类型只做类型转换。它是目前唯一实现 FullCapturer 的源：
// 起点单榜能翻完（total 500~1000），所以整榜快照是完整可信的，
// 新上榜与掉榜判定才有意义。
type qidianSource struct{}

func newQidianSource() Source { return qidianSource{} }

func (qidianSource) ID() string   { return "qidian" }
func (qidianSource) Name() string { return "起点中文网" }
func (qidianSource) Desc() string { return "数据来自起点榜单（移动端公开接口）" }

func (qidianSource) Groups() ([]Group, error) {
	catalogs := qidian.Catalogs()
	out := make([]Group, 0, len(catalogs))
	for _, c := range catalogs {
		items := make([]GroupItem, 0, len(c.Items))
		for _, it := range c.Items {
			items = append(items, GroupItem{
				ID:      qidian.TabID(c.Gender, it.Code),
				Name:    it.Name,
				Filters: toFilterDims(it.Filters),
			})
		}
		out = append(out, Group{Title: c.Name, Items: items})
	}
	return out, nil
}

func (qidianSource) Books(q BookQuery) ([]Book, error) {
	books, err := qidian.Books(q.Tab, q.Filters, q.Offset, q.Limit)
	if err != nil {
		return nil, err
	}
	return toBooks(books), nil
}

// BooksFull 实现 FullCapturer。
// 没翻到榜尾就返回错误：截断的「整榜」会把尾部误报成掉榜，宁可这次不落快照。
func (qidianSource) BooksFull(q BookQuery, maxPages int) ([]Book, error) {
	res, err := qidian.BooksFull(q.Tab, q.Filters, maxPages)
	if err != nil {
		return nil, err
	}
	if !res.Complete {
		return nil, fmt.Errorf("起点整榜被抓取页数上限截断（%d 页），本次不落快照", maxPages)
	}
	return toBooks(res.Books), nil
}

// WorthFullCapture 只在「各筛选维度都是默认值」（即全站视图）时才值得整榜抓取。
// 起点的题材有 15 种、还叠加月份/周期，逐一变体全抓会打出上百次请求；
// 而掉榜/新上榜真正要看的也是全站榜。筛选视图退化为当前页快照（幅度仍准，新上榜不做判断）。
func (qidianSource) WorthFullCapture(q BookQuery) bool {
	for _, c := range qidian.Catalogs() {
		for _, it := range c.Items {
			if qidian.TabID(c.Gender, it.Code) != q.Tab {
				continue
			}
			for _, d := range it.Filters {
				cur, chosen := q.Filters[d.Key]
				if !chosen || cur == "" {
					continue // 没选等于用缺省
				}
				if cur != d.DefaultID {
					return false
				}
			}
			return true
		}
	}
	return false
}

func toFilterDims(in []qidian.FilterDim) []FilterDim {
	if len(in) == 0 {
		return nil
	}
	out := make([]FilterDim, 0, len(in))
	for _, d := range in {
		opts := make([]FilterOption, 0, len(d.Options))
		for _, o := range d.Options {
			opts = append(opts, FilterOption{ID: o.ID, Name: o.Name})
		}
		out = append(out, FilterDim{
			Key:       d.Key,
			Name:      d.Name,
			Options:   opts,
			DefaultID: d.DefaultID,
		})
	}
	return out
}

func toBooks(in []qidian.Book) []Book {
	out := make([]Book, 0, len(in))
	for _, b := range in {
		out = append(out, Book{
			Rank:         b.Rank,
			Title:        b.Title,
			Author:       b.Author,
			Category:     b.Category,
			Status:       b.Status,
			ReadingCount: b.ReadingCount,
			WordCount:    b.WordCount,
			CoverURL:     b.CoverURL,
			BookID:       b.BookID,
			Link:         b.Link,
			// 起点榜单接口不给名次变动，连方向都没有：
			// RankDiffMode 留空，等 internal/ranksnap 用快照算幅度；没快照就显示「—」。
			RankDiffMode: "",
			Synopsis:     b.Synopsis,
			MetricLabel:  b.MetricLabel,
		})
	}
	return out
}
