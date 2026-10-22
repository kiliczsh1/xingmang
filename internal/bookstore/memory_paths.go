package bookstore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 记忆文件的路径规则（docs/books-memory-storage.md §4）。
//
//	<书>/memory/chats/20261018-2240-列车开场-a1b2c3.md
//	<书>/memory/agent/…
//	<书>/memory/graph/3-初版.md
//	<书>/memory/analysis/20261018-2245-拆书-77c1ab.md
//	<书>/memory/history/2026-10/20261018-2310-续写-5b2e90.md
//
// 与章节一致：文件是真相，DB 只存相对书目录的路径。
const memoryRootDir = "memory"

// 记忆类型（既用于目录分派，也用于 flusher 注册与 DB 的 kind）。
const (
	MemoryKindChat       = "chat"
	MemoryKindAgent      = "agent"
	MemoryKindGraph      = "graph"
	MemoryKindAnalysis   = "analysis"
	MemoryKindGeneration = "generation"
)

// memorySubdirs 是类型 → memory/ 下的子目录名。
var memorySubdirs = map[string]string{
	MemoryKindChat:       "chats",
	MemoryKindAgent:      "agent",
	MemoryKindGraph:      "graph",
	MemoryKindAnalysis:   "analysis",
	MemoryKindGeneration: "history",
}

// MemorySubdirs 是 memory/ 下需要预建的全部子目录。
var MemorySubdirs = []string{"chats", "agent", "graph", "analysis", "history"}

// MemoryKindDir 返回类型对应的子目录名（相对 memory/）。
func MemoryKindDir(kind string) (string, error) {
	if sub, ok := memorySubdirs[strings.ToLower(strings.TrimSpace(kind))]; ok {
		return sub, nil
	}
	return "", unknownKindErr(kind)
}

// MemoryRelDir 返回类型对应的相对书目录的目录（history 按月再分一层）。
func MemoryRelDir(kind string, at time.Time) (string, error) {
	sub, err := MemoryKindDir(kind)
	if err != nil {
		return "", err
	}
	if strings.ToLower(kind) == MemoryKindGeneration {
		if at.IsZero() {
			at = time.Now()
		}
		return filepath.ToSlash(filepath.Join(memoryRootDir, sub, at.Format("2006-01"))), nil
	}
	return memoryRootDir + "/" + sub, nil
}

// MemoryShortCode 是文件名里的稳定短码：类型 + 主键的 sha256 前 6 位。
//
// 之所以要短码：文件主名带创建时间与标题，同名不同记录会撞；改标题又不重命名文件，
// 于是需要一个与内容无关、与记录一一对应的后缀来保证唯一。
func MemoryShortCode(kind string, id int) string {
	sum := sha256.Sum256([]byte(strings.ToLower(kind) + "#" + strconv.Itoa(id)))
	return hex.EncodeToString(sum[:])[:6]
}

// MemoryMainName 生成文件主名（不含扩展名）：20261018-2240-列车开场-a1b2c3。
func MemoryMainName(at time.Time, title, kind string, id int) string {
	if at.IsZero() {
		at = time.Now()
	}
	clean := SanitizeName(TruncateRunes(strings.TrimSpace(title), memMaxTitleRunes))
	return fmt.Sprintf("%s-%s-%s", at.Format(memFileNameTimeForm), clean, MemoryShortCode(kind, id))
}

// ParseMemoryMainName 从文件名解析出创建时间、标题与短码。
//
// 标题里可能含 "-"，所以按"前两段是日期与时间、最后一段是短码"来切。
func ParseMemoryMainName(fileName string) (created string, title string, code string, ok bool) {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	parts := strings.Split(base, "-")
	if len(parts) < 4 {
		return "", "", "", false
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || !isAllDigits(parts[0]) || !isAllDigits(parts[1]) {
		return "", "", "", false
	}
	code = parts[len(parts)-1]
	title = strings.Join(parts[2:len(parts)-1], "-")
	return parts[0] + "-" + parts[1], title, code, true
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsMemoryRelPath 判断相对书目录的路径是否落在 memory/ 下。
func IsMemoryRelPath(rel string) bool {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	return rel == memoryRootDir || strings.HasPrefix(rel, memoryRootDir+"/")
}

// MemoryKindOfRelPath 从相对路径推断记忆类型；不是记忆路径时 ok=false。
func MemoryKindOfRelPath(rel string) (string, bool) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if !IsMemoryRelPath(rel) {
		return "", false
	}
	rest := strings.TrimPrefix(rel, memoryRootDir+"/")
	sub := strings.SplitN(rest, "/", 2)[0]
	for kind, dir := range memorySubdirs {
		if sub == dir {
			return kind, true
		}
	}
	return "", false
}

// MemoryAbsDir 返回（并创建）某类型记忆的绝对目录。
//
// 走 JoinWithin，与其它落盘路径同一套越界防护。
func (s *Service) MemoryAbsDir(bookID int, kind string, at time.Time) (string, error) {
	rel, err := MemoryRelDir(kind, at)
	if err != nil {
		return "", err
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", err
	}
	abs, err := JoinWithin(bookDir, rel)
	if err != nil {
		return "", err
	}
	if err := ensureDir(abs); err != nil {
		return "", err
	}
	return abs, nil
}

// MemoryRelPath 返回相对书目录的路径，用于写进 DB 的 file_path。
func MemoryRelPath(kind string, at time.Time, title string, id int) (string, error) {
	relDir, err := MemoryRelDir(kind, at)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(relDir, MemoryMainName(at, title, kind, id)+".md")), nil
}

// MemoryGlobalDir 返回库根下的全局记忆目录（未绑定书的 Agent 会话、无归属的历史）。
func (s *Service) MemoryGlobalDir(kind string, at time.Time) (string, error) {
	root, err := s.Root()
	if err != nil {
		return "", err
	}
	rel, err := MemoryRelDir(kind, at)
	if err != nil {
		return "", err
	}
	// memory/agent → _memory/agent
	rel = strings.Replace(rel, memoryRootDir, "_memory", 1)
	abs, err := JoinWithin(root, rel)
	if err != nil {
		return "", err
	}
	if err := ensureDir(abs); err != nil {
		return "", err
	}
	return abs, nil
}

// MemoryMainNameOf 是字符串主键版的文件主名（Agent 会话 id 是 32 位十六进制）。
func MemoryMainNameOf(at time.Time, title, kind, id string) string {
	if at.IsZero() {
		at = time.Now()
	}
	clean := SanitizeName(TruncateRunes(strings.TrimSpace(title), memMaxTitleRunes))
	return fmt.Sprintf("%s-%s-%s", at.Format(memFileNameTimeForm), clean, MemoryShortCodeOf(kind, id))
}

// MemoryRelPathOf 是字符串主键版的相对路径。
func MemoryRelPathOf(kind string, at time.Time, title, id string) (string, error) {
	relDir, err := MemoryRelDir(kind, at)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(relDir, MemoryMainNameOf(at, title, kind, id)+".md")), nil
}

// ReconcileMemoryFileName 让文件名里的标题段与当前标题一致，时间前缀与短码保持不变。
//
// 与章节同一规则：改标题就改文件名。短码是稳定标识，所以改名不会让文件"失去身份"
// （凭短码仍能定位到同一条记录），文件名则始终等于标题。
//
// 只在文件名确实符合 `<时间>-<标题>-<短码>.<ext>` 时才动：用户自己重命名过就尊重他。
func ReconcileMemoryFileName(rel, title string) (string, bool) {
	dir := filepath.Dir(filepath.FromSlash(rel))
	base := filepath.Base(filepath.FromSlash(rel))
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	// 时间戳本身含一个短横线（`20261019-0102`），所以必须按前两段还原，
	// 不能按 "-" 切完就取 parts[0]。标题里也可能有短横线，因此短码取最后一段、标题取中间。
	parts := strings.Split(stem, "-")
	if len(parts) < 4 {
		return rel, false
	}
	stamp, code := parts[0]+"-"+parts[1], parts[len(parts)-1]
	if _, err := time.ParseInLocation(memFileNameTimeForm, stamp, time.Local); err != nil {
		return rel, false // 前两段不是时间，说明这不是程序起的名字
	}
	if len([]rune(code)) != 6 || len(parts[2:len(parts)-1]) == 0 {
		return rel, false
	}

	clean := SanitizeName(TruncateRunes(strings.TrimSpace(title), memMaxTitleRunes))
	want := stamp + "-" + clean + "-" + code + ext
	if want == base {
		return rel, false
	}
	if dir == "." {
		return filepath.ToSlash(want), true
	}
	return filepath.ToSlash(filepath.Join(dir, want)), true
}

// RenameMemoryFile 把文件按新的相对路径改名；目标已存在则一动不动。
//
// 返回新的绝对/相对路径，以及是否真的改了。base 由当前 abs 去掉 rel 前缀推得，
// 所以书内文件与库根 _memory/ 都适用，不需要知道记录属于谁。
func RenameMemoryFile(abs, rel, wantRel string) (string, string, bool) {
	if wantRel == "" || filepath.ToSlash(wantRel) == filepath.ToSlash(rel) {
		return abs, rel, false
	}
	base := strings.TrimSuffix(abs, filepath.FromSlash(rel))
	newAbs, err := JoinWithin(base, wantRel)
	if err != nil {
		return abs, rel, false
	}
	// 目标已存在：宁可文件名旧一点，也不能覆盖另一个文件
	if _, err := os.Stat(newAbs); err == nil {
		return abs, rel, false
	}
	if err := renameWithRetry(abs, newAbs); err != nil {
		return abs, rel, false
	}
	return newAbs, wantRel, true
}
