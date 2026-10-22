package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 创作对话的接口（docs/books-memory-storage.md §5.1、§8.1）。
//
// 从 M1 起：**对话正文在 memory/chats/*.md 里，文件是真相**。
// 这张表只剩下索引字段（book_id / title / file_path / content_hash / missing / parse_error），
// `conversation_messages` 不再写入 —— 消息只存在于文件，读取一律走文件解析。

const conversationSelectCols = `SELECT id, book_id, COALESCE(title,''), COALESCE(created_at,''), COALESCE(updated_at,''),
	COALESCE(file_path,''), COALESCE(missing,0), COALESCE(parse_error,'') FROM conversations`

func scanConversation(row interface{ Scan(...any) error }) (models.Conversation, error) {
	var c models.Conversation
	err := row.Scan(&c.ID, &c.BookID, &c.Title, &c.CreatedAt, &c.UpdatedAt, &c.FilePath, &c.Missing, &c.ParseError)
	return c, err
}

// GetConversations returns all conversations, optionally filtered by book_id.
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("book_id")

	var rows *sql.Rows
	var err error

	if bookID != "" {
		rows, err = h.db.Query(conversationSelectCols+" WHERE book_id = ? ORDER BY updated_at DESC", bookID)
	} else {
		rows, err = h.db.Query(conversationSelectCols + " ORDER BY updated_at DESC")
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Conversation
	for rows.Next() {
		c, err := scanConversation(rows)
		if err != nil {
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

	c, err := scanConversation(h.db.QueryRow(conversationSelectCols+" WHERE id = ?", id))
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

// CreateConversation creates a new conversation and its file.
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

	// 立刻建文件：让用户在资源管理器里马上能看到这个对话。
	// 建不出来就整条回滚，避免留下"有索引没文件"的对话。
	if err := bookstore.NewService(h.db).CreateChatFile(c.ID); err != nil {
		h.db.Exec("DELETE FROM conversations WHERE id=?", c.ID)
		respondFail(w, 500, "创建对话文件失败: "+err.Error())
		return
	}

	saved, err := scanConversation(h.db.QueryRow(conversationSelectCols+" WHERE id = ?", c.ID))
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, saved)
}

// UpdateConversation updates a conversation's title（同时改写文件里的 front-matter）。
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

	// 文件是真值：先改文件，再改索引。顺序反了会出现"库里是新名、磁盘是旧名"。
	if err := bookstore.NewService(h.db).RenameChat(id, c.Title); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if _, err := h.db.Exec("UPDATE conversations SET title=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", c.Title, id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	c.ID = id
	respondOK(w, c)
}

// DeleteConversation deletes a conversation: 文件进回收站，再删索引行。
func (h *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	if err := bookstore.NewService(h.db).TrashChatFile(id); err != nil {
		respondFail(w, 500, "移入回收站失败: "+err.Error())
		return
	}
	if _, err := h.db.Exec("DELETE FROM conversations WHERE id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// GetMessages returns all messages for a conversation（从文件解析）。
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的对话ID")
		return
	}

	doc, err := bookstore.NewService(h.db).LoadChat(conversationID)
	if err != nil {
		// 解析失败/文件缺失都如实返回：界面上要能区分"没有消息"和"文件坏了"。
		respondFail(w, 409, err.Error())
		return
	}

	list := make([]models.ConversationMessage, 0, len(doc.Messages))
	for _, m := range doc.Messages {
		msg := models.ConversationMessage{
			ID:             m.ID,
			ConversationID: conversationID,
			Role:           m.Role,
			Content:        m.Content,
			CreatedAt:      m.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if m.Display != "" && m.Display != m.Content {
			d := m.Display
			msg.DisplayContent = &d
		}
		list = append(list, msg)
	}
	respondOK(w, list)
}

// SaveMessage appends a message to the conversation file.
// 前端把 conversation_id 放在 URL path（POST /api/conversations/{id}/messages），
// body 通常只带 role/content/displayContent。
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

	display := ""
	if m.DisplayContent != nil {
		display = *m.DisplayContent
	}
	saved, err := bookstore.NewService(h.db).AppendChatMessage(conversationID, m.Role, m.Content, display)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	m.ID = saved.ID
	m.CreatedAt = saved.CreatedAt.Format("2006-01-02 15:04:05")
	respondOK(w, m)
}

// ClearMessages deletes all messages of a conversation（文件保留，内容清空）。
func (h *Handler) ClearMessages(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的对话ID")
		return
	}
	if err := bookstore.NewService(h.db).ClearChatMessages(conversationID); err != nil {
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

	rows, err := h.db.Query(conversationSelectCols+" WHERE book_id = ? ORDER BY updated_at DESC", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Conversation
	for rows.Next() {
		c, err := scanConversation(rows)
		if err != nil {
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

	if err := bookstore.NewService(h.db).DeleteChatMessage(conversationID, messageID); err != nil {
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

	display := ""
	if m.DisplayContent != nil {
		display = *m.DisplayContent
	}
	if err := bookstore.NewService(h.db).UpdateChatMessage(conversationID, messageID, m.Content, display); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, m)
}
