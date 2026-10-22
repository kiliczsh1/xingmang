package sharestore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func newTestStore(t *testing.T, floor int) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := New(dir, floor)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, dir
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// 文件帖 id 必须从「数据库水位之上」开始，绝不能与旧帖 id 撞车。
func TestCreateAllocatesIdAboveFloor(t *testing.T) {
	s, dir := newTestStore(t, 5)

	e, err := s.Create(models.ExperienceShare{Title: "第一帖", Content: "<p>hi</p>"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if e.ID != 6 {
		t.Fatalf("期望 id=6（DB 最大 5 之上），实际 %d", e.ID)
	}
	if _, err := os.Stat(filepath.Join(dir, "6.json")); err != nil {
		t.Fatalf("帖子文件未落盘: %v", err)
	}

	e2, _ := s.Create(models.ExperienceShare{Title: "第二帖"})
	if e2.ID != 7 {
		t.Fatalf("期望 id=7，实际 %d", e2.ID)
	}
}

// created_at 格式必须与数据库 CURRENT_TIMESTAMP 一致，否则与旧帖混排会乱。
func TestTimestampMatchesDatabaseFormat(t *testing.T) {
	s, _ := newTestStore(t, 0)
	e, _ := s.Create(models.ExperienceShare{Title: "时间格式"})

	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	if !re.MatchString(e.CreatedAt) {
		t.Fatalf("created_at 格式不符（应为 2006-01-02T15:04:05Z，与接口输出一致）: %q", e.CreatedAt)
	}
	if e.CreatedAt != e.UpdatedAt {
		t.Fatalf("新建时 updated_at 应等于 created_at: %q vs %q", e.CreatedAt, e.UpdatedAt)
	}
}

// 重启后水位不能回退：从磁盘 + _meta.json 恢复。
func TestNextIdSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	s1, _ := New(dir, 5)
	if _, err := s1.Create(models.ExperienceShare{Title: "a"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s1.Create(models.ExperienceShare{Title: "b"}); err != nil {
		t.Fatal(err)
	}

	s2, _ := New(dir, 5)
	e, _ := s2.Create(models.ExperienceShare{Title: "c"})
	if e.ID != 8 {
		t.Fatalf("重启后应继续分配 id=8，实际 %d", e.ID)
	}
}

// 丢 _meta.json 也不能撞 id（靠扫目录兜底）。
func TestNextIdFallsBackToDirScan(t *testing.T) {
	dir := t.TempDir()
	s1, _ := New(dir, 0)
	if _, err := s1.Create(models.ExperienceShare{Title: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, metaFile)); err != nil {
		t.Fatal(err)
	}

	s2, _ := New(dir, 0)
	e, _ := s2.Create(models.ExperienceShare{Title: "b"})
	if e.ID != 2 {
		t.Fatalf("meta 丢失后应扫目录得到 id=2，实际 %d", e.ID)
	}
}

// Update 只改请求体里出现过的键（未传字段必须保留）。
func TestUpdateOnlyTouchesProvidedKeys(t *testing.T) {
	s, _ := newTestStore(t, 0)
	created, _ := s.Create(models.ExperienceShare{
		Title:     "原标题",
		Summary:   strPtr("原简介"),
		Content:   "<p>正文</p>",
		Version:   strPtr("v1.0.0"),
		SectionID: intPtr(3),
		LikeCount: 7,
	})

	updated, err := s.Update(created.ID, map[string]json.RawMessage{
		"title": json.RawMessage(`"新标题"`),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "新标题" {
		t.Fatalf("title 未更新: %q", updated.Title)
	}
	if updated.Summary == nil || *updated.Summary != "原简介" {
		t.Fatalf("未传的 summary 被清掉了: %v", updated.Summary)
	}
	if updated.Version == nil || *updated.Version != "v1.0.0" {
		t.Fatalf("未传的 version 被清掉了: %v", updated.Version)
	}
	if updated.SectionID == nil || *updated.SectionID != 3 {
		t.Fatalf("未传的 section_id 被清掉了: %v", updated.SectionID)
	}
	if updated.LikeCount != 7 {
		t.Fatalf("未传的 like_count 被清掉了: %d", updated.LikeCount)
	}
	if updated.Content != "<p>正文</p>" {
		t.Fatalf("正文被清掉了: %q", updated.Content)
	}
}

func TestUpdateCanClearNullableFields(t *testing.T) {
	s, _ := newTestStore(t, 0)
	created, _ := s.Create(models.ExperienceShare{Title: "t", SectionID: intPtr(2), CoverURL: strPtr("/api/x.png")})

	updated, err := s.Update(created.ID, map[string]json.RawMessage{
		"section_id": json.RawMessage(`null`),
		"cover_url":  json.RawMessage(`null`),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.SectionID != nil {
		t.Fatalf("section_id 应被置空: %v", *updated.SectionID)
	}
	if updated.CoverURL != nil {
		t.Fatalf("cover_url 应被置空")
	}
}

func TestLikeFeaturedSection(t *testing.T) {
	s, _ := newTestStore(t, 0)
	c, _ := s.Create(models.ExperienceShare{Title: "t"})

	if e, _ := s.AdjustLike(c.ID, 1); e.LikeCount != 1 {
		t.Fatalf("点赞后应为 1，实际 %d", e.LikeCount)
	}
	// 不能出现负数
	if e, _ := s.AdjustLike(c.ID, -5); e.LikeCount != 0 {
		t.Fatalf("取消过头应夹到 0，实际 %d", e.LikeCount)
	}
	if e, _ := s.SetFeatured(c.ID, 1); e.IsFeatured != 1 {
		t.Fatalf("加精失败: %d", e.IsFeatured)
	}
	if e, _ := s.SetSection(c.ID, intPtr(9)); e.SectionID == nil || *e.SectionID != 9 {
		t.Fatalf("改分区失败: %v", e.SectionID)
	}
	if e, _ := s.SetSection(c.ID, nil); e.SectionID != nil {
		t.Fatalf("移除分区失败")
	}
}

func TestListSortedByCreatedAtDesc(t *testing.T) {
	s, _ := newTestStore(t, 0)
	a, _ := s.Create(models.ExperienceShare{Title: "早"})
	b, _ := s.Create(models.ExperienceShare{Title: "晚"})
	// 人为把 a 的时间调早（同秒创建时靠 id 兜底排序，这里显式拉开）
	if _, err := s.Update(a.ID, map[string]json.RawMessage{"title": json.RawMessage(`"早"`)}); err != nil {
		t.Fatal(err)
	}
	list := s.List()
	if len(list) != 2 {
		t.Fatalf("应有 2 条，实际 %d", len(list))
	}
	if list[0].ID != b.ID {
		t.Fatalf("列表应按创建时间倒序，首条应为 %d，实际 %d", b.ID, list[0].ID)
	}
}

func TestDeleteAndClearSection(t *testing.T) {
	s, _ := newTestStore(t, 0)
	a, _ := s.Create(models.ExperienceShare{Title: "a", SectionID: intPtr(4)})
	b, _ := s.Create(models.ExperienceShare{Title: "b", SectionID: intPtr(4)})

	if n := s.ClearSection(4); n != 2 {
		t.Fatalf("ClearSection 应影响 2 条，实际 %d", n)
	}
	for _, id := range []int{a.ID, b.ID} {
		if e, ok := s.Get(id); !ok || e.SectionID != nil {
			t.Fatalf("id=%d 的分区未清空", id)
		}
	}

	if err := s.Delete(a.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := s.Get(a.ID); ok {
		t.Fatalf("删除后仍能读到")
	}
	// 幂等
	if err := s.Delete(a.ID); err != nil {
		t.Fatalf("重复删除应无错: %v", err)
	}
}

func TestFileContentIsReadableJSON(t *testing.T) {
	s, dir := newTestStore(t, 0)
	c, _ := s.Create(models.ExperienceShare{Title: "可读性", Content: "<p>正文</p>"})

	raw, err := os.ReadFile(filepath.Join(dir, "1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("文件内容不是合法 JSON: %v", err)
	}
	if parsed["title"] != "可读性" || parsed["content"] != "<p>正文</p>" {
		t.Fatalf("字段名/值不符（应为 snake_case）: %v", parsed)
	}
	if int(parsed["id"].(float64)) != c.ID {
		t.Fatalf("id 未写入文件")
	}
}
