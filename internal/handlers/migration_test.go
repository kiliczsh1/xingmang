package handlers

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/xingmang/xingmang-desktop/internal/database"
)

// TestLegacyDBMigrationPreservesData 模拟「旧版库升级」：
// 用旧表结构建库并写入数据，跑一次 database.Init，验证
//  1. 新列（config_type / param_groups / default_params / 渠道扩展列）都补上
//  2. 老数据一行不少、一值不改
//  3. category 正确回填进 config_type
//  4. 新列可写可读
func TestLegacyDBMigrationPreservesData(t *testing.T) {
	p := filepath.Join(t.TempDir(), "legacy.db")

	raw, err := sql.Open("sqlite", p)
	if err != nil {
		t.Fatal(err)
	}
	// 旧版表结构：api_models 缺 config_type 及之后的所有列
	if _, err := raw.Exec(`
CREATE TABLE api_providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, provider_type TEXT NOT NULL,
  api_key TEXT, api_url TEXT NOT NULL, api_format TEXT DEFAULT 'openai', use_full_url INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE api_models (
  id INTEGER PRIMARY KEY AUTOINCREMENT, provider_id INTEGER NOT NULL, name TEXT NOT NULL, model TEXT NOT NULL,
  temperature REAL DEFAULT 0.7, max_tokens INTEGER DEFAULT 2000, top_p REAL DEFAULT 0.9,
  frequency_penalty REAL DEFAULT 0.0, description TEXT DEFAULT '', is_default INTEGER DEFAULT 0,
  enabled INTEGER DEFAULT 1, sort_order INTEGER DEFAULT 0, category TEXT DEFAULT '',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
INSERT INTO api_providers (name, provider_type, api_url) VALUES ('老渠道', 'other', 'https://old.example.com/v1');
INSERT INTO api_models (provider_id, name, model, category) VALUES
  (1, '老模型A', 'old-model-a', 'multimodal'),
  (1, '老模型B', 'old-model-b', 'image');
`); err != nil {
		t.Fatal(err)
	}
	raw.Close()

	db, err := database.Init(p)
	if err != nil {
		t.Fatalf("旧库升级失败: %v", err)
	}
	defer db.Close()

	// 1) 老数据无损
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM api_models`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("升级后模型数 = %d，期望 2（数据丢失）", n)
	}

	// 2) category 回填：multimodal → chat，image → image
	cases := map[string]string{"old-model-a": "chat", "old-model-b": "image"}
	for model, want := range cases {
		var got string
		if err := db.QueryRow(`SELECT COALESCE(config_type,'') FROM api_models WHERE model = ?`, model).Scan(&got); err != nil {
			t.Fatalf("读 %s 失败: %v", model, err)
		}
		if got != want {
			t.Fatalf("%s 的 config_type = %q，期望 %q", model, got, want)
		}
	}

	// 3) 新列可写可读（走 handler 的列清单，保证列名与顺序都对得上）
	if _, err := db.Exec(`UPDATE api_models SET param_groups = ?, default_params = ? WHERE model = ?`,
		`[{"name":"duration","type":"number","options":[{"label":"5s","value":5}]}]`,
		`{"duration":5}`, "old-model-a"); err != nil {
		t.Fatalf("param_groups 不可写: %v", err)
	}

	got, err := scanModelWithProvider(db.QueryRow("SELECT "+apiModelColumnsWithProvider+`
		FROM api_models m JOIN api_providers p ON m.provider_id = p.id WHERE m.model = ?`, "old-model-a"))
	if err != nil {
		t.Fatalf("扫描失败: %v", err)
	}
	if got.ParamGroups == "" || got.DefaultParams == "" {
		t.Fatalf("参数组新列未生效: param_groups=%q default_params=%q", got.ParamGroups, got.DefaultParams)
	}
	// 渠道侧扩展列：auth_type 应回填默认值 api_key
	if got.AuthType != "api_key" {
		t.Fatalf("provider.auth_type 默认值 = %q，期望 api_key", got.AuthType)
	}
	if got.Concurrency != 1 {
		t.Fatalf("provider.concurrency 默认值 = %d，期望 1", got.Concurrency)
	}
}
