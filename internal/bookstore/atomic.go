package bookstore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// FileMeta 是落盘后回填索引的元数据。
type FileMeta struct {
	MTime     int64  `json:"mtime"`      // Unix 秒
	Size      int64  `json:"size"`       // 字节数
	Hash      string `json:"hash"`       // sha256(归一化正文)
	WordCount int    `json:"word_count"` // 字符数
}

// NormalizeText 统一文本形态：去 BOM、CRLF/CR 归一为 LF。
//
// 写入前必过；读取时也算一次 hash，保证"同样的内容得到同样的 hash"，
// 否则 Windows 上外部编辑器把 LF 换成 CRLF 会被误判成内容冲突。
func NormalizeText(s string) string {
	s = strings.TrimPrefix(s, "\ufeff")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// HashText 返回归一化正文的 sha256（hex）。
func HashText(s string) string {
	sum := sha256.Sum256([]byte(NormalizeText(s)))
	return hex.EncodeToString(sum[:])
}

// WordCount 按字符数统计，与数据库里 LENGTH(content) 的口径一致。
//
// 不用"词"：中文没有空格分词，按词算会和用户看到的字数差一个量级。
func WordCount(s string) int {
	return utf8.RuneCountInString(NormalizeText(s))
}

// ReadTextFile 读取文件并返回归一化正文与元数据。
func ReadTextFile(path string) (string, FileMeta, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", FileMeta{}, err
	}
	text := NormalizeText(string(raw))
	st, err := os.Stat(path)
	if err != nil {
		return "", FileMeta{}, err
	}
	return text, FileMeta{
		MTime:     st.ModTime().Unix(),
		Size:      st.Size(),
		Hash:      HashText(text),
		WordCount: WordCount(text),
	}, nil
}

// WriteFileAtomic 以"临时文件 + rename"的方式写文件。
//
// 为什么不用 os.WriteFile：
//   - 直接覆盖时进程崩溃/断电会留下半截文件，而这是用户的正文；
//   - rename 在同目录内是原子的，读者只会看到旧版本或新版本。
//
// 失败重试：Windows 上杀软/索引服务会短暂占用目标文件，
// 50/150/400ms 三次退避能覆盖绝大多数瞬时占用。
func WriteFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".xm-*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() {
		tmp.Close()
		os.Remove(tmpName)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("写入失败: %w", err)
	}
	// fsync：保证 rename 之前数据真的落盘，而不是只进了页缓存。
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("刷盘失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	delays := []time.Duration{50 * time.Millisecond, 150 * time.Millisecond, 400 * time.Millisecond}
	var lastErr error
	for attempt := 0; attempt <= len(delays); attempt++ {
		if err := os.Rename(tmpName, path); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < len(delays) {
			time.Sleep(delays[attempt])
		}
	}
	os.Remove(tmpName)
	return fmt.Errorf("替换文件失败（可能被其他程序占用）: %w", lastErr)
}

// WriteTextFile 归一化后原子写入，并返回元数据。
func WriteTextFile(path, text string) (FileMeta, error) {
	normalized := NormalizeText(text)
	if err := WriteFileAtomic(path, []byte(normalized)); err != nil {
		return FileMeta{}, err
	}
	st, err := os.Stat(path)
	if err != nil {
		return FileMeta{}, err
	}
	return FileMeta{
		MTime:     st.ModTime().Unix(),
		Size:      st.Size(),
		Hash:      HashText(normalized),
		WordCount: WordCount(normalized),
	}, nil
}
