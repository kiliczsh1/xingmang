package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 工作流「文件输入」节点用的文本读取接口。
//
// 安全口径（可在需要时收紧）：
// - 只读文本，不写不删；默认单文件上限 5MB（可用 max_bytes 调整，硬上限 32MB）
// - 拒绝目录与空路径；不跟随软链接目标做额外限制（本机单用户桌面应用场景）
// - 编码：默认先试 UTF-8，失败再按 GBK 解码（中文老文件常见）
const (
	fileReadDefaultMax = 5 << 20
	fileReadHardMax    = 32 << 20
)

// ReadTextFile 读取本地文本文件：POST /api/files/read-text
//
// 请求：{ path, encoding?, max_bytes? }
// 响应：{ path, size, encoding, truncated, content }
func (h *Handler) ReadTextFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path     string `json:"path"`
		Encoding string `json:"encoding"`
		MaxBytes int    `json:"max_bytes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	path := strings.TrimSpace(req.Path)
	if path == "" {
		respondFail(w, 400, "path 不能为空")
		return
	}
	maxBytes := req.MaxBytes
	if maxBytes <= 0 {
		maxBytes = fileReadDefaultMax
	}
	if maxBytes > fileReadHardMax {
		maxBytes = fileReadHardMax
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			respondFail(w, 404, "文件不存在: "+path)
			return
		}
		respondFail(w, 500, "无法访问文件: "+err.Error())
		return
	}
	if info.IsDir() {
		respondFail(w, 400, "这是一个目录，不是文件: "+path)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		respondFail(w, 500, "打开文件失败: "+err.Error())
		return
	}
	defer f.Close()

	raw := make([]byte, maxBytes)
	n, readErr := f.Read(raw)
	if readErr != nil && readErr.Error() != "EOF" && n == 0 {
		respondFail(w, 500, "读取文件失败: "+readErr.Error())
		return
	}
	raw = raw[:n]
	truncated := info.Size() > int64(n)

	encoding := strings.ToLower(strings.TrimSpace(req.Encoding))
	if encoding == "" || encoding == "auto" {
		if utf8.Valid(raw) {
			encoding = "utf-8"
		} else {
			encoding = "gbk"
		}
	}

	content := string(raw)
	switch encoding {
	case "utf-8", "utf8":
		encoding = "utf-8"
		content = strings.ToValidUTF8(string(raw), "")
	case "gbk", "gb18030":
		decoded, _, derr := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), raw)
		if derr != nil {
			respondFail(w, 400, fmt.Sprintf("按 GBK 解码失败: %v", derr))
			return
		}
		encoding = "gbk"
		content = string(decoded)
	default:
		respondFail(w, 400, "不支持的编码: "+encoding+"（可选 auto / utf-8 / gbk）")
		return
	}

	respondOK(w, map[string]interface{}{
		"path":      path,
		"size":      info.Size(),
		"encoding":  encoding,
		"truncated": truncated,
		"content":   content,
	})
}
