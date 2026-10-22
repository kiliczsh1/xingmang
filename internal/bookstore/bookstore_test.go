package bookstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"第一章 落雪", "第一章 落雪"},
		{`a/b\c:d*e?f"g<h>i|j`, "a_b_c_d_e_f_g_h_i_j"},
		{"  前后空格  ", "前后空格"},
		{"结尾点...", "结尾点"},
		{"结尾空格   ", "结尾空格"},
		{"", "未命名"},
		{"   ", "未命名"},
		{"CON", "_CON"},
		{"com1", "_com1"},
		{"带\t制表符", "带_制表符"},
	}
	for _, c := range cases {
		if got := SanitizeName(c.in); got != c.want {
			t.Fatalf("SanitizeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSanitizeName_Length(t *testing.T) {
	long := strings.Repeat("字", 80)
	got := SanitizeName(long)
	if n := len([]rune(got)); n != maxNameRunes {
		t.Fatalf("超长标题应截到 %d 字符，实得 %d", maxNameRunes, n)
	}
}

func TestChapterFileName(t *testing.T) {
	if got := ChapterFileName(2, "第一章 落雪", ".md"); got != "002 第一章 落雪.md" {
		t.Fatalf("got %q", got)
	}
	if got := ChapterFileName(1000, "很长了", ".txt"); got != "1000 很长了.txt" {
		t.Fatalf("got %q", got)
	}
	if got := ChapterFileName(1, "x", ".exe"); got != "001 x.md" {
		t.Fatalf("非 md/txt 应回落到 .md，got %q", got)
	}
}

func TestParseOrderedName(t *testing.T) {
	order, title, ok := ParseOrderedName("007 第七章 血战.md")
	if !ok || order != 7 || title != "第七章 血战" {
		t.Fatalf("got %d %q %v", order, title, ok)
	}

	order, title, ok = ParseOrderedName("1000 尾声.md")
	if !ok || order != 1000 || title != "尾声" {
		t.Fatalf("got %d %q %v", order, title, ok)
	}

	if _, title, ok = ParseOrderedName("序章.md"); ok || title != "序章" {
		t.Fatalf("无序号应 ok=false，实得 %q %v", title, ok)
	}
}

func TestSortEntries(t *testing.T) {
	list := []Entry{
		{Name: "010 第十章.md", Order: 10},
		{Name: "002 第二章.md", Order: 2},
		{Name: "1000 尾声.md", Order: 1000},
		{Name: "001 第一章.md", Order: 1},
	}
	SortEntries(list)
	want := []string{"001 第一章.md", "002 第二章.md", "010 第十章.md", "1000 尾声.md"}
	for i := range want {
		if list[i].Name != want[i] {
			t.Fatalf("第 %d 项 = %q, want %q", i, list[i].Name, want[i])
		}
	}
}

func TestJoinWithin(t *testing.T) {
	root := filepath.Join("D:", "books", "太荒神诀")
	if _, err := JoinWithin(root, "第一卷/001 楔子.md"); err != nil {
		t.Fatalf("正常路径不应报错: %v", err)
	}
	if _, err := JoinWithin(root, "../secret.md"); err == nil {
		t.Fatal("越界路径必须拒绝")
	}
	if _, err := JoinWithin(root, "a/../../b.md"); err == nil {
		t.Fatal("归一化后越界也必须拒绝")
	}
	if _, err := JoinWithin(root, "D:/other/x.md"); err == nil {
		t.Fatal("绝对路径必须拒绝")
	}
	if _, err := JoinWithin(root, ""); err == nil {
		t.Fatal("空路径必须拒绝")
	}
}

func TestNormalizeText(t *testing.T) {
	in := "\ufeff第一行\r\n第二行\r第三行"
	want := "第一行\n第二行\n第三行"
	if got := NormalizeText(in); got != want {
		t.Fatalf("got %q", got)
	}
}

func TestWordCount(t *testing.T) {
	if got := WordCount("你好，世界\nabc"); got != 9 {
		t.Fatalf("按字符计数应为 9（中文 5 + 换行 1 + 英文 3），实得 %d", got)
	}
}

func TestWriteTextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "第一卷", "001 测试.md")

	meta, err := WriteTextFile(path, "旧内容\r\n第二行")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Hash != HashText("旧内容\n第二行") {
		t.Fatal("hash 应基于归一化正文")
	}

	// 覆盖写
	meta2, err := WriteTextFile(path, "新内容")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "新内容" {
		t.Fatalf("覆盖写失败: %q", raw)
	}
	if meta2.Hash == meta.Hash {
		t.Fatal("内容变了 hash 应变化")
	}

	// 目录下不应残留临时文件
	entries, _ := os.ReadDir(filepath.Dir(path))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".xm-") {
			t.Fatalf("残留临时文件: %s", e.Name())
		}
	}
}

func TestFrontMatter_PreservesUnknownFields(t *testing.T) {
	src := "---\nsummary: 旧摘要\nmood: 紧张\n---\n\n正文第一段\n"
	out := SetFrontMatterField(src, "summary", "新摘要")

	if !strings.Contains(out, "mood: 紧张") {
		t.Fatalf("自定义字段被吃掉了: %q", out)
	}
	if !strings.Contains(out, "summary: 新摘要") {
		t.Fatalf("摘要未更新: %q", out)
	}
	if !strings.Contains(out, "正文第一段") {
		t.Fatalf("正文丢了: %q", out)
	}
}

func TestFrontMatter_EmptyRemovesBlock(t *testing.T) {
	src := "---\nsummary: 只有摘要\n---\n\n正文"
	out := SetFrontMatterField(src, "summary", "")
	if strings.HasPrefix(out, "---") {
		t.Fatalf("字段清空后不应保留围栏: %q", out)
	}
	if BodyOf(out) != "正文" {
		t.Fatalf("正文应保留，实得 %q", BodyOf(out))
	}
}

func TestFrontMatter_UnclosedIsBody(t *testing.T) {
	src := "---\nsummary: 没有闭合\n\n正文"
	fm := ParseFrontMatter(src)
	if fm.Has {
		t.Fatal("未闭合的围栏应视为正文")
	}
	if fm.Body != src {
		t.Fatalf("正文应原样返回")
	}
}

func TestUniqueName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "序章.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := UniqueName(dir, "序章.md"); got != "序章 (2).md" {
		t.Fatalf("got %q", got)
	}
	if got := UniqueName(dir, "新章.md"); got != "新章.md" {
		t.Fatalf("无冲突不应改名，got %q", got)
	}
}
