package ranksnap

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/database"
	"github.com/xingmang/xingmang-desktop/internal/rank"
)

const testKey = "male_hotsales|catId=-1"

// newTestDB 走真实的 database.Init，保证测试里的表结构与生产一致（DDL 只在 db.go 一处维护）。
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), "snap.db"))
	if err != nil {
		t.Fatalf("建库失败: %v", err)
	}
	t.Cleanup(database.Close)
	return db
}

// seed 写入一份历史快照。daysAgo=1 即「昨天」。
func seed(t *testing.T, db *sql.DB, source, scope string, daysAgo int, ranks map[string]int) string {
	t.Helper()
	date := time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02")
	for id, pos := range ranks {
		if _, err := db.Exec(
			`INSERT INTO rank_snapshots (source, tab, book_id, rank, scope, captured_date)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			source, testKey, id, pos, scope, date,
		); err != nil {
			t.Fatalf("写历史快照失败: %v", err)
		}
	}
	return date
}

// shape 造一本七猫口径的书（原生只有方向）。
func shape(id string, pos int) rank.Book {
	return rank.Book{BookID: id, Rank: pos, RankDiff: 1, RankDiffMode: rank.RankDiffModeDir}
}

func byID(books []rank.Book) map[string]rank.Book {
	out := map[string]rank.Book{}
	for _, b := range books {
		out[b.BookID] = b
	}
	return out
}

func TestEnrichNoHistoryKeepsDirection(t *testing.T) {
	db := newTestDB(t)
	books := Enrich(db, "qimao", testKey, ScopePage, []rank.Book{shape("a", 1), shape("b", 2)})

	for _, b := range books {
		if b.RankDiffMode != rank.RankDiffModeDir {
			t.Errorf("%s 无历史时不应改口径，得到 %q", b.BookID, b.RankDiffMode)
		}
		if b.IsNew {
			t.Errorf("%s 首次采集不应标成新上榜（没有可对比的上一份快照）", b.BookID)
		}
		if b.DaysOnList != 1 {
			t.Errorf("%s 首次出现应为在榜 1 天，得到 %d", b.BookID, b.DaysOnList)
		}
	}
}

func TestEnrichComputesMagnitudeAndNew(t *testing.T) {
	db := newTestDB(t)
	prev := map[string]int{}
	for i := 1; i <= 30; i++ {
		prev[bid(i)] = i
	}
	seed(t, db, "qimao", ScopePage, 1, prev)

	got := byID(Enrich(db, "qimao", testKey, ScopePage, []rank.Book{
		shape("b5", 3),     // 昨天第 5 → 今天第 3（上升 2）
		shape("b8", 10),    // 昨天第 8 → 今天第 10（下降 2）
		shape("new6", 6),   // 昨天窗口内没有 → 新上榜
		shape("out45", 45), // 名次在昨天窗口(1~30)之外 → 不能判为"新"
	}))

	if b := got["b5"]; b.RankDiff != 2 || b.RankDiffMode != rank.RankDiffModeStep {
		t.Errorf("b5 期望 ▲2/step，得到 %d/%q", b.RankDiff, b.RankDiffMode)
	}
	if b := got["b8"]; b.RankDiff != -2 {
		t.Errorf("b8 期望 ▼2，得到 %d", b.RankDiff)
	}
	if !got["new6"].IsNew {
		t.Error("new6 落在昨天覆盖过的名次窗口内且不在快照里，应标为新上榜")
	}
	if got["out45"].IsNew {
		t.Error("out45 在昨天覆盖窗口之外，不能判定为新上榜")
	}
}

func TestEnrichKeepsNativeStepDiff(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "fanqie", ScopeFull, 1, map[string]int{"f1": 2})

	in := []rank.Book{{BookID: "f1", Rank: 4, RankDiff: 7, RankDiffMode: rank.RankDiffModeStep}}
	books := Enrich(db, "fanqie", testKey, ScopePage, in)

	if books[0].RankDiff != 7 {
		t.Errorf("原生 step 口径不应被快照覆盖，期望 7，得到 %d", books[0].RankDiff)
	}
	if books[0].IsNew {
		t.Error("上一份快照里已经有 f1，不应标为新上榜")
	}
}

func TestEnrichSameDayIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "qimao", ScopePage, 1, map[string]int{"x": 5})

	first := Enrich(db, "qimao", testKey, ScopePage, []rank.Book{shape("x", 3)})
	if first[0].RankDiff != 2 {
		t.Fatalf("首次对比期望 ▲2，得到 %d", first[0].RankDiff)
	}

	// 同一天再抓一次：基准仍应是「昨天」，不能拿今天的行跟自己比
	second := Enrich(db, "qimao", testKey, ScopePage, []rank.Book{shape("x", 3)})
	if second[0].RankDiff != 2 {
		t.Errorf("同一天重复抓取应仍与昨天对比（▲2），得到 %d", second[0].RankDiff)
	}

	var rows int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM rank_snapshots WHERE source='qimao' AND book_id='x'`,
	).Scan(&rows); err != nil {
		t.Fatalf("统计行数失败: %v", err)
	}
	if rows != 2 {
		t.Errorf("同一天应只留一行当天快照（加昨天共 2 行），实际 %d 行", rows)
	}
}

// 整榜快照让「上份里没有这本」直接可信，不再需要名次窗口守卫。
func TestFullScopeMakesNewCredibleOutsideWindow(t *testing.T) {
	db := newTestDB(t)
	prev := map[string]int{}
	for i := 1; i <= 30; i++ {
		prev[bid(i)] = i
	}
	seed(t, db, "qidian", ScopeFull, 1, prev)

	got := byID(Enrich(db, "qidian", testKey, ScopePage, []rank.Book{
		{BookID: "brand-new", Rank: 45}, // 名次远在"昨天只看了前 30"的窗口之外
	}))

	if !got["brand-new"].IsNew {
		t.Error("整榜快照覆盖完整，榜内没有的书应判为新上榜（不受窗口限制）")
	}
}

func TestDaysOnListCountsDistinctDays(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "qimao", ScopePage, 2, map[string]int{"x": 5, "y": 9})
	seed(t, db, "qimao", ScopePage, 1, map[string]int{"x": 4})

	got := byID(Enrich(db, "qimao", testKey, ScopePage, []rank.Book{shape("x", 3), shape("y", 8)}))

	if got["x"].DaysOnList != 3 {
		t.Errorf("x 三天都有，期望在榜 3 天，得到 %d", got["x"].DaysOnList)
	}
	if got["y"].DaysOnList != 2 {
		t.Errorf("y 隔天缺一次但共出现 2 天，期望 2，得到 %d", got["y"].DaysOnList)
	}
}

func TestDropsNeedsTodayFullSnapshot(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "qidian", ScopeFull, 1, map[string]int{"old": 3})

	// 今天还没整榜 → 不能判定
	res, err := Drops(db, "qidian", testKey)
	if err != nil {
		t.Fatalf("Drops 失败: %v", err)
	}
	if res.Ready {
		t.Error("今天没有整榜快照时 Ready 应为 false，不能拿半份数据当基准")
	}
	if len(res.Drops) != 0 {
		t.Errorf("Ready=false 时不应给掉榜名单，得到 %d 条", len(res.Drops))
	}
}

func TestDropsFindsBooksMissingFromToday(t *testing.T) {
	db := newTestDB(t)
	base := seed(t, db, "qidian", ScopeFull, 1, map[string]int{"keep": 1, "gone": 2, "also-gone": 7})
	seed(t, db, "qidian", ScopeFull, 0, map[string]int{"keep": 1, "fresh": 2})

	res, err := Drops(db, "qidian", testKey)
	if err != nil {
		t.Fatalf("Drops 失败: %v", err)
	}
	if !res.Ready {
		t.Fatal("今天已有整榜快照，Ready 应为 true")
	}
	if res.BaseDate != base {
		t.Errorf("基准日应为 %s，得到 %s", base, res.BaseDate)
	}
	if len(res.Drops) != 2 {
		t.Fatalf("期望 2 本掉榜，得到 %d 本: %+v", len(res.Drops), res.Drops)
	}
	// 按昨天的名次排序
	if res.Drops[0].BookID != "gone" || res.Drops[0].Rank != 2 {
		t.Errorf("掉榜应按昨天名次排序，第一条应为 gone/2，得到 %s/%d", res.Drops[0].BookID, res.Drops[0].Rank)
	}
	if res.Drops[1].BookID != "also-gone" {
		t.Errorf("第二条应为 also-gone，得到 %s", res.Drops[1].BookID)
	}
}

// 只比同 scope：当前页快照不该被当成整榜的对比基准。
func TestDropsIgnoresPageScope(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "qidian", ScopePage, 1, map[string]int{"only-in-page": 3})
	seed(t, db, "qidian", ScopeFull, 0, map[string]int{"keep": 1})

	res, err := Drops(db, "qidian", testKey)
	if err != nil {
		t.Fatalf("Drops 失败: %v", err)
	}
	if !res.Ready {
		t.Error("今天已有整榜快照，Ready 应为 true")
	}
	if res.BaseDate != "" {
		t.Errorf("昨天只有当前页快照，不能当整榜基准，期望基准日为空，得到 %s", res.BaseDate)
	}
	if len(res.Drops) != 0 {
		t.Errorf("没有整榜基准时不应给掉榜名单，得到 %d 条", len(res.Drops))
	}
}

func bid(n int) string { return "b" + strconv.Itoa(n) }
