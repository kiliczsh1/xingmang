package bookstore

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// volumeDirName 返回某卷自己的目录名（不含父级）。
func (s *Service) volumeDirName(bookID, volumeID int) (string, error) {
	var title string
	var order int
	err := s.DB.QueryRow(`SELECT title, order_num FROM volumes WHERE id=? AND book_id=?`, volumeID, bookID).Scan(&title, &order)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("卷不存在")
		}
		return "", err
	}
	return VolumeDirName(order, title), nil
}

// volumePath 返回卷的完整相对路径（支持嵌套：父卷目录/子卷目录）。
func (s *Service) volumePath(bookID, volumeID int) (string, error) {
	var parts []string
	cur := volumeID
	for depth := 0; cur > 0; depth++ {
		if depth > 16 {
			return "", fmt.Errorf("卷层级过深或存在循环引用")
		}
		var title string
		var order int
		var parent sql.NullInt64
		err := s.DB.QueryRow(`SELECT title, order_num, parent_id FROM volumes WHERE id=? AND book_id=?`, cur, bookID).
			Scan(&title, &order, &parent)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("卷不存在")
			}
			return "", err
		}
		parts = append([]string{VolumeDirName(order, title)}, parts...)
		if !parent.Valid || parent.Int64 <= 0 {
			break
		}
		cur = int(parent.Int64)
	}
	return filepath.ToSlash(filepath.Join(parts...)), nil
}

// chapterDir 返回章节应写入的绝对目录与相对前缀（未分卷前缀为空）。
func (s *Service) chapterDir(bookID int, volumeID *int) (string, string, error) {
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", "", err
	}
	if volumeID == nil || *volumeID <= 0 {
		return bookDir, "", nil
	}
	rel, err := s.volumePath(bookID, *volumeID)
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(bookDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	return dir, rel + "/", nil
}

// CreateVolume 建卷目录并插库。
func (s *Service) CreateVolume(v *models.Volume) error {
	if strings.TrimSpace(v.Title) == "" {
		v.Title = "未命名卷"
	}
	bookDir, err := s.BookDir(v.BookID)
	if err != nil {
		return err
	}
	var parentID int
	if v.ParentID != nil {
		parentID = *v.ParentID
	}
	if v.OrderNum <= 0 {
		var maxOrder int
		s.DB.QueryRow(`SELECT COALESCE(MAX(order_num),0) FROM volumes
			WHERE book_id=? AND COALESCE(parent_id,0)=?`, v.BookID, parentID).Scan(&maxOrder)
		v.OrderNum = maxOrder + 1
	}
	// 目录名必须能由 (order,title) 复现，否则 chapterDir 会找不到它。
	// 所以重名时递增 order，而不是加 " (2)" 后缀。
	parentPath := ""
	if parentID > 0 {
		p, err := s.volumePath(v.BookID, parentID)
		if err != nil {
			return err
		}
		parentPath = p
	}
	var dirName string
	for i := 0; i < 500; i++ {
		dirName = VolumeDirName(v.OrderNum, v.Title)
		if _, err := os.Lstat(filepath.Join(bookDir, filepath.FromSlash(parentPath), dirName)); os.IsNotExist(err) {
			break
		}
		v.OrderNum++
	}
	full := filepath.Join(bookDir, filepath.FromSlash(parentPath), dirName)
	if err := os.MkdirAll(full, 0o755); err != nil {
		return err
	}
	res, err := s.DB.Exec(`INSERT INTO volumes (book_id, parent_id, title, order_num) VALUES (?, ?, ?, ?)`,
		v.BookID, v.ParentID, v.Title, v.OrderNum)
	if err != nil {
		os.Remove(full)
		return err
	}
	id, _ := res.LastInsertId()
	v.ID = int(id)
	return nil
}

// UpdateVolume 改名/改序号/换父卷时同步移动目录，并修正受影响的章节路径。
func (s *Service) UpdateVolume(v *models.Volume) error {
	var oldTitle string
	var oldOrder, bookID int
	var oldParent sql.NullInt64
	if err := s.DB.QueryRow(`SELECT book_id, title, order_num, parent_id FROM volumes WHERE id=?`, v.ID).
		Scan(&bookID, &oldTitle, &oldOrder, &oldParent); err != nil {
		return err
	}
	if v.BookID == 0 {
		v.BookID = bookID
	}
	if v.OrderNum <= 0 {
		v.OrderNum = oldOrder
	}
	if v.Title == "" {
		v.Title = oldTitle
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	oldRel, err := s.volumePath(bookID, v.ID) // 仍是旧标题/旧序号/旧父卷
	if err != nil {
		return err
	}
	newName := VolumeDirName(v.OrderNum, v.Title)
	parentPath := ""
	if v.ParentID != nil && *v.ParentID > 0 {
		if *v.ParentID == v.ID {
			return fmt.Errorf("卷不能以自己为父级")
		}
		p, err := s.volumePath(bookID, *v.ParentID)
		if err != nil {
			return err
		}
		parentPath = p
	}
	newRel := filepath.ToSlash(filepath.Join(parentPath, newName))

	if oldRel != newRel {
		oldDir := filepath.Join(bookDir, filepath.FromSlash(oldRel))
		newDir := filepath.Join(bookDir, filepath.FromSlash(newRel))
		if _, err := os.Stat(newDir); err == nil {
			return fmt.Errorf("目标卷目录已存在: %s", newRel)
		}
		if err := os.MkdirAll(filepath.Dir(newDir), 0o755); err != nil {
			return err
		}
		if _, err := os.Stat(oldDir); err == nil {
			if err := renameWithRetry(oldDir, newDir); err != nil {
				return fmt.Errorf("移动卷目录失败: %w", err)
			}
		}
		// 前缀替换覆盖整棵子树（嵌套子卷里的章节也跟着走）
		if err := s.replacePathPrefix(bookID, oldRel+"/", newRel+"/"); err != nil {
			return err
		}
	}

	_, err = s.DB.Exec(`UPDATE volumes SET book_id=?, parent_id=?, title=?, order_num=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		v.BookID, v.ParentID, v.Title, v.OrderNum, v.ID)
	return err
}

// replacePathPrefix 批量改写章节 file_path 前缀（同事务，避免改一半）。
func (s *Service) replacePathPrefix(bookID int, oldPrefix, newPrefix string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT id, COALESCE(file_path,'') FROM chapters WHERE book_id=? AND file_path LIKE ?`,
		bookID, oldPrefix+"%")
	if err != nil {
		return err
	}
	type upd struct {
		id   int
		path string
	}
	var list []upd
	for rows.Next() {
		var id int
		var p string
		if err := rows.Scan(&id, &p); err != nil {
			rows.Close()
			return err
		}
		list = append(list, upd{id, newPrefix + strings.TrimPrefix(p, oldPrefix)})
	}
	rows.Close()
	for _, u := range list {
		if _, err := tx.Exec(`UPDATE chapters SET file_path=? WHERE id=?`, u.path, u.id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DeleteVolume 只允许删除空卷（无章节、无子卷）。
func (s *Service) DeleteVolume(id int) error {
	var bookID int
	var title string
	var order int
	if err := s.DB.QueryRow(`SELECT book_id, title, order_num FROM volumes WHERE id=?`, id).Scan(&bookID, &title, &order); err != nil {
		return err
	}
	var n, children int
	s.DB.QueryRow(`SELECT COUNT(*) FROM chapters WHERE book_id=? AND volume_id=?`, bookID, id).Scan(&n)
	if n > 0 {
		return fmt.Errorf("卷内还有 %d 章，请先移出或删除章节", n)
	}
	s.DB.QueryRow(`SELECT COUNT(*) FROM volumes WHERE parent_id=?`, id).Scan(&children)
	if children > 0 {
		return fmt.Errorf("还有 %d 个子卷，请先删除或移出", children)
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	rel, err := s.volumePath(bookID, id)
	if err != nil {
		return err
	}
	dir := filepath.Join(bookDir, filepath.FromSlash(rel))
	if _, err := os.Stat(dir); err == nil {
		if err := moveToRecycleBin(dir); err != nil {
			trash := filepath.Join(bookDir, "_trash")
			os.MkdirAll(trash, 0o755)
			dst := UniqueName(trash, filepath.Base(dir))
			if err := os.Rename(dir, filepath.Join(trash, dst)); err != nil {
				return err
			}
		}
	}
	_, err = s.DB.Exec(`DELETE FROM volumes WHERE id=?`, id)
	return err
}
