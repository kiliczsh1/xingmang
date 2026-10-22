package handlers

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/agent"
	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/ai/providers"
	"github.com/xingmang/xingmang-desktop/internal/aitokens"
	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// getModelConfig gets model config with provider info
func (h *Handler) getModelConfig(configID *int) (*models.APIModelWithProvider, error) {
	var query string
	var args []interface{}

	baseQuery := "SELECT " + apiModelColumnsWithProvider + `
		FROM api_models m LEFT JOIN api_providers p ON m.provider_id = p.id
		WHERE m.enabled = 1`

	if configID != nil && *configID > 0 {
		query = baseQuery + " AND m.id = ?"
		args = append(args, *configID)
	} else {
		query = baseQuery + " AND m.is_default = 1 LIMIT 1"
	}

	c, err := scanModelWithProvider(h.db.QueryRow(query, args...))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("未找到 API 配置")
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// recordUsage records usage statistics
func (h *Handler) recordUsage(modelID int, modelName, providerName string) {
	today := time.Now().Format("2006-01-02")
	var exists int
	h.db.QueryRow("SELECT 1 FROM usage_stats WHERE date = ? AND model_id = ?", today, modelID).Scan(&exists)
	if exists == 1 {
		h.db.Exec("UPDATE usage_stats SET usage_count = usage_count + 1, updated_at = CURRENT_TIMESTAMP WHERE date = ? AND model_id = ?", today, modelID)
	} else {
		h.db.Exec("INSERT INTO usage_stats (date, model_id, model_name, provider_name, usage_count) VALUES (?, ?, ?, ?, 1)", today, modelID, modelName, providerName)
	}
}

// writeSSEError 以合法 JSON 推送 SSE 错误事件（错误文本可能含引号，必须走 json.Marshal 转义）
// GetMaxTokensCache 返回自适应 max_tokens 的探测结果：GET /api/ai/max-tokens-cache
//
// 键形如 `<模型行id>::<渠道id>::<模型串>`，值是**实测可行过**的档位。
// 它不代表模型的物理上限，只代表"这条配置上跑通过的最大值"——界面据此展示
// 「自适应模式下实际探测到的输出长度」，而不是继续显示一个可能过时的配置值。
func (h *Handler) GetMaxTokensCache(w http.ResponseWriter, r *http.Request) {
	respondOK(w, aitokens.All())
}

func writeSSEError(w http.ResponseWriter, msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (h *Handler) AIChat(w http.ResponseWriter, r *http.Request) {
	var req models.AIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	config, err := h.getModelConfig(req.ConfigID)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	// Build messages
	var chatMessages []models.ChatMessage
	var systemContents []string

	if len(req.SystemPrompts) > 0 {
		systemContents = append(systemContents, strings.Join(req.SystemPrompts, "\n\n"))
	}
	if len(req.RelatedContent) > 0 {
		var parts []string
		for _, rc := range req.RelatedContent {
			parts = append(parts, fmt.Sprintf("[%s] %s\n%s", rc.Type, rc.Title, rc.Content))
		}
		systemContents = append(systemContents, "参考内容：\n"+strings.Join(parts, "\n\n"))
	}
	if len(systemContents) > 0 {
		chatMessages = append(chatMessages, models.ChatMessage{
			Role:    "system",
			Content: strings.Join(systemContents, "\n\n---\n\n"),
		})
	}
	chatMessages = append(chatMessages, req.Messages...)
	// 空 content 的消息（content:""）会被中转站转成 Gemini 未初始化的 Part，
	// 上游直接 400：contents[n].parts[0].data: required oneof field 'data'...
	// 这里统一丢弃；全被丢光时给出明确错误而不是透传上游报错。
	filtered := chatMessages[:0]
	for _, m := range chatMessages {
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		filtered = append(filtered, m)
	}
	chatMessages = filtered
	if len(chatMessages) == 0 {
		respondFail(w, 400, "消息内容为空，无法发起生成")
		return
	}

	apiFormat := config.APIFormat
	if apiFormat == "" {
		apiFormat = "openai"
	}
	useFullUrl := config.UseFullURL == 1
	apiUrl := ai.BuildApiUrl(config.APIUrl, apiFormat, useFullUrl)
	requestBody := ai.BuildRequestBody(*config, chatMessages, true)
	headers := ai.BuildAuthHeadersForModel(config.APIKey, apiFormat, config.AuthType, config.AccessKey, config.SecretKey, config.GroupID)
	// 追加 extra_config 自定义头与 OpenCode Go 必需的会话头（缺失会被上游拒绝）
	ai.ApplyProviderHeaders(headers, config.APIUrl, config.ProviderID, config.ExtraConfig)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Make streaming request
	//
	// 自适应 max_tokens：auto 模式下从家族上限逐档下探，只把"token 超限"当作
	// 继续降档的信号（判据见 internal/aitokens）。探测**只发生在响应头到达之前** ——
	// 一旦开始转发 SSE 就不能再重试，否则用户会看到重复输出。
	client := &http.Client{Timeout: 180 * time.Second}
	hint := 0
	if req.MaxTokens != nil {
		hint = *req.MaxTokens
	}
	candidates := aitokens.Candidates(config, hint)

	var resp *http.Response
	var lastStatus int
	var lastBody string
	// cacheable 记录本次降档是否值得写缓存。链路上只要出现过一次「上下文超限」，
	// 就说明失败是**本次输入太长**造成的、与模型能力无关 —— 换句短话本来就能用大值，
	// 把它写进缓存等于把偶发当成永久结论（这正是缓存「只降不升」的成因之一）。
	cacheable := true
	for i, mt := range candidates {
		applied := aitokens.ApplyToBody(requestBody, mt, config)
		bodyBytes, _ := json.Marshal(requestBody)
		httpReq, reqErr := http.NewRequest("POST", apiUrl, strings.NewReader(string(bodyBytes)))
		if reqErr != nil {
			writeSSEError(w, reqErr.Error())
			fmt.Fprintf(w, "data: [DONE]\n\n")
			return
		}
		for k, v := range headers {
			httpReq.Header.Set(k, v)
		}

		resp, err = client.Do(httpReq)
		if err != nil {
			writeSSEError(w, err.Error())
			fmt.Fprintf(w, "data: [DONE]\n\n")
			return
		}
		if resp.StatusCode < 400 {
			// 探测成功：记下来，后续请求直接从这一档起步，省掉一次必然失败的往返。
			// applied == 0 表示本次没发送该字段（交给上游决定），没有值可记。
			//
			// ⚠ 带 hint 的请求（调用方自己指定了上限，如图谱抽取）**不能写缓存**：
			// hint 是「这次我想省点额度」的业务选择，不是「上游最多只能给这么多」的结论。
			// 写进缓存会让该模型此后所有请求都从这个偏小的值起步，等于把一次调用方的
			// 局部限制永久固化成模型能力 —— 实际发生过：图谱抽取传 8192，
			// 缓存记成 8192，之后连普通对话也被压在 8192。
			if cacheable && applied > 0 && hint <= 0 {
				aitokens.Remember(aitokens.CacheKey(config), applied)
			}
			break
		}

		lastStatus = resp.StatusCode
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastBody = string(body)
		resp = nil

		kind := aitokens.Classify(lastStatus, lastBody)
		if !aitokens.CanCacheAfter(kind) {
			cacheable = false
		}
		if i < len(candidates)-1 && kind != aitokens.FailureOther {
			log.Printf("max_tokens=%d 被上游拒绝(%d, %s)，降档重试: %s",
				applied, lastStatus, kind, lastBody[:min(200, len(lastBody))])
			continue
		}
		writeSSEError(w, fmt.Sprintf("API错误(%d): %s", lastStatus, lastBody))
		fmt.Fprintf(w, "data: [DONE]\n\n")
		return
	}
	if resp == nil {
		// 所有档位都被拒（或候选为空），把最后一发的错误原样报给前端。
		writeSSEError(w, fmt.Sprintf("API错误(%d): %s", lastStatus, lastBody))
		fmt.Fprintf(w, "data: [DONE]\n\n")
		return
	}
	defer resp.Body.Close()

	// Forward stream。
	// 思考内容（delta.reasoning_content / reasoning 字段，以及 content 里
	// 混入的 <think>…</think> 块，部分中转会把两者混进 content）单独以
	// {"reasoning": ...} 事件下发，正文仍是 {"content": ...}；前端据此把
	// 思考过程折叠进独立面板，历史/复制/导入只保留正文（无痕）。
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var split agent.ThinkSplitter
	var pendingContent, pendingReasoning strings.Builder
	lastContentEmit, lastReasonEmit := time.Time{}, time.Time{}

	writeSSEField := func(field, text string) {
		data, _ := json.Marshal(map[string]string{field: text})
		fmt.Fprintf(w, "data: %s\n\n", data)
	}
	flushPending := func(force bool) {
		now := time.Now()
		if pendingContent.Len() > 0 && (force || now.Sub(lastContentEmit) >= 350*time.Millisecond) {
			writeSSEField("content", pendingContent.String())
			pendingContent.Reset()
			lastContentEmit = now
		}
		if pendingReasoning.Len() > 0 && (force || now.Sub(lastReasonEmit) >= 350*time.Millisecond) {
			writeSSEField("reasoning", pendingReasoning.String())
			pendingReasoning.Reset()
			lastReasonEmit = now
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		message := strings.TrimPrefix(line, "data:")
		message = strings.TrimSpace(message)
		if message == "[DONE]" {
			flushPending(true)
			h.recordUsage(config.ID, config.Name, config.ProviderName)
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return
		}
		// Parse and extract content
		var chunk struct {
			Error   any `json:"error"`
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
					Reasoning        string `json:"reasoning"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(message), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil {
			errData, _ := json.Marshal(map[string]interface{}{"error": chunk.Error})
			fmt.Fprintf(w, "data: %s\n\n", errData)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			continue
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			// 标准推理字段（DeepSeek / OpenRouter 等）直接进思考流
			if rc := delta.ReasoningContent; rc != "" {
				pendingReasoning.WriteString(rc)
			}
			if rc := delta.Reasoning; rc != "" {
				pendingReasoning.WriteString(rc)
			}
			// content 里可能夹着 <think> 推理块，必须分离
			if delta.Content != "" {
				ans, th := split.Feed(delta.Content)
				if ans != "" {
					pendingContent.WriteString(ans)
				}
				if th != "" {
					pendingReasoning.WriteString(th)
				}
			}
			flushPending(false)
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	// 流结束：吐出分流器残留（未闭合的 <think> 内容一律算思考）
	if ans, th := split.Flush(); ans != "" || th != "" {
		if ans != "" {
			pendingContent.WriteString(ans)
		}
		if th != "" {
			pendingReasoning.WriteString(th)
		}
	}
	flushPending(true)

	if err := scanner.Err(); err != nil {
		writeSSEError(w, fmt.Sprintf("流读取错误: %s", err.Error()))
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (h *Handler) GenerateDescription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		PromptID *int   `json:"promptId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	promptContent := "你是一个专业的网络小说编辑，请根据书名生成一个吸引人的简介，字数控制在 100-200 字之间。"
	if req.PromptID != nil {
		var p models.Prompt
		h.db.QueryRow("SELECT content FROM prompts WHERE id = ?", *req.PromptID).Scan(&p.Content)
		if p.Content != "" {
			promptContent = p.Content
		}
	}

	config, err := h.getModelConfig(nil)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	apiFormat := config.APIFormat
	if apiFormat == "" {
		apiFormat = "openai"
	}
	apiUrl := ai.BuildApiUrl(config.APIUrl, apiFormat, config.UseFullURL == 1)
	messages := []models.ChatMessage{
		{Role: "system", Content: promptContent},
		{Role: "user", Content: "书名：" + req.Title},
	}
	requestBody := ai.BuildRequestBody(*config, messages, false)
	headers := ai.BuildAuthHeadersForModel(config.APIKey, apiFormat, config.AuthType, config.AccessKey, config.SecretKey, config.GroupID)
	// 追加 extra_config 自定义头与 OpenCode Go 必需的会话头（缺失会被上游拒绝）
	ai.ApplyProviderHeaders(headers, config.APIUrl, config.ProviderID, config.ExtraConfig)

	result, err := ai.RequestJSON(apiUrl, requestBody, headers, 120*time.Second)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	respondOK(w, ai.ExtractResponseContent(result))
}

// GenerateImage 封面 / 插图用的一次性文生图：POST /api/ai/generate-image
//
// 请求：{ messages, configId?, size?, quality?, style?, reference_images?, bookId? }
// 响应：{ url } —— url 是落盘后的站内地址，可直接用于 <img src> 与封面预览
//
// 与 /api/media/image 共用同一套适配器与选模型逻辑（config_type='image'），
// 差别只在于：这里取 messages 里最后一条 user 文本当提示词，并强制把产出落到本地。
//
// 落点由 bookId 决定：
//   - 带 bookId → books/<书名>/gallery/，返回 /api/books/{id}/gallery/<name>
//     （和"被顶掉的封面"同一个池子，前端画廊里一起展示）；
//   - 不带     → 共享缓存目录 data/generated-images（没有作品上下文的调用）。
//
// 之所以要分：生成图是"某个作品的素材"，跟着作品走才符合"删作品=删文件夹"的口径；
// 留在共享缓存里的话，作品被删/被备份时这些图都带不走。
func (h *Handler) GenerateImage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Messages        []models.ChatMessage       `json:"messages"`
		ConfigID        *int                       `json:"configId"`
		BookID          *int                       `json:"bookId"`
		Size            string                     `json:"size"`
		Quality         string                     `json:"quality"`
		Style           string                     `json:"style"`
		ReferenceImages []providers.ReferenceImage `json:"reference_images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	prompt := lastUserContent(req.Messages)
	if strings.TrimSpace(prompt) == "" {
		respondFail(w, 400, "提示词为空，无法生成图片")
		return
	}

	svc := bookstore.NewService(h.db)
	sink := globalImageSink
	if req.BookID != nil && *req.BookID > 0 {
		if _, err := svc.BookDir(*req.BookID); err != nil {
			respondFail(w, 400, "作品目录不可用: "+err.Error())
			return
		}
		sink = galleryImageSink{svc: svc, bookID: *req.BookID}
	}

	config, err := h.mediaModelConfig(req.ConfigID, "image")
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	adapter := providers.ResolveImage(config)
	result, err := adapter.Generate(r.Context(), config, providers.ImageRequest{
		Prompt:          prompt,
		Size:            req.Size,
		Count:           1,
		ReferenceImages: req.ReferenceImages,
	})
	if err != nil {
		respondFail(w, 502, err.Error())
		return
	}
	if len(result.Images) == 0 {
		respondFail(w, 502, "上游返回里没有图片数据")
		return
	}

	url := persistGeneratedImage(result.Images[0], sink)
	if url == "" {
		respondFail(w, 502, "图片落盘失败")
		return
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	respondOK(w, map[string]interface{}{
		"url":      url,
		"model":    config.Model,
		"provider": config.ProviderName,
		"adapter":  adapter.Name(),
	})
}

// lastUserContent 取最后一条 user 消息的文本（图片生成的提示词来源）。
func lastUserContent(messages []models.ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" && strings.TrimSpace(messages[i].Content) != "" {
			return messages[i].Content
		}
	}
	// 兜底：没有 user 角色时，取任意最后一条非空消息
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.TrimSpace(messages[i].Content) != "" {
			return messages[i].Content
		}
	}
	return ""
}

// RecognizeCharacters handles character recognition requests (placeholder).
func (h *Handler) RecognizeCharacters(w http.ResponseWriter, r *http.Request) {
	respondFail(w, 501, "角色识别功能尚未实现")
}

// GetRankReferenceCategories returns available rank reference categories.
func (h *Handler) GetRankReferenceCategories(w http.ResponseWriter, r *http.Request) {
	categories := []map[string]string{
		{"key": "genre", "label": "题材类型"},
		{"key": "setting", "label": "世界观设定"},
		{"key": "power", "label": "力量体系"},
		{"key": "faction", "label": "势力阵营"},
		{"key": "role", "label": "角色定位"},
		{"key": "relationship", "label": "关系类型"},
	}
	respondOK(w, categories)
}

// GetRankReferenceBooks returns rank reference books for a given category.
func (h *Handler) GetRankReferenceBooks(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	bookIDStr := r.URL.Query().Get("bookId")

	var chapters []map[string]interface{}
	query := "SELECT id, title, content FROM chapters"
	var args []interface{}

	if bookIDStr != "" {
		query += " WHERE book_id = ?"
		args = append(args, bookIDStr)
	}
	query += " ORDER BY order_num ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondOK(w, []interface{}{})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title, content string
		if err := rows.Scan(&id, &title, &content); err != nil {
			continue
		}
		chapters = append(chapters, map[string]interface{}{
			"id":      id,
			"title":   title,
			"content": content,
		})
	}

	_ = category
	respondOK(w, chapters)
}
