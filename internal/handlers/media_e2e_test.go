package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/database"
)

// setupMediaTestDB 建临时库，塞一个指向 mock 上游的指定类型模型。
func setupMediaTestDB(t *testing.T, upstreamURL, configType string) *Handler {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	res, err := db.Exec(`INSERT INTO api_providers (name, provider_type, api_url, api_format, auth_type) VALUES (?, ?, ?, ?, ?)`,
		"mock", "other", upstreamURL, "openai", "api_key")
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := res.LastInsertId()

	if _, err := db.Exec(`INSERT INTO api_models (provider_id, name, model, config_type, enabled, is_default) VALUES (?, ?, ?, ?, 1, 1)`,
		pid, "测试模型", "mock-model-1", configType); err != nil {
		t.Fatal(err)
	}
	return &Handler{db: db}
}

func decodeEnvelope(t *testing.T, raw []byte) map[string]interface{} {
	t.Helper()
	var env struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
		Message string                 `json:"message"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("响应不是合法 JSON: %v (%s)", err, string(raw))
	}
	if !env.Success {
		t.Fatalf("接口失败: %s", env.Message)
	}
	return env.Data
}

// TestGenerateMediaImageEndToEnd 验证 /api/media/image 走适配器 + 命中 OpenAI 兼容生图端点。
func TestGenerateMediaImageEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/images/generations") {
			t.Errorf("打到了非预期路径: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{{"url": "https://cdn.example.com/x.png"}},
		})
	}))
	defer srv.Close()

	h := setupMediaTestDB(t, srv.URL, "image")

	req := httptest.NewRequest("POST", "/api/media/image", strings.NewReader(`{"prompt":"一只猫","size":"600x800"}`))
	rec := httptest.NewRecorder()
	h.GenerateMediaImage(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	data := decodeEnvelope(t, rec.Body.Bytes())
	images, _ := data["images"].([]interface{})
	if len(images) != 1 {
		t.Fatalf("images=%v", data["images"])
	}
	if data["adapter"] != "openai_image" {
		t.Fatalf("adapter=%v, want openai_image", data["adapter"])
	}
}

// TestGenerateImagePersistsLocally 验证封面用的 /api/ai/generate-image 会把 b64 产出落盘成站内 URL。
func TestGenerateImagePersistsLocally(t *testing.T) {
	t.Chdir(t.TempDir())

	raw := []byte("fake-png-bytes")
	b64 := base64.StdEncoding.EncodeToString(raw)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{{"b64_json": b64}},
		})
	}))
	defer srv.Close()

	h := setupMediaTestDB(t, srv.URL, "image")

	body := `{"messages":[{"role":"user","content":"画一只猫"}],"size":"600x800"}`
	req := httptest.NewRequest("POST", "/api/ai/generate-image", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.GenerateImage(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	data := decodeEnvelope(t, rec.Body.Bytes())
	url, _ := data["url"].(string)
	if !strings.HasPrefix(url, generatedImageURLPrefix) {
		t.Fatalf("url=%q, want prefix %s", url, generatedImageURLPrefix)
	}

	// 文件确实落到了磁盘
	saved := filepath.Join(generatedImageDir, strings.TrimPrefix(url, generatedImageURLPrefix))
	if _, err := os.Stat(saved); err != nil {
		t.Fatalf("落盘文件不存在: %v", err)
	}
}

// TestGenerateImageNoImageModel 验证没有 image 类型模型时会退回默认模型而不是直接崩。
func TestGenerateImageEmptyPrompt(t *testing.T) {
	h := setupMediaTestDB(t, "http://127.0.0.1:1", "image")

	req := httptest.NewRequest("POST", "/api/ai/generate-image", strings.NewReader(`{"messages":[]}`))
	rec := httptest.NewRecorder()
	h.GenerateImage(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status=%d, want 400 (提示词为空)", rec.Code)
	}
}

// TestMediaVideoNoAdapter 验证视频模型没配 request_rules 时返回明确 501，而不是静默失败。
func TestMediaVideoNoAdapter(t *testing.T) {
	h := setupMediaTestDB(t, "http://127.0.0.1:1", "video")

	req := httptest.NewRequest("POST", "/api/media/video", strings.NewReader(`{"prompt":"一只猫在跑"}`))
	rec := httptest.NewRecorder()
	h.GenerateMediaVideo(rec, req)

	if rec.Code != 501 {
		t.Fatalf("status=%d body=%s, want 501", rec.Code, rec.Body.String())
	}
}
