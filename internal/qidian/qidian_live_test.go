package qidian

import (
	"encoding/json"
	"os"
	"testing"
)

// TestLiveProbe 直连起点接口做端到端验证。
// 默认跳过（需要联网），设置 QIDIAN_LIVE_TEST=1 开启。
// 留着它是因为起点接口的字段类型会变：filters 的 value 一开始按 int 解析，
// 直到 yearmonth 给出 "202608" 这种字符串才炸——这种坑只有打真接口才发现。
func TestLiveProbe(t *testing.T) {
	if os.Getenv("QIDIAN_LIVE_TEST") == "" {
		t.Skip("需要联网验证起点接口，设置 QIDIAN_LIVE_TEST=1 开启")
	}
	catalogs := Catalogs()
	t.Logf("频道数 = %d", len(catalogs))
	for _, c := range catalogs {
		t.Logf("── %s (%s) 榜单数=%d", c.Name, c.Gender, len(c.Items))
		for _, it := range c.Items {
			desc := ""
			for _, f := range it.Filters {
				names := make([]string, 0, len(f.Options))
				for _, o := range f.Options {
					names = append(names, o.ID+"="+o.Name)
				}
				b, _ := json.Marshal(names)
				desc += f.Key + "(" + f.Name + ",默认 " + f.DefaultID + ")" + string(b) + " "
			}
			if desc == "" {
				desc = "(无筛选器)"
			}
			t.Logf("   %-22s %s", it.Code+"/"+it.Name, desc)
		}
	}

	t.Logf("")
	for _, tc := range []struct {
		tab     string
		filters map[string]string
	}{
		{"male_hotsales", map[string]string{"catId": "-1"}},
		{"male_yuepiao", map[string]string{"catId": "-1"}},
		{"male_yuepiao", map[string]string{"catId": "-1", "yearmonth": "202608"}},
		{"female_hotsales", nil},
		{"female_yuepiao", map[string]string{"catId": "21"}},
		{"male_hotsales", map[string]string{"catId": "21"}},
	} {
		books, err := Books(tc.tab, tc.filters, 0, 30)
		if err != nil {
			t.Errorf("Books(%s, %v) ERR %v", tc.tab, tc.filters, err)
			continue
		}
		t.Logf("Books(%s, %v) -> %d 本", tc.tab, tc.filters, len(books))
		for i, b := range books {
			if i >= 2 {
				break
			}
			t.Logf("   #%d %s | %s | %s %s | %s | %s | %s",
				b.Rank, b.Title, b.Author, b.MetricLabel, b.ReadingCount, b.WordCount, b.Category, b.CoverURL)
		}
	}

	// 翻页折算
	pg, err := Books("male_hotsales", map[string]string{"catId": "-1"}, 30, 30)
	if err != nil {
		t.Fatalf("分页 ERR %v", err)
	}
	t.Logf("offset=30 -> %d 本, 首 #%d %s, 末 #%d", len(pg), pg[0].Rank, pg[0].Title, pg[len(pg)-1].Rank)

	// 整榜抓取：挑一个短榜（女频新人作者榜 total≈169 → 9 页）
	res, err := BooksFull("female_newauthor", nil, 60)
	if err != nil {
		t.Fatalf("BooksFull ERR %v", err)
	}
	t.Logf("BooksFull(female_newauthor) -> %d 本, complete=%v", len(res.Books), res.Complete)
	if !res.Complete {
		t.Error("女频新人作者榜应该能翻完（total≈169）")
	}
	if len(res.Books) > 0 {
		t.Logf("   首 #%d %s / 末 #%d %s", res.Books[0].Rank, res.Books[0].Title,
			res.Books[len(res.Books)-1].Rank, res.Books[len(res.Books)-1].Title)
	}
}
