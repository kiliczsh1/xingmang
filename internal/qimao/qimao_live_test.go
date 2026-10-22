package qimao

import (
	"os"
	"testing"
)

// TestLiveProbe 直连七猫分类接口做端到端验证。
// 默认跳过（需要联网），设置 QIMAO_LIVE_TEST=1 开启。
//
// 留着是因为这条线有两个容易判断错的地方：
//  1. 榜单接口「看起来」也接受题材参数——实际完全忽略（六种写法返回值一模一样），
//     所以题材必须走分类接口；
//  2. 分类接口缺任一参数就报 40001002，「全部」要用 "a" 而不是 0。
func TestLiveProbe(t *testing.T) {
	if os.Getenv("QIMAO_LIVE_TEST") == "" {
		t.Skip("需要联网验证七猫接口，设置 QIMAO_LIVE_TEST=1 开启")
	}

	opts, err := FetchOptions()
	if err != nil {
		t.Fatalf("选项拉取失败: %v", err)
	}
	t.Logf("频道 %d / 排序 %d / 一级分类 %d", len(opts.Channels), len(opts.Orders), len(opts.Categories))
	for _, c := range opts.Channels {
		t.Logf("   频道 %-4s %s", c.Value, c.Label)
	}
	for _, o := range opts.Orders {
		t.Logf("   排序 %-12s %s", o.Value, o.Label)
	}
	for i, c := range opts.Categories {
		if i >= 4 {
			t.Logf("   … 其余 %d 个一级分类", len(opts.Categories)-4)
			break
		}
		kids := make([]string, 0, len(c.Children))
		for _, k := range c.Children {
			kids = append(kids, k.Name)
		}
		t.Logf("   分类 %-6s %s -> %v", c.ID, c.Name, kids)
	}

	// 题材清单按频道过滤：分类树是跨频道的一份全集，每个分类的 Type 标明它属于哪个频道。
	// 这里不只比数量，而是**逐个题材真取一次书**——清单里不该出现取不到书的组合。
	for _, tc := range []struct {
		channel string
		want    int
	}{
		{"a", 23}, // 全部：全集
		{"0", 11}, // 男生原创
		{"1", 6},  // 女生原创
		{"2", 6},  // 出版图书
	} {
		genres := Genres(tc.channel)
		if len(genres) != tc.want {
			t.Errorf("Genres(%s) = %d 个，期望 %d", tc.channel, len(genres), tc.want)
		}
		empty := 0
		for _, g := range genres {
			books, err := Books(tc.channel+"_click", map[string]string{"cat": g.ID}, 0, 1)
			if err != nil {
				t.Errorf("Genres(%s) 里的 %s 取书失败: %v", tc.channel, g.Name, err)
				continue
			}
			if len(books) == 0 {
				empty++
				t.Errorf("Genres(%s) 里的 %s(%s) 实际没有书——频道过滤失效",
					tc.channel, g.Name, g.ID)
			}
		}
		t.Logf("频道 %s: 题材 %d 个，其中空组合 %d 个", tc.channel, len(genres), empty)
	}

	// 反向确认：不属于该频道的题材确实取不到书（否则说明过滤是多余的、或者 Type 字段已失效）
	if got, err := Books("0_click", map[string]string{"cat": "1_a"}, 0, 1); err != nil {
		t.Errorf("跨频道组合请求失败: %v", err)
	} else if len(got) != 0 {
		t.Errorf("男生原创·现代言情 本应为空，实际返回 %d 本——Type 字段可能已失效", len(got))
	} else {
		t.Log("跨频道组合（男生原创·现代言情）确认返回空，说明过滤是必要的")
	}

	t.Logf("")
	for _, tc := range []struct {
		tab string
		cat string
	}{
		{"0_click", ""},        // 男生原创·点击榜·全部题材
		{"0_click", "203_a"},   // 男生原创·点击榜·都市（一级）
		{"1_click", "1_a"},     // 女生原创·点击榜·现代言情（一级）
		{"0_click", "203_219"}, // 顺带确认二级题材也能用（UI 现在只发一级）
		{"1_favorite_uv", ""},  // 女生原创·收藏榜
		{"0_update_time", ""},  // 男生原创·更新榜
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
		cats := map[string]int{}
		for _, b := range books {
			cats[b.Category]++
		}
		t.Logf("tab=%-14s cat=%-8s -> %d 本, 题材=%v", tc.tab, tc.cat, len(books), cats)
		for i, b := range books {
			if i >= 2 {
				break
			}
			t.Logf("   #%-3d %-20s 作者=%-10s 题材=%-8s 状态=%-6s 字数=%-10s 更新=%s",
				b.Rank, b.Title, b.Author, b.Category, b.Status, b.WordCount, b.Updated)
		}
	}

	// 翻页折算：offset=30 应拿到第 31 名起
	pg, err := Books("0_click", map[string]string{"cat": "203_a"}, 30, 30)
	if err != nil {
		t.Fatalf("分页 ERR %v", err)
	}
	if len(pg) == 0 {
		t.Fatal("offset=30 应该还有书")
	}
	t.Logf("offset=30 -> %d 本, 首 #%d %s, 末 #%d", len(pg), pg[0].Rank, pg[0].Title, pg[len(pg)-1].Rank)

	// 单题材总量：确认比榜单接口的 100 本封顶更多
	_, total, err := ClassifyBooks(ClassifyQuery{Channel: "0", Category1: "203", Category2: "a", Order: "click", Page: 1})
	if err != nil {
		t.Fatalf("ClassifyBooks ERR %v", err)
	}
	t.Logf("都市(一级)·点击榜 总数 = %d", total)
	if total <= 100 {
		t.Errorf("分类接口单题材应有 100 本以上，实际 %d", total)
	}
}
