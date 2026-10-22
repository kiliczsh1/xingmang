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
	// FolderPath 是书目录相对 books_root 的路径（作品文件夹化后的唯一定位）。
	FolderPath string `json:"folder_path"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
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
	ID       int    `json:"id"`
	BookID   int    `json:"book_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Summary  string `json:"summary"`
	OrderNum int    `json:"order_num"`
	Type     string `json:"type"`
	VolumeID *int   `json:"volume_id"` // nullable
	// FilePath 是章节文件相对书目录的路径；空表示该行还没有落盘文件。
	FilePath  string `json:"file_path"`
	Missing   int    `json:"missing,omitempty"`
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
	IsLocked      int     `json:"is_locked"` // 锁定标记：0=未锁定，1=已锁定
	Insight       *string `json:"insight"`   // 使用心得（卡片上的自由笔记）
	CreatorName   *string `json:"creator_name"`
	Version       *string `json:"version"`
	Metadata      *string `json:"metadata"` // JSON（灵犀式扩展信息：适用作品/使用模式/生成参数等）
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	// 以下两个是只读聚合字段（来自 prompt_insights 子查询），供卡片徽标免额外请求渲染；
	// 请求体里带这两个键不会写库。
	InsightCount int     `json:"insight_count"`
	InsightAvg   float64 `json:"insight_avg"`
}

// PromptInsight 是提示词卡片上的一条使用心得。一张卡片可以有多条，按时间倒序构成时间线。
type PromptInsight struct {
	ID        int     `json:"id"`
	PromptID  int     `json:"prompt_id"`
	Content   string  `json:"content"`
	Rating    int     `json:"rating"` // 0=未评分，1-5
	Tags      *string `json:"tags"`   // JSON 数组字符串，前端在 API 层解析
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
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
	// is_locked 兼容 bool（true/false）与数字（0/1）两种写法
	if v, ok := raw["is_locked"]; ok {
		var locked bool
		if json.Unmarshal(v, &locked) == nil {
			if locked {
				p.IsLocked = 1
			}
		} else {
			var fi FlexibleInt
			if json.Unmarshal(v, &fi) == nil {
				p.IsLocked = int(fi)
			}
		}
	}
	p.Description = flexibleStringPtr(raw, "description")
	p.Fields = flexibleStringPtr(raw, "fields")
	p.Subcategories = flexibleStringPtr(raw, "subcategories")
	p.Password = flexibleStringPtr(raw, "password")
	p.Insight = flexibleStringPtr(raw, "insight")
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
	// 所属吧：tutorial = 教程吧，experience = 经验分享吧（缺省）。
	// ⚠ 老版本导出文件里没有这个字段，读出空串由 handlers.normalizeBoard 兜成 experience。
	Board string `json:"board"`
	// 贴吧式分区：NULL = 未归类；分区被删除时回落为 NULL（不加外键，存量库是 ALTER 补的列）
	SectionID *int `json:"section_id"`
	// 精华标记（1 = 加精，跨分区生效，与所属分区无关）
	IsFeatured int `json:"is_featured"`
	// 置顶标记（1 = 置顶，每个吧最多 3 条；置顶帖排在列表最前）
	IsPinned int `json:"is_pinned"`
	// 主楼赞（楼层赞仍存在浏览器 localStorage 的 bbs-replies-<id> 里，两边相加才是帖子总赞）
	LikeCount int    `json:"like_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ExperienceSection 经验分享的贴吧式自定义分区。
// 「精华」与「全部」是内置入口，不落这张表；自定义分区上限 15 个（由 handler 校验，**按吧各自计数**）。
type ExperienceSection struct {
	ID int `json:"id"`
	// 所属吧，与 ExperienceShare.Board 同口径；分区按吧隔离，两个吧可以有同名分区
	Board     string `json:"board"`
	Name      string `json:"name"`
	OrderNum  int    `json:"order_num"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// 帖子所属吧的合法取值。除 BoardTutorial 之外的任何输入一律当 BoardExperience。
const (
	BoardExperience = "experience"
	BoardTutorial   = "tutorial"
)

// NormalizeBoard 把任意 board 输入收成两个合法值之一。
//
// ⚠ 唯一需要兜底的场景是**用户从旧版本导出的文件**（当时还没有 board 字段，读出来是空串），
// 以及手改过的帖子 JSON。本机存量数据在数据库迁移时已经刷平，不靠这里。
func NormalizeBoard(b string) string {
	if strings.EqualFold(strings.TrimSpace(b), BoardTutorial) {
		return BoardTutorial
	}
	return BoardExperience
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
	// ⚠ 三个新字段必须显式挑出来，否则前端回传的分区/精华/赞会被静默丢掉
	if v, ok := raw["section_id"]; ok {
		json.Unmarshal(v, &e.SectionID)
	}
	if v, ok := raw["is_featured"]; ok {
		json.Unmarshal(v, &e.IsFeatured)
	}
	if v, ok := raw["is_pinned"]; ok {
		json.Unmarshal(v, &e.IsPinned)
	}
	if v, ok := raw["like_count"]; ok {
		json.Unmarshal(v, &e.LikeCount)
	}
	// ⚠ 同理：board 不加进这里 → 前端传了也读不到，帖子会静默落到默认吧
	if v, ok := raw["board"]; ok {
		json.Unmarshal(v, &e.Board)
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
	// 多模态补齐（2026-10-15）：渠道 / 扩展认证 / 并发 / 长尾配置
	AuthType    string  `json:"auth_type"`
	AccessKey   *string `json:"access_key"`
	SecretKey   *string `json:"secret_key"`
	GroupID     *string `json:"group_id"`
	EndpointID  *string `json:"endpoint_id"`
	Concurrency int     `json:"concurrency"`
	ChannelName string  `json:"channel_name"`
	ExtraConfig *string `json:"extra_config"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
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
	if v, ok := raw["auth_type"]; ok {
		json.Unmarshal(v, &a.AuthType)
	}
	if v, ok := raw["channel_name"]; ok {
		json.Unmarshal(v, &a.ChannelName)
	}
	if v, ok := raw["concurrency"]; ok {
		json.Unmarshal(v, &a.Concurrency)
	}
	if v, ok := raw["created_at"]; ok {
		json.Unmarshal(v, &a.CreatedAt)
	}
	if v, ok := raw["updated_at"]; ok {
		json.Unmarshal(v, &a.UpdatedAt)
	}
	a.AccessKey = flexibleStringPtr(raw, "access_key")
	a.SecretKey = flexibleStringPtr(raw, "secret_key")
	a.GroupID = flexibleStringPtr(raw, "group_id")
	a.EndpointID = flexibleStringPtr(raw, "endpoint_id")
	a.ExtraConfig = flexibleStringPtr(raw, "extra_config")
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
	// 多模态补齐（2026-10-15）：强类型 config_type 与高级参数。
	// config_type 取代 category 作为选模型依据；category 保留为兼容字段。
	ConfigType      string   `json:"config_type"`
	Prompt          string   `json:"prompt"`
	PromptRole      string   `json:"prompt_role"`
	TopK            *int     `json:"top_k"`
	PresencePenalty *float64 `json:"presence_penalty"`
	Stop            string   `json:"stop"`
	MaxTokensMode   string   `json:"max_tokens_mode"`
	// MaxTokensSource 记录 MaxTokens 这个值的来源，供界面区分真值与猜测：
	// upstream（上游 /models 明确返回）/ catalog（第三方目录）/ table（本地家族表）
	// / manual（用户手填）/ probed（请求时探测得到）/ ""（未知，存量数据）。
	MaxTokensSource        string  `json:"max_tokens_source"`
	ThinkingMode           int     `json:"thinking_mode"`
	ThinkingEffort         string  `json:"thinking_effort"`
	GenerationConfigSource string  `json:"generation_config_source"`
	RequestRules           *string `json:"request_rules"`
	// 参数组 / 默认参数（JSON 文本）：图片 / 视频模型的可变生成参数
	// （如 duration / resolution / ratio），由「API 广场」预设带出或用户自定义。
	ParamGroups   string   `json:"param_groups"`
	DefaultParams string   `json:"default_params"`
	IsFree        int      `json:"is_free"`
	InputRatio    *float64 `json:"input_ratio"`
	OutputRatio   *float64 `json:"output_ratio"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// APIModelWithProvider is APIModel joined with provider info (for AI requests)
type APIModelWithProvider struct {
	APIModel
	APIKey       *string `json:"api_key"`
	APIUrl       string  `json:"api_url"`
	APIFormat    string  `json:"api_format"`
	UseFullURL   int     `json:"use_full_url"`
	ProviderName string  `json:"provider_name"`
	// provider 侧的渠道 / 扩展认证字段（适配器构造请求时需要）
	AuthType    string  `json:"auth_type"`
	AccessKey   *string `json:"access_key"`
	SecretKey   *string `json:"secret_key"`
	GroupID     *string `json:"group_id"`
	EndpointID  *string `json:"endpoint_id"`
	Concurrency int     `json:"concurrency"`
	ChannelName string  `json:"channel_name"`
	ExtraConfig *string `json:"extra_config"`
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
	if v, ok := raw["config_type"]; ok {
		json.Unmarshal(v, &a.APIModel.ConfigType)
	}
	if v, ok := raw["prompt"]; ok {
		json.Unmarshal(v, &a.APIModel.Prompt)
	}
	if v, ok := raw["prompt_role"]; ok {
		json.Unmarshal(v, &a.APIModel.PromptRole)
	}
	if v, ok := raw["top_k"]; ok {
		json.Unmarshal(v, &a.APIModel.TopK)
	}
	if v, ok := raw["presence_penalty"]; ok {
		json.Unmarshal(v, &a.APIModel.PresencePenalty)
	}
	if v, ok := raw["stop"]; ok {
		json.Unmarshal(v, &a.APIModel.Stop)
	}
	if v, ok := raw["max_tokens_mode"]; ok {
		json.Unmarshal(v, &a.APIModel.MaxTokensMode)
	}
	if v, ok := raw["max_tokens_source"]; ok {
		json.Unmarshal(v, &a.APIModel.MaxTokensSource)
	}
	if v, ok := raw["thinking_mode"]; ok {
		json.Unmarshal(v, &a.APIModel.ThinkingMode)
	}
	if v, ok := raw["thinking_effort"]; ok {
		json.Unmarshal(v, &a.APIModel.ThinkingEffort)
	}
	if v, ok := raw["generation_config_source"]; ok {
		json.Unmarshal(v, &a.APIModel.GenerationConfigSource)
	}
	if v, ok := raw["param_groups"]; ok {
		json.Unmarshal(v, &a.APIModel.ParamGroups)
	}
	if v, ok := raw["default_params"]; ok {
		json.Unmarshal(v, &a.APIModel.DefaultParams)
	}
	if v, ok := raw["is_free"]; ok {
		json.Unmarshal(v, &a.APIModel.IsFree)
	}
	if v, ok := raw["input_ratio"]; ok {
		json.Unmarshal(v, &a.APIModel.InputRatio)
	}
	if v, ok := raw["output_ratio"]; ok {
		json.Unmarshal(v, &a.APIModel.OutputRatio)
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
	if v, ok := raw["auth_type"]; ok {
		json.Unmarshal(v, &a.AuthType)
	}
	if v, ok := raw["channel_name"]; ok {
		json.Unmarshal(v, &a.ChannelName)
	}
	if v, ok := raw["concurrency"]; ok {
		json.Unmarshal(v, &a.Concurrency)
	}
	a.AccessKey = flexibleStringPtr(raw, "access_key")
	a.SecretKey = flexibleStringPtr(raw, "secret_key")
	a.GroupID = flexibleStringPtr(raw, "group_id")
	a.EndpointID = flexibleStringPtr(raw, "endpoint_id")
	a.ExtraConfig = flexibleStringPtr(raw, "extra_config")
	a.APIModel.RequestRules = flexibleStringPtr(raw, "request_rules")
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
	// 以下三个是记忆落盘后的索引字段：正文在 memory/chats/*.md 里，DB 只记路径与状态。
	FilePath   string `json:"file_path,omitempty"`
	Missing    int    `json:"missing,omitempty"`
	ParseError string `json:"parse_error,omitempty"`
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
	if c.DisplayContent == nil {
		// 前端历史上发的是 displayContent（驼峰），只认下划线会让它被静默丢掉，
		// 表现为"重新载入对话后，用户消息变成注入提示词后的原文"。
		c.DisplayContent = flexibleStringPtr(raw, "displayContent")
	}
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

// ===== 工作流画布 =====

// Workflow is a node-graph workflow. Graph / Config are raw JSON strings:
// 执行完全在前端，Go 侧不解析节点语义。
type Workflow struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Category 分类名，空串表示「默认分类」（前端按此分组过滤）
	Category  string `json:"category"`
	Graph     string `json:"graph"`
	Config    string `json:"config"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	RunCount  int    `json:"run_count,omitempty"`
}

// WorkflowRun is one execution of a workflow.
type WorkflowRun struct {
	ID         int    `json:"id"`
	WorkflowID int    `json:"workflow_id"`
	Status     string `json:"status"`
	Logs       string `json:"logs"`
	Context    string `json:"context"`
	Error      string `json:"error"`
	OutputDir  string `json:"output_dir"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
}

// WorkflowOutput indexes one file written under workflow_outputs/<runId>/.
// RelPath is always slash-separated and relative to the workflow output root.
type WorkflowOutput struct {
	ID         int    `json:"id"`
	RunID      int    `json:"run_id"`
	WorkflowID int    `json:"workflow_id"`
	Kind       string `json:"kind"`
	RelPath    string `json:"rel_path"`
	Size       int    `json:"size"`
	CreatedAt  string `json:"created_at"`
}

// BookDoc 是书内自由文档（大纲 / 灵感）。正文在 outlines/ 与 ideas/ 的 .md 里，
// 这张表只是索引；kind = outline | idea。
type BookDoc struct {
	ID        int    `json:"id"`
	BookID    int    `json:"book_id"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	FilePath  string `json:"file_path"`
	Content   string `json:"content,omitempty"`
	WordCount int    `json:"word_count"`
	Pinned    int    `json:"pinned"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
