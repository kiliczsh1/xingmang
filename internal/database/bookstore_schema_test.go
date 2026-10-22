package database

import (
	"path/filepath"
	"testing"
)

// 作品文件夹化相关列/表必须真出现在库里 —— 这些列一旦漏建，
// 运行期表现为"读到空路径"，比报错更难查。
func TestBookFolderSchema(t *testing.T) {
	db, err := Init(filepath.Join(t.TempDir(), "schema.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	want := map[string][]string{
		"books":      {"folder_path"},
		"chapters":   {"file_path", "word_count", "content_hash", "file_mtime", "missing"},
		"characters": {"file_path"},
		"entries":    {"file_path"},
		"memos":      {"book_id", "file_path"},
	}
	for table, cols := range want {
		existing, err := tableColumns(db, table)
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		for _, c := range cols {
			if !existing[c] {
				t.Fatalf("%s 缺列 %s", table, c)
			}
		}
	}

	for _, table := range []string{"book_docs", "app_settings"} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("表 %s 未创建: %v", table, err)
		}
	}
}

// 记忆落盘（docs/books-memory-storage.md）的索引列与新表。
func TestMemorySchema(t *testing.T) {
	db, err := Init(filepath.Join(t.TempDir(), "memory.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	indexCols := []string{"file_path", "content_hash", "missing", "parse_error", "synced_at"}
	for _, table := range []string{"conversations", "agent_sessions", "graph_versions"} {
		existing, err := tableColumns(db, table)
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		for _, c := range indexCols {
			if !existing[c] {
				t.Fatalf("%s 缺记忆索引列 %s", table, c)
			}
		}
	}

	recCols, err := tableColumns(db, "memory_records")
	if err != nil {
		t.Fatalf("memory_records: %v", err)
	}
	for _, c := range append([]string{"book_id", "kind", "external_id", "title", "source", "status", "word_count"}, indexCols...) {
		if !recCols[c] {
			t.Fatalf("memory_records 缺列 %s", c)
		}
	}
	// v3 决策：数据库不保存正文副本。这两列一旦出现，说明又走回了"双份存储"。
	for _, banned := range []string{"content", "payload_json", "data_path"} {
		if recCols[banned] {
			t.Fatalf("memory_records 不应有正文副本列 %s", banned)
		}
	}

	for _, idx := range []string{"idx_memory_external", "idx_memory_book"} {
		var name string
		if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='index' AND name=?`, idx).Scan(&name); err != nil {
			t.Fatalf("索引 %s 未创建: %v", idx, err)
		}
	}

	// 唯一索引真的能挡住重复导入
	if _, err := db.Exec(`INSERT INTO memory_records (book_id, kind, external_id, file_path) VALUES (1,'analysis','abc','memory/analysis/x.md')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO memory_records (book_id, kind, external_id, file_path) VALUES (1,'analysis','abc','memory/analysis/y.md')`); err == nil {
		t.Fatal("同一个 (book_id, kind, external_id) 不应能插两次")
	}
}
