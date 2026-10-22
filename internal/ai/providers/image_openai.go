package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// openAICompatibleImageAdapter 走 OpenAI 兼容的 /v1/images/generations，作为兜底适配器。
type openAICompatibleImageAdapter struct{}

func (a *openAICompatibleImageAdapter) Name() string { return "openai_image" }

// Match 恒真：注册表里排最后，任何没被前面适配器接住的模型都走这里。
func (a *openAICompatibleImageAdapter) Match(*models.APIModelWithProvider) bool { return true }

func (a *openAICompatibleImageAdapter) Generate(_ context.Context, m *models.APIModelWithProvider, req ImageRequest) (*ImageResult, error) {
	url := ai.BuildImageApiUrl(m.APIUrl, m.APIFormat, m.UseFullURL == 1)
	headers := ai.BuildAuthHeadersForModel(m.APIKey, m.APIFormat, m.AuthType, m.AccessKey, m.SecretKey, m.GroupID)
	ai.ApplyProviderHeaders(headers, m.APIUrl, m.ProviderID, m.ExtraConfig)

	count := req.Count
	if count <= 0 {
		count = 1
	}
	size := req.Size
	if size == "" {
		size = "1024x1024"
	}
	modelName := req.Model
	if modelName == "" {
		modelName = m.Model
	}

	body := map[string]interface{}{
		"model":  modelName,
		"prompt": req.Prompt,
		"n":      count,
		"size":   size,
	}
	if req.NegativePrompt != "" {
		body["negative_prompt"] = req.NegativePrompt
	}
	if len(req.ReferenceImages) > 0 {
		body["reference_images"] = req.ReferenceImages
	}

	raw, err := ai.RequestJSON(url, body, headers, 300*time.Second)
	if err != nil {
		return nil, fmt.Errorf("生图失败: %w", err)
	}
	images := NormalizeImagePayload(raw)
	if len(images) == 0 {
		return nil, fmt.Errorf("上游返回里没有图片数据")
	}
	return &ImageResult{Images: images}, nil
}
