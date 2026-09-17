package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetConversations returns all conversations, optionally filtered by book_id.
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("book_id")

	var rows *sql.Rows
	var err error

	if bookID != "" {
		rows, err = h.db.Query("SELECT id, book_id, title, created_at, updated_at FROM conversations WHERE book_id = ? ORDER BY updated_at DESC", bookID)
	} else {
		rows, err = h.db.Query("SELECT id, book_id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC")
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Conversation
	for rows.Next() {
		var c models.Conversation
		if err := rows.Scan(&c.ID, &c.BookID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, c)
	}
	respondOK(w, list)
}

// GetConversation returns a single conversation by id.
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var c models.Conversation
	err := h.db.QueryRow("SELECT id, book_id, title, created_at, updated_at FROM conversations WHERE id = ?", id).
		Scan(&c.ID, &c.BookID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "对话不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, c)
}

// CreateConversation creates a new conversation.
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var c models.Conversation
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec("INSERT INTO conversations (book_id, title) VALUES (?, ?)", c.BookID, c.Title)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	c.ID = int(id)
	respondOK(w, c)
}

// UpdateConversation updates a conversation's title.
func (h *Handler) UpdateConversation(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var c models.Conversation
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE conversations SET title=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", c.Title, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	c.ID = id
	respondOK(w, c)
}

// DeleteConversation deletes a conversation by id.
func (h *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// GetMessages returns all messages for a conversation.
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的对话ID")
		return
	}

	rows, err := h.db.Query("SELECT id, conversation_id, role, content, display_content, created_at FROM conversation_messages WHERE conversation_id = ? ORDER BY created_at ASC", conversationID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.ConversationMessage
	for rows.Next() {
		var m models.ConversationMessage
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.DisplayContent, &m.CreatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// SaveMessage inserts a new message into a conversation.
// 前端把 conversation_id 放在 URL path（POST /api/conversations/{id}/messages），
// body 通常只带 role/content。这里优先从 path 读取，body 里的 conversation_id 仅作兼容。
func (h *Handler) SaveMessage(w http.ResponseWriter, r *http.Request) {
	var m models.ConversationMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	conversationID := m.ConversationID
	if conversationID == 0 {
		if id, ok := idParam(r, "id"); ok {
			conversationID = id
		}
	}
	if conversationID == 0 {
		respondFail(w, 400, "缺少对话ID")
		return
	}
	m.ConversationID = conversationID

	result, err := h.db.Exec("INSERT INTO conversation_messages (conversation_id, role, content, display_content) VALUES (?, ?, ?, ?)",
		m.ConversationID, m.Role, m.Content, m.DisplayContent)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	m.ID = int(id)

	// Update conversation's updated_at timestamp
	h.db.Exec("UPDATE conversations SET updated_at=CURRENT_TIMESTAMP WHERE id=?", m.ConversationID)

	respondOK(w, m)
}

// ClearMessages deletes all messages for a conversation.
func (h *Handler) ClearMessages(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的对话ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM conversation_messages WHERE conversation_id = ?", conversationID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "清空成功")
}

// GetConversationsByBook returns all conversations for a specific book.
func (h *Handler) GetConversationsByBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	rows, err := h.db.Query("SELECT id, book_id, title, created_at, updated_at FROM conversations WHERE book_id = ? ORDER BY updated_at DESC", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Conversation
	for rows.Next() {
		var c models.Conversation
		if err := rows.Scan(&c.ID, &c.BookID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, c)
	}
	respondOK(w, list)
}

// DeleteMessage deletes a specific message by conversation_id and message_id.
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageID, ok := idParam(r, "messageId")
	if !ok {
		respondFail(w, 400, "无效的消息ID")
		return
	}
	conversationID, ok := idParam(r, "conversationId")
	if !ok {
		respondFail(w, 400, "无效的对话ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM conversation_messages WHERE id = ? AND conversation_id = ?", messageID, conversationID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// UpdateMessage updates a message's content.
func (h *Handler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	messageID, ok := idParam(r, "messageId")
	if !ok {
		respondFail(w, 400, "无效的消息ID")
		return
	}
	conversationID, ok := idParam(r, "conversationId")
	if !ok {
		respondFail(w, 400, "无效的对话ID")
		return
	}

	var m models.ConversationMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE conversation_messages SET content=?, display_content=? WHERE id=? AND conversation_id=?",
		m.Content, m.DisplayContent, messageID, conversationID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, m)
}
