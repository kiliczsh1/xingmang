// Package ranksnap 记录榜单名次快照，用来算出真实的名次变动幅度，
// 并在数据源能抓完整榜时给出可信的「新上榜」与「掉榜」。
//
// 背景：番茄接口自带 rankPosDiff（真实差值），七猫只给 index_change（-1/0/1 方向），
// 起点连方向都没有。本包按天落快照、下次抓取时对比，把三个源的口径统一成「有幅度的名次变动」。
//
// 口径约定（改动时必须和前端 AiRanking.vue 的 diffHtml 一起看）：
//   - 对比基准是「同一 source+榜标识下、日期早于今天的最近一份快照」，不是严格意义的"昨天"。
//     中间隔几天没打开也照样对得上前一次，语义是「相比上次采集」。
//   - 同一天内重复抓取只覆盖当天的行，不产生变动 —— 榜单以天为粒度。
//   - 原生就有幅度的源（番茄，模式 step）不被覆盖。
//   - 快照分两种 scope：对比时**优先用整榜快照**（覆盖完整，能直接判定"上一份里没有这本"）；
//     只有当前页快照时才退化成窗口守卫，宁可不标也不给假「新上榜」。
package ranksnap

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/rank"
)

// 快照范围。
const (
	// ScopePage 是用户实际翻过的那几页，任何数据源都能落。
	ScopePage = "page"
	// ScopeFull 是整榜翻完，只有可穷尽的数据源（起点）能落。
	// 有了它才能判定「新上榜」和「掉榜」——否则没翻过的页会被误判成新上榜。
	ScopeFull = "full"
)

// retentionDays 快照保留天数。按天 100~1000 行量级估算，半年足够回溯。
const retentionDays = 180

// pruned 保证清理每个进程只跑一次，不让每个请求都带上一条 DELETE。
var pruned atomic.Bool

// Enrich 把本次抓到的 books 落一份当日快照（scope 标明来源范围），
// 并回填 RankDiff / RankDiffMode / IsNew / DaysOnList。
// 任何数据库问题都不会让请求失败：出错时原样返回 books，只打日志。
func Enrich(db *sql.DB, source, key, scope string, books []rank.Book) []rank.Book {
	if db == nil || len(books) == 0 {
		return books
	}
	if scope == "" {
		scope = ScopePage
	}

	today := time.Now().Format("2006-01-02")
	prev, err := loadPrevious(db, source, key, today)
	if err != nil {
		log.Printf("榜单快照读取失败 (%s/%s): %v", source, key, err)
		return books
	}

	for i := range books {
		b := &books[i]
		if b.BookID == "" {
			continue
		}
		prevRank, seen := prev.ranks[b.BookID]
		switch {
		case seen:
			// 有上次的名次 → 直接算幅度（原生已是 step 的源保留平台自己的口径）
			if b.RankDiffMode != rank.RankDiffModeStep {
				b.RankDiff = prevRank - b.Rank
				b.RankDiffMode = rank.RankDiffModeStep
			}
		case prev.date != "" && prev.credibleFor(b.Rank):
			b.IsNew = true
			if b.RankDiffMode != rank.RankDiffModeStep {
				b.RankDiff = 0
				b.RankDiffMode = rank.RankDiffModeStep
			}
		}
	}

	if err := save(db, source, key, scope, today, books); err != nil {
		log.Printf("榜单快照写入失败 (%s/%s): %v", source, key, err)
	}

	// 在榜天数在写入之后统计，这样今天也算一天（首次出现即「在榜 1 天」）
	if days, err := loadDaysOnList(db, source, key, books); err != nil {
		log.Printf("在榜天数统计失败 (%s/%s): %v", source, key, err)
	} else {
		for i := range books {
			books[i].DaysOnList = days[books[i].BookID]
		}
	}
	return books
}

// previous 是上一份快照。
type previous struct {
	date  string
	scope string
	ranks map[string]int
	// winMin/winMax 是上一份快照覆盖过的名次窗口，只有 page 快照需要它。
	winMin int
	winMax int
}

// credibleFor 判断「上一份快照里没这本书」是否可信。
//
//   - 整榜快照覆盖完整，落在榜内就是可信的新上榜；
//   - 当前页快照只记了用户翻过的行，只有名次落在覆盖过的窗口 [min,max] 内才敢下结论，
//     窗口外无法判断，宁可不标也不给假「新上榜」。
func (p previous) credibleFor(rankPos int) bool {
	if p.scope == ScopeFull {
		return true
	}
	return rankPos >= p.winMin && rankPos <= p.winMax
}

// loadPrevious 取「日期早于今天的最近一份快照」，整榜快照优先。
func loadPrevious(db *sql.DB, source, key, today string) (previous, error) {
	var prev previous

	// 先找整榜；没有整榜再退回当前页
	for _, scope := range []string{ScopeFull, ScopePage} {
		err := db.QueryRow(
			`SELECT captured_date FROM rank_snapshots
			  WHERE source = ? AND tab = ? AND scope = ? AND captured_date < ?
			  ORDER BY captured_date DESC LIMIT 1`,
			source, key, scope, today,
		).Scan(&prev.date)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return previous{}, err
		}
		prev.scope = scope
		break
	}
	if prev.date == "" {
		return prev, nil
	}

	rows, err := db.Query(
		`SELECT book_id, rank FROM rank_snapshots
		  WHERE source = ? AND tab = ? AND scope = ? AND captured_date = ?`,
		source, key, prev.scope, prev.date,
	)
	if err != nil {
		return previous{}, err
	}
	defer rows.Close()

	prev.ranks = map[string]int{}
	hasRow := false
	for rows.Next() {
		var id string
		var pos int
		if err := rows.Scan(&id, &pos); err != nil {
			return previous{}, err
		}
		prev.ranks[id] = pos
		if !hasRow || pos < prev.winMin {
			prev.winMin = pos
		}
		if !hasRow || pos > prev.winMax {
			prev.winMax = pos
		}
		hasRow = true
	}
	return prev, rows.Err()
}

// loadDaysOnList 统计每本在快照里出现过的天数（含今天）。
func loadDaysOnList(db *sql.DB, source, key string, books []rank.Book) (map[string]int, error) {
	ids := make([]string, 0, len(books))
	seen := map[string]bool{}
	for _, b := range books {
		if b.BookID != "" && !seen[b.BookID] {
			seen[b.BookID] = true
			ids = append(ids, b.BookID)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	q := `SELECT book_id, COUNT(DISTINCT captured_date) FROM rank_snapshots
	      WHERE source = ? AND tab = ? AND book_id IN (` + placeholders + `)
	      GROUP BY book_id`

	args := make([]interface{}, 0, len(ids)+2)
	args = append(args, source, key)
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}

// save 当日快照写入（同键覆盖当天旧值）。
func save(db *sql.DB, source, key, scope, today string, books []rank.Book) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO rank_snapshots (source, tab, book_id, rank, title, author, heat, scope, captured_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (source, tab, book_id, captured_date, scope)
		 DO UPDATE SET rank = excluded.rank, title = excluded.title,
		               author = excluded.author, heat = excluded.heat`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range books {
		if b.BookID == "" {
			continue
		}
		if _, err := stmt.Exec(source, key, b.BookID, b.Rank, b.Title, b.Author, b.ReadingCount, scope, today); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	prune(db, today)
	return nil
}

// ─── 掉榜 ─────────────────────────────────────────────────────────────

// Drop 是一本掉榜的书。
type Drop struct {
	BookID string `json:"bookId"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Rank   int    `json:"rank"` // 最后一次出现在整榜时的名次
}

// DropResult 是掉榜查询结果。
type DropResult struct {
	// Ready 表示今天已有整榜快照。没有它就不能判定掉榜（不能拿半份数据当基准）。
	Ready bool `json:"ready"`
	// BaseDate 是对比基准日（上一份整榜的日期）。
	BaseDate string `json:"baseDate"`
	Drops    []Drop `json:"drops"`
}

// Drops 找出「上一份整榜里有、今天整榜里没有」的书。
//
// 只有今天的整榜快照抓完之后才有意义：Ready=false 时前端应显示"整榜抓取中"而不是 0。
// 基准日为空（今天第一次整榜）时 Drops 为空，属于正常。
func Drops(db *sql.DB, source, key string) (DropResult, error) {
	res := DropResult{Drops: []Drop{}}
	if db == nil {
		return res, nil
	}
	today := time.Now().Format("2006-01-02")

	var todayFull sql.NullString
	if err := db.QueryRow(
		`SELECT MAX(captured_date) FROM rank_snapshots
		  WHERE source = ? AND tab = ? AND scope = ? AND captured_date = ?`,
		source, key, ScopeFull, today,
	).Scan(&todayFull); err != nil {
		return res, err
	}
	if !todayFull.Valid {
		return res, nil // 今天的整榜还没抓完
	}
	res.Ready = true

	// MAX() 在无匹配行时返回 NULL，必须用 NullString 接，否则扫进 string 会报错
	var base sql.NullString
	if err := db.QueryRow(
		`SELECT MAX(captured_date) FROM rank_snapshots
		  WHERE source = ? AND tab = ? AND scope = ? AND captured_date < ?`,
		source, key, ScopeFull, today,
	).Scan(&base); err != nil {
		return res, err
	}
	res.BaseDate = base.String
	if res.BaseDate == "" {
		return res, nil // 今天第一次整榜，没有可比基准
	}

	rows, err := db.Query(
		`SELECT p.book_id, p.title, p.author, p.rank
		   FROM rank_snapshots p
		  WHERE p.source = ? AND p.tab = ? AND p.scope = ? AND p.captured_date = ?
		    AND p.book_id NOT IN (
		          SELECT t.book_id FROM rank_snapshots t
		           WHERE t.source = ? AND t.tab = ? AND t.scope = ? AND t.captured_date = ?
		    )
		  ORDER BY p.rank`,
		source, key, ScopeFull, res.BaseDate,
		source, key, ScopeFull, today,
	)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Drop
		if err := rows.Scan(&d.BookID, &d.Title, &d.Author, &d.Rank); err != nil {
			return res, err
		}
		res.Drops = append(res.Drops, d)
	}
	return res, rows.Err()
}

// prune 清理超期快照。每个进程只跑一次。
func prune(db *sql.DB, today string) {
	if !pruned.CompareAndSwap(false, true) {
		return
	}
	day, err := time.Parse("2006-01-02", today)
	if err != nil {
		return
	}
	cutoff := day.AddDate(0, 0, -retentionDays).Format("2006-01-02")
	if _, err := db.Exec(`DELETE FROM rank_snapshots WHERE captured_date < ?`, cutoff); err != nil {
		log.Printf("榜单快照清理失败: %v", err)
	}
}
