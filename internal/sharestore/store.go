// Package sharestore 提供经验分享帖子的本地文件存储（不占数据库空间）。
//
// 存储形态：
//
//	data/experience-shares/<id>.json   一帖一文件，内容与 models.ExperienceShare 的 JSON 一致
//	data/experience-shares/_meta.json  { "nextId": N } —— 文件帖的自增 id 水位
//
// 设计要点：
// - **id 与数据库帖不冲突**：nextId 取「DB 最大 id、已有文件最大 id、_meta.nextId」三者最大值 + 1
// - 写入原子化：先写 <id>.json.tmp 再 rename，避免读到半截文件
// - 目录随进程工作目录（与 xingmang.db 同级），备份/迁移时一起带走
//
// 上层（handlers）采用「数据库 + 文件」双读：旧帖还在库里照常读写，新帖一律落到文件。
package sharestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// Dir 帖子文件目录（相对进程工作目录）
const Dir = "data/experience-shares"

const metaFile = "_meta.json"

type meta struct {
	NextID int `json:"nextId"`
}

// Store 帖子文件存储。读写加互斥锁，单机单进程场景足够。
type Store struct {
	mu   sync.Mutex
	dir  string
	next int // 下一个可用 id（内存水位，落盘在 _meta.json）
}

// New 创建存储；floorID 传入数据库当前最大 id，避免文件帖 id 与库内旧帖撞车。
func New(dir string, floorID int) (*Store, error) {
	if dir == "" {
		dir = Dir
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("创建帖子目录失败: %w", err)
	}
	s := &Store{dir: dir, next: floorID + 1}

	// 读 _meta.json
	if raw, err := os.ReadFile(filepath.Join(dir, metaFile)); err == nil {
		var m meta
		if json.Unmarshal(raw, &m) == nil && m.NextID > s.next {
			s.next = m.NextID
		}
	}
	// 再扫一遍文件名兜底（手改过文件/丢过 meta 时也能避开冲突）
	for _, id := range s.idsOnDisk() {
		if id >= s.next {
			s.next = id + 1
		}
	}
	return s, nil
}

func (s *Store) path(id int) string {
	return filepath.Join(s.dir, strconv.Itoa(id)+".json")
}

func (s *Store) idsOnDisk() []int {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil
	}
	var ids []int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || e.Name() == metaFile {
			continue
		}
		if id, err := strconv.Atoi(strings.TrimSuffix(e.Name(), ".json")); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (s *Store) saveMetaLocked() error {
	raw, err := json.Marshal(meta{NextID: s.next})
	if err != nil {
		return err
	}
	tmp := filepath.Join(s.dir, metaFile+".tmp")
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(s.dir, metaFile))
}

// writeLocked 原子写单个帖子文件
func (s *Store) writeLocked(e models.ExperienceShare) error {
	raw, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	final := s.path(e.ID)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// List 返回全部文件帖（按创建时间倒序，与列表页排序一致）
func (s *Store) List() []models.ExperienceShare {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]models.ExperienceShare, 0, 16)
	for _, id := range s.idsOnDisk() {
		e, err := s.readLocked(id)
		if err != nil {
			continue // 单个文件损坏不该拖垮整个列表
		}
		// 出口统一归一：存量文件里没有 board 键，读出来是空串
		e.Board = models.NormalizeBoard(e.Board)
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		// 置顶优先，其次创建时间倒序（与库内帖同一口径）
		pi, pj := out[i].IsPinned == 1, out[j].IsPinned == 1
		if pi != pj {
			return pi
		}
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].ID > out[j].ID
	})
	return out
}

// BackfillBoard 给存量文件帖补上 board 字段（早期帖子的 JSON 里没有这个键）。
//
// 只处理 board 为空的文件，幂等；升级后第一次启动跑一遍即可。
// ⚠ 走 writeLocked 的原子写（.tmp + rename），不要直接覆盖原文件。
func (s *Store) BackfillBoard() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fixed := 0
	for _, id := range s.idsOnDisk() {
		e, err := s.readLocked(id)
		if err != nil {
			continue // 坏文件跳过，不阻断启动
		}
		if strings.TrimSpace(e.Board) != "" {
			continue
		}
		e.Board = models.BoardExperience
		if err := s.writeLocked(e); err != nil {
			return fixed, err
		}
		fixed++
	}
	return fixed, nil
}

func (s *Store) readLocked(id int) (models.ExperienceShare, error) {
	var e models.ExperienceShare
	raw, err := os.ReadFile(s.path(id))
	if err != nil {
		return e, err
	}
	if err := json.Unmarshal(raw, &e); err != nil {
		return e, err
	}
	return e, nil
}

// Get 读取单个文件帖
func (s *Store) Get(id int) (models.ExperienceShare, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, err := s.readLocked(id)
	if err != nil {
		return models.ExperienceShare{}, false
	}
	e.Board = models.NormalizeBoard(e.Board)
	return e, true
}

// Has 判断该 id 是否为文件帖
func (s *Store) Has(id int) bool {
	_, err := os.Stat(s.path(id))
	return err == nil
}

// nowString 必须与「库内帖经接口输出的格式」完全一致：RFC3339 UTC（2026-09-18T14:05:08Z）。
//
// ⚠ 踩过一次：数据库里存的原始文本是「2026-09-18 14:05:08」（CURRENT_TIMESTAMP），
// 但驱动 Scan 到 string 时会转成 RFC3339，接口对外给的是带 T/Z 的那种。
// 列表是「库内帖 + 文件帖」合并后按 created_at 字符串排序（前端 localeCompare、后端直接比字符串），
// 两边格式不同会让混合排序错乱（' ' 0x20 < 'T' 0x54），所以这里以**接口输出**为准。
func nowString() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}
