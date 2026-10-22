package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 正文插图本地落盘。
//
// 口径与封面一致：前端把文件读成 base64 传上来 → 后端写入本地目录 →
// 返回站内 URL 给编辑器 <img src>。图片不进数据库（DB 里只有正文 HTML 里的一行地址）。
//
// 目录随进程工作目录（与 xingmang.db 同级），备份/迁移数据时一起带走。
const (
	editorImageDir       = "data/editor-images"
	editorImageURLPrefix = "/api/editor-images/"
	editorImageMaxBytes  = 10 << 20
)

var imageExtByMime = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
	"image/bmp":  ".bmp",
	"image/avif": ".avif",
}

// saveLocalImage 通用落盘：解码 base64 → 写 <dir>/<时间戳>_<随机>.<ext> → 返回可访问 URL。
// label 只用于错误文案（「封面」「图片」），urlPrefix 决定对外地址。
func saveLocalImage(dir, urlPrefix, label, name, mime, dataBase64 string, maxBytes int) (string, error) {
	raw := strings.TrimSpace(dataBase64)
	// 容错：前端可能直接传完整 data URL
	if i := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && i > 0 {
		header := raw[:i]
		if j := strings.Index(header, ":"); j >= 0 {
			km := header[j+1:]
			if k := strings.Index(km, ";"); k >= 0 {
				km = km[:k]
			}
			if km != "" {
				mime = km
			}
		}
		raw = raw[i+1:]
	}
	if raw == "" {
		return "", fmt.Errorf("%s数据为空", label)
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		// 兼容 URL-safe base64
		decoded, err = base64.RawURLEncoding.DecodeString(raw)
		if err != nil {
			return "", fmt.Errorf("%s base64 解码失败: %w", label, err)
		}
	}
	if len(decoded) == 0 {
		return "", fmt.Errorf("%s数据为空", label)
	}
	if len(decoded) > maxBytes {
		return "", fmt.Errorf("%s过大（%d 字节，上限 %d）", label, len(decoded), maxBytes)
	}

	ext := imageExtByMime[strings.ToLower(strings.TrimSpace(mime))]
	if ext == "" {
		ext = filepath.Ext(name)
	}
	return saveLocalImageBytes(dir, urlPrefix, label, ext, decoded, maxBytes)
}

// normalizeImageExt 归一图片扩展名：认得的一律小写，.jpeg 归 .jpg，其余按 .png 存。
func normalizeImageExt(ext string) string {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".png", ".gif", ".webp", ".bmp", ".avif":
		return strings.ToLower(strings.TrimSpace(ext))
	case ".jpg", ".jpeg":
		return ".jpg"
	default:
		// 看不出类型时按 png 存，浏览器仍能按内容嗅探
		return ".png"
	}
}

// saveLocalImageBytes 把已解码的字节写入 dir 并返回可访问 URL。
// 供「下载远端图片后落盘」复用（无需再走 base64 往返）。
func saveLocalImageBytes(dir, urlPrefix, label, ext string, data []byte, maxBytes int) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("%s数据为空", label)
	}
	if len(data) > maxBytes {
		return "", fmt.Errorf("%s过大（%d 字节，上限 %d）", label, len(data), maxBytes)
	}

	ext = normalizeImageExt(ext)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建%s目录失败: %w", label, err)
	}

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成%s文件名失败: %w", label, err)
	}
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hex.EncodeToString(buf), ext)

	if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
		return "", fmt.Errorf("写入%s失败: %w", label, err)
	}
	return urlPrefix + filename, nil
}

// serveLocalFile 只允许读取 dir 下的单层文件名，防目录穿越。
// localMediaPath 把站内静态图片 URL 还原成本地绝对路径。
//
// 只接受 `<urlPrefix><一个文件名>` 这一种形态：多一层目录、`..`、反斜杠一律拒绝。
// 复用它的地方（比如"把 AI 生成的封面设为当前封面"）都需要读原文件，
// 所以把路径校验收在一处，避免各处自己拼路径时漏掉越界检查。
func localMediaPath(dir, urlPrefix, url string) (string, string, error) {
	if !strings.HasPrefix(url, urlPrefix) {
		return "", "", fmt.Errorf("不是本站图片地址: %s", url)
	}
	name := strings.TrimPrefix(url, urlPrefix)
	// 去掉可能的查询串（缓存戳），但不能因此放过路径穿越
	if i := strings.IndexAny(name, "?#"); i >= 0 {
		name = name[:i]
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", "", fmt.Errorf("非法的图片地址")
	}
	full := filepath.Join(dir, name)
	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		return "", "", fmt.Errorf("图片已不存在")
	}
	return full, name, nil
}

func serveLocalFile(w http.ResponseWriter, r *http.Request, dir, urlPrefix string) {
	name := filepath.Base(strings.TrimPrefix(r.URL.Path, urlPrefix))
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		http.NotFound(w, r)
		return
	}
	full := filepath.Join(dir, name)
	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, full)
}

// UploadImage 正文插图上传：POST /api/images
// 请求：{ name, mime?, data_base64 }  → 响应：{ url }
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		Mime       string `json:"mime"`
		MimeType   string `json:"mime_type"`
		DataBase64 string `json:"data_base64"`
		Data       string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	data := req.DataBase64
	if strings.TrimSpace(data) == "" {
		data = req.Data
	}
	mime := req.Mime
	if mime == "" {
		mime = req.MimeType
	}

	url, err := saveLocalImage(editorImageDir, editorImageURLPrefix, "图片", req.Name, mime, data, editorImageMaxBytes)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"url": url, "name": req.Name})
}

// ServeEditorImage 正文插图静态访问：GET /api/editor-images/{file}
func (h *Handler) ServeEditorImage(w http.ResponseWriter, r *http.Request) {
	serveLocalFile(w, r, editorImageDir, editorImageURLPrefix)
}
