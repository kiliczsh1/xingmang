package providers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// customJSONVideoAdapter 按 request_rules 走「提交任务 + 轮询」两段式协议。
// 视频上游各家差异大，这里只提供通用规则驱动的实现，具体供应商靠填规则接入。
type customJSONVideoAdapter struct{}

func (a *customJSONVideoAdapter) Name() string { return "custom_json_video" }

// Match：显式 api_format=custom_json，或配了 request_rules 就走自定义规则。
func (a *customJSONVideoAdapter) Match(m *models.APIModelWithProvider) bool {
	return isCustomJSON(m) || hasRequestRules(m)
}

func (a *customJSONVideoAdapter) Generate(ctx context.Context, m *models.APIModelWithProvider, req VideoRequest) (*VideoResult, error) {
	rules, err := parseMediaRules(m)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(rules.Submit.URL) == "" {
		return nil, fmt.Errorf("request_rules 缺少 submit.url（生成提交路径）")
	}

	modelName := req.Model
	if modelName == "" {
		modelName = m.Model
	}
	vars := map[string]string{
		"prompt":     req.Prompt,
		"model":      modelName,
		"image_url":  req.ImageURL,
		"duration":   strconv.Itoa(req.Duration),
		"resolution": req.Resolution,
	}

	headers := ai.BuildAuthHeadersForModel(m.APIKey, m.APIFormat, m.AuthType, m.AccessKey, m.SecretKey, m.GroupID)
	ai.ApplyProviderHeaders(headers, m.APIUrl, m.ProviderID, m.ExtraConfig)

	var body interface{}
	if len(rules.Submit.Body) > 0 {
		body = applyVarsDeep(rules.Submit.Body, vars)
	} else {
		def := map[string]interface{}{"model": modelName, "prompt": req.Prompt}
		if req.ImageURL != "" {
			def["image_url"] = req.ImageURL
		}
		if req.Duration > 0 {
			def["duration"] = req.Duration
		}
		if req.Resolution != "" {
			def["resolution"] = req.Resolution
		}
		body = def
	}

	raw, err := ai.RequestJSON(resolveURL(m, rules.Submit.URL), body, headers, 120*time.Second)
	if err != nil {
		return nil, fmt.Errorf("视频任务提交失败: %w", err)
	}

	taskID := firstString(extractByPath(raw, rules.Response.TaskID))

	// 提交即出片
	if direct := firstString(extractByPath(raw, rules.Response.VideoURL)); direct != "" {
		return &VideoResult{URL: direct, TaskID: taskID}, nil
	}
	if strings.TrimSpace(rules.Poll.URL) == "" || taskID == "" {
		return nil, fmt.Errorf("未拿到视频地址或 task_id（检查 request_rules 的 response.video_url / response.task_id）")
	}

	interval := time.Duration(rules.Poll.IntervalMs) * time.Millisecond
	if interval <= 0 {
		interval = 5 * time.Second
	}
	timeout := time.Duration(rules.Poll.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	deadline := time.Now().Add(timeout)

	pollVars := map[string]string{"task_id": taskID}
	for k, v := range vars {
		pollVars[k] = v
	}
	pollURL := resolveURL(m, applyVars(rules.Poll.URL, pollVars))

	doneValues := rules.Response.DoneValues
	if len(doneValues) == 0 {
		doneValues = []string{"succeeded", "success", "completed", "done"}
	}

	for {
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("视频生成超时（%s）", timeout)
		}

		pollRaw, err := ai.RequestJSONGet(pollURL, headers, 30*time.Second)
		if err != nil {
			// 轮询期间的偶发网络错误不致命，继续下一轮
			continue
		}
		if vurl := firstString(extractByPath(pollRaw, rules.Response.VideoURL)); vurl != "" {
			return &VideoResult{URL: vurl, TaskID: taskID}, nil
		}
		if rules.Response.Status != "" {
			st := strings.ToLower(firstString(extractByPath(pollRaw, rules.Response.Status)))
			if isDoneStatus(st, doneValues) {
				return nil, fmt.Errorf("任务状态为 %q 但未返回视频地址", st)
			}
		}
	}
}

// isDoneStatus 判断状态是否落在规则声明的完成值里。
func isDoneStatus(status string, doneValues []string) bool {
	for _, d := range doneValues {
		if strings.EqualFold(strings.TrimSpace(d), status) {
			return true
		}
	}
	return false
}
