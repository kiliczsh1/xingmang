package bookstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func TestVolumeChapterPlacement(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "卷测试", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}

	v := &models.Volume{BookID: b.ID, Title: "第一卷"}
	if err := svc.CreateVolume(v); err != nil {
		t.Fatal(err)
	}
	if v.OrderNum != 1 || v.ID == 0 {
		t.Fatalf("卷序号/ID 不对: %+v", v)
	}
	if st, err := os.Stat(filepath.Join(root, "卷测试", "001 第一卷")); err != nil || !st.IsDir() {
		t.Fatalf("卷目录未创建: %v", err)
	}

	c := &models.Chapter{BookID: b.ID, Title: "楔子", Content: "正文", VolumeID: &v.ID}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}
	if c.FilePath != "001 第一卷/001 楔子.md" {
		t.Fatalf("章节应落在卷目录: %q", c.FilePath)
	}
	if _, err := os.Stat(filepath.Join(root, "卷测试", c.FilePath)); err != nil {
		t.Fatalf("文件不存在: %v", err)
	}

	// 移出卷：文件应搬到书根，卷内不再有它
	c.VolumeID = nil
	c.Content = "移出后的正文"
	if err := svc.UpdateChapter(c); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(c.FilePath, "第一卷") {
		t.Fatalf("移出卷后路径应不含卷目录: %q", c.FilePath)
	}
	if _, err := os.Stat(filepath.Join(root, "卷测试", c.FilePath)); err != nil {
		t.Fatalf("移动后的文件不存在: %v", err)
	}
}

func TestDeleteVolume_RefusesNonEmpty(t *testing.T) {
	svc, _ := setupService(t)
	b := &models.Book{Title: "书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	v := &models.Volume{BookID: b.ID, Title: "第一卷"}
	if err := svc.CreateVolume(v); err != nil {
		t.Fatal(err)
	}
	c := &models.Chapter{BookID: b.ID, Title: "章", Content: "x", VolumeID: &v.ID}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteVolume(v.ID); err == nil {
		t.Fatal("非空卷不应允许删除")
	}
}

func TestCover_ArchiveAndActivate(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "封面书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}

	png := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
	if err := svc.SetCover(b.ID, png); err != nil {
		t.Fatal(err)
	}
	cover := filepath.Join(root, "封面书", "covers", "cover.png")
	if _, err := os.Stat(cover); err != nil {
		t.Fatalf("当前封面未写入: %v", err)
	}

	// 同一张不重复归档
	if err := svc.SetCover(b.ID, png); err != nil {
		t.Fatal(err)
	}
	items, _ := svc.ListGallery(b.ID)
	if len(items) != 0 {
		t.Fatalf("同图不应产生画廊记录: %+v", items)
	}

	// 换一张（换内容）→ 旧图进画廊
	png2 := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	if err := svc.SetCover(b.ID, png2); err != nil {
		t.Fatal(err)
	}
	items, _ = svc.ListGallery(b.ID)
	if len(items) != 1 {
		t.Fatalf("旧封面应进画廊，实得 %d", len(items))
	}

	// 从画廊回切
	if err := svc.ActivateGallery(b.ID, items[0].Name); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(cover)
	base := strings.TrimSuffix(items[0].Name, filepath.Ext(items[0].Name))
	parts := strings.Split(base, "-")
	if hash8(raw) != parts[len(parts)-1] {
		t.Fatalf("回切后的封面内容不对: name=%s hash=%s", items[0].Name, hash8(raw))
	}
}

func TestNestedVolumes(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "嵌套", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	parent := &models.Volume{BookID: b.ID, Title: "第一部"}
	if err := svc.CreateVolume(parent); err != nil {
		t.Fatal(err)
	}
	child := &models.Volume{BookID: b.ID, Title: "上卷", ParentID: &parent.ID}
	if err := svc.CreateVolume(child); err != nil {
		t.Fatal(err)
	}
	c := &models.Chapter{BookID: b.ID, Title: "章", Content: "x", VolumeID: &child.ID}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}
	if c.FilePath != "001 第一部/001 上卷/001 章.md" {
		t.Fatalf("嵌套路径不对: %q", c.FilePath)
	}

	// 改父卷名：子目录与章节路径都要跟着走
	parent.Title = "第一部 重制"
	if err := svc.UpdateVolume(parent); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM chapters WHERE id=?`, c.ID).Scan(&rel)
	if rel != "001 第一部 重制/001 上卷/001 章.md" {
		t.Fatalf("重命名父卷后路径未更新: %q", rel)
	}
	if _, err := os.Stat(filepath.Join(root, "嵌套", rel)); err != nil {
		t.Fatalf("文件未随目录移动: %v", err)
	}

	// 有子卷时不允许删父卷
	if err := svc.DeleteVolume(parent.ID); err == nil {
		t.Fatal("有子卷时不应允许删除")
	}
}

func TestAddGallery_DoesNotChangeCover(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "画廊书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	png := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
	if _, err := svc.AddGallery(b.ID, png); err != nil {
		t.Fatal(err)
	}
	items, _ := svc.ListGallery(b.ID)
	if len(items) != 1 {
		t.Fatalf("画廊应有 1 张，实得 %d", len(items))
	}
	if _, _, err := svc.CoverFile(b.ID); err == nil {
		t.Fatal("加入画廊不应设置当前封面")
	}
	if _, err := os.Stat(filepath.Join(root, "画廊书", "gallery")); err != nil {
		t.Fatalf("gallery 目录不存在: %v", err)
	}
}
