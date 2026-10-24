package handlers

import (
	"path/filepath"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/database"
)

// 空库首次启动应种入 3 条（经验吧 1 + 教程吧 2）；重复调用不再插。
// ⚠ 同一 board 有多条种子时，「按标题判重」必须成立，否则第二条永远种不进去。
func TestSeedSystemPosts(t *testing.T) {
	db, err := database.Init(filepath.Join(t.TempDir(), "seed-test.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer db.Close()

	n, err := SeedSystemPosts(db)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if n != 3 {
		t.Fatalf("首次应种入 3 条，实际 %d", n)
	}

	var tutorialCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM experience_shares WHERE COALESCE(board,'experience') = 'tutorial'`).Scan(&tutorialCount); err != nil {
		t.Fatalf("count: %v", err)
	}
	if tutorialCount != 2 {
		t.Fatalf("教程吧应有 2 条种子，实际 %d", tutorialCount)
	}

	var title string
	if err := db.QueryRow(`SELECT title FROM experience_shares WHERE title LIKE '如何拆书%'`).Scan(&title); err != nil {
		t.Fatalf("查新种子失败: %v", err)
	}
	t.Logf("新种子标题: %s", title)

	// 打包版新装：种子 id 从 1 开始按 seedPosts 顺序编号
	rows, err := db.Query(`SELECT id, COALESCE(board,'experience') FROM experience_shares ORDER BY id`)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	defer rows.Close()
	wantBoards := []string{BoardExperience, BoardTutorial, BoardTutorial}
	idx := 0
	for rows.Next() {
		var id int
		var board string
		if err := rows.Scan(&id, &board); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if id != idx+1 {
			t.Fatalf("第 %d 条种子 id 应为 %d，实际 %d", idx+1, idx+1, id)
		}
		if board != wantBoards[idx] {
			t.Fatalf("第 %d 条种子 board 应为 %s，实际 %s", idx+1, wantBoards[idx], board)
		}
		idx++
	}
	if idx != 3 {
		t.Fatalf("应有 3 条种子，实际 %d", idx)
	}

	// 文件帖 id 水位从 max(id)+1 起算（server.go 里 sharestore.New 的口径）
	var maxID int
	if err := db.QueryRow(`SELECT COALESCE(MAX(id),0) FROM experience_shares`).Scan(&maxID); err != nil {
		t.Fatalf("max id: %v", err)
	}
	if maxID != 3 {
		t.Fatalf("种子最大 id 应为 3，实际 %d", maxID)
	}

	again, err := SeedSystemPosts(db)
	if err != nil {
		t.Fatalf("second seed: %v", err)
	}
	if again != 0 {
		t.Fatalf("第二次应跳过，实际插入 %d", again)
	}
}
