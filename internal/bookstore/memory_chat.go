package bookstore

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 创作对话的文件存储（docs/books-memory-storage.md §5.1）。
//
// 形态：
//
//	---
//	kind: chat
//	conversation_id: 12
//	book_id: 22
//	title: 列车开场怎么写
//	created_at: 2026-10-18 22:40:03
//	updated_at: 2026-10-18 23:02:11
//	---
//
//	<!-- xm:chat v=1 -->
//
//	# 列车开场怎么写
//
//	## 用户 · 22:40
//
//	<!-- xm:msg id=31 role=user at="2026-10-18 22:40:12" -->
//
//	正文
//
//	<!-- xm:msg-end -->
//
// 两条重要约定：
//  1. `## 用户 · 22:40` 这类标题是**装饰**，机器不解析它，用户随便改都不影响数据；
//     真正的结构在 xm:msg 标记里。
//  2. `conversation_messages` 表**不再写入**，消息只存在于文件 —— 文件是真相。

// ChatMessage 是一条对话消息。
type ChatMessage struct {
	ID        int
	Role      string
	Content   string
	Display   string // display_content；为空或与 Content 相同表示没有额外版本
	CreatedAt time.Time
}

// ChatDoc 是一份对话文件解析后的内容。
type ChatDoc struct {
	ConversationID int
	BookID         int
	Title          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Messages       []ChatMessage
	Extra          [][2]string // 用户自己加的 front-matter 键，写回时保留
	RelPath        string
}

// chatFrontKeys 是程序负责的 front-matter 键，其余一律视为用户自定义字段保留。
var chatFrontKeys = []string{"kind", "conversation_id", "book_id", "title", "created_at", "updated_at"}

// 对话文件的写锁：读-改-写必须串行，否则两条并发消息会互相覆盖。
//
// 键是**对话 id** 而不是路径。改成标题会重命名文件，若按路径加锁，
// 改名瞬间两个 goroutine 会锁到不同的键上，于是同时写同一个对话。
// 锁在记录 id 上，路径怎么变都不影响互斥。
var chatWriteLocks sync.Map

func lockChatConv(convID int) func() {
	v, _ := chatWriteLocks.LoadOrStore(convID, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// ---------------- 路径 ----------------

// ChatFilePath 返回对话文件的绝对路径与相对书目录的路径。
//
// 索引里没有 file_path（老数据）时按创建时间与标题现算一个并写回索引，
// 这样同一个对话不会每次换一个名字。
func (s *Service) ChatFilePath(convID int) (string, string, error) {
	var bookID int
	var title string
	var rel string
	var createdRaw, updatedRaw string
	err := s.DB.QueryRow(`SELECT book_id, COALESCE(title,''), COALESCE(file_path,''),
		COALESCE(created_at,''), COALESCE(updated_at,'') FROM conversations WHERE id=?`, convID).
		Scan(&bookID, &title, &rel, &createdRaw, &updatedRaw)
	if err != nil {
		return "", "", fmt.Errorf("对话不存在: %w", err)
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(rel) == "" {
		// 数据库时间是 UTC，先换算成本地时间再取日期做文件名，
		// 否则跨零点的对话会被编到前一天。
		at := ParseDBTime(createdRaw)
		if at.IsZero() {
			at = time.Now()
		}
		rel, err = MemoryRelPath(MemoryKindChat, at, title, convID)
		if err != nil {
			return "", "", err
		}
		if _, err := s.DB.Exec(`UPDATE conversations SET file_path=? WHERE id=?`, rel, convID); err != nil {
			return "", "", err
		}
	}
	abs, err := JoinWithin(bookDir, rel)
	if err != nil {
		return "", "", err
	}
	if err := ensureDir(filepath.Dir(abs)); err != nil {
		return "", "", err
	}
	return abs, rel, nil
}

// ---------------- 解析与渲染 ----------------

func roleLabel(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "user":
		return "用户"
	case "assistant":
		return "助手"
	case "system":
		return "系统"
	case "prompt":
		return "提示词"
	default:
		if strings.TrimSpace(role) == "" {
			return "消息"
		}
		return role
	}
}

// RenderChatDoc 把对话渲染成文件内容。
func RenderChatDoc(d *ChatDoc) string {
	known := [][2]string{
		{"kind", MemoryKindChat},
		{"conversation_id", strconv.Itoa(d.ConversationID)},
		{"book_id", strconv.Itoa(d.BookID)},
		{"title", d.Title},
		{"created_at", FormatMemTime(d.CreatedAt)},
		{"updated_at", FormatMemTime(d.UpdatedAt)},
	}
	var b strings.Builder
	b.WriteString(RenderMemoryFrontMatter(known, d.Extra))
	b.WriteString("\n")
	b.WriteString(RenderMarkerLine("chat", [][2]string{{"v", "1"}}))
	b.WriteString("\n\n")

	title := strings.TrimSpace(d.Title)
	if title == "" {
		title = "新对话"
	}
	b.WriteString("# " + title + "\n\n")
	b.WriteString("<!-- 自动生成，请勿手改标记；正文可以随便改。 -->\n\n")

	for _, m := range d.Messages {
		at := m.CreatedAt
		if at.IsZero() {
			at = d.UpdatedAt
		}
		b.WriteString("## " + roleLabel(m.Role) + " · " + at.Format("15:04") + "\n\n")
		b.WriteString(RenderPairedMarker("msg", [][2]string{
			{"id", strconv.Itoa(m.ID)},
			{"role", m.Role},
			{"at", FormatMemTime(at)},
		}, m.Content))
		if strings.TrimSpace(m.Display) != "" && m.Display != m.Content {
			b.WriteString("\n")
			b.WriteString(RenderPairedMarker("display", [][2]string{{"id", strconv.Itoa(m.ID)}}, m.Display))
		}
	}
	return b.String()
}

// ParseChatDoc 解析对话文件。失败返回 MemoryParseError（原因可展示给用户）。
func ParseChatDoc(text string) (*ChatDoc, error) {
	fm, err := ParseMemoryFrontMatter(text)
	if err != nil {
		return nil, err
	}
	pairs := fm.Pairs()
	vals := KnownFields(pairs, chatFrontKeys...)
	d := &ChatDoc{
		Extra:   ExtraFields(pairs, chatFrontKeys...),
		Title:   vals["title"],
		RelPath: "",
	}
	if n, err := strconv.Atoi(strings.TrimSpace(vals["conversation_id"])); err == nil {
		d.ConversationID = n
	}
	if n, err := strconv.Atoi(strings.TrimSpace(vals["book_id"])); err == nil {
		d.BookID = n
	}
	d.CreatedAt = ParseMemTime(vals["created_at"])
	d.UpdatedAt = ParseMemTime(vals["updated_at"])

	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		return nil, err
	}
	// display 块与消息用 id 关联（标记平铺，不嵌套）
	displays := map[int]string{}
	for _, mk := range FindMarkers(markers, "display") {
		displays[mk.IntAttr("id")] = mk.Body
	}
	for _, mk := range FindMarkers(markers, "msg") {
		m := ChatMessage{
			ID:        mk.IntAttr("id"),
			Role:      mk.Attr("role"),
			Content:   mk.Body,
			CreatedAt: ParseMemTime(mk.Attr("at")),
		}
		if v, ok := displays[m.ID]; ok {
			m.Display = v
		}
		if m.CreatedAt.IsZero() {
			m.CreatedAt = d.CreatedAt
		}
		d.Messages = append(d.Messages, m)
	}
	return d, nil
}

// ---------------- 读写 ----------------

// newChatDocFromRecord 在文件不存在时按索引行造一份空对话。
//
// 语义上这等于"文件被删了"：既然文件是真相，重建出来的对话就只有之后写入的消息。
func (s *Service) newChatDocFromRecord(convID int, rel string) (*ChatDoc, error) {
	var bookID int
	var title, createdRaw, updatedRaw string
	err := s.DB.QueryRow(`SELECT book_id, COALESCE(title,''), COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM conversations WHERE id=?`, convID).Scan(&bookID, &title, &createdRaw, &updatedRaw)
	if err != nil {
		return nil, fmt.Errorf("对话不存在: %w", err)
	}
	created := ParseDBTime(createdRaw)
	if created.IsZero() {
		created = time.Now()
	}
	updated := ParseDBTime(updatedRaw)
	if updated.IsZero() {
		updated = created
	}
	return &ChatDoc{
		ConversationID: convID,
		BookID:         bookID,
		Title:          title,
		CreatedAt:      created,
		UpdatedAt:      updated,
		RelPath:        rel,
	}, nil
}

// LoadChat 读取并解析对话文件。
//
// file 是真相：解析成功后顺手把 index 同步到文件（标题、哈希、清除 missing/parse_error）。
func (s *Service) LoadChat(convID int) (*ChatDoc, error) {
	abs, rel, err := s.ChatFilePath(convID)
	if err != nil {
		return nil, err
	}
	text, meta, err := ReadTextFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			s.DB.Exec(`UPDATE conversations SET missing=1, parse_error='' WHERE id=?`, convID)
			return nil, fmt.Errorf("对话文件不存在")
		}
		return nil, err
	}
	doc, err := ParseChatDoc(text)
	if err != nil {
		s.markChatParseError(convID, err)
		return nil, err
	}
	doc.RelPath = rel
	s.DB.Exec(`UPDATE conversations SET file_path=?, content_hash=?, missing=0, parse_error='' WHERE id=?`,
		rel, meta.Hash, convID)
	if strings.TrimSpace(doc.Title) != "" {
		s.DB.Exec(`UPDATE conversations SET title=? WHERE id=? AND title<>?`, doc.Title, convID, doc.Title)
	}
	return doc, nil
}

// CreateChatFile 建对话文件（已存在则不动）。
func (s *Service) CreateChatFile(convID int) error {
	abs, rel, err := s.ChatFilePath(convID)
	if err != nil {
		return err
	}
	unlock := lockChatConv(convID)
	defer unlock()

	if _, err := os.Stat(abs); err == nil {
		return nil
	}
	doc, err := s.newChatDocFromRecord(convID, rel)
	if err != nil {
		return err
	}
	meta, err := WriteTextFile(abs, RenderChatDoc(doc))
	if err != nil {
		return err
	}
	s.syncChatIndex(convID, rel, meta)
	return nil
}

// mutateChat 是"读文件 → 改 → 写回"的统一入口。
//
// 并发安全靠两件事：
//  1. 同一路径的互斥锁（进程内串行）；
//  2. 写前再读一次比对哈希，若期间被外部改动则重来，绝不覆盖别人的改动。
//
// 这也是对话不返回 409 的原因：对话是追加型操作，重读重放比让用户选版本更合适。
func (s *Service) mutateChat(convID int, mutate func(*ChatDoc) error) (*ChatDoc, error) {
	abs, rel, err := s.ChatFilePath(convID)
	if err != nil {
		return nil, err
	}
	unlock := lockChatConv(convID)
	defer unlock()

	for attempt := 0; attempt < 3; attempt++ {
		text, _, rerr := ReadTextFile(abs)
		exists := rerr == nil
		if rerr != nil && !os.IsNotExist(rerr) {
			return nil, rerr
		}
		var doc *ChatDoc
		if exists {
			doc, err = ParseChatDoc(text)
			if err != nil {
				// 文件被改坏：不覆盖、不猜，如实记录原因
				s.markChatParseError(convID, err)
				return nil, fmt.Errorf("对话文件解析失败: %w", err)
			}
		} else {
			doc, err = s.newChatDocFromRecord(convID, rel)
			if err != nil {
				return nil, err
			}
		}
		doc.RelPath = rel
		if err := mutate(doc); err != nil {
			return nil, err
		}
		if doc.CreatedAt.IsZero() {
			doc.CreatedAt = time.Now()
		}
		doc.UpdatedAt = time.Now()

		if exists {
			cur, _, cerr := ReadTextFile(abs)
			if cerr == nil && HashText(cur) != HashText(text) {
				continue // 期间被别的写入/外部编辑改过：重读重放
			}
		}

		// 改名要连同文件一起改：文件名里的标题段始终等于当前标题（与章节同一规则）。
		abs, rel = s.reconcileChatName(convID, abs, rel, doc.Title)

		meta, werr := WriteTextFile(abs, RenderChatDoc(doc))
		if werr != nil {
			return nil, werr
		}
		s.syncChatIndex(convID, rel, meta)
		return doc, nil
	}
	return nil, fmt.Errorf("对话文件被反复并发修改，请重试")
}

// reconcileChatName 把文件重命名成"当前标题"，并把新路径写回索引。
//
// 调用方已持有该对话的写锁，所以这里不需要额外同步。
func (s *Service) reconcileChatName(convID int, abs, rel, title string) (string, string) {
	wantRel, changed := ReconcileMemoryFileName(rel, title)
	if !changed {
		return abs, rel
	}
	newAbs, newRel, ok := RenameMemoryFile(abs, rel, wantRel)
	if !ok {
		return abs, rel
	}
	s.DB.Exec(`UPDATE conversations SET file_path=? WHERE id=?`, newRel, convID)
	return newAbs, newRel
}

func (s *Service) syncChatIndex(convID int, rel string, meta FileMeta) {
	s.DB.Exec(`UPDATE conversations SET file_path=?, content_hash=?, missing=0, parse_error='',
		synced_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=?`, rel, meta.Hash, convID)
}

func (s *Service) markChatParseError(convID int, err error) {
	s.DB.Exec(`UPDATE conversations SET parse_error=? WHERE id=?`, MemoryParseReason(err), convID)
}

func nextChatMessageID(d *ChatDoc) int {
	max := 0
	for _, m := range d.Messages {
		if m.ID > max {
			max = m.ID
		}
	}
	return max + 1
}

// ---------------- 业务操作 ----------------

// AppendChatMessage 追加一条消息，返回落盘后的消息（含分配到的 id）。
func (s *Service) AppendChatMessage(convID int, role, content, display string) (ChatMessage, error) {
	var saved ChatMessage
	_, err := s.mutateChat(convID, func(d *ChatDoc) error {
		saved = ChatMessage{
			ID:        nextChatMessageID(d),
			Role:      role,
			Content:   content,
			Display:   display,
			CreatedAt: time.Now(),
		}
		d.Messages = append(d.Messages, saved)
		return nil
	})
	if err != nil {
		return ChatMessage{}, err
	}
	return saved, nil
}

// UpdateChatMessage 改写某条消息（编辑器里改完保存走这里）。
func (s *Service) UpdateChatMessage(convID, msgID int, content, display string) error {
	found := false
	_, err := s.mutateChat(convID, func(d *ChatDoc) error {
		for i := range d.Messages {
			if d.Messages[i].ID == msgID {
				d.Messages[i].Content = content
				d.Messages[i].Display = display
				found = true
				return nil
			}
		}
		return fmt.Errorf("消息不存在（id=%d）", msgID)
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("消息不存在（id=%d）", msgID)
	}
	return nil
}

// DeleteChatMessage 删除某条消息。
func (s *Service) DeleteChatMessage(convID, msgID int) error {
	return s.mutateVoid(convID, func(d *ChatDoc) error {
		out := d.Messages[:0]
		for _, m := range d.Messages {
			if m.ID != msgID {
				out = append(out, m)
			}
		}
		d.Messages = out
		return nil
	})
}

// ClearChatMessages 清空对话里的全部消息（文件保留）。
func (s *Service) ClearChatMessages(convID int) error {
	return s.mutateVoid(convID, func(d *ChatDoc) error {
		d.Messages = nil
		return nil
	})
}

// RenameChat 改标题（只改 front-matter，文件名不动）。
func (s *Service) RenameChat(convID int, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("标题不能为空")
	}
	return s.mutateVoid(convID, func(d *ChatDoc) error {
		d.Title = title
		return nil
	})
}

func (s *Service) mutateVoid(convID int, mutate func(*ChatDoc) error) error {
	_, err := s.mutateChat(convID, mutate)
	return err
}

// TrashChatFile 把对话文件移入系统回收站（不可用时回退 _trash）。
func (s *Service) TrashChatFile(convID int) error {
	var bookID int
	var rel string
	if err := s.DB.QueryRow(`SELECT book_id, COALESCE(file_path,'') FROM conversations WHERE id=?`, convID).
		Scan(&bookID, &rel); err != nil {
		return nil // 行都没了，没什么可清的
	}
	if strings.TrimSpace(rel) == "" {
		return nil
	}
	return s.trashBookFile(bookID, rel)
}
