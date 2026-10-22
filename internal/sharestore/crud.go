package sharestore

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// ---- 请求体小工具（与 handlers 包同口径：null/缺失 → nil；非字符串 → 原样 JSON 串） ----

func rawString(v json.RawMessage) string {
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return s
	}
	return ""
}

func rawStringPtr(v json.RawMessage) *string {
	if string(v) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return &s
	}
	rawStr := string(v)
	return &rawStr
}

func rawInt(v json.RawMessage) int {
	var n int
	if err := json.Unmarshal(v, &n); err == nil {
		return n
	}
	return 0
}

func rawIntPtr(v json.RawMessage) *int {
	if string(v) == "null" {
		return nil
	}
	var n int
	if err := json.Unmarshal(v, &n); err == nil {
		return &n
	}
	return nil
}

// Create 写入一条新帖：分配 id、补时间戳，返回落盘后的完整对象。
func (s *Store) Create(e models.ExperienceShare) (models.ExperienceShare, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.next <= 0 {
		s.next = 1
	}
	e.ID = s.next
	s.next++

	// 落盘前归一，保证文件里一定是合法值（前端没传 board 时默认经验分享吧）
	e.Board = models.NormalizeBoard(e.Board)

	now := nowString()
	if strings.TrimSpace(e.CreatedAt) == "" {
		e.CreatedAt = now
	}
	e.UpdatedAt = now

	if err := s.writeLocked(e); err != nil {
		s.next-- // 落盘失败就把水位还回去
		return models.ExperienceShare{}, err
	}
	// 水位落盘失败不影响本次创建：下次启动扫目录会纠正
	_ = s.saveMetaLocked()
	return e, nil
}

// Update 按「请求体里出现过的键」打补丁（与数据库版语义一致：整行覆盖会误清字段）。
func (s *Store) Update(id int, patch map[string]json.RawMessage) (models.ExperienceShare, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, err := s.readLocked(id)
	if err != nil {
		return models.ExperienceShare{}, err
	}

	if v, ok := patch["title"]; ok {
		e.Title = rawString(v)
	}
	if v, ok := patch["summary"]; ok {
		e.Summary = rawStringPtr(v)
	}
	if v, ok := patch["content"]; ok {
		e.Content = rawString(v)
	}
	if v, ok := patch["content_render_mode"]; ok {
		e.ContentRenderMode = rawString(v)
	}
	if v, ok := patch["cover_url"]; ok {
		e.CoverURL = rawStringPtr(v)
	}
	if v, ok := patch["pdf_file_url"]; ok {
		e.PdfFileURL = rawStringPtr(v)
	}
	if v, ok := patch["pdf_file_name"]; ok {
		e.PdfFileName = rawStringPtr(v)
	}
	if v, ok := patch["pdf_file_size"]; ok {
		e.PdfFileSize = rawIntPtr(v)
	}
	if v, ok := patch["create_type"]; ok {
		e.CreateType = rawString(v)
	}
	if v, ok := patch["author_id"]; ok {
		e.AuthorID = rawIntPtr(v)
	}
	if v, ok := patch["author_name"]; ok {
		e.AuthorName = rawString(v)
	}
	if v, ok := patch["status"]; ok {
		e.Status = rawString(v)
	}
	if v, ok := patch["pdf_parse_status"]; ok {
		e.PdfParseStatus = rawStringPtr(v)
	}
	if v, ok := patch["pdf_parse_result"]; ok {
		e.PdfParseResult = rawStringPtr(v)
	}
	if v, ok := patch["source_file_name"]; ok {
		e.SourceFileName = rawStringPtr(v)
	}
	if v, ok := patch["version"]; ok {
		e.Version = rawStringPtr(v)
	}
	if v, ok := patch["section_id"]; ok {
		e.SectionID = rawIntPtr(v)
	}
	if v, ok := patch["is_featured"]; ok {
		e.IsFeatured = rawInt(v)
	}
	if v, ok := patch["is_pinned"]; ok {
		e.IsPinned = rawInt(v)
	}
	if v, ok := patch["like_count"]; ok {
		e.LikeCount = rawInt(v)
	}
	// ⚠ board 必须在这里显式列出来：Update 是手写字段白名单，漏了就是「改吧不生效」
	if v, ok := patch["board"]; ok {
		e.Board = models.NormalizeBoard(rawString(v))
	}

	e.UpdatedAt = nowString()
	if err := s.writeLocked(e); err != nil {
		return models.ExperienceShare{}, err
	}
	return e, nil
}

// Delete 删除帖子文件（不存在也返回 nil，保持幂等）。
func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.path(id)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// AdjustLike 主楼点赞增减（下限 0）。
func (s *Store) AdjustLike(id int, delta int) (models.ExperienceShare, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, err := s.readLocked(id)
	if err != nil {
		return models.ExperienceShare{}, err
	}
	e.LikeCount += delta
	if e.LikeCount < 0 {
		e.LikeCount = 0
	}
	e.UpdatedAt = nowString()
	if err := s.writeLocked(e); err != nil {
		return models.ExperienceShare{}, err
	}
	return e, nil
}

// SetFeatured 加精 / 取消加精。
func (s *Store) SetFeatured(id int, featured int) (models.ExperienceShare, error) {
	return s.Update(id, map[string]json.RawMessage{
		"is_featured": json.RawMessage(fmt.Sprintf("%d", featured)),
	})
}

// SetPinned 置顶 / 取消置顶（名额上限由 handler 校验，这里只落盘）。
func (s *Store) SetPinned(id int, pinned int) (models.ExperienceShare, error) {
	return s.Update(id, map[string]json.RawMessage{
		"is_pinned": json.RawMessage(fmt.Sprintf("%d", pinned)),
	})
}

// SetSection 移动到分区（sectionID 为 nil 表示未归类）。
func (s *Store) SetSection(id int, sectionID *int) (models.ExperienceShare, error) {
	raw := json.RawMessage("null")
	if sectionID != nil {
		raw = json.RawMessage(fmt.Sprintf("%d", *sectionID))
	}
	return s.Update(id, map[string]json.RawMessage{"section_id": raw})
}

// BatchSetSection 批量改分区，返回实际改动条数。
func (s *Store) BatchSetSection(ids []int, sectionID *int) int {
	affected := 0
	for _, id := range ids {
		if !s.Has(id) {
			continue
		}
		if _, err := s.SetSection(id, sectionID); err == nil {
			affected++
		}
	}
	return affected
}

// ClearSection 把某分区下的文件帖置为未归类（删分区时调用）。
func (s *Store) ClearSection(sectionID int) int {
	affected := 0
	for _, e := range s.List() {
		if e.SectionID != nil && *e.SectionID == sectionID {
			if _, err := s.SetSection(e.ID, nil); err == nil {
				affected++
			}
		}
	}
	return affected
}
