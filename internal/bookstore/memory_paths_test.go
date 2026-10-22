package bookstore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func TestMemoryShortCode_StableAndDistinct(t *testing.T) {
	a := MemoryShortCode(MemoryKindChat, 12)
	b := MemoryShortCode(MemoryKindChat, 12)
	if a != b {
		t.Fatalf("同一记录两次取短码应一致: %s / %s", a, b)
	}
	if len(a) != 6 {
		t.Fatalf("短码应为 6 位: %s", a)
	}
	if MemoryShortCode(MemoryKindChat, 13) == a {
		t.Fatal("不同记录不应撞短码")
	}
	if MemoryShortCode(MemoryKindGraph, 12) == a {
		t.Fatal("不同类型不应撞短码")
	}
}

func TestMemoryMainName_RoundTrip(t *testing.T) {
	at := time.Date(2026, 10, 18, 22, 40, 3, 0, time.Local)
	name := MemoryMainName(at, "列车开场 / 怎么写", MemoryKindChat, 12)
	if filepath.Ext(name) != "" {
		t.Fatalf("主名不应带扩展名: %s", name)
	}
	if !isAllDigits(name[:8]) || name[8] != '-' {
		t.Fatalf("主名应以 YYYYMMDD- 开头: %s", name)
	}
	// 非法字符必须被清洗掉，否则落盘直接失败
	if filepath.Base(name) != name {
		t.Fatalf("主名含路径分隔符: %s", name)
	}

	created, title, code, ok := ParseMemoryMainName(name + ".md")
	if !ok {
		t.Fatalf("应能解析自己生成的主名: %s", name)
	}
	if created != "20261018-2240" {
		t.Fatalf("创建时间解析错: %s", created)
	}
	if title != "列车开场 _ 怎么写" {
		t.Fatalf("标题解析错（应保留清洗后的形态）: %q", title)
	}
	if code != MemoryShortCode(MemoryKindChat, 12) {
		t.Fatalf("短码解析错: %s", code)
	}
}

func TestMemoryMainName_TitleWithDashes(t *testing.T) {
	at := time.Date(2026, 10, 18, 22, 40, 3, 0, time.Local)
	name := MemoryMainName(at, "卷一-开场-冲突", MemoryKindAnalysis, 5)
	_, title, code, ok := ParseMemoryMainName(name + ".md")
	if !ok {
		t.Fatalf("标题含连字符也应能解析: %s", name)
	}
	if title != "卷一-开场-冲突" {
		t.Fatalf("标题里的连字符应原样保留: %q", title)
	}
	if code != MemoryShortCode(MemoryKindAnalysis, 5) {
		t.Fatalf("短码错: %s", code)
	}
}

func TestMemoryRelDir(t *testing.T) {
	at := time.Date(2026, 10, 18, 22, 40, 3, 0, time.Local)
	cases := map[string]string{
		MemoryKindChat:       "memory/chats",
		MemoryKindAgent:      "memory/agent",
		MemoryKindGraph:      "memory/graph",
		MemoryKindAnalysis:   "memory/analysis",
		MemoryKindGeneration: "memory/history/2026-10",
	}
	for kind, want := range cases {
		got, err := MemoryRelDir(kind, at)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if got != want {
			t.Fatalf("%s 目录应为 %s，实得 %s", kind, want, got)
		}
	}
	if _, err := MemoryRelDir("unknown", at); err == nil {
		t.Fatal("未知类型应报错")
	}
}

func TestMemoryKindOfRelPath(t *testing.T) {
	if kind, ok := MemoryKindOfRelPath("memory/history/2026-10/x.md"); !ok || kind != MemoryKindGeneration {
		t.Fatalf("history 路径应识别为 generation: %s %v", kind, ok)
	}
	if kind, ok := MemoryKindOfRelPath("memory/chats/x.md"); !ok || kind != MemoryKindChat {
		t.Fatalf("chats 路径应识别为 chat: %s %v", kind, ok)
	}
	if _, ok := MemoryKindOfRelPath("001 第一章.md"); ok {
		t.Fatal("章节路径不应被当成记忆")
	}
	if !IsMemoryRelPath("memory") {
		t.Fatal("memory 自身也算记忆路径")
	}
	if IsMemoryRelPath("memoryx/a.md") {
		t.Fatal("memoryx 不是记忆目录")
	}
}

func TestCreateBook_MemoryDirs(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "记忆目录测试", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	for _, sub := range MemorySubdirs {
		dir := filepath.Join(root, "记忆目录测试", "memory", sub)
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			t.Fatalf("缺少目录 %s: %v", dir, err)
		}
	}
	if !IsSystemDir("memory") {
		t.Fatal("memory 必须进系统目录白名单，否则会被当成卷")
	}
}
