package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// APIResponse is the standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// StringList is a []string that can be unmarshaled from a JSON string or array,
// and always marshals to a JSON array. Used for fields stored as JSON text in DB
// but consumed as arrays by the frontend (e.g. book tags).
type StringList []string

func (s *StringList) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = nil
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		str = strings.TrimSpace(str)
		if str == "" || str == "null" {
			*s = nil
			return nil
		}
		// 兼容 DB 中存的 JSON 数组字符串，如 "[\"a\",\"b\"]"
		if strings.HasPrefix(str, "[") {
			var inner []string
			if err := json.Unmarshal([]byte(str), &inner); err == nil {
				*s = inner
				return nil
			}
		}
		// 逗号分隔的纯文本
		parts := strings.Split(str, ",")
		cleaned := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				cleaned = append(cleaned, t)
			}
		}
		*s = cleaned
		return nil
	}
	*s = nil
	return nil
}

// ToDBValue returns the JSON array text to store in the TEXT column.
func (s StringList) ToDBValue() string {
	if len(s) == 0 {
		return "[]"
	}
	data, err := json.Marshal(s)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// Book represents a book
type Book struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Cover        string     `json:"cover"`
	Author       string     `json:"author"`
	Category     string     `json:"category"`
	Tags         StringList `json:"tags"` // stored as JSON array text in DB
	Status       string     `json:"status"`
	IsPublic     int        `json:"is_public"`
	WordCount    int        `json:"word_count"`
	ChapterCount int        `json:"chapter_count"`
	TargetWords  int        `json:"target_words"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

// FlexibleInt is an int that can be unmarshaled from bool or int JSON values.
type FlexibleInt int

func (f *FlexibleInt) UnmarshalJSON(data []byte) error {
	// Try bool first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		if b {
			*f = 1
		} else {
			*f = 0
		}
		return nil
	}
	// Try int
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleInt(n)
		return nil
	}
	// Default
	*f = 0
	return nil
}

// FlexibleString is a *string that can be unmarshaled from any JSON value.
// Arrays/objects are stored as their raw JSON string representation.
// null becomes nil.
type FlexibleString struct {
	Value *string
}

func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		f.Value = nil
		return nil
	}
	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		f.Value = &s
		return nil
	}
	// It's an array/object/number/bool — store raw JSON as string
	raw := string(data)
	f.Value = &raw
	return nil
}

// ToPtr returns the underlying *string
func (f *FlexibleString) ToPtr() *string {
	return f.Value
}

// JSONText 以 JSON 文本存进 TEXT 列，但在 HTTP 层按 JSON 值收发。
// graph_entities.metadata 这类字段前端按对象发送、数据库按字符串存储，
// 用 string 接收对象会直接报 "cannot unmarshal object into Go struct field"。
// 非法 JSON 退化为字符串输出，保证响应不因脏数据失败。
type JSONText string

func (j *JSONText) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*j = "{}"
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			s = "{}"
		}
		*j = JSONText(s)
		return nil
	}
	// 对象/数组/数字/布尔：原样保留 JSON 文本
	*j = JSONText(string(data))
	return nil
}

func (j JSONText) MarshalJSON() ([]byte, error) {
	if j == "" {
		return []byte("{}"), nil
	}
	if json.Valid([]byte(j)) {
		return []byte(j), nil
	}
	return json.Marshal(string(j))
}

// Scan 让 JSONText 可直接作为 database/sql 的扫描目标
func (j *JSONText) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		*j = ""
	case []byte:
		*j = JSONText(v)
	case string:
		*j = JSONText(v)
	default:
		return fmt.Errorf("无法将 %T 读作 JSONText", src)
	}
	return nil
}

// Value 让 JSONText 可直接作为 database/sql 的写入参数
func (j JSONText) Value() (driver.Value, error) {
	return string(j), nil
}

// UnmarshalJSON handles flexible tags and is_public fields for Book
func (b *Book) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &b.ID)
	}
	if v, ok := raw["title"]; ok {
		json.Unmarshal(v, &b.Title)
	}
	if v, ok := raw["description"]; ok {
		json.Unmarshal(v, &b.Description)
	}
	if v, ok := raw["cover"]; ok {
		json.Unmarshal(v, &b.Cover)
	}
	if v, ok := raw["author"]; ok {
		json.Unmarshal(v, &b.Author)
	}
	if v, ok := raw["category"]; ok {
		json.Unmarshal(v, &b.Category)
	}
	if v, ok := raw["status"]; ok {
		json.Unmarshal(v, &b.Status)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &b.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &b.UpdatedAt)
	}
	if v, ok := raw["target_words"]; ok {
		json.Unmarshal(v, &b.TargetWords)
	}
	// Tags: accept string or array (StringList 自带灵活解析)
	if v, ok := raw["tags"]; ok {
		var tags StringList
		if err := json.Unmarshal(v, &tags); err == nil {
			b.Tags = tags
		}
	}
	// IsPublic: accept bool or int
	if v, ok := raw["is_public"]; ok {
		var fi FlexibleInt
		if json.Unmarshal(v, &fi) == nil {
			b.IsPublic = int(fi)
		}
	}
	return nil
}

// Chapter represents a chapter
type Chapter struct {
	ID        int    `json:"id"`
	BookID    int    `json:"book_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Summary   string `json:"summary"`
	OrderNum  int    `json:"order_num"`
	Type      string `json:"type"`
	VolumeID  *int   `json:"volume_id"` // nullable
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Prompt represents a prompt
type Prompt struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	Content       string  `json:"content"`
	Category      string  `json:"category"`
	OrderNum      int     `json:"order_num"`
	Fields        *string `json:"fields"`        // JSON
	Subcategories *string `json:"subcategories"` // JSON
	CardType      string  `json:"card_type"`
	Password      *string `json:"password"`
	CreatorName   *string `json:"creator_name"`
	Version       *string `json:"version"`
	Metadata      *string `json:"metadata"` // JSON（灵犀式扩展信息：适用作品/使用模式/生成参数等）
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// UnmarshalJSON handles flexible fields for Prompt (accepts string/number for order_num,
// arrays/objects for string pointer fields)
func (p *Prompt) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &p.ID)
	}
	if v, ok := raw["name"]; ok {
		json.Unmarshal(v, &p.Name)
	}
	if v, ok := raw["content"]; ok {
		json.Unmarshal(v, &p.Content)
	}
	if v, ok := raw["category"]; ok {
		json.Unmarshal(v, &p.Category)
	}
	if v, ok := raw["card_type"]; ok {
		json.Unmarshal(v, &p.CardType)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &p.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &p.UpdatedAt)
	}
	if v, ok := raw["order_num"]; ok {
		var fi FlexibleInt
		if json.Unmarshal(v, &fi) == nil {
			p.OrderNum = int(fi)
		}
	}
	p.Description = flexibleStringPtr(raw, "description")
	p.Fields = flexibleStringPtr(raw, "fields")
	p.Subcategories = flexibleStringPtr(raw, "subcategories")
	p.Password = flexibleStringPtr(raw, "password")
	p.CreatorName = flexibleStringPtr(raw, "creator_name")
	p.Version = flexibleStringPtr(raw, "version")
	p.Metadata = flexibleStringPtr(raw, "metadata")
	return nil
}

// flexibleStringPtr extracts a *string from a raw JSON map, handling arrays/objects/null
func flexibleStringPtr(raw map[string]json.RawMessage, key string) *string {
	v, ok := raw[key]
	if !ok || string(v) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return &s
	}
	rawStr := string(v)
	return &rawStr
}

// Memo represents a memo
type Memo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Category  string `json:"category"`
	OrderNum  int    `json:"order_num"`
	Tags      string `json:"tags"`
	IsPinned  int    `json:"is_pinned"`
	WordCount int    `json:"word_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ExperienceShare represents an experience share
type ExperienceShare struct {
	ID                int     `json:"id"`
	Title             string  `json:"title"`
	Summary           *string `json:"summary"`
	Content           string  `json:"content"`
	ContentRenderMode string  `json:"content_render_mode"`
	CoverURL          *string `json:"cover_url"`
	PdfFileURL        *string `json:"pdf_file_url"`
	PdfFileName       *string `json:"pdf_file_name"`
	PdfFileSize       *int    `json:"pdf_file_size"`
	CreateType        string  `json:"create_type"`
	AuthorID          *int    `json:"author_id"`
	AuthorName        string  `json:"author_name"`
	Status            string  `json:"status"`
	PdfParseStatus    *string `json:"pdf_parse_status"`
	PdfParseResult    *string `json:"pdf_parse_result"`
	SourceFileName    *string `json:"source_file_name"`
	Version           *string `json:"version"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// UnmarshalJSON handles flexible fields for ExperienceShare (accepts arrays/objects for string pointer fields)
func (e *ExperienceShare) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &e.ID)
	}
	if v, ok := raw["title"]; ok {
		json.Unmarshal(v, &e.Title)
	}
	if v, ok := raw["content"]; ok {
		json.Unmarshal(v, &e.Content)
	}
	if v, ok := raw["content_render_mode"]; ok {
		json.Unmarshal(v, &e.ContentRenderMode)
	}
	if v, ok := raw["create_type"]; ok {
		json.Unmarshal(v, &e.CreateType)
	}
	if v, ok := raw["author_name"]; ok {
		json.Unmarshal(v, &e.AuthorName)
	}
	if v, ok := raw["status"]; ok {
		json.Unmarshal(v, &e.Status)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &e.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &e.UpdatedAt)
	}
	if v, ok := raw["pdf_file_size"]; ok {
		json.Unmarshal(v, &e.PdfFileSize)
	}
	if v, ok := raw["author_id"]; ok {
		json.Unmarshal(v, &e.AuthorID)
	}
	e.Summary = flexibleStringPtr(raw, "summary")
	e.CoverURL = flexibleStringPtr(raw, "cover_url")
	e.PdfFileURL = flexibleStringPtr(raw, "pdf_file_url")
	e.PdfFileName = flexibleStringPtr(raw, "pdf_file_name")
	e.PdfParseStatus = flexibleStringPtr(raw, "pdf_parse_status")
	e.PdfParseResult = flexibleStringPtr(raw, "pdf_parse_result")
	e.SourceFileName = flexibleStringPtr(raw, "source_file_name")
	e.Version = flexibleStringPtr(raw, "version")
	return nil
}

// Volume represents a volume
type Volume struct {
	ID        int    `json:"id"`
	BookID    int    `json:"book_id"`
	ParentID  *int   `json:"parent_id"`
	Title     string `json:"title"`
	OrderNum  int    `json:"order_num"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// APIProvider represents an API provider
type APIProvider struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	ProviderType string  `json:"provider_type"`
	APIKey       *string `json:"api_key"`
	APIUrl       string  `json:"api_url"`
	APIFormat    string  `json:"api_format"`
	UseFullURL   int     `json:"use_full_url"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// UnmarshalJSON handles flexible fields for APIProvider (accepts arrays/objects for string pointer fields)
func (a *APIProvider) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &a.ID)
	}
	if v, ok := raw["name"]; ok {
		json.Unmarshal(v, &a.Name)
	}
	if v, ok := raw["provider_type"]; ok {
		json.Unmarshal(v, &a.ProviderType)
	}
	if v, ok := raw["api_url"]; ok {
		json.Unmarshal(v, &a.APIUrl)
	}
	if v, ok := raw["api_format"]; ok {
		json.Unmarshal(v, &a.APIFormat)
	}
	if v, ok := raw["use_full_url"]; ok {
		json.Unmarshal(v, &a.UseFullURL)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &a.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &a.UpdatedAt)
	}
	a.APIKey = flexibleStringPtr(raw, "api_key")
	return nil
}

// APIModel represents an API model config
type APIModel struct {
	ID               int     `json:"id"`
	ProviderID       int     `json:"provider_id"`
	Name             string  `json:"name"`
	Model            string  `json:"model"`
	Temperature      float64 `json:"temperature"`
	MaxTokens        int     `json:"max_tokens"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	Description      string  `json:"description"`
	IsDefault        int     `json:"is_default"`
	Enabled          int     `json:"enabled"`
	SortOrder        int     `json:"sort_order"`
	Category         string  `json:"category"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// APIModelWithProvider is APIModel joined with provider info (for AI requests)
type APIModelWithProvider struct {
	APIModel
	APIKey       *string `json:"api_key"`
	APIUrl       string  `json:"api_url"`
	APIFormat    string  `json:"api_format"`
	UseFullURL   int     `json:"use_full_url"`
	ProviderName string  `json:"provider_name"`
}

// UnmarshalJSON handles flexible fields for APIModelWithProvider (accepts arrays/objects for string pointer fields)
func (a *APIModelWithProvider) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// Embedded APIModel fields
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &a.APIModel.ID)
	}
	if v, ok := raw["provider_id"]; ok {
		json.Unmarshal(v, &a.APIModel.ProviderID)
	}
	if v, ok := raw["name"]; ok {
		json.Unmarshal(v, &a.APIModel.Name)
	}
	if v, ok := raw["model"]; ok {
		json.Unmarshal(v, &a.APIModel.Model)
	}
	if v, ok := raw["temperature"]; ok {
		json.Unmarshal(v, &a.APIModel.Temperature)
	}
	if v, ok := raw["max_tokens"]; ok {
		json.Unmarshal(v, &a.APIModel.MaxTokens)
	}
	if v, ok := raw["top_p"]; ok {
		json.Unmarshal(v, &a.APIModel.TopP)
	}
	if v, ok := raw["frequency_penalty"]; ok {
		json.Unmarshal(v, &a.APIModel.FrequencyPenalty)
	}
	if v, ok := raw["description"]; ok {
		json.Unmarshal(v, &a.APIModel.Description)
	}
	if v, ok := raw["is_default"]; ok {
		json.Unmarshal(v, &a.APIModel.IsDefault)
	}
	if v, ok := raw["enabled"]; ok {
		json.Unmarshal(v, &a.APIModel.Enabled)
	}
	if v, ok := raw["sort_order"]; ok {
		json.Unmarshal(v, &a.APIModel.SortOrder)
	}
	if v, ok := raw["category"]; ok {
		json.Unmarshal(v, &a.APIModel.Category)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &a.APIModel.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &a.APIModel.UpdatedAt)
	}
	// Own fields
	if v, ok := raw["api_url"]; ok {
		json.Unmarshal(v, &a.APIUrl)
	}
	if v, ok := raw["api_format"]; ok {
		json.Unmarshal(v, &a.APIFormat)
	}
	if v, ok := raw["use_full_url"]; ok {
		json.Unmarshal(v, &a.UseFullURL)
	}
	if v, ok := raw["provider_name"]; ok {
		json.Unmarshal(v, &a.ProviderName)
	}
	a.APIKey = flexibleStringPtr(raw, "api_key")
	return nil
}

// Conversation represents a conversation
type Conversation struct {
	ID        int    `json:"id"`
	BookID    int    `json:"book_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ConversationMessage represents a conversation message
type ConversationMessage struct {
	ID             int     `json:"id"`
	ConversationID int     `json:"conversation_id"`
	Role           string  `json:"role"`
	Content        string  `json:"content"`
	DisplayContent *string `json:"display_content"`
	CreatedAt      string  `json:"created_at"`
}

// UnmarshalJSON handles flexible fields for ConversationMessage (accepts arrays/objects for string pointer fields)
func (c *ConversationMessage) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &c.ID)
	}
	if v, ok := raw["conversation_id"]; ok {
		json.Unmarshal(v, &c.ConversationID)
	}
	if v, ok := raw["role"]; ok {
		json.Unmarshal(v, &c.Role)
	}
	if v, ok := raw["content"]; ok {
		json.Unmarshal(v, &c.Content)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &c.CreatedAt)
	}
	c.DisplayContent = flexibleStringPtr(raw, "display_content")
	return nil
}

// Generator represents a generator
type Generator struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        string  `json:"icon"`
	CorePrompt  *string `json:"core_prompt"`
	Remark      *string `json:"remark"`
	OrderNum    int     `json:"order_num"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// UnmarshalJSON handles flexible fields for Generator (accepts arrays/objects for string pointer fields)
func (g *Generator) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &g.ID)
	}
	if v, ok := raw["name"]; ok {
		json.Unmarshal(v, &g.Name)
	}
	if v, ok := raw["icon"]; ok {
		json.Unmarshal(v, &g.Icon)
	}
	if v, ok := raw["order_num"]; ok {
		json.Unmarshal(v, &g.OrderNum)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &g.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &g.UpdatedAt)
	}
	g.Description = flexibleStringPtr(raw, "description")
	g.CorePrompt = flexibleStringPtr(raw, "core_prompt")
	g.Remark = flexibleStringPtr(raw, "remark")
	return nil
}

// Character represents a character
type Character struct {
	ID          int     `json:"id"`
	BookID      int     `json:"book_id"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Personality *string `json:"personality"`
	Info        *string `json:"info"`
	Folder      string  `json:"folder"`
	Folders     string  `json:"folders"` // stored as JSON array string
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// UnmarshalJSON handles flexible fields for Character (accepts both string and array for folders, arrays/objects for string pointer fields)
func (c *Character) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &c.ID)
	}
	if v, ok := raw["book_id"]; ok {
		json.Unmarshal(v, &c.BookID)
	}
	if v, ok := raw["name"]; ok {
		json.Unmarshal(v, &c.Name)
	}
	if v, ok := raw["gender"]; ok {
		json.Unmarshal(v, &c.Gender)
	}
	if v, ok := raw["folder"]; ok {
		json.Unmarshal(v, &c.Folder)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &c.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &c.UpdatedAt)
	}
	// Folders: accept string or array
	if v, ok := raw["folders"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			c.Folders = s
		} else {
			c.Folders = string(v)
		}
	}
	c.Personality = flexibleStringPtr(raw, "personality")
	c.Info = flexibleStringPtr(raw, "info")
	return nil
}

// UsageStat represents usage statistics
type UsageStat struct {
	ID           int     `json:"id"`
	Date         string  `json:"date"`
	ModelID      *int    `json:"model_id"`
	ModelName    string  `json:"model_name"`
	ProviderName *string `json:"provider_name"`
	UsageCount   int     `json:"usage_count"`
	TotalTokens  int     `json:"total_tokens"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// UnmarshalJSON handles flexible fields for UsageStat (accepts arrays/objects for string pointer fields)
func (u *UsageStat) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &u.ID)
	}
	if v, ok := raw["date"]; ok {
		json.Unmarshal(v, &u.Date)
	}
	if v, ok := raw["model_id"]; ok {
		json.Unmarshal(v, &u.ModelID)
	}
	if v, ok := raw["model_name"]; ok {
		json.Unmarshal(v, &u.ModelName)
	}
	if v, ok := raw["usage_count"]; ok {
		json.Unmarshal(v, &u.UsageCount)
	}
	if v, ok := raw["total_tokens"]; ok {
		json.Unmarshal(v, &u.TotalTokens)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &u.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &u.UpdatedAt)
	}
	u.ProviderName = flexibleStringPtr(raw, "provider_name")
	return nil
}

// GraphEntity represents a knowledge graph entity
type GraphEntity struct {
	ID          int      `json:"id"`
	BookID      int      `json:"book_id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Metadata    JSONText `json:"metadata"`
	VersionID   *int     `json:"version_id"`
	// PrimaryID 是实体合并系统的主/副词条标记：
	//   nil  = 主词条（画布上真正渲染的节点）
	//   非空 = 副词条，值指向同版本内主词条的 id；读接口会把它的关系端点重定向到主词条
	PrimaryID *int `json:"primary_id"`
	// ChapterIDs 只用于导入请求：声明该实体出自哪些章节，落库到 graph_chapter_refs。
	// 读接口的 SELECT 不填它，因此查询响应里不会出现。
	ChapterIDs []int  `json:"chapter_ids,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// GraphRelation represents a knowledge graph relation
type GraphRelation struct {
	ID           int     `json:"id"`
	BookID       int     `json:"book_id"`
	SourceID     int     `json:"source_id"`
	TargetID     int     `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Description  string  `json:"description"`
	Evidence     string  `json:"evidence"`
	Confidence   float64 `json:"confidence"`
	VersionID    *int    `json:"version_id"`
	// ChapterIDs 同 GraphEntity.ChapterIDs：只用于导入请求
	ChapterIDs []int  `json:"chapter_ids,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// GraphVersion represents a knowledge graph version
type GraphVersion struct {
	ID            int    `json:"id"`
	BookID        int    `json:"book_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Scope         string `json:"scope"`
	ChapterCount  int    `json:"chapter_count"`
	EntityCount   int    `json:"entity_count"`
	RelationCount int    `json:"relation_count"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// GraphChapterRef 是按章节分组的图谱关联（服务端聚合后的响应体）
type GraphChapterRef struct {
	ChapterID   int   `json:"chapter_id"`
	EntityIDs   []int `json:"entity_ids"`
	RelationIDs []int `json:"relation_ids"`
}

// EntityMergeCandidate 是「疑似重复实体」候选对，供后续的人工合并检测使用。
// 服务端只做启发式发现并给出可解释的理由，最终是否合并由人决定。
type EntityMergeCandidate struct {
	A       GraphEntity `json:"a"`
	B       GraphEntity `json:"b"`
	Score   float64     `json:"score"`
	Reasons []string    `json:"reasons"`
	// SuggestPrimaryID / SuggestDuplicateID 是服务端建议的合并方向（把 Duplicate 并入 Primary）
	SuggestPrimaryID   int `json:"suggest_primary_id"`
	SuggestDuplicateID int `json:"suggest_duplicate_id"`
}

// KnowledgeGraphData is the full graph response
type KnowledgeGraphData struct {
	Entities  []GraphEntity   `json:"entities"`
	Relations []GraphRelation `json:"relations"`
}

// EntryCategory represents an entry category
type EntryCategory struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Entry represents an entry (world book item)
type Entry struct {
	ID            int     `json:"id"`
	BookID        *int    `json:"book_id"`
	CategoryID    int     `json:"category_id"`
	CategoryType  string  `json:"category_type"`
	Name          string  `json:"name"`
	Avatar        *string `json:"avatar"`
	Tags          string  `json:"tags"`
	Appearance    string  `json:"appearance"`
	Background    string  `json:"background"`
	Personality   *string `json:"personality"`
	Relationships *string `json:"relationships"`
	CustomFields  string  `json:"custom_fields"`
	Description   string  `json:"description"`
	IsPublic      int     `json:"is_public"`
	CreatorName   *string `json:"creator_name"`
	CreatorAvatar *string `json:"creator_avatar"`
	UseCount      int     `json:"use_count"`
	LikeCount     int     `json:"like_count"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// UnmarshalJSON handles flexible fields for Entry (accepts arrays/objects for string pointer fields)
func (e *Entry) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"]; ok {
		json.Unmarshal(v, &e.ID)
	}
	if v, ok := raw["book_id"]; ok {
		json.Unmarshal(v, &e.BookID)
	}
	if v, ok := raw["category_id"]; ok {
		json.Unmarshal(v, &e.CategoryID)
	}
	if v, ok := raw["category_type"]; ok {
		json.Unmarshal(v, &e.CategoryType)
	}
	if v, ok := raw["name"]; ok {
		json.Unmarshal(v, &e.Name)
	}
	if v, ok := raw["tags"]; ok {
		json.Unmarshal(v, &e.Tags)
	}
	if v, ok := raw["appearance"]; ok {
		json.Unmarshal(v, &e.Appearance)
	}
	if v, ok := raw["background"]; ok {
		json.Unmarshal(v, &e.Background)
	}
	if v, ok := raw["custom_fields"]; ok {
		json.Unmarshal(v, &e.CustomFields)
	}
	if v, ok := raw["description"]; ok {
		json.Unmarshal(v, &e.Description)
	}
	if v, ok := raw["is_public"]; ok {
		var fi FlexibleInt
		if json.Unmarshal(v, &fi) == nil {
			e.IsPublic = int(fi)
		}
	}
	if v, ok := raw["use_count"]; ok {
		json.Unmarshal(v, &e.UseCount)
	}
	if v, ok := raw["like_count"]; ok {
		json.Unmarshal(v, &e.LikeCount)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &e.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &e.UpdatedAt)
	}
	e.Avatar = flexibleStringPtr(raw, "avatar")
	e.Personality = flexibleStringPtr(raw, "personality")
	e.Relationships = flexibleStringPtr(raw, "relationships")
	e.CreatorName = flexibleStringPtr(raw, "creator_name")
	e.CreatorAvatar = flexibleStringPtr(raw, "creator_avatar")
	return nil
}

// ChatMessage for AI chat requests
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RelatedContent for AI chat requests
type RelatedContent struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// AIChatRequest is the request body for /api/ai/chat
type AIChatRequest struct {
	Messages       []ChatMessage    `json:"messages"`
	ConfigID       *int             `json:"configId"`
	SystemPrompts  []string         `json:"systemPrompts"`
	RelatedContent []RelatedContent `json:"relatedContent"`
	// MaxTokens 覆盖模型配置里的 max_tokens（>0 时生效），用于 AI 图谱分析这类需要长输出的场景
	MaxTokens *int `json:"maxTokens"`
}

// Workflow represents a saved workflow (工作流工作台条目)
// 字段与 workflows 表一一对应（见 internal/database/db.go 的建表语句）。
//   - scene 存的是整张画布的 JSON 快照（前端 WorkflowScene），用 JSONText 直接往返对象，
//     不能再当普通字符串——前端 createWorkflow/saveWorkflowScene 传的都是对象，按 string 解会 400。
//   - createdAt/updatedAt 用驼峰：前端 web/src/utils/workflow.ts 是按驼峰读取的。
type Workflow struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scene       JSONText `json:"scene"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// WorkflowNodeLibraryItem 是工作流节点模板库条目（表 workflow_node_library）。
// 对外暴露的 id 是 template_id（前端 delete 也按它删），自增主键只在库内使用；
// template 是节点模板的 JSON 快照。
type WorkflowNodeLibraryItem struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Variant     string   `json:"variant"`
	Category    string   `json:"category"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Outputs     string   `json:"outputs"`
	Generated   bool     `json:"generated"`
	Template    JSONText `json:"template"`
}
