package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/xingmang/xingmang-desktop/internal/agent"
	"github.com/xingmang/xingmang-desktop/internal/aitokens"
	"github.com/xingmang/xingmang-desktop/internal/models"
	"github.com/xingmang/xingmang-desktop/internal/sharestore"
)

// Handler holds the database connection and provides handler methods.
type Handler struct {
	db    *sql.DB
	agent *agent.Manager
	// shareStore 经验分享的文件存储。帖子采用「数据库 + 文件」双读：
	// 旧帖继续留在库里照常读写，新建的帖子一律落到 data/experience-shares/，不占数据库空间。
	shareStore *sharestore.Store
}

// NewHandler creates a new Handler with the given database connection.
//
// 这里顺带初始化自适应 max_tokens 的探测缓存（internal/aitokens）——
// 它同时被对话链路与 Agent 链路使用，用包级单例避免把 *sql.DB 穿透到每个构造函数。
func NewHandler(db *sql.DB) *Handler {
	aitokens.Init(db)
	return &Handler{db: db, agent: agent.NewManager(db)}
}

// NewWithShareStore 带文件存储的构造（server 启动时用；存储初始化失败也不该让整个服务起不来）。
func NewWithShareStore(db *sql.DB, store *sharestore.Store) *Handler {
	h := NewHandler(db)
	h.shareStore = store
	return h
}

// New is an alias for NewHandler.
func New(db *sql.DB) *Handler {
	return NewHandler(db)
}

// respondOK writes a success JSON response.
func respondOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: data})
}

// respondFail writes an error JSON response with the given status code.
func respondFail(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{Success: false, Message: message})
}

// readAllBody 一次性读出请求体。
// 多处 PUT 需要先看「原始请求体里出现了哪些键」，再决定拼哪些 SET 子句
// （整行覆盖会把调用方没传的可选字段静默清零）。
func readAllBody(r *http.Request) ([]byte, error) {
	return io.ReadAll(r.Body)
}

// touchBook 更新书的 updated_at，让「最近更新」排序与展示反映章节编辑。
func (h *Handler) touchBook(bookID int) {
	if bookID <= 0 {
		return
	}
	h.db.Exec("UPDATE books SET updated_at=CURRENT_TIMESTAMP WHERE id=?", bookID)
}

// pathParamKey is the context key type for path parameters set by dispatch handlers.
type pathParamKey string

// setPathValue stores a path parameter value in the request context.
func setPathValue(r *http.Request, key, value string) *http.Request {
	ctx := context.WithValue(r.Context(), pathParamKey(key), value)
	return r.WithContext(ctx)
}

// idParam extracts an integer ID from the URL path parameter.
// It first tries r.PathValue (Go 1.22 ServeMux patterns), then falls back
// to context values set by dispatch handlers.
func idParam(r *http.Request, key string) (int, bool) {
	val := r.PathValue(key)
	if val == "" {
		// Fallback: try context value (set by dispatch handlers)
		if cv := r.Context().Value(pathParamKey(key)); cv != nil {
			if s, ok := cv.(string); ok {
				val = s
			}
		}
	}
	if val == "" {
		return 0, false
	}
	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return id, true
}

// queryInt returns a query parameter as int, or the default value.
func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
