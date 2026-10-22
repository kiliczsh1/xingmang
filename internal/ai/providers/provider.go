// Package providers 是媒体生成（图片 / 视频）的供应商适配器层。
//
// 设计口径：
//   - 只关心「怎么发请求、怎么解析结果」，协议与鉴权工具复用 internal/ai
//   - 图片解析顺序：自定义 JSON 规则优先，OpenAI 兼容兜底（永远有可用适配器）
//   - 视频无通用协议：只保留「自定义 JSON 规则」适配器，未匹配时明确报错
//   - 新增供应商只需实现接口并注册，不改 handler
package providers

import (
	"context"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// ReferenceImage 是参考图（图生图 / 首尾帧用）。
type ReferenceImage struct {
	Name       string `json:"name"`
	MimeType   string `json:"mime_type"`
	DataBase64 string `json:"data_base64"`
}

// ImageRequest 是文生图的统一入参。
type ImageRequest struct {
	Prompt          string
	NegativePrompt  string
	Size            string
	Count           int
	Model           string
	ReferenceImages []ReferenceImage
}

// GeneratedImage 是单张产出（b64 或 url 二选一）。
type GeneratedImage struct {
	B64JSON       string `json:"b64_json,omitempty"`
	URL           string `json:"url,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageResult 是文生图的统一出参。
type ImageResult struct {
	Images []GeneratedImage
}

// ImageAdapter 是图片生成适配器。
type ImageAdapter interface {
	Name() string
	Match(m *models.APIModelWithProvider) bool
	Generate(ctx context.Context, m *models.APIModelWithProvider, req ImageRequest) (*ImageResult, error)
}

// VideoRequest 是视频生成的统一入参。
type VideoRequest struct {
	Prompt     string
	ImageURL   string
	Duration   int
	Resolution string
	Model      string
}

// VideoResult 是视频生成的统一出参。
type VideoResult struct {
	URL    string
	TaskID string
}

// VideoAdapter 是视频生成适配器。
type VideoAdapter interface {
	Name() string
	Match(m *models.APIModelWithProvider) bool
	Generate(ctx context.Context, m *models.APIModelWithProvider, req VideoRequest) (*VideoResult, error)
}

// imageAdapters 是图片适配器注册表：顺序即优先级，OpenAI 兼容兜底必须最后。
var imageAdapters = []ImageAdapter{
	&customJSONImageAdapter{},
	&openAICompatibleImageAdapter{},
}

// videoAdapters 是视频适配器注册表：顺序即优先级，无兜底。
var videoAdapters = []VideoAdapter{
	&customJSONVideoAdapter{},
}

// ResolveImage 选出能处理该模型的图片适配器（永远有值，最后一项是 OpenAI 兼容兜底）。
func ResolveImage(m *models.APIModelWithProvider) ImageAdapter {
	for _, a := range imageAdapters {
		if a.Match(m) {
			return a
		}
	}
	return imageAdapters[len(imageAdapters)-1]
}

// ResolveVideo 选出能处理该模型的视频适配器；无匹配返回 nil（调用方据此明确报错）。
func ResolveVideo(m *models.APIModelWithProvider) VideoAdapter {
	for _, a := range videoAdapters {
		if a.Match(m) {
			return a
		}
	}
	return nil
}
