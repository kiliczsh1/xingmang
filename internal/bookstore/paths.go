// Package bookstore 负责作品在磁盘上的路径规则、原子读写与目录扫描。
//
// 设计约束见 docs/books-folder-storage.md：
//   - 正文是唯一真相，数据库只做索引；
//   - 书目录 = 书名，章节标题 = 文件名，卷 = 子目录；
//   - 所有磁盘写入必须原子（tmp + rename），失败不能静默。
package bookstore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// maxNameRunes 限制单个文件/目录名的长度。
// 加上序号前缀与书目录后仍要留在 Windows 260 字符以内，所以只给 60。
const maxNameRunes = 60

var (
	illegalChars = regexp.MustCompile(`[\\/:*?"<>|]`)
	// 形如 "001 第一章"、"12 序章"；允许 1–4 位数字，避免超过 999 章后乱序。
	orderedName = regexp.MustCompile(`^(\d{1,4})\s+(.+)$`)
)

// SystemDirs 是书根下不参与"卷"判定的系统目录。
// 加新类型（比如以后的 outlines 变体）只需进这里，不会和用户的卷目录冲突。
var SystemDirs = []string{
	"covers", "gallery", "characters", "world", "outlines", "ideas", "memos", "assets",
	"memory",
}

var systemDirSet = func() map[string]bool {
	m := make(map[string]bool, len(SystemDirs))
	for _, d := range SystemDirs {
		m[d] = true
	}
	return m
}()

// reservedNames 是 Windows 保留设备名，作为文件名会直接失败。
var reservedNames = func() map[string]bool {
	m := map[string]bool{"CON": true, "PRN": true, "AUX": true, "NUL": true}
	for i := 1; i <= 9; i++ {
		m[fmt.Sprintf("COM%d", i)] = true
		m[fmt.Sprintf("LPT%d", i)] = true
	}
	return m
}()

// IsSystemDir 判断书根下的目录是否是系统目录（非卷）。
func IsSystemDir(name string) bool {
	return systemDirSet[strings.ToLower(name)]
}

// SanitizeName 把任意标题清洗成可用的文件名/目录名。
//
// 不抛错：调用方拿到的一定是能落盘的名字。若发生替换，调用方应从返回值是否
// 与输入相等来判断要不要提示用户"已重命名"。
func SanitizeName(raw string) string {
	s := strings.TrimSpace(raw)
	s = illegalChars.ReplaceAllString(s, "_")
	// 控制字符（含制表符、换行）一律替换，否则 Windows 直接拒绝
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '_'
		}
		return r
	}, s)
	s = strings.TrimRight(s, " .")
	s = strings.TrimSpace(s)
	if s == "" {
		return "未命名"
	}
	if utf8.RuneCountInString(s) > maxNameRunes {
		s = string([]rune(s)[:maxNameRunes])
		s = strings.TrimRight(s, " .")
	}
	if reservedNames[strings.ToUpper(s)] {
		s = "_" + s
	}
	return s
}

// FormatOrder 把序号格式化成固定宽度前缀：1 → "001"，1000 → "1000"。
func FormatOrder(n int) string {
	if n <= 0 {
		n = 1
	}
	if n <= 999 {
		return fmt.Sprintf("%03d", n)
	}
	return fmt.Sprintf("%04d", n)
}

// NormalizeExt 只保留 .md / .txt，其余一律按 .md 处理。
func NormalizeExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".txt":
		return ".txt"
	default:
		return ".md"
	}
}

// ChapterFileName 生成章节文件名：`NNN 标题.md`。
func ChapterFileName(order int, title, ext string) string {
	return fmt.Sprintf("%s %s%s", FormatOrder(order), SanitizeName(title), NormalizeExt(ext))
}

// VolumeDirName 生成卷目录名：`NNN 卷名`。
func VolumeDirName(order int, title string) string {
	return fmt.Sprintf("%s %s", FormatOrder(order), SanitizeName(title))
}

// ParseOrderedName 从章节/卷名里解析序号与标题。
//
// 输入可以是 "001 第一章.md" 或 "第一卷"（后者没有序号，ok=false）。
// 显示标题不含扩展名与序号；改标题时序号保持不变。
func ParseOrderedName(name string) (order int, title string, ok bool) {
	base := name
	if i := strings.LastIndex(base, "."); i > 0 {
		base = base[:i]
	}
	m := orderedName.FindStringSubmatch(base)
	if m == nil {
		return 0, base, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0, base, false
	}
	return n, strings.TrimSpace(m[2]), true
}

// DisplayTitle 返回给人看的标题：去掉序号前缀与扩展名。
func DisplayTitle(fileName string) string {
	_, title, _ := ParseOrderedName(fileName)
	if title == "" {
		return fileName
	}
	return title
}

// JoinWithin 把相对路径安全地拼到书目录下，拒绝任何逃逸。
//
// 所有 file_path 都必须经过这里；禁止调用方直接 filepath.Join(root, rel)。
func JoinWithin(root, rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("路径为空")
	}
	clean := filepath.Clean(filepath.FromSlash(rel))
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("不允许绝对路径: %s", rel)
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("路径越界: %s", rel)
	}
	full := filepath.Join(root, clean)
	back, err := filepath.Rel(root, full)
	if err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("路径越界: %s", rel)
	}
	return full, nil
}

// UniqueName 在 dir 下为 name 找一个未被占用的文件名（忽略大小写）。
//
// 冲突时追加 " (2)"、" (3)"…，扩展名保持在最后。
// 已存在同名文件时返回“已改名”的完整名字，调用方应把它写进索引。
func UniqueName(dir, name string) string {
	if !exists(filepath.Join(dir, name)) {
		return name
	}
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for i := 2; i < 1000; i++ {
		cand := fmt.Sprintf("%s (%d)%s", stem, i, ext)
		if !exists(filepath.Join(dir, cand)) {
			return cand
		}
	}
	return name
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// Entry 是扫描出的一个目录项，供排序与对账使用。
type Entry struct {
	Name  string // 文件名（不含目录）
	Order int    // 序号；无序号为 0
	IsDir bool
}

// SortEntries 按"有编号的在前、编号升序、再按名称自然序"排列。
//
// 人物/世界观目录没有编号，会全部落到名称比较这一档。
func SortEntries(list []Entry) {
	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return naturalLess(a.Name, b.Name)
	})
}

// naturalLess 做自然序比较：第2章 < 第10章。
func naturalLess(a, b string) bool {
	an, bn := digitsPrefix(a), digitsPrefix(b)
	if an != bn {
		return an < bn
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

func digitsPrefix(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
		if n > 1<<30 {
			return 1 << 30
		}
	}
	return n
}

// ---------------- 外部改动拦截 ----------------

// ExternalEditReason 是"文件被外部改过、程序拒绝覆盖"时写进索引的原因。
//
// 这条规则是"文件为真值"的必然推论：文件赢，程序不覆盖。
// 但必须留下痕迹，否则用户手改完发现没生效、又看不出为什么。
const ExternalEditReason = "文件被程序之外改过，本次未覆盖（改回原样或删除该文件即可让程序重建）"

// FileEditedExternally 判断磁盘文件是否在程序之外被改过。
//
// recorded 是索引里记的上次由程序写入的哈希；为空表示没有基线（老数据），
// 此时不判断、按可写处理 —— 否则会把自己刚建的正常文件误判成"外部改动"。
func FileEditedExternally(abs, recorded string) (bool, string, error) {
	if strings.TrimSpace(abs) == "" || strings.TrimSpace(recorded) == "" {
		return false, "", nil
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", nil
		}
		return false, "", err
	}
	cur := HashText(string(raw))
	return cur != recorded, cur, nil
}
