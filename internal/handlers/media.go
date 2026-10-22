package handlers

import (
	"bufio"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/ai/providers"
	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 媒体生成接口（工作流媒体节点 + 封面 / 插图共用）。
//
// 设计口径：
// - 复用星芒既有的 provider / api_models 体系（api_url / api_key / api_format / use_full_url / 扩展认证）
// - 选模型：请求带 config_id 就用它；否则按 config_type 找（chat / image / video / vision / embed），
//   再退回默认模型 —— 不再依赖前端猜出来的 category
// - 具体请求与解析交给 internal/ai/providers 适配器层：自定义 JSON 规则优先，OpenAI 兼容兜底
// - 产出统一落到本地 data/generated-images，返回站内 URL（远端 url 会尽力下载后落盘）

const (
	generatedImageDir       = "data/generated-images"
	generatedImageURLPrefix = "/api/generated-images/"
	generatedImageMaxBytes  = 20 << 20
)

// ServeGeneratedImage 生成图片静态访问：GET /api/generated-images/{file}
func (h *Handler) ServeGeneratedImage(w http.ResponseWriter, r *http.Request) {
	serveLocalFile(w, r, generatedImageDir, generatedImageURLPrefix)
}

// mediaModelConfig 按 config_type 找模型；config_type 为空则退回默认模型。
func (h *Handler) mediaModelConfig(configID *int, configType string) (*models.APIModelWithProvider, error) {
	if configID != nil && *configID > 0 {
		return h.getModelConfig(configID)
	}
	if strings.TrimSpace(configType) == "" {
		return h.getModelConfig(nil)
	}

	c, err := scanModelWithProvider(h.db.QueryRow("SELECT "+apiModelColumnsWithProvider+`
		FROM api_models m LEFT JOIN api_providers p ON m.provider_id = p.id
		WHERE m.enabled = 1 AND COALESCE(m.config_type, 'chat') = ?
		ORDER BY m.is_default DESC, m.sort_order ASC, m.id ASC LIMIT 1`, configType))
	if err == sql.ErrNoRows {
		// 没有该类型的模型就退回默认模型，避免功能直接不可用
		return h.getModelConfig(nil)
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GenerateMediaImage 文生图：POST /api/media/image
//
// 请求：{ config_id?, model?, prompt, size?, count?, negative_prompt? }
// 响应：{ images: [{ b64?, url?, revised_prompt? }], model, provider, adapter }
func (h *Handler) GenerateMediaImage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfigID       *int   `json:"config_id"`
		Prompt         string `json:"prompt"`
		Size           string `json:"size"`
		Count          int    `json:"count"`
		NegativePrompt string `json:"negative_prompt"`
		Model          string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		respondFail(w, 400, "prompt 不能为空")
		return
	}
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 8 {
		req.Count = 8
	}

	config, err := h.mediaModelConfig(req.ConfigID, "image")
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	adapter := providers.ResolveImage(config)
	result, err := adapter.Generate(r.Context(), config, providers.ImageRequest{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Size:           req.Size,
		Count:          req.Count,
		Model:          req.Model,
	})
	if err != nil {
		respondFail(w, 502, err.Error())
		return
	}

	images := make([]map[string]interface{}, 0, len(result.Images))
	for _, img := range result.Images {
		entry := map[string]interface{}{}
		if img.B64JSON != "" {
			entry["b64"] = img.B64JSON
		}
		if img.URL != "" {
			entry["url"] = img.URL
		}
		if img.RevisedPrompt != "" {
			entry["revised_prompt"] = img.RevisedPrompt
		}
		if len(entry) > 0 {
			images = append(images, entry)
		}
	}
	if len(images) == 0 {
		respondFail(w, 502, "上游返回里没有图片数据")
		return
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	respondOK(w, map[string]interface{}{
		"images":   images,
		"model":    config.Model,
		"provider": config.ProviderName,
		"adapter":  adapter.Name(),
	})
}

// imageSink 描述"生成出来的图该落到哪里"。
//
// 绝大多数调用（工作流媒体节点、插图等）没有作品上下文，落共享缓存目录；
// 封面生成会带上 bookId，于是**直接落进该作品的 gallery/** —— 与"被顶掉的封面"
// 同一个池子，前端画廊里一起展示，图和作品也绑在一起（删作品=删文件夹、
// 备份作品=拷文件夹，共享缓存目录不会再越攒越多）。
type imageSink interface {
	Save(raw []byte, ext string) (string, error)
}

// dirImageSink 落到固定目录 + 固定 URL 前缀（没有作品上下文时的共享缓存目录）。
type dirImageSink struct{ dir, prefix string }

func (s dirImageSink) Save(raw []byte, ext string) (string, error) {
	return saveLocalImageBytes(s.dir, s.prefix, "生成图片", ext, raw, generatedImageMaxBytes)
}

// globalImageSink 是没有作品上下文时的落点（进程级缓存目录）。
var globalImageSink imageSink = dirImageSink{dir: generatedImageDir, prefix: generatedImageURLPrefix}

// galleryImageSink 落到某作品的 gallery/。
//
// 走 bookstore 的 AddGalleryBytes 而不是直接写目录，是为了拿到与归档封面一致的
// 命名（<时间戳>-<内容指纹><扩展名>）与去重 —— 画廊里新旧图排序才会统一。
type galleryImageSink struct {
	svc    *bookstore.Service
	bookID int
}

func (s galleryImageSink) Save(raw []byte, ext string) (string, error) {
	name, err := s.svc.AddGalleryBytes(s.bookID, raw, ext)
	if err != nil {
		return "", err
	}
	return bookGalleryURLPrefix(s.bookID) + name, nil
}

// persistGeneratedImage 把适配器产出的图片落到本地，返回站内 URL。
// b64 / data URL 直接落盘；只有远端 url 时尽力下载后落盘，下载失败则退回原 url。
func persistGeneratedImage(img providers.GeneratedImage, sink imageSink) string {
	if b64 := strings.TrimSpace(img.B64JSON); b64 != "" {
		if raw, mime, err := decodeImageData(b64, ""); err == nil {
			if url, err := sink.Save(raw, extForImageMime(mime, "")); err == nil {
				return url
			}
		}
	}

	u := strings.TrimSpace(img.URL)
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "data:") {
		if raw, mime, err := decodeImageData(u, ""); err == nil {
			if url, err := sink.Save(raw, extForImageMime(mime, "")); err == nil {
				return url
			}
		}
		return ""
	}
	// 已经是本程序在管的站内图片（缓存目录、插图、某作品的画廊）：直接用，别再下一遍
	if isSiteImageURL(u) {
		return u
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return u
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return u
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, int64(generatedImageMaxBytes)+1))
	if err != nil || len(data) == 0 || len(data) > generatedImageMaxBytes {
		return u
	}

	ext := filepath.Ext(strings.SplitN(u, "?", 2)[0])
	if ext == "" {
		mime := strings.TrimSpace(strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)[0])
		ext = imageExtByMime[strings.ToLower(mime)]
	}
	local, err := sink.Save(data, ext)
	if err != nil {
		return u
	}
	return local
}

// decodeImageData 解码 base64 / data URL 形态的图片，返回原始字节与 MIME。
func decodeImageData(s, fallbackMime string) ([]byte, string, error) {
	mime := fallbackMime
	if mime == "" {
		mime = "image/png"
	}
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "data:") {
		comma := strings.Index(s, ",")
		if comma < 0 {
			return nil, "", fmt.Errorf("图片数据格式错误")
		}
		if i := strings.Index(s[len("data:"):comma], ";"); i >= 0 {
			mime = s[len("data:") : len("data:")+i]
		}
		s = s[comma+1:]
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		raw, err = base64.RawURLEncoding.DecodeString(strings.TrimSpace(s))
		if err != nil {
			return nil, "", fmt.Errorf("图片 base64 解码失败: %w", err)
		}
	}
	if len(raw) == 0 {
		return nil, "", fmt.Errorf("图片数据为空")
	}
	if len(raw) > generatedImageMaxBytes {
		return nil, "", fmt.Errorf("图片过大（%d 字节，上限 %d）", len(raw), generatedImageMaxBytes)
	}
	return raw, mime, nil
}

// extForImageMime 按 MIME 猜扩展名，认不出时退回文件名自带的扩展名。
func extForImageMime(mime, name string) string {
	if ext := imageExtByMime[strings.ToLower(strings.TrimSpace(mime))]; ext != "" {
		return ext
	}
	return filepath.Ext(name)
}

// AnalyzeMediaImage 图片分析：POST /api/media/analyze（SSE 流式）
//
// 请求：{ config_id?, image_base64?, image_url?, prompt?, detail? }
// 流：data: {"content":"..."} … data: [DONE]
func (h *Handler) AnalyzeMediaImage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfigID    *int   `json:"config_id"`
		ImageBase64 string `json:"image_base64"`
		ImageURL    string `json:"image_url"`
		Prompt      string `json:"prompt"`
		Detail      string `json:"detail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	imageURL := strings.TrimSpace(req.ImageURL)
	if imageURL == "" && strings.TrimSpace(req.ImageBase64) != "" {
		b64 := strings.TrimSpace(req.ImageBase64)
		if strings.HasPrefix(b64, "data:") {
			imageURL = b64
		} else {
			imageURL = "data:image/png;base64," + b64
		}
	}
	if imageURL == "" {
		respondFail(w, 400, "必须提供 image_base64 或 image_url")
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = "请描述这张图片的内容。"
	}

	config, err := h.mediaModelConfig(req.ConfigID, "vision")
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	apiFormat := config.APIFormat
	if apiFormat == "" {
		apiFormat = "openai"
	}
	apiUrl := ai.BuildApiUrl(config.APIUrl, apiFormat, config.UseFullURL == 1)
	headers := ai.BuildAuthHeadersForModel(config.APIKey, apiFormat, config.AuthType, config.AccessKey, config.SecretKey, config.GroupID)
	ai.ApplyProviderHeaders(headers, config.APIUrl, config.ProviderID, config.ExtraConfig)

	// OpenAI 兼容多模态 messages
	detail := req.Detail
	if detail == "" {
		detail = "auto"
	}
	messages := []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{"type": "text", "text": prompt},
				{"type": "image_url", "image_url": map[string]interface{}{"url": imageURL, "detail": detail}},
			},
		},
	}
	// max_tokens 为空时**不发送该字段**，与 ai.BuildRequestBody 一致：
	// 编一个 2000 兜底只会在模型本可以写更长时静默截断输出。
	body := map[string]interface{}{
		"model":    config.Model,
		"messages": messages,
		"stream":   true,
	}
	if config.MaxTokens > 0 {
		body["max_tokens"] = config.MaxTokens
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	bodyBytes, _ := json.Marshal(body)
	client := &http.Client{Timeout: 300 * time.Second}
	httpReq, _ := http.NewRequest("POST", apiUrl, strings.NewReader(string(bodyBytes)))
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		writeSSEError(w, err.Error())
		fmt.Fprint(w, "data: [DONE]\n\n")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		writeSSEError(w, fmt.Sprintf("API错误(%d): %s", resp.StatusCode, string(raw)))
		fmt.Fprint(w, "data: [DONE]\n\n")
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		message := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if message == "[DONE]" {
			break
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(message), &parsed); err != nil {
			continue
		}
		if errMsg, ok := parsed["error"]; ok {
			data, _ := json.Marshal(map[string]interface{}{"error": errMsg})
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			continue
		}
		if choices, ok := parsed["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok && content != "" {
						data, _ := json.Marshal(map[string]string{"content": content})
						fmt.Fprintf(w, "data: %s\n\n", data)
					}
				}
			}
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	if err := scanner.Err(); err != nil {
		writeSSEError(w, fmt.Sprintf("流读取错误: %s", err.Error()))
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	fmt.Fprint(w, "data: [DONE]\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// GenerateMediaVideo 文生视频：POST /api/media/video
// 请求：{ config_id?, model?, prompt, duration?, resolution? }
// 响应：{ url, task_id?, model, provider, adapter }
func (h *Handler) GenerateMediaVideo(w http.ResponseWriter, r *http.Request) {
	h.handleVideoGeneration(w, r, false)
}

// MediaImageToVideo 图生视频：POST /api/media/image-to-video
// 请求：{ config_id?, model?, prompt?, image_url, duration?, resolution? }
func (h *Handler) MediaImageToVideo(w http.ResponseWriter, r *http.Request) {
	h.handleVideoGeneration(w, r, true)
}

// handleVideoGeneration 文生视频 / 图生视频共用逻辑。
// 视频上游没有通用协议，统一交给 providers 里的自定义 JSON 适配器（提交 + 轮询）；
// 模型没配 request_rules 时明确报错，而不是静默失败。
func (h *Handler) handleVideoGeneration(w http.ResponseWriter, r *http.Request, imageToVideo bool) {
	var req struct {
		ConfigID   *int   `json:"config_id"`
		Model      string `json:"model"`
		Prompt     string `json:"prompt"`
		ImageURL   string `json:"image_url"`
		Duration   int    `json:"duration"`
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if imageToVideo && strings.TrimSpace(req.ImageURL) == "" {
		respondFail(w, 400, "image_url 不能为空")
		return
	}
	if !imageToVideo && strings.TrimSpace(req.Prompt) == "" {
		respondFail(w, 400, "prompt 不能为空")
		return
	}

	config, err := h.mediaModelConfig(req.ConfigID, "video")
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	adapter := providers.ResolveVideo(config)
	if adapter == nil {
		respondFail(w, 501, "没有匹配的视频适配器：该模型未配置 request_rules。"+
			"视频上游协议各异（多为提交任务 + 轮询），请在「API 配置」里为该模型填写自定义 JSON 规则")
		return
	}

	result, err := adapter.Generate(r.Context(), config, providers.VideoRequest{
		Prompt:     req.Prompt,
		ImageURL:   req.ImageURL,
		Duration:   req.Duration,
		Resolution: req.Resolution,
		Model:      req.Model,
	})
	if err != nil {
		respondFail(w, 502, err.Error())
		return
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	respondOK(w, map[string]interface{}{
		"url":      result.URL,
		"task_id":  result.TaskID,
		"model":    config.Model,
		"provider": config.ProviderName,
		"adapter":  adapter.Name(),
	})
}
