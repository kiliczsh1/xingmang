package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 经验分享封面图本地落盘。
//
// 背景：前端「上传图片」走的是 { name, size, data_base64 } 的 cover_image 字段，
// 但后端此前完全没有处理它 —— 前端传了、后端忽略，只有手填的 cover_url 字符串会入库。
// 这就造成一条脏数据：cover_url 存着本地磁盘路径（D:\Downloads\xxx.png），
// 浏览器访问不了，破图按原始尺寸撑爆卡片。
//
// 这里的口径：
// - 收到 cover_image.data_base64 → 解码写入 data/experience-covers/<随机名>.<ext>
// - 返回可直接给 <img src> 用的 URL：/api/experience-covers/<文件名>
// - 目录随进程工作目录（与 xingmang.db 同级），保持单机桌面应用的数据一起搬走
const experienceCoverDir = "data/experience-covers"

// experienceCoverURLPrefix 封面对外访问前缀
const experienceCoverURLPrefix = "/api/experience-covers/"

// experienceCoverMaxBytes 单张封面硬上限（前端已限 5MB，这里留一倍冗余防绕过）
const experienceCoverMaxBytes = 10 << 20

// saveExperienceCoverImage 把 base64 封面写入本地并返回可访问 URL。
// 落盘逻辑已抽到 saveLocalImage（正文插图共用），这里只固定目录/前缀/上限。
func saveExperienceCoverImage(name, mime, dataBase64 string) (string, error) {
	return saveLocalImage(experienceCoverDir, experienceCoverURLPrefix, "封面", name, mime, dataBase64, experienceCoverMaxBytes)
}

// deleteExperienceCoverByURL 删除由本接口落盘的封面（只认自家目录，绝不动外部路径）。
func deleteExperienceCoverByURL(url string) {
	const prefix = "/api/experience-covers/"
	u := strings.TrimSpace(url)
	if !strings.HasPrefix(u, prefix) {
		return
	}
	name := filepath.Base(strings.TrimPrefix(u, prefix))
	if name == "" || name == "." || name == ".." {
		return
	}
	_ = os.Remove(filepath.Join(experienceCoverDir, name))
}

// saveCoverFromRaw 从请求体里取 cover_image（{ name, mime?, data_base64 }）并落盘。
// 没有该字段或字段为空时返回 ("", nil)，调用方按原逻辑走。
func saveCoverFromRaw(raw map[string]json.RawMessage) (string, error) {
	v, ok := raw["cover_image"]
	if !ok || string(v) == "null" {
		return "", nil
	}
	var payload struct {
		Name       string `json:"name"`
		Mime       string `json:"mime"`
		MimeType   string `json:"mime_type"`
		DataBase64 string `json:"data_base64"`
		Data       string `json:"data"`
	}
	if err := json.Unmarshal(v, &payload); err != nil {
		return "", fmt.Errorf("cover_image 格式错误: %w", err)
	}
	// 兼容 size 等多余字段：Unmarshal 会忽略未知键，无需处理
	data := payload.DataBase64
	if strings.TrimSpace(data) == "" {
		data = payload.Data
	}
	if strings.TrimSpace(data) == "" {
		return "", nil
	}
	mime := payload.Mime
	if mime == "" {
		mime = payload.MimeType
	}
	return saveExperienceCoverImage(payload.Name, mime, data)
}

// rawBool 宽松解析请求体布尔（true / "true" / 1 / "1"）。
func rawBool(v json.RawMessage) bool {
	s := strings.TrimSpace(string(v))
	switch s {
	case "true", `"true"`, "1", `"1"`:
		return true
	case "", "null":
		return false
	}
	var b bool
	if err := json.Unmarshal(v, &b); err == nil {
		return b
	}
	var n int
	if err := json.Unmarshal(v, &n); err == nil {
		return n != 0
	}
	return false
}

// ServeExperienceCover 提供封面静态访问：GET /api/experience-covers/{file}
//
// 只允许读取 experienceCoverDir 下的单层文件名，防目录穿越。
func (h *Handler) ServeExperienceCover(w http.ResponseWriter, r *http.Request) {
	serveLocalFile(w, r, experienceCoverDir, experienceCoverURLPrefix)
}

// 远程图片代理。
//
// 背景：贴吧等站点的图床有防盗链，非自家 Referer 一律 403（实测 tiebapic.baidu.com
// 带 localhost Referer 的请求恒返回 403）→ 用户在外链里贴了这类图片地址时，
// 前端 <img> 必然破图。由后端代理抓取（不带 Referer / 用目标站自身 Referer），
// 绕开防盗链，也让前端拿到同源地址、不受跨域与 mixed-content 影响。
//
// 口径：
// - 只允许 http/https，拒内网/环回地址（SSRF 防护）
// - 只回图片类 content-type，响应体上限 8MB
// - 不落盘，纯转发（缓存交给浏览器）
const imageProxyMaxBytes = 8 << 20

func (h *Handler) ProxyImage(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if raw == "" {
		respondFail(w, 400, "缺少 url 参数")
		return
	}

	parsed, err := neturl.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		respondFail(w, 400, "仅支持 http/https 图片地址")
		return
	}
	host := parsed.Hostname()
	if host == "" || isBlockedImageHost(host) {
		respondFail(w, 400, "该地址不允许代理")
		return
	}

	// ---- 本地磁盘缓存：首次抓取后落盘，之后直接读本地文件（外链图片只慢第一次）----
	// 缓存键 = 整个 URL 的 sha256；扩展名在首次落盘时按响应 Content-Type 确定，
	// 查找用 Glob "<hash>.*"（每 URL 最多一个文件，目录很小，扫一下无所谓）。
	if cached := findImageCache(parsed.String()); cached != "" {
		// http.ServeFile 会按扩展名自动设置 Content-Type（.jpg → image/jpeg 等）
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, cached)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, parsed.String(), nil)
	if err != nil {
		respondFail(w, 400, "请求创建失败")
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	// 关键：Referer 设成目标站自身，而不是留空（部分站点空 Referer 也拦）
	req.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		respondFail(w, 502, "拉取远程图片失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondFail(w, 502, fmt.Sprintf("远程图片返回 %d", resp.StatusCode))
		return
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "image/") {
		respondFail(w, 400, "远程资源不是图片")
		return
	}

	// 先读全量（有 8MB 上限）→ 落盘缓存 → 再回给客户端。
	// 好处：缓存写入失败不影响本次响应；坏尾巴不会进缓存。
	body, err := io.ReadAll(io.LimitReader(resp.Body, imageProxyMaxBytes))
	if err != nil || len(body) == 0 {
		respondFail(w, 502, "读取远程图片失败")
		return
	}
	cachePath := storeImageCache(parsed.String(), ct, body) // 失败仅放弃缓存，不阻断

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Image-Cache", map[bool]string{true: "miss", false: "skip"}[cachePath != ""])
	w.Write(body)
}

// imageCacheDir 外链图片的本地缓存目录（随进程 CWD，与数据库/封面同一套 data/ 口径）。
const imageCacheDir = "data/image-cache"

// findImageCache 按 URL 的 sha256 查缓存文件，命中返回完整路径，未命中返回空串。
func findImageCache(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	prefix := filepath.Join(imageCacheDir, hex.EncodeToString(sum[:]))
	matches, err := filepath.Glob(prefix + ".*")
	if err != nil || len(matches) == 0 {
		return ""
	}
	if info, err := os.Stat(matches[0]); err == nil && !info.IsDir() && info.Size() > 0 {
		return matches[0]
	}
	return ""
}

// storeImageCache 把抓到的图片写入缓存，返回缓存路径；失败返回空串（不阻断响应）。
// 原子写：先写 <hash>.tmp 再 rename 成 <hash>.<ext>，避免并发请求读到半截文件。
func storeImageCache(rawURL, contentType string, body []byte) string {
	ext := extFromContentType(contentType)
	sum := sha256.Sum256([]byte(rawURL))
	hash := hex.EncodeToString(sum[:])

	if err := os.MkdirAll(imageCacheDir, 0o755); err != nil {
		return ""
	}
	tmp := filepath.Join(imageCacheDir, hash+".tmp")
	final := filepath.Join(imageCacheDir, hash+ext)
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return ""
	}
	if err := os.Rename(tmp, final); err != nil {
		_ = os.Remove(tmp)
		return ""
	}
	return final
}

// extFromContentType 由 Content-Type 推缓存文件扩展名（供 ServeFile 回填正确 Content-Type）。
func extFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "bmp"):
		return ".bmp"
	case strings.Contains(ct, "avif"):
		return ".avif"
	default:
		return ".img"
	}
}

// isBlockedImageHost 拦截内网/环回/元数据地址，避免 SSRF。
func isBlockedImageHost(host string) bool {
	h := strings.ToLower(host)
	if h == "localhost" || h == "127.0.0.1" || h == "::1" || h == "0.0.0.0" {
		return true
	}
	// 169.254.0.0/16（含云元数据 169.254.169.254）与 10./192.168./172.16-31. 内网段
	if strings.HasPrefix(h, "169.254.") || strings.HasPrefix(h, "10.") ||
		strings.HasPrefix(h, "192.168.") || strings.HasPrefix(h, "127.") {
		return true
	}
	if strings.HasPrefix(h, "172.") {
		parts := strings.Split(h, ".")
		if len(parts) >= 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil && n >= 16 && n <= 31 {
				return true
			}
		}
	}
	return false
}
