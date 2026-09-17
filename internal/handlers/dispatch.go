package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

// NotFound returns 404
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	respondFail(w, 404, "未找到")
}

// DispatchChapters handles /api/chapters/* routing manually
// Routes: GET book/{bookId}, POST import-book, POST import-file, GET/PUT/DELETE {id}
func (h *Handler) DispatchChapters(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/chapters/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) >= 2 && parts[0] == "book" {
		bookId, err := strconv.Atoi(parts[1])
		if err != nil {
			respondFail(w, 400, "无效的书籍ID")
			return
		}
		r = setPathValue(r, "bookId", parts[1])
		h.GetChaptersByBook(w, r)
		_ = bookId
		return
	}
	if len(parts) >= 1 && parts[0] == "import-book" && r.Method == "POST" {
		h.ImportChapters(w, r)
		return
	}
	if len(parts) >= 1 && parts[0] == "import-file" && r.Method == "POST" {
		h.ImportFile(w, r)
		return
	}
	if len(parts) >= 1 {
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			respondFail(w, 400, "无效的ID")
			return
		}
		r = setPathValue(r, "id", parts[0])
		_ = id
		switch r.Method {
		case "GET":
			h.GetChapter(w, r)
		case "PUT":
			h.UpdateChapter(w, r)
		case "DELETE":
			h.DeleteChapter(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	respondFail(w, 404, "未找到")
}

// DispatchMemos handles /api/memos/* routing manually
// Routes: GET search, POST batch4batch, GET/PUT/DELETE {id}
func (h *Handler) DispatchMemos(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/memos/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) >= 1 && parts[0] == "search" && r.Method == "GET" {
		h.SearchMemos(w, r)
		return
	}
	if len(parts) >= 1 && parts[0] == "batch4batch" && r.Method == "POST" {
		h.BatchMemos(w, r)
		return
	}
	if len(parts) >= 1 {
		r = setPathValue(r, "id", parts[0])
		switch r.Method {
		case "GET":
			h.GetMemo(w, r)
		case "PUT":
			h.UpdateMemo(w, r)
		case "DELETE":
			h.DeleteMemo(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	respondFail(w, 404, "未找到")
}

// DispatchConfig handles /api/config/* routing manually
// Routes: GET default, GET provider/{providerId}, GET/PUT/DELETE {id}, PUT {id}/toggle, PUT reorder/all
func (h *Handler) DispatchConfig(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/config/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) >= 1 && parts[0] == "default" && r.Method == "GET" {
		h.GetDefaultConfig(w, r)
		return
	}
	if len(parts) >= 2 && parts[0] == "provider" {
		r = setPathValue(r, "providerId", parts[1])
		h.GetConfigsByProvider(w, r)
		return
	}
	if len(parts) >= 2 && parts[0] == "reorder" && parts[1] == "all" && r.Method == "PUT" {
		h.ReorderModels(w, r)
		return
	}
	if len(parts) >= 2 && parts[1] == "toggle" && r.Method == "PUT" {
		r = setPathValue(r, "id", parts[0])
		h.ToggleModel(w, r)
		return
	}
	if len(parts) >= 1 {
		r = setPathValue(r, "id", parts[0])
		switch r.Method {
		case "GET":
			h.GetAllConfigs(w, r)
		case "PUT":
			h.UpdateConfig(w, r)
		case "DELETE":
			h.DeleteConfig(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	respondFail(w, 404, "未找到")
}

// DispatchConversations handles /api/conversations/* routing manually
// Routes: GET book/{bookId}, GET/PUT/DELETE {id},
//
//	GET {id}/messages, POST {id}/messages, DELETE {id}/messages,
//	DELETE {id}/messages/{messageId}, PUT {id}/messages/{messageId}
func (h *Handler) DispatchConversations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/conversations/")
	parts := strings.SplitN(path, "/", 4)

	if len(parts) >= 2 && parts[0] == "book" {
		r = setPathValue(r, "bookId", parts[1])
		h.GetConversationsByBook(w, r)
		return
	}
	if len(parts) >= 2 && parts[1] == "messages" {
		r = setPathValue(r, "id", parts[0])
		if len(parts) >= 3 && parts[2] != "" {
			// {id}/messages/{messageId}
			r = setPathValue(r, "conversationId", parts[0])
			r = setPathValue(r, "messageId", parts[2])
			switch r.Method {
			case "DELETE":
				h.DeleteMessage(w, r)
			case "PUT":
				h.UpdateMessage(w, r)
			default:
				respondFail(w, 405, "方法不允许")
			}
			return
		}
		// {id}/messages
		switch r.Method {
		case "GET":
			h.GetMessages(w, r)
		case "POST":
			h.SaveMessage(w, r)
		case "DELETE":
			h.ClearMessages(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	if len(parts) >= 1 && parts[0] != "" {
		r = setPathValue(r, "id", parts[0])
		switch r.Method {
		case "GET":
			h.GetConversation(w, r)
		case "PUT":
			h.UpdateConversation(w, r)
		case "DELETE":
			h.DeleteConversation(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	respondFail(w, 404, "未找到")
}

// DispatchCharacters handles /api/characters/* routing manually
// Routes: GET book/{bookId}, GET folders/{bookId}, GET/PUT/DELETE {id},
//
//	POST batch-delete, POST batch-move
func (h *Handler) DispatchCharacters(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/characters/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) >= 2 && parts[0] == "book" {
		r = setPathValue(r, "bookId", parts[1])
		h.GetCharactersByBook(w, r)
		return
	}
	if len(parts) >= 2 && parts[0] == "folders" {
		r = setPathValue(r, "bookId", parts[1])
		h.GetFolders(w, r)
		return
	}
	if len(parts) >= 1 && parts[0] == "batch-delete" && r.Method == "POST" {
		h.BatchDeleteCharacters(w, r)
		return
	}
	if len(parts) >= 1 && parts[0] == "batch-move" && r.Method == "POST" {
		h.BatchMoveCharacters(w, r)
		return
	}
	if len(parts) >= 1 {
		r = setPathValue(r, "id", parts[0])
		switch r.Method {
		case "GET":
			h.GetCharacter(w, r)
		case "PUT":
			h.UpdateCharacter(w, r)
		case "DELETE":
			h.DeleteCharacter(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	respondFail(w, 404, "未找到")
}

// DispatchVolumes handles /api/volumes/* routing manually
// Routes: GET book/{bookId}, GET {id}
func (h *Handler) DispatchVolumes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/volumes/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) >= 2 && parts[0] == "book" {
		r = setPathValue(r, "bookId", parts[1])
		h.GetVolumesByBook(w, r)
		return
	}
	if len(parts) >= 1 && parts[0] != "" {
		r = setPathValue(r, "id", parts[0])
		switch r.Method {
		case "GET":
			h.GetVolume(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	respondFail(w, 404, "未找到")
}

// DispatchKgEntities handles /api/knowledge-graph/entities/* routing manually
// Routes: GET book/{bookId}, GET {id}
func (h *Handler) DispatchKgEntities(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/knowledge-graph/entities/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) >= 2 && parts[0] == "book" {
		r = setPathValue(r, "bookId", parts[1])
		h.GetEntities(w, r)
		return
	}
	if len(parts) >= 1 && parts[0] != "" {
		r = setPathValue(r, "id", parts[0])
		h.GetEntity(w, r)
		return
	}
	respondFail(w, 404, "未找到")
}

// DispatchEntries handles /api/entries/* routing manually
// Routes: GET categories, GET/PUT/DELETE {id}, POST {id}/add-to-book, POST {id}/ai-auto-complete
func (h *Handler) DispatchEntries(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/entries/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) >= 1 && parts[0] == "categories" && r.Method == "GET" {
		h.GetCategories(w, r)
		return
	}
	if len(parts) >= 2 && parts[1] == "add-to-book" && r.Method == "POST" {
		r = setPathValue(r, "id", parts[0])
		h.AddToBook(w, r)
		return
	}
	if len(parts) >= 2 && parts[1] == "ai-auto-complete" && r.Method == "POST" {
		r = setPathValue(r, "id", parts[0])
		h.AIAutoComplete(w, r)
		return
	}
	if len(parts) >= 1 {
		r = setPathValue(r, "id", parts[0])
		switch r.Method {
		case "GET":
			h.GetEntry(w, r)
		case "PUT":
			h.UpdateEntry(w, r)
		case "DELETE":
			h.DeleteEntry(w, r)
		default:
			respondFail(w, 405, "方法不允许")
		}
		return
	}
	respondFail(w, 404, "未找到")
}
