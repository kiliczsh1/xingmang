package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// Handler holds the database connection and provides handler methods.
type Handler struct {
	db *sql.DB
}

// NewHandler creates a new Handler with the given database connection.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
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
