package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"

	"github.com/xingmang/xingmang-desktop/internal/database"
)

func main() {
	sourcePath := flag.String("source", "", "源数据库文件路径")
	targetPath := flag.String("target", "xingmang.db", "目标数据库文件路径")
	flag.Parse()

	if *sourcePath == "" {
		fmt.Println("用法: migrate --source <源数据库路径> [--target <目标数据库路径>]")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  migrate --source ../xingmang/server/database/xingmang.db")
		os.Exit(1)
	}

	if _, err := os.Stat(*sourcePath); os.IsNotExist(err) {
		log.Fatalf("源数据库文件不存在: %s", *sourcePath)
	}

	// Open source database (read-only)
	sourceDB, err := sql.Open("sqlite", *sourcePath+"?mode=ro")
	if err != nil {
		log.Fatalf("无法打开源数据库: %v", err)
	}
	defer sourceDB.Close()

	// Initialize target database
	targetDB, err := database.Init(*targetPath)
	if err != nil {
		log.Fatalf("无法初始化目标数据库: %v", err)
	}
	defer database.Close()

	// Tables to migrate (in dependency order)
	tables := []string{
		"books",
		"volumes",
		"chapters",
		"prompts",
		"memos",
		"experience_shares",
		"api_providers",
		"api_models",
		"conversations",
		"conversation_messages",
		"generators",
		"characters",
		"usage_stats",
		"graph_versions",
		"graph_entities",
		"graph_relations",
		"entry_categories",
		"entries",
	}

	totalRows := 0
	for _, table := range tables {
		count, err := migrateTable(sourceDB, targetDB, table)
		if err != nil {
			log.Printf("⚠ 迁移表 %s 失败: %v", table, err)
			continue
		}
		if count > 0 {
			fmt.Printf("✓ %s: %d 行\n", table, count)
		} else {
			fmt.Printf("  %s: 空 (跳过)\n", table)
		}
		totalRows += count
	}

	fmt.Printf("\n迁移完成! 共迁移 %d 行数据\n", totalRows)
}

func migrateTable(source, target *sql.DB, table string) (int, error) {
	// Count rows in source
	var count int
	err := source.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询源表失败: %w", err)
	}
	if count == 0 {
		return 0, nil
	}

	// Get column names from source
	rows, err := source.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table))
	if err != nil {
		return 0, fmt.Errorf("查询源表结构失败: %w", err)
	}
	columns, err := rows.Columns()
	rows.Close()
	if err != nil {
		return 0, fmt.Errorf("获取列名失败: %w", err)
	}

	// Build INSERT statement
	colCount := len(columns)
	placeholders := ""
	for i := 0; i < colCount; i++ {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}
	colList := ""
	for i, col := range columns {
		if i > 0 {
			colList += ", "
		}
		colList += col
	}
	insertSQL := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (%s)", table, colList, placeholders)

	// Read all rows from source and insert into target
	sourceRows, err := source.Query(fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		return 0, fmt.Errorf("读取源数据失败: %w", err)
	}
	defer sourceRows.Close()

	stmt, err := target.Prepare(insertSQL)
	if err != nil {
		return 0, fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	migrated := 0
	values := make([]interface{}, colCount)
	valuePtrs := make([]interface{}, colCount)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	tx, err := target.Begin()
	if err != nil {
		return 0, fmt.Errorf("开始事务失败: %w", err)
	}
	txStmt := tx.Stmt(stmt)

	for sourceRows.Next() {
		if err := sourceRows.Scan(valuePtrs...); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("读取行失败: %w", err)
		}
		// Convert []byte to string for TEXT columns
		args := make([]interface{}, colCount)
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				args[i] = string(b)
			} else {
				args[i] = v
			}
		}
		if _, err := txStmt.Exec(args...); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("插入行失败: %w", err)
		}
		migrated++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return migrated, nil
}
