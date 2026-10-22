package providers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// mediaRules 是模型 request_rules 字段的 JSON 结构（图片/视频共用）。
//
// 示例（自定义 JSON 视频供应商）：
//
//	{
//	  "submit":   { "url": "/v1/video/submit", "method": "POST", "body": {"prompt": "{{prompt}}", "model": "{{model}}"} },
//	  "poll":     { "url": "/v1/video/tasks/{{task_id}}", "method": "GET", "interval_ms": 5000, "timeout_ms": 300000 },
//	  "response": { "task_id": "data.task_id", "video_url": "data.video_url", "status": "data.status",
//	                "done_values": ["succeeded", "completed"] }
//	}
type mediaRules struct {
	Submit struct {
		URL    string                 `json:"url"`
		Method string                 `json:"method"`
		Body   map[string]interface{} `json:"body"`
	} `json:"submit"`
	Poll struct {
		URL        string `json:"url"`
		Method     string `json:"method"`
		IntervalMs int    `json:"interval_ms"`
		TimeoutMs  int    `json:"timeout_ms"`
	} `json:"poll"`
	Response struct {
		Images     string   `json:"images"`
		TaskID     string   `json:"task_id"`
		VideoURL   string   `json:"video_url"`
		Status     string   `json:"status"`
		DoneValues []string `json:"done_values"`
	} `json:"response"`
}

// isCustomJSON 判断模型是否显式声明走自定义 JSON 协议。
func isCustomJSON(m *models.APIModelWithProvider) bool {
	return strings.EqualFold(strings.TrimSpace(m.APIFormat), "custom_json")
}

// hasRequestRules 判断模型是否配了非空的 request_rules。
func hasRequestRules(m *models.APIModelWithProvider) bool {
	return m.RequestRules != nil && strings.TrimSpace(*m.RequestRules) != ""
}

// parseMediaRules 解析模型的 request_rules；为空返回空规则（调用方按默认行为走）。
func parseMediaRules(m *models.APIModelWithProvider) (*mediaRules, error) {
	rules := &mediaRules{}
	if !hasRequestRules(m) {
		return rules, nil
	}
	raw := strings.TrimSpace(*m.RequestRules)
	if err := json.Unmarshal([]byte(raw), rules); err != nil {
		return nil, fmt.Errorf("request_rules 不是合法 JSON: %w", err)
	}
	return rules, nil
}

// resolveURL 把规则里的路径解析成绝对 URL：已是 http(s) 的原样用，否则拼到 provider 的 Base URL 后面。
func resolveURL(m *models.APIModelWithProvider, path string) string {
	p := strings.TrimSpace(path)
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p
	}
	base := strings.TrimRight(strings.TrimSpace(m.APIUrl), "/")
	if p == "" {
		return base
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return base + p
}

// applyVars 把字符串里的 {{key}} 替换成 vars[key]。
func applyVars(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

// applyVarsDeep 递归替换 map/数组/字符串里的模板变量。
func applyVarsDeep(v interface{}, vars map[string]string) interface{} {
	switch t := v.(type) {
	case string:
		return applyVars(t, vars)
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			out[k] = applyVarsDeep(val, vars)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, val := range t {
			out[i] = applyVarsDeep(val, vars)
		}
		return out
	default:
		return v
	}
}

// extractByPath 按极简路径取值，支持点号与 [] 数组展开，例如 "data[].url"、"data.task_id"。
// 返回值一律是「叶子节点列表」，找不到则为空。
func extractByPath(root interface{}, path string) []interface{} {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	cur := []interface{}{root}
	for _, seg := range strings.Split(path, ".") {
		iter := strings.HasSuffix(seg, "[]")
		seg = strings.TrimSuffix(seg, "[]")

		var next []interface{}
		for _, node := range cur {
			if seg == "" {
				next = append(next, node)
				continue
			}
			switch t := node.(type) {
			case map[string]interface{}:
				if v, ok := t[seg]; ok {
					next = append(next, v)
				}
			case []interface{}:
				for _, item := range t {
					if mm, ok := item.(map[string]interface{}); ok {
						if v, ok := mm[seg]; ok {
							next = append(next, v)
						}
					}
				}
			}
		}
		if iter {
			var flat []interface{}
			for _, n := range next {
				if arr, ok := n.([]interface{}); ok {
					flat = append(flat, arr...)
				} else {
					flat = append(flat, n)
				}
			}
			next = flat
		}
		cur = next
	}
	return cur
}

// firstString 取叶子列表里第一个非空字符串。
func firstString(nodes []interface{}) string {
	for _, n := range nodes {
		if s, ok := n.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

// NormalizeImagePayload 把各种上游返回归一成 []GeneratedImage。
// 兼容：data[].b64_json / data[].url / images[]（字符串或对象）/ 顶层 {url|b64_json}。
func NormalizeImagePayload(raw map[string]interface{}) []GeneratedImage {
	var out []GeneratedImage

	if data, ok := raw["data"].([]interface{}); ok {
		for _, item := range data {
			if img, ok := imageFromNode(item); ok {
				out = append(out, img)
			}
		}
	}
	if len(out) == 0 {
		if imgs, ok := raw["images"].([]interface{}); ok {
			for _, item := range imgs {
				if img, ok := imageFromNode(item); ok {
					out = append(out, img)
				}
			}
		}
	}
	if len(out) == 0 {
		if img, ok := imageFromNode(raw); ok {
			out = append(out, img)
		}
	}
	return out
}

// imageFromNode 从单个节点里提取图片：字符串当 URL，对象取 b64_json / url。
func imageFromNode(node interface{}) (GeneratedImage, bool) {
	switch v := node.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return GeneratedImage{}, false
		}
		return GeneratedImage{URL: v}, true
	case map[string]interface{}:
		img := GeneratedImage{}
		if s, ok := v["b64_json"].(string); ok {
			img.B64JSON = s
		}
		if s, ok := v["url"].(string); ok {
			img.URL = s
		}
		if s, ok := v["revised_prompt"].(string); ok {
			img.RevisedPrompt = s
		}
		if img.B64JSON == "" && img.URL == "" {
			return GeneratedImage{}, false
		}
		return img, true
	}
	return GeneratedImage{}, false
}
