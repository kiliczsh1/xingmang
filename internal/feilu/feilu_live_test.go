package feilu

import (
	"os"
	"testing"
	"time"
)

// TestLiveProbe 直连飞卢做端到端验证。
// 默认跳过（需要联网），设置 FEILU_LIVE_TEST=1 开启。
//
// 留着是因为这个站有两处很容易判断错：
//  1. 302 不是封禁而是 cookie 挑战（C3VK），没 cookie jar 会无限重定向；
//  2. 榜单页导航只显示附近几页，照它推断"最多 5 页"是错的，实际能翻 40 页以上。
func TestLiveProbe(t *testing.T) {
	if os.Getenv("FEILU_LIVE_TEST") == "" {
		t.Skip("需要联网验证飞卢接口，设置 FEILU_LIVE_TEST=1 开启")
	}

	cats := Catalogs()
	t.Logf("题材数 = %d", len(cats))
	for i, c := range cats {
		if i >= 8 {
			t.Logf("   … 其余 %d 个", len(cats)-8)
			break
		}
		t.Logf("   %-8s %s", c.ID, c.Name)
	}

	t.Logf("")
	t.Logf("榜单分组 = %d 组", len(Groups()))
	for _, g := range Groups() {
		t.Logf("   %s: %s", g.Title, codes(g.Items))
	}

	t.Logf("")
	for _, tc := range []struct {
		tab string
		cat string
	}{
		{"15", ""},      // 月票榜·全部题材
		{"15", "3_19"},  // 月票榜·架空历史
		{"15", "44_69"}, // 月票榜·动漫同人
		{"1", ""},       // 周点击榜
		{"7", ""},       // 周收藏榜（页面不给收藏数）
		{"10", ""},      // 总字数榜
	} {
		filters := map[string]string{}
		if tc.cat != "" {
			filters["cat"] = tc.cat
		}
		books, err := Books(tc.tab, filters, 0, 30)
		if err != nil {
			t.Errorf("Books(%s, %v) ERR %v", tc.tab, filters, err)
			continue
		}
		t.Logf("tab=%-3s cat=%-6s -> %d 本", tc.tab, tc.cat, len(books))
		for i, b := range books {
			if i >= 3 {
				break
			}
			t.Logf("   #%-3d %-22s 作者=%-10s 题材=%-8s %s=%s 字数=%s 封面=%v",
				b.Rank, b.Title, b.Author, b.Category, b.MetricName, b.MetricValue, b.WordCount, b.CoverURL != "")
		}
	}

	// 翻页折算：offset=30 应拿到第 31 名起
	pg, err := Books("15", nil, 30, 30)
	if err != nil {
		t.Fatalf("分页 ERR %v", err)
	}
	t.Logf("offset=30 -> %d 本, 首 #%d %s, 末 #%d", len(pg), pg[0].Rank, pg[0].Title, pg[len(pg)-1].Rank)

	// 深翻页：验证不是"最多 5 页"
	deep, err := Books("15", nil, 900, 30)
	if err != nil {
		t.Fatalf("深翻页 ERR %v", err)
	}
	if len(deep) == 0 {
		t.Error("offset=900 应该还有书（榜单实际能翻 40 页以上），拿到 0 本")
	} else {
		t.Logf("offset=900 -> %d 本, 首 #%d %s", len(deep), deep[0].Rank, deep[0].Title)
	}
}

func codes(items []rankType) string {
	out := ""
	for _, it := range items {
		out += it.Name + "(" + it.Code + ") "
	}
	return out
}

var _ = time.Second
