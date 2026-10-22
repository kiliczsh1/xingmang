package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// experienceShareSelectCols 统一 SELECT 列清单。
// ⚠ 新增列必须同时改这里 + scanExperienceShare + CREATE TABLE + migrateExperienceSharesColumns + INSERT 清单，
// 四处不一致会直接 Scan 报「expected N destination arguments」。
const experienceShareSelectCols = `id, title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
		create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version,
		COALESCE(board, 'experience') AS board,
		section_id, COALESCE(is_featured, 0) AS is_featured, COALESCE(like_count, 0) AS like_count,
		COALESCE(is_pinned, 0) AS is_pinned,
		created_at, updated_at`

// scanExperienceShare 把一行扫成 ExperienceShare；SectionID 允许 NULL（未归类）。
func scanExperienceShare(row interface {
	Scan(dest ...interface{}) error
}) (models.ExperienceShare, error) {
	var e models.ExperienceShare
	var featured, likes, pinned sql.NullInt64
	err := row.Scan(&e.ID, &e.Title, &e.Summary, &e.Content, &e.ContentRenderMode, &e.CoverURL, &e.PdfFileURL, &e.PdfFileName, &e.PdfFileSize,
		&e.CreateType, &e.AuthorID, &e.AuthorName, &e.Status, &e.PdfParseStatus, &e.PdfParseResult, &e.SourceFileName, &e.Version,
		&e.Board, &e.SectionID, &featured, &likes, &pinned, &e.CreatedAt, &e.UpdatedAt)
	if featured.Valid {
		e.IsFeatured = int(featured.Int64)
	}
	if likes.Valid {
		e.LikeCount = int(likes.Int64)
	}
	if pinned.Valid {
		e.IsPinned = int(pinned.Int64)
	}
	e.Board = normalizeBoard(e.Board)
	return e, err
}

// ---- 请求体解析小工具（与 models.flexibleStringPtr 同口径：null/缺失 → nil；非字符串 → 原样 JSON 串） ----

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

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return false
	}
	return true
}

// decodeRawBody 把请求体读成 raw map，便于同一份 body 既喂模型又喂自定义键（如 cover_image）。
func decodeRawBody(w http.ResponseWriter, r *http.Request) (map[string]json.RawMessage, bool) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return nil, false
	}
	return raw, true
}

// rawToModel 把 raw map 重新序列化后交给模型的 UnmarshalJSON（逐键挑字段逻辑保持不变）。
func rawToModel(raw map[string]json.RawMessage, dst interface{}) error {
	buf, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, dst)
}

// GetAllExperienceShares returns all experience shares.
// ?board=tutorial|experience（缺省 experience）。库与文件两边各自过滤后再合并。
func (h *Handler) GetAllExperienceShares(w http.ResponseWriter, r *http.Request) {
	board, ok := parseBoardParam(r)
	if !ok {
		respondFail(w, 400, "board 只能是 experience 或 tutorial")
		return
	}
	list, err := h.shareListByBoard(board)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, list)
}

// GetExperienceShare returns a single experience share by id.
func (h *Handler) GetExperienceShare(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	e, err := scanExperienceShare(h.db.QueryRow(`SELECT `+experienceShareSelectCols+` FROM experience_shares WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		// 新帖在文件存储里
		if fileShare, ok := h.storeShareByID(id); ok {
			respondOK(w, fileShare)
			return
		}
		respondFail(w, 404, "经验分享不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, e)
}

// CreateExperienceShare creates a new experience share.
func (h *Handler) CreateExperienceShare(w http.ResponseWriter, r *http.Request) {
	raw, ok := decodeRawBody(w, r)
	if !ok {
		return
	}
	var e models.ExperienceShare
	if err := rawToModel(raw, &e); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	// board 非法值直接拒绝：拼错参数却静默落到默认吧，会让用户以为帖子丢了
	if v, ok := raw["board"]; ok {
		b := strings.ToLower(strings.TrimSpace(rawString(v)))
		if b != "" && b != BoardExperience && b != BoardTutorial {
			respondFail(w, 400, "board 只能是 experience 或 tutorial")
			return
		}
	}
	e.Board = normalizeBoard(e.Board)
	// cover_image 不在模型字段里（模型逐键挑字段），单独处理
	cover, err := saveCoverFromRaw(raw)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	if cover != "" {
		e.CoverURL = &cover
	}

	// 新帖写文件存储（不占数据库空间）；存储不可用时才回落数据库
	if h.shareStore != nil {
		created, err := h.shareStore.Create(e)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		respondOK(w, created)
		return
	}

	result, err := h.db.Exec(`INSERT INTO experience_shares (title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
		create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version, board, section_id, is_featured, like_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Title, e.Summary, e.Content, e.ContentRenderMode, e.CoverURL, e.PdfFileURL, e.PdfFileName, e.PdfFileSize,
		e.CreateType, e.AuthorID, e.AuthorName, e.Status, e.PdfParseStatus, e.PdfParseResult, e.SourceFileName, e.Version,
		e.Board, e.SectionID, boolToInt(e.IsFeatured != 0), e.LikeCount)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	e.ID = int(id)

	// 回显落库后的完整行（cover_url 已换成新地址），前端不必再自己拼
	if saved, err := scanExperienceShare(h.db.QueryRow(`SELECT `+experienceShareSelectCols+` FROM experience_shares WHERE id = ?`, e.ID)); err == nil {
		respondOK(w, saved)
		return
	}
	respondOK(w, e)
}

// UpdateExperienceShare updates an existing experience share.
//
// ⚠ 这个 PUT 被「编辑弹窗」「改分区」「加精/取消加精」等多种调用共用，
// 必须按「请求体里出现过的键」拼 SET —— 整行覆盖会把别的调用方没传的字段静默清零。
func (h *Handler) UpdateExperienceShare(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var raw map[string]json.RawMessage
	if !decodeJSONBody(w, r, &raw) {
		return
	}

	// 换吧：必须同时把分区解绑，否则帖子会挂到原吧的分区 id 上（跨吧脏关联）。
	// 两条分支（文件帖 / 库内帖）都吃这里补的 section_id=null。
	if v, ok := raw["board"]; ok {
		b := strings.ToLower(strings.TrimSpace(rawString(v)))
		if b != BoardExperience && b != BoardTutorial {
			respondFail(w, 400, "board 只能是 experience 或 tutorial")
			return
		}
		raw["board"] = json.RawMessage(`"` + b + `"`)
		if _, hasSection := raw["section_id"]; !hasSection {
			raw["section_id"] = json.RawMessage("null")
		}
	}

	// 文件帖：封面先在各自分支里落盘，再把补丁交给文件存储
	if h.shareInStore(id) {
		var oldCover *string
		if old, ok := h.storeShareByID(id); ok {
			oldCover = old.CoverURL
		}
		patch, err := h.resolveCoverPatch(raw, oldCover)
		if err != nil {
			respondFail(w, 400, err.Error())
			return
		}
		if h.shareStore == nil {
			respondFail(w, 500, "帖子存储不可用")
			return
		}
		updated, err := h.shareStore.Update(id, patch)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		respondOK(w, updated)
		return
	}

	sets := make([]string, 0, 20)
	args := make([]interface{}, 0, 20)
	add := func(col string, val interface{}) {
		sets = append(sets, col+"=?")
		args = append(args, val)
	}

	if v, ok := raw["title"]; ok {
		add("title", rawString(v))
	}
	if v, ok := raw["summary"]; ok {
		add("summary", rawStringPtr(v))
	}
	if v, ok := raw["content"]; ok {
		add("content", rawString(v))
	}
	if v, ok := raw["content_render_mode"]; ok {
		add("content_render_mode", rawString(v))
	}
	if v, ok := raw["cover_url"]; ok {
		add("cover_url", rawStringPtr(v))
	}
	// 上传型封面：新图落盘换 URL；remove_cover=true 则清空并删掉旧文件
	if cover, err := saveCoverFromRaw(raw); err != nil {
		respondFail(w, 400, err.Error())
		return
	} else if cover != "" {
		if old, err := scanExperienceShare(h.db.QueryRow(`SELECT `+experienceShareSelectCols+` FROM experience_shares WHERE id = ?`, id)); err == nil && old.CoverURL != nil {
			deleteExperienceCoverByURL(*old.CoverURL)
		}
		add("cover_url", cover)
	} else if v, ok := raw["remove_cover"]; ok && rawBool(v) {
		if old, err := scanExperienceShare(h.db.QueryRow(`SELECT `+experienceShareSelectCols+` FROM experience_shares WHERE id = ?`, id)); err == nil && old.CoverURL != nil {
			deleteExperienceCoverByURL(*old.CoverURL)
		}
		add("cover_url", nil)
	}
	if v, ok := raw["pdf_file_url"]; ok {
		add("pdf_file_url", rawStringPtr(v))
	}
	if v, ok := raw["pdf_file_name"]; ok {
		add("pdf_file_name", rawStringPtr(v))
	}
	if v, ok := raw["pdf_file_size"]; ok {
		add("pdf_file_size", rawIntPtr(v))
	}
	if v, ok := raw["create_type"]; ok {
		add("create_type", rawString(v))
	}
	if v, ok := raw["author_id"]; ok {
		add("author_id", rawIntPtr(v))
	}
	if v, ok := raw["author_name"]; ok {
		add("author_name", rawString(v))
	}
	if v, ok := raw["status"]; ok {
		add("status", rawString(v))
	}
	if v, ok := raw["pdf_parse_status"]; ok {
		add("pdf_parse_status", rawStringPtr(v))
	}
	if v, ok := raw["pdf_parse_result"]; ok {
		add("pdf_parse_result", rawStringPtr(v))
	}
	if v, ok := raw["source_file_name"]; ok {
		add("source_file_name", rawStringPtr(v))
	}
	if v, ok := raw["version"]; ok {
		add("version", rawStringPtr(v))
	}
	if v, ok := raw["section_id"]; ok {
		add("section_id", rawIntPtr(v))
	}
	if v, ok := raw["board"]; ok {
		add("board", normalizeBoard(rawString(v)))
	}
	if v, ok := raw["is_featured"]; ok {
		add("is_featured", boolToInt(rawInt(v) != 0))
	}
	if v, ok := raw["like_count"]; ok {
		add("like_count", rawInt(v))
	}

	if len(sets) == 0 {
		respondFail(w, 400, "没有需要更新的字段")
		return
	}

	sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
	args = append(args, id)

	if _, err := h.db.Exec("UPDATE experience_shares SET "+strings.Join(sets, ", ")+" WHERE id=?", args...); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// 回显整行（含 created_at/updated_at），前端不必再自己拼
	e, err := scanExperienceShare(h.db.QueryRow(`SELECT `+experienceShareSelectCols+` FROM experience_shares WHERE id = ?`, id))
	if err != nil {
		respondOK(w, map[string]interface{}{"id": id})
		return
	}
	respondOK(w, e)
}

// LikeExperienceShare 主楼点赞 / 取消赞（delta: 1 或 -1）。
// 楼层赞仍在浏览器 localStorage，两边相加才是帖子总赞。
func (h *Handler) LikeExperienceShare(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var req struct {
		Delta int `json:"delta"`
	}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	delta := req.Delta
	if delta == 0 {
		delta = 1
	}
	if delta > 1 {
		delta = 1
	}
	if delta < -1 {
		delta = -1
	}

	if h.shareInStore(id) {
		updated, err := h.shareStore.AdjustLike(id, delta)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		respondOK(w, map[string]interface{}{"id": id, "like_count": updated.LikeCount})
		return
	}

	if _, err := h.db.Exec(`UPDATE experience_shares SET like_count = MAX(0, COALESCE(like_count, 0) + ?), updated_at=CURRENT_TIMESTAMP WHERE id = ?`, delta, id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	var likeCount int
	if err := h.db.QueryRow(`SELECT COALESCE(like_count, 0) FROM experience_shares WHERE id = ?`, id).Scan(&likeCount); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"id": id, "like_count": likeCount})
}

// SetExperienceShareFeatured 加精 / 取消加精。
func (h *Handler) SetExperienceShareFeatured(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var raw map[string]json.RawMessage
	if !decodeJSONBody(w, r, &raw) {
		return
	}
	v, ok := raw["is_featured"]
	if !ok {
		respondFail(w, 400, "缺少 is_featured")
		return
	}
	featured := boolToInt(rawInt(v) != 0)

	if h.shareInStore(id) {
		if _, err := h.shareStore.SetFeatured(id, featured); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		respondOK(w, map[string]interface{}{"id": id, "is_featured": featured})
		return
	}

	if _, err := h.db.Exec(`UPDATE experience_shares SET is_featured=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, featured, id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"id": id, "is_featured": featured})
}

// shareBoard 取帖子所属吧（库内帖优先，其次文件帖）；不存在返回空串。
func (h *Handler) shareBoard(id int) string {
	var b string
	if err := h.db.QueryRow(`SELECT COALESCE(board, 'experience') FROM experience_shares WHERE id = ?`, id).Scan(&b); err == nil {
		return normalizeBoard(b)
	}
	if e, ok := h.storeShareByID(id); ok {
		return normalizeBoard(e.Board)
	}
	return ""
}

// sectionBelongsToBoard 校验分区存在且属于指定吧。
// ⚠ 只校验「存在」会让教程吧的帖子挂到经验分享吧的分区上（分区 id 全局唯一，跨吧能查到）。
func (h *Handler) sectionBelongsToBoard(sectionID int, board string) (bool, error) {
	var n int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM experience_sections WHERE id = ? AND COALESCE(board, 'experience') = ?`,
		sectionID, board).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// SetExperienceSharePinned 置顶 / 取消置顶。
//
// 规则：**每个吧最多 3 条置顶**。已置顶的再点一次只更新状态，不占额外名额。
func (h *Handler) SetExperienceSharePinned(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var raw map[string]json.RawMessage
	if !decodeJSONBody(w, r, &raw) {
		return
	}
	v, ok := raw["is_pinned"]
	if !ok {
		respondFail(w, 400, "缺少 is_pinned")
		return
	}
	pinned := boolToInt(rawInt(v) != 0)

	board := h.shareBoard(id)
	if board == "" {
		respondFail(w, 404, "经验分享不存在")
		return
	}

	// 置顶时才查名额；取消置顶随便点
	if pinned == 1 {
		// 库内帖（排除自己）+ 文件帖（排除自己），两边都要算，否则会突破上限
		var count int
		if err := h.db.QueryRow(
			`SELECT COUNT(*) FROM experience_shares WHERE COALESCE(board,'experience') = ? AND COALESCE(is_pinned,0) = 1 AND id <> ?`,
			board, id).Scan(&count); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if h.shareStore != nil {
			for _, e := range h.shareStore.List() {
				if e.ID != id && normalizeBoard(e.Board) == board && e.IsPinned == 1 {
					count++
				}
			}
		}
		if count >= maxPinnedShares {
			respondFail(w, 400, "每个吧最多置顶 3 条，先取消一条再置顶")
			return
		}
	}

	if h.shareInStore(id) {
		if _, err := h.shareStore.SetPinned(id, pinned); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		respondOK(w, map[string]interface{}{"id": id, "is_pinned": pinned})
		return
	}

	if _, err := h.db.Exec(`UPDATE experience_shares SET is_pinned=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, pinned, id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"id": id, "is_pinned": pinned})
}

// maxPinnedShares 每个吧的置顶上限
const maxPinnedShares = 3

// SetExperienceShareSection 把帖子移到某个分区（section_id = null 表示移出到「未归类」）。
func (h *Handler) SetExperienceShareSection(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var raw map[string]json.RawMessage
	if !decodeJSONBody(w, r, &raw) {
		return
	}
	v, ok := raw["section_id"]
	if !ok {
		respondFail(w, 400, "缺少 section_id")
		return
	}
	sectionID := rawIntPtr(v)
	if sectionID != nil {
		board := h.shareBoard(id)
		if board == "" {
			respondFail(w, 404, "经验分享不存在")
			return
		}
		matched, err := h.sectionBelongsToBoard(*sectionID, board)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if !matched {
			respondFail(w, 400, "分区不存在，或不属于该吧")
			return
		}
	}

	if h.shareInStore(id) {
		if _, err := h.shareStore.SetSection(id, sectionID); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		respondOK(w, map[string]interface{}{"id": id, "section_id": sectionID})
		return
	}

	if _, err := h.db.Exec(`UPDATE experience_shares SET section_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, sectionID, id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"id": id, "section_id": sectionID})
}

// BatchUpdateExperienceShareSection 批量移动分区（列表页「批量选择」后一次性归类）。
func (h *Handler) BatchUpdateExperienceShareSection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs       []int `json:"ids"`
		SectionID *int  `json:"section_id"`
	}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if len(req.IDs) == 0 {
		respondFail(w, 400, "请先选择帖子")
		return
	}
	if req.SectionID != nil {
		// 目标分区属于哪个吧，就以那个吧为准校验被移动的帖子（防跨吧批量串帖）
		var sectionBoard string
		if err := h.db.QueryRow(`SELECT COALESCE(board, 'experience') FROM experience_sections WHERE id = ?`, *req.SectionID).Scan(&sectionBoard); err != nil {
			if err == sql.ErrNoRows {
				respondFail(w, 400, "分区不存在")
				return
			}
			respondFail(w, 500, err.Error())
			return
		}
		sectionBoard = normalizeBoard(sectionBoard)
		for _, id := range req.IDs {
			if b := h.shareBoard(id); b != "" && b != sectionBoard {
				respondFail(w, 400, "选中的帖子里有属于别的吧的（id="+strconv.Itoa(id)+"），不能批量移到该分区")
				return
			}
		}
	}

	// 先分拣：文件帖交给文件存储，其余走数据库事务
	var dbIDs, storeIDs []int
	for _, id := range req.IDs {
		if h.shareInStore(id) {
			storeIDs = append(storeIDs, id)
		} else {
			dbIDs = append(dbIDs, id)
		}
	}

	affected := 0
	if h.shareStore != nil && len(storeIDs) > 0 {
		affected += h.shareStore.BatchSetSection(storeIDs, req.SectionID)
	}

	if len(dbIDs) > 0 {
		tx, err := h.db.Begin()
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		defer tx.Rollback()

		for _, id := range dbIDs {
			res, err := tx.Exec(`UPDATE experience_shares SET section_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, req.SectionID, id)
			if err != nil {
				respondFail(w, 500, err.Error())
				return
			}
			n, _ := res.RowsAffected()
			affected += int(n)
		}
		if err := tx.Commit(); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}
	respondOK(w, map[string]interface{}{"affected": affected})
}

// DeleteExperienceShare deletes an experience share by id.
func (h *Handler) DeleteExperienceShare(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	if h.shareInStore(id) {
		if err := h.shareStore.Delete(id); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		respondOK(w, "删除成功")
		return
	}

	_, err := h.db.Exec("DELETE FROM experience_shares WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// ImportPdf handles PDF file import for experience shares.
func (h *Handler) ImportPdf(w http.ResponseWriter, r *http.Request) {
	respondFail(w, 501, "PDF 导入功能尚未实现，请使用其他格式")
}

// exportPayload 导出文件的顶层结构（export_version 2 起）。
//
// ⚠ 与旧版（裸数组、只有 6 个字段）不兼容是**方向性**的：只要求新版能读旧文件，
// 不要求旧版能读新文件（旧版本已不维护）。导入端四种外壳全认，见 decodeImportPayload。
type exportPayload struct {
	ExportVersion int                      `json:"export_version"`
	ExportedAt    string                   `json:"exported_at"`
	Boards        []string                 `json:"boards"`
	Cards         []models.ExperienceShare `json:"cards"`
}

// ExportAll 导出全部帖子（两个吧全导，备份语义）。
//
// ⚠ 这里给的是**完整对象**：封面地址、PDF 附件元信息、分区、精华、点赞、创建时间全在。
// 附件文件本体（封面图片 / PDF 二进制）不在 JSON 里，所以它不等于整机备份。
func (h *Handler) ExportAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.shareListByBoard("") // 空 = 不过滤，两吧全量
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []models.ExperienceShare{}
	}

	payload := exportPayload{
		ExportVersion: 2,
		ExportedAt:    time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Boards:        []string{BoardTutorial, BoardExperience},
		Cards:         list,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=experience-shares.json")
	json.NewEncoder(w).Encode(payload)
}

// decodeImportPayload 解析导入文件，**四种外壳全认**：
//
//	[ {...} ]                          旧版前端导出（裸数组）
//	{ "cards": [...] }                 现前端发给后端的形状
//	{ "shares": [...] }                后端历史版本认过的键
//	{ "success": true, "data": [...] } 后端 /export 的原始响应
//	{ "export_version": 2, "cards": [...] } 新格式
//
// 认不出来就明确报错，不要静默当成 0 条（用户会以为导入成功）。
func decodeImportPayload(raw json.RawMessage) ([]models.ExperienceShare, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("文件内容为空")
	}

	if trimmed[0] == '[' {
		var list []models.ExperienceShare
		if err := json.Unmarshal(trimmed, &list); err != nil {
			return nil, fmt.Errorf("帖子数组解析失败: %w", err)
		}
		return list, nil
	}
	if trimmed[0] != '{' {
		return nil, fmt.Errorf("文件格式错误：应为帖子数组，或带 cards / shares / data 字段的对象")
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return nil, fmt.Errorf("文件解析失败: %w", err)
	}
	for _, key := range []string{"cards", "shares", "data"} {
		v, ok := obj[key]
		if !ok {
			continue
		}
		v = bytes.TrimSpace(v)
		if len(v) == 0 || string(v) == "null" {
			continue
		}
		if v[0] != '[' {
			continue
		}
		var list []models.ExperienceShare
		if err := json.Unmarshal(v, &list); err != nil {
			return nil, fmt.Errorf("%s 字段解析失败: %w", key, err)
		}
		return list, nil
	}
	return nil, fmt.Errorf("文件里没有找到帖子数组（支持裸数组，或 cards / shares / data 字段）")
}

// importError 单条导入失败的原因，原样返回给前端展示。
// ⚠ 用户的数据不能无声消失：失败条数必须和原因一起报出来。
type importError struct {
	Index  int    `json:"index"`
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

// remapImportSection 旧产物里的 section_id 是**另一台机器**的分区 id，跨机导入必然错位。
// 只有「目标库里确实存在该 id，且属于同一个吧」才保留，否则置 NULL 落「未分区」。
// ⚠ 绝不能因为分区对不上就整条失败 —— 分区是次要信息，帖子本体才是用户的数据。
//
// ⚠ 分区归属必须在**事务外**一次性预读（sectionBoards），不能在写事务里再查库：
// SQLite 单写事务会锁库，事务内另开连接查询会一直等到超时（实测导入直接挂住 15s+）。
func remapImportSection(sectionID *int, board string, sectionBoards map[int]string) *int {
	if sectionID == nil {
		return nil
	}
	if b, ok := sectionBoards[*sectionID]; ok && b == board {
		return sectionID
	}
	return nil
}

// loadSectionBoards 一次性读出「分区 id → 所属吧」映射。
func (h *Handler) loadSectionBoards() map[int]string {
	out := map[int]string{}
	rows, err := h.db.Query(`SELECT id, COALESCE(board, 'experience') FROM experience_sections`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var board string
		if err := rows.Scan(&id, &board); err == nil {
			out[id] = normalizeBoard(board)
		}
	}
	return out
}

// importTitleFallback 标题缺失时用正文前 20 字兜底（旧产物可能没标题）。
func importTitleFallback(content string) string {
	runes := []rune(strings.TrimSpace(content))
	if len(runes) == 0 {
		return ""
	}
	if len(runes) > 20 {
		runes = runes[:20]
	}
	return string(runes)
}

// ImportAll 导入帖子（兼容旧版本导出产物）。
//
// 字段容错口径见 docs/tutorial-hub-plan.md §3.2.3：
// board 缺失 → experience；created_at 缺失 → 当前时间；渲染模式 / 作者 / 状态等缺失 → 默认值；
// section_id 对不上 → NULL。只有 content 为空才判为失败（记进 errors 返回）。
func (h *Handler) ImportAll(w http.ResponseWriter, r *http.Request) {
	var rawBody json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	cards, err := decodeImportPayload(rawBody)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	if len(cards) == 0 {
		respondOK(w, map[string]interface{}{
			"insertedCount": 0,
			"failedCount":   0,
			"errors":        []importError{},
		})
		return
	}

	// ⚠ 必须在 Begin **之前**读：写事务里再查库会自己等自己的锁（见 remapImportSection 注释）
	sectionBoards := h.loadSectionBoards()

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	inserted := 0
	failures := []importError{}

	for i, e := range cards {
		content := e.Content
		if strings.TrimSpace(content) == "" {
			failures = append(failures, importError{Index: i, Title: e.Title, Reason: "正文为空"})
			continue
		}

		title := strings.TrimSpace(e.Title)
		if title == "" {
			title = importTitleFallback(content)
		}

		// 逐字段兜底：旧版本导出的文件只有 title/summary/content/渲染模式/作者/创建类型
		renderMode := strings.TrimSpace(e.ContentRenderMode)
		if renderMode == "" {
			renderMode = "markdown"
		}
		authorName := strings.TrimSpace(e.AuthorName)
		if authorName == "" {
			authorName = "星芒用户"
		}
		status := strings.TrimSpace(e.Status)
		if status == "" {
			status = "published"
		}
		createType := strings.TrimSpace(e.CreateType)
		if createType == "" {
			createType = "manual"
		}
		createdAt := strings.TrimSpace(e.CreatedAt)
		if createdAt == "" {
			createdAt = now
		}
		board := normalizeBoard(e.Board)
		sectionID := remapImportSection(e.SectionID, board, sectionBoards)

		_, err := tx.Exec(`INSERT INTO experience_shares (title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
			create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version, board, section_id, is_featured, is_pinned, like_count,
			created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			title, e.Summary, content, renderMode, e.CoverURL, e.PdfFileURL, e.PdfFileName, e.PdfFileSize,
			createType, e.AuthorID, authorName, status, e.PdfParseStatus, e.PdfParseResult, e.SourceFileName, e.Version,
			board, sectionID, boolToInt(e.IsFeatured != 0), boolToInt(e.IsPinned != 0), e.LikeCount, createdAt)
		if err != nil {
			failures = append(failures, importError{Index: i, Title: title, Reason: err.Error()})
			continue
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{
		"insertedCount": inserted,
		"failedCount":   len(failures),
		"errors":        failures,
	})
}

// boolToInt 只在需要把「非 0 即真」收成 0/1 时用（SQLite 没有 boolean）
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
