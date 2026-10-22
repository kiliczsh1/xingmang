package providers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// customJSONImageAdapter 按模型 request_rules 里的 JSON 规则发请求并解析结果。
// 适用于不遵循 OpenAI 图片协议的自建/聚合服务。
type customJSONImageAdapter struct{}

func (a *customJSONImageAdapter) Name() string { return "custom_json_image" }

// Match：显式 api_format=custom_json，或配了 request_rules 就走自定义规则。
func (a *customJSONImageAdapter) Match(m *models.APIModelWithProvider) bool {
	return isCustomJSON(m) || hasRequestRules(m)
}

func (a *customJSONImageAdapter) Generate(_ context.Context, m *models.APIModelWithProvider, req ImageRequest) (*ImageResult, error) {
	rules, err := parseMediaRules(m)
	if err != nil {
		return nil, err
	}

	modelName := req.Model
	if modelName == "" {
		modelName = m.Model
	}
	count := req.Count
	if count <= 0 {
		count = 1
	}
	size := req.Size
	if size == "" {
		size = "1024x1024"
	}

	vars := map[string]string{
		"prompt":          req.Prompt,
		"negative_prompt": req.NegativePrompt,
		"model":           modelName,
		"size":            size,
		"count":           strconv.Itoa(count),
	}

	// 没配 submit.url 时，退回 OpenAI 兼容端点与请求体，保证功能可用
	submitPath := rules.Submit.URL
	if submitPath == "" {
		submitPath = ai.BuildImageApiUrl(m.APIUrl, "openai", m.UseFullURL == 1)
	}
	url := resolveURL(m, submitPath)

	var body interface{}
	if len(rules.Submit.Body) > 0 {
		body = applyVarsDeep(rules.Submit.Body, vars)
	} else {
		body = map[string]interface{}{
			"model":  modelName,
			"prompt": req.Prompt,
			"n":      count,
			"size":   size,
		}
	}

	headers := ai.BuildAuthHeadersForModel(m.APIKey, m.APIFormat, m.AuthType, m.AccessKey, m.SecretKey, m.GroupID)
	ai.ApplyProviderHeaders(headers, m.APIUrl, m.ProviderID, m.ExtraConfig)

	raw, err := ai.RequestJSON(url, body, headers, 300*time.Second)
	if err != nil {
		return nil, fmt.Errorf("生图失败: %w", err)
	}

	// 优先按规则里声明的 images 路径取；取不到再走通用归一
	var images []GeneratedImage
	if rules.Response.Images != "" {
		for _, leaf := range extractByPath(raw, rules.Response.Images) {
			if img, ok := imageFromNode(leaf); ok {
				images = append(images, img)
			}
		}
	}
	if len(images) == 0 {
		images = NormalizeImagePayload(raw)
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("上游返回里没有图片数据")
	}
	return &ImageResult{Images: images}, nil
}
