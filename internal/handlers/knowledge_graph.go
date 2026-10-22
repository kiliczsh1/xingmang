package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// graphEntityColumns 是 graph_entities 的读列清单，顺序必须与 scanGraphEntity 一一对应。
// 集中一处是为了不再出现「某个读接口漏带 primary_id」这类漂移。
const graphEntityColumns = "id, book_id, name, type, description, metadata, version_id, primary_id, created_at, updated_at"

// scanGraphEntity 按 graphEntityColumns 的顺序扫描一行实体。
// *sql.Rows 与 *sql.Row 都满足这个接口，读多行/读单行可以共用。
func scanGraphEntity(src interface{ Scan(...interface{}) error }, e *models.GraphEntity) error {
	return src.Scan(&e.ID, &e.BookID, &e.Name, &e.Type, &e.Description, &e.Metadata,
		&e.VersionID, &e.PrimaryID, &e.CreatedAt, &e.UpdatedAt)
}

// normalizePrimaryID 把 0/负数当作「不是副词条」，避免往库里写无意义的 0。
func normalizePrimaryID(id *int) *int {
	if id == nil || *id <= 0 {
		return nil
	}
	return id
}

// GetGraph returns entities and relations for a book.
// 可选 query 参数 versionId：只返回该版本的实体/关系。
// 不带 versionId 时保持历史行为（返回该书全部版本的数据），供「无版本图谱」的回退路径使用。
// ===== 图谱落盘（docs/books-memory-storage.md §5.3）=====
//
// 图谱写操作只负责标脏，真正的"库 → Markdown"渲染由后台合并写入器完成。
// 未分版本的实体/关系（前端 legacyData 那批）落到同一本书的 `0-未分版本.md`。

// markGraphDirty 标脏某个版本；versionID 为空时按"整书未分版本"处理。
func (h *Handler) markGraphDirty(bookID int, versionID *int) {
	vid := 0
	if versionID != nil {
		vid = *versionID
	}
	bookstore.MarkMemoryDirty(bookstore.MemoryKindGraph, bookstore.GraphDirtyKey(vid, bookID))
}

// markAllGraphVersionsDirty 标脏某本书的全部版本 + 未分版本那份。
// 用在"整书级"操作上（清空图谱、批量导入）。
func (h *Handler) markAllGraphVersionsDirty(bookID int) {
	rows, err := h.db.Query(`SELECT id FROM graph_versions WHERE book_id=?`, bookID)
	if err == nil {
		var ids []int
		for rows.Next() {
			var id int
			if rows.Scan(&id) == nil {
				ids = append(ids, id)
			}
		}
		rows.Close()
		for _, id := range ids {
			v := id
			h.markGraphDirty(bookID, &v)
		}
	}
	h.markGraphDirty(bookID, nil)
}

// graphRowScope 查出某个**实体或关系**所属的书与版本，用于删除类操作标脏。
//
// 只对 graph_entities / graph_relations 有效：这两张表有 version_id 列。
// graph_versions 本身没有这一列，查它只会静默失败——版本级操作请直接查 book_id。
func (h *Handler) graphRowScope(table string, id int) (int, *int) {
	var bookID int
	var versionID sql.NullInt64
	if err := h.db.QueryRow(`SELECT book_id, version_id FROM `+table+` WHERE id=?`, id).
		Scan(&bookID, &versionID); err != nil {
		return 0, nil
	}
	if versionID.Valid && versionID.Int64 > 0 {
		v := int(versionID.Int64)
		return bookID, &v
	}
	return bookID, nil
}

func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	// 版本过滤。前端 store.loadVersion() 一直有传 ?versionId=，
	// 但这里早期漏了它 —— 于是切到任一版本，返回的都是「全书所有版本」的实体/关系，
	// 同名实体被叠加渲染（book16 实测 179 行里只有 85 个不同名字，陈奕/大春等各 5 份），
	// 视觉上就像两份图谱叠着画。语义与 GetVersionData 保持一致：只认精确匹配的版本。
	entityWhere := "book_id = ?"
	relationWhere := "book_id = ?"
	args := []interface{}{bookID}
	if raw := r.URL.Query().Get("versionId"); raw != "" {
		if versionID, err := strconv.Atoi(raw); err == nil && versionID > 0 {
			entityWhere += " AND version_id = ?"
			relationWhere += " AND version_id = ?"
			args = append(args, versionID)
		}
	}

	// Get entities
	entityRows, err := h.db.Query("SELECT "+graphEntityColumns+" FROM graph_entities WHERE "+entityWhere, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer entityRows.Close()

	var entities []models.GraphEntity
	for entityRows.Next() {
		var e models.GraphEntity
		if err := scanGraphEntity(entityRows, &e); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		entities = append(entities, e)
	}
	if err := entityRows.Err(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// Get relations
	relationRows, err := h.db.Query("SELECT id, book_id, source_id, target_id, relation_type, description, evidence, confidence, version_id, created_at, updated_at FROM graph_relations WHERE "+relationWhere, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer relationRows.Close()

	var relations []models.GraphRelation
	for relationRows.Next() {
		var rel models.GraphRelation
		if err := relationRows.Scan(&rel.ID, &rel.BookID, &rel.SourceID, &rel.TargetID, &rel.RelationType, &rel.Description, &rel.Evidence, &rel.Confidence, &rel.VersionID, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		relations = append(relations, rel)
	}
	if err := relationRows.Err(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	respondOK(w, models.KnowledgeGraphData{Entities: entities, Relations: relations})
}

// GetEntities returns all entities for a book.
func (h *Handler) GetEntities(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	rows, err := h.db.Query("SELECT "+graphEntityColumns+" FROM graph_entities WHERE book_id = ?", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.GraphEntity
	for rows.Next() {
		var e models.GraphEntity
		if err := scanGraphEntity(rows, &e); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, e)
	}
	respondOK(w, list)
}

// GetEntity returns a single entity by id.
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var e models.GraphEntity
	err := scanGraphEntity(h.db.QueryRow("SELECT "+graphEntityColumns+" FROM graph_entities WHERE id = ?", id), &e)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "实体不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, e)
}

// CreateEntity creates a new graph entity.
func (h *Handler) CreateEntity(w http.ResponseWriter, r *http.Request) {
	var e models.GraphEntity
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec("INSERT INTO graph_entities (book_id, name, type, description, metadata, version_id, primary_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		e.BookID, e.Name, e.Type, e.Description, e.Metadata, e.VersionID, normalizePrimaryID(e.PrimaryID))
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	e.ID = int(id)
	h.markGraphDirty(e.BookID, e.VersionID)
	respondOK(w, e)
}

// UpdateEntity updates an existing graph entity.
func (h *Handler) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var e models.GraphEntity
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE graph_entities SET book_id=?, name=?, type=?, description=?, metadata=?, version_id=?, primary_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		e.BookID, e.Name, e.Type, e.Description, e.Metadata, e.VersionID, normalizePrimaryID(e.PrimaryID), id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	e.ID = id
	h.markGraphDirty(e.BookID, e.VersionID)
	respondOK(w, e)
}

// DeleteEntity deletes a graph entity by id.
// 若它本身是主词条，先把挂在它名下的副词条解绑（变成主词条），
// 否则那些副词条的 primary_id 会指向一个不存在的实体（表里刻意没加外键，不会自动处理）。
func (h *Handler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	if _, err = tx.Exec("UPDATE graph_entities SET primary_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE primary_id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if _, err = tx.Exec("DELETE FROM graph_entities WHERE id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if err = tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if bookID, vid := h.graphRowScope("graph_entities", id); bookID > 0 {
		h.markGraphDirty(bookID, vid)
	}
	respondOK(w, "删除成功")
}

// GetRelations returns all relations for a book.
func (h *Handler) GetRelations(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	rows, err := h.db.Query("SELECT id, book_id, source_id, target_id, relation_type, description, evidence, confidence, version_id, created_at, updated_at FROM graph_relations WHERE book_id = ?", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.GraphRelation
	for rows.Next() {
		var rel models.GraphRelation
		if err := rows.Scan(&rel.ID, &rel.BookID, &rel.SourceID, &rel.TargetID, &rel.RelationType, &rel.Description, &rel.Evidence, &rel.Confidence, &rel.VersionID, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, rel)
	}
	respondOK(w, list)
}

// CreateRelation creates a new graph relation.
func (h *Handler) CreateRelation(w http.ResponseWriter, r *http.Request) {
	var rel models.GraphRelation
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec("INSERT INTO graph_relations (book_id, source_id, target_id, relation_type, description, evidence, confidence, version_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		rel.BookID, rel.SourceID, rel.TargetID, rel.RelationType, rel.Description, rel.Evidence, rel.Confidence, rel.VersionID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	rel.ID = int(id)
	h.markGraphDirty(rel.BookID, rel.VersionID)
	respondOK(w, rel)
}

// UpdateRelation updates an existing graph relation.
func (h *Handler) UpdateRelation(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var rel models.GraphRelation
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE graph_relations SET book_id=?, source_id=?, target_id=?, relation_type=?, description=?, evidence=?, confidence=?, version_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		rel.BookID, rel.SourceID, rel.TargetID, rel.RelationType, rel.Description, rel.Evidence, rel.Confidence, rel.VersionID, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	rel.ID = id
	h.markGraphDirty(rel.BookID, rel.VersionID)
	respondOK(w, rel)
}

// DeleteRelation deletes a graph relation by id.
func (h *Handler) DeleteRelation(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	bookID, vid := h.graphRowScope("graph_relations", id)
	_, err := h.db.Exec("DELETE FROM graph_relations WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if bookID > 0 {
		h.markGraphDirty(bookID, vid)
	}
	respondOK(w, "删除成功")
}

// ClearGraph deletes all entities and relations for a book.
func (h *Handler) ClearGraph(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM graph_relations WHERE book_id = ?", bookID); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if _, err := tx.Exec("DELETE FROM graph_entities WHERE book_id = ?", bookID); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.markAllGraphVersionsDirty(bookID)
	respondOK(w, "清空成功")
}

// GetVersions returns all graph versions for a book.
func (h *Handler) GetVersions(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	rows, err := h.db.Query("SELECT id, book_id, name, description, scope, chapter_count, entity_count, relation_count, created_at, updated_at FROM graph_versions WHERE book_id = ? ORDER BY created_at DESC", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.GraphVersion
	for rows.Next() {
		var v models.GraphVersion
		if err := rows.Scan(&v.ID, &v.BookID, &v.Name, &v.Description, &v.Scope, &v.ChapterCount, &v.EntityCount, &v.RelationCount, &v.CreatedAt, &v.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, v)
	}
	respondOK(w, list)
}

// CreateVersion creates a new graph version.
func (h *Handler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	// 请求体用的是驼峰 bookId（见 web/src/api/index.ts 的 createVersion 与 store.saveAsNewVersion），
	// 而响应里的 GraphVersion 是 snake_case。所以请求必须单独解成下面这个结构：
	// 早期直接用 models.GraphVersion 解码，book_id 收不到值恒为 0，
	// 触发 graph_versions.book_id 的外键校验失败 —— 「新建版本」「另存为新版本」于是全部报错。
	var req struct {
		BookID      int    `json:"bookId"`
		BookIDSnake int    `json:"book_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	bookID := req.BookID
	if bookID == 0 {
		bookID = req.BookIDSnake
	}
	if bookID <= 0 {
		respondFail(w, 400, "无效的书籍ID")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "未命名版本"
	}
	scope := req.Scope
	if scope == "" {
		scope = "all"
	}

	// 计数从 0 起步：真实数字由 ImportGraph 按落库行数回写
	result, err := h.db.Exec(
		"INSERT INTO graph_versions (book_id, name, description, scope, chapter_count, entity_count, relation_count) VALUES (?, ?, ?, ?, 0, 0, 0)",
		bookID, name, req.Description, scope)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := result.LastInsertId()

	// 读回整行，否则 created_at 等字段会是空值
	var v models.GraphVersion
	err = h.db.QueryRow(
		"SELECT id, book_id, name, description, scope, chapter_count, entity_count, relation_count, created_at, updated_at FROM graph_versions WHERE id = ?",
		id).Scan(&v.ID, &v.BookID, &v.Name, &v.Description, &v.Scope,
		&v.ChapterCount, &v.EntityCount, &v.RelationCount, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.markGraphDirty(v.BookID, &v.ID)
	respondOK(w, v)
}

// VersionData is a version with its entities and relations.
type VersionData struct {
	models.GraphVersion
	Entities  []models.GraphEntity   `json:"entities"`
	Relations []models.GraphRelation `json:"relations"`
}

// GetVersionData returns a version with its entities and relations.
func (h *Handler) GetVersionData(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "versionId")
	if !ok {
		respondFail(w, 400, "无效的版本ID")
		return
	}

	var v models.GraphVersion
	err := h.db.QueryRow("SELECT id, book_id, name, description, scope, chapter_count, entity_count, relation_count, created_at, updated_at FROM graph_versions WHERE id = ?", id).
		Scan(&v.ID, &v.BookID, &v.Name, &v.Description, &v.Scope, &v.ChapterCount, &v.EntityCount, &v.RelationCount, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "版本不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// Get entities for this version
	entityRows, err := h.db.Query("SELECT "+graphEntityColumns+" FROM graph_entities WHERE version_id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer entityRows.Close()

	var entities []models.GraphEntity
	for entityRows.Next() {
		var e models.GraphEntity
		if err := scanGraphEntity(entityRows, &e); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		entities = append(entities, e)
	}

	// Get relations for this version
	relationRows, err := h.db.Query("SELECT id, book_id, source_id, target_id, relation_type, description, evidence, confidence, version_id, created_at, updated_at FROM graph_relations WHERE version_id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer relationRows.Close()

	var relations []models.GraphRelation
	for relationRows.Next() {
		var rel models.GraphRelation
		if err := relationRows.Scan(&rel.ID, &rel.BookID, &rel.SourceID, &rel.TargetID, &rel.RelationType, &rel.Description, &rel.Evidence, &rel.Confidence, &rel.VersionID, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		relations = append(relations, rel)
	}

	respondOK(w, VersionData{GraphVersion: v, Entities: entities, Relations: relations})
}

// RenameVersion updates a version's name.
func (h *Handler) RenameVersion(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE graph_versions SET name=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", req.Name, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	// 名字写在 front-matter 与文件主名里，改了就要重渲染。
	// 这里不能用 graphRowScope：graph_versions 表**没有** version_id 列，
	// 拿它去查会静默失败（返回 0 导致标脏被跳过，表现成"改名没生效"）。
	var bookID int
	if err := h.db.QueryRow(`SELECT book_id FROM graph_versions WHERE id=?`, id).Scan(&bookID); err == nil && bookID > 0 {
		vid := id
		h.markGraphDirty(bookID, &vid)
	}
	respondOK(w, "重命名成功")
}

// DeleteVersion deletes a graph version by id.
// 路由是 DELETE /api/knowledge-graph/versions/{bookId}/{versionId}，路径参数名是 versionId。
// 早期这里取的是 "id" —— PathValue 拿不到值，于是恒返回 400「无效的ID」，
// 也就是说「删除版本」这个功能从来没能用。顺带用 bookId 校验归属，避免误删别书的版本。
// graph_versions 上有 ON DELETE CASCADE，删除会一并清掉该版本的实体/关系/章节关联（外键已开启）。
func (h *Handler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}
	versionID, ok := idParam(r, "versionId")
	if !ok {
		respondFail(w, 400, "无效的版本ID")
		return
	}

	// 先归档文件再删行：删行是级联的，反过来做文件就成了找不到主人的孤儿
	if err := bookstore.NewService(h.db).TrashGraphFile(versionID); err != nil {
		respondFail(w, 500, "移入回收站失败: "+err.Error())
		return
	}
	result, err := h.db.Exec("DELETE FROM graph_versions WHERE id = ? AND book_id = ?", versionID, bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		respondFail(w, 404, "版本不存在或不属于该书籍")
		return
	}
	respondOK(w, "删除成功")
}

// AnalyzeGraph analyzes a knowledge graph for a book (placeholder).
func (h *Handler) AnalyzeGraph(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID int `json:"bookId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"message": "知识图谱分析功能尚未实现", "bookId": req.BookID})
}

// AISupplement uses AI to supplement the knowledge graph (placeholder).
func (h *Handler) AISupplement(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID int `json:"bookId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"message": "AI 补充知识图谱功能尚未实现", "bookId": req.BookID})
}

// ImportGraph imports graph data from JSON.
// 语义是「整批 INSERT 成一份快照」：写入指定的 version_id（或逐行携带的 version_id），
// 旧版本数据一律不动。所以「AI 分析结果合并到当前图谱」只要把合并结果写成新版本即可，天然可回退。
func (h *Handler) ImportGraph(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID    int                    `json:"bookId"`
		VersionID *int                   `json:"versionId"`
		Entities  []models.GraphEntity   `json:"entities"`
		Relations []models.GraphRelation `json:"relations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	// 顶层 versionId 作为逐行 version_id 缺失时的兜底
	fallback := req.VersionID
	versionOf := func(v *int) *int {
		if v != nil && *v > 0 {
			return v
		}
		if fallback != nil && *fallback > 0 {
			return fallback
		}
		return nil
	}

	// 本次导入涉及的版本集合，提交前统一回写计数
	versionIDs := make(map[int]bool)

	entityCount := 0
	entityIDMap := make(map[int]int) // old id -> new id
	for _, e := range req.Entities {
		version := versionOf(e.VersionID)
		// primary_id 先留空：它指向的是「入参里的旧 id」，要等 entityIDMap 建好后再回填
		result, err := tx.Exec("INSERT INTO graph_entities (book_id, name, type, description, metadata, version_id, primary_id) VALUES (?, ?, ?, ?, ?, ?, NULL)",
			req.BookID, e.Name, e.Type, e.Description, e.Metadata, version)
		if err != nil {
			continue
		}
		newID, _ := result.LastInsertId()
		entityIDMap[e.ID] = int(newID)
		entityCount++
		if version != nil {
			versionIDs[*version] = true
		}
		// 章节关联必须一并落库：前端抽取结果里带着 chapter_ids（来源章节），
		// 这里丢掉的话，左栏章节覆盖标记与按章节过滤会全部失效。
		entityID := int(newID)
		if _, err := insertChapterRefs(tx, req.BookID, version, &entityID, nil, e.ChapterIDs); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}

	// 回填副词条的主词条指向（old id -> new id）。主词条不在本批次里就放弃，
	// 宁可让这条变成主词条，也不要写出指向不存在实体的 primary_id。
	for _, e := range req.Entities {
		if e.PrimaryID == nil || *e.PrimaryID <= 0 {
			continue
		}
		newID, ok := entityIDMap[e.ID]
		if !ok {
			continue
		}
		newPrimary, ok := entityIDMap[*e.PrimaryID]
		if !ok || newPrimary == newID {
			continue
		}
		if _, err := tx.Exec("UPDATE graph_entities SET primary_id = ? WHERE id = ?", newPrimary, newID); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}

	relationCount := 0
	for _, rel := range req.Relations {
		version := versionOf(rel.VersionID)
		sourceID, ok := entityIDMap[rel.SourceID]
		if !ok {
			sourceID = rel.SourceID
		}
		targetID, ok := entityIDMap[rel.TargetID]
		if !ok {
			targetID = rel.TargetID
		}
		result, err := tx.Exec("INSERT INTO graph_relations (book_id, source_id, target_id, relation_type, description, evidence, confidence, version_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			req.BookID, sourceID, targetID, rel.RelationType, rel.Description, rel.Evidence, rel.Confidence, version)
		if err != nil {
			continue
		}
		newID, _ := result.LastInsertId()
		relationCount++
		if version != nil {
			versionIDs[*version] = true
		}
		relationID := int(newID)
		if _, err := insertChapterRefs(tx, req.BookID, version, nil, &relationID, rel.ChapterIDs); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}

	// 回写版本计数：前端 api/index.ts 的 importGraph 注释明确期望「后端会同时回写该版本的
	// 实体/关系计数」，版本列表展示的就是这两个字段——不同步就会出现「列表写着 59 条、图上 63 条」。
	// 按实际落库行数重算而不是累加，重复导入或中途失败都不会把数字带偏。
	for versionID := range versionIDs {
		if _, err := tx.Exec(
			`UPDATE graph_versions SET
			   entity_count   = (SELECT COUNT(*) FROM graph_entities  WHERE version_id = ?),
			   relation_count = (SELECT COUNT(*) FROM graph_relations WHERE version_id = ?),
			   updated_at     = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			versionID, versionID, versionID); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	for versionID := range versionIDs {
		v := versionID
		h.markGraphDirty(req.BookID, &v)
	}
	// 入参里没有版本号的行会写进"未分版本"那份
	if fallback == nil || *fallback <= 0 {
		h.markGraphDirty(req.BookID, nil)
	}
	respondOK(w, map[string]interface{}{"entityCount": entityCount, "relationCount": relationCount})
}

// ---------------------------------------------------------------- 章节关联
// graph_chapter_refs 记录「实体/关系出自哪些章节」，是左栏章节覆盖标记与按章节过滤的数据源。
// 表上有 CHECK ((entity_id IS NOT NULL) <> (relation_id IS NOT NULL))，一行只能挂实体或关系。

// chapterRefBody 是设置章节关联的请求体。
type chapterRefBody struct {
	ChapterIDs []int `json:"chapterIds"`
}

// insertChapterRefs 在已有事务里为某个实体或关系写入章节关联行。
// entityID / relationID 必须恰好给一个（表上有 CHECK 约束）；versionID 为 nil 时写 NULL。
// 自动跳过非法与重复的章节 id，返回实际写入行数。
func insertChapterRefs(tx *sql.Tx, bookID int, versionID *int, entityID, relationID *int, chapterIDs []int) (int, error) {
	var version interface{}
	if versionID != nil && *versionID > 0 {
		version = *versionID
	}

	seen := make(map[int]bool, len(chapterIDs))
	count := 0
	for _, chapterID := range chapterIDs {
		if chapterID <= 0 || seen[chapterID] {
			continue
		}
		seen[chapterID] = true
		if _, err := tx.Exec(
			"INSERT INTO graph_chapter_refs (book_id, version_id, chapter_id, entity_id, relation_id) VALUES (?, ?, ?, ?, ?)",
			bookID, version, chapterID, entityID, relationID); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// replaceChapterRefs 用 chapterIDs 整体替换某个实体/关系已有的关联行。
// entityID 与 relationID 必须恰好给一个；versionID 为 0 时写入 NULL。
func (h *Handler) replaceChapterRefs(bookID, versionID int, entityID, relationID *int, chapterIDs []int) (int, error) {
	tx, err := h.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if entityID != nil {
		if _, err = tx.Exec("DELETE FROM graph_chapter_refs WHERE entity_id = ?", *entityID); err != nil {
			return 0, err
		}
	} else {
		if _, err = tx.Exec("DELETE FROM graph_chapter_refs WHERE relation_id = ?", *relationID); err != nil {
			return 0, err
		}
	}

	count, err := insertChapterRefs(tx, bookID, &versionID, entityID, relationID, chapterIDs)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

// SetEntityChapters 整体替换实体与章节的关联（PUT /api/knowledge-graph/entities/{id}/chapters）
func (h *Handler) SetEntityChapters(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var body chapterRefBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	var bookID, versionID int
	err := h.db.QueryRow("SELECT book_id, COALESCE(version_id, 0) FROM graph_entities WHERE id = ?", id).
		Scan(&bookID, &versionID)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "实体不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	count, err := h.replaceChapterRefs(bookID, versionID, &id, nil, body.ChapterIDs)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if versionID > 0 {
		h.markGraphDirty(bookID, &versionID)
	} else {
		h.markGraphDirty(bookID, nil)
	}
	respondOK(w, map[string]interface{}{"chapterRefCount": count})
}

// SetRelationChapters 整体替换关系与章节的关联（PUT /api/knowledge-graph/relations/{id}/chapters）
func (h *Handler) SetRelationChapters(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var body chapterRefBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	var bookID, versionID int
	err := h.db.QueryRow("SELECT book_id, COALESCE(version_id, 0) FROM graph_relations WHERE id = ?", id).
		Scan(&bookID, &versionID)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "关系不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	count, err := h.replaceChapterRefs(bookID, versionID, nil, &id, body.ChapterIDs)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if versionID > 0 {
		h.markGraphDirty(bookID, &versionID)
	} else {
		h.markGraphDirty(bookID, nil)
	}
	respondOK(w, map[string]interface{}{"chapterRefCount": count})
}

// GetChapterRefs 按章节分组返回该书的实体/关系关联（GET /api/knowledge-graph/chapter-refs/{bookId}）
// 可选 query 参数 versionId，只取指定版本的关联。
func (h *Handler) GetChapterRefs(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	query := "SELECT chapter_id, entity_id, relation_id FROM graph_chapter_refs WHERE book_id = ?"
	args := []interface{}{bookID}
	if raw := r.URL.Query().Get("versionId"); raw != "" {
		if versionID, err := strconv.Atoi(raw); err == nil && versionID > 0 {
			query += " AND version_id = ?"
			args = append(args, versionID)
		}
	}
	query += " ORDER BY chapter_id, id"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	order := make([]int, 0)
	groups := make(map[int]*models.GraphChapterRef)
	for rows.Next() {
		var chapterID int
		var entityID, relationID sql.NullInt64
		if err := rows.Scan(&chapterID, &entityID, &relationID); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		ref, exists := groups[chapterID]
		if !exists {
			ref = &models.GraphChapterRef{EntityIDs: []int{}, RelationIDs: []int{}}
			groups[chapterID] = ref
			order = append(order, chapterID)
		}
		if entityID.Valid {
			ref.EntityIDs = append(ref.EntityIDs, int(entityID.Int64))
		}
		if relationID.Valid {
			ref.RelationIDs = append(ref.RelationIDs, int(relationID.Int64))
		}
	}
	if err := rows.Err(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	list := make([]models.GraphChapterRef, 0, len(order))
	for _, chapterID := range order {
		ref := groups[chapterID]
		ref.ChapterID = chapterID
		list = append(list, *ref)
	}
	respondOK(w, list)
}

// ================================================================ 实体合并系统
//
// 主词条 / 副词条（graph_entities.primary_id）：
//	primary_id IS NULL     -> 主词条，画布上真正渲染的节点
//	primary_id = 某个主词条 -> 副词条（别名），读接口照原样带出该字段，由前端折进主词条渲染
//
// 合并只改 primary_id，**不动关系、不动章节关联**，所以解除合并能完全还原（可逆、不丢数据）。
// 检测只做启发式并给出可解释理由，最终是否合并由人决定（人工合并检测）。

// sameVersion 判断两个可空 version_id 是否指同一版本。
func sameVersion(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// normalizeEntityName 与前端 normalizeName 同口径：去首尾空白、折叠内部空白、转小写。
func normalizeEntityName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}

// charBigrams 取字符二元组集合。中文没有词边界，按字符切比按空格切稳。
func charBigrams(s string) map[string]bool {
	runes := []rune(strings.Join(strings.Fields(s), ""))
	out := make(map[string]bool, len(runes))
	for i := 0; i+1 < len(runes); i++ {
		out[string(runes[i:i+2])] = true
	}
	return out
}

func bigramJaccard(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	for k := range a {
		if b[k] {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func intSetIntersect(a, b map[int]bool) int {
	if len(a) > len(b) {
		a, b = b, a
	}
	n := 0
	for k := range a {
		if b[k] {
			n++
		}
	}
	return n
}

func intSetJaccard(a, b map[int]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := intSetIntersect(a, b)
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

// mergeCandidateThreshold 低于该分数就不当成候选，免得把噪声堆给人看。
const mergeCandidateThreshold = 0.5

// isNameContainment 判断两个归一化名字是否构成「一个包含另一个」的别名候选。
// 两个限制专门用来砍误报（实测 v10：不加限制会产出 41 对，其中绝大多数是噪声）：
//  1. 只比同类型 —— 「老刘」vs「老刘认识聚灵丹」、「飞虎寨」vs「飞虎寨二当家」这类
//     把短名嵌进去的复合体（事件 / 线索 / 势力），类型通常与短名不同；
//  2. 长度比 ≥ 0.5 —— 砍掉「陈奕」vs「陈奕的女儿」这种「短名 + 关系后缀」，
//     它指的是另一个人而不是别名。
func isNameContainment(a, b string) bool {
	if a == b {
		return false
	}
	short, long := a, b
	if len([]rune(short)) > len([]rune(long)) {
		short, long = long, short
	}
	shortRunes, longRunes := []rune(short), []rune(long)
	if len(shortRunes) < 2 || !strings.Contains(long, short) {
		return false
	}
	return float64(len(shortRunes))/float64(len(longRunes)) >= 0.5
}

// scoreCandidatePair 给一对实体打「疑似重复」分，并给出可读理由与建议的合并方向。
// 返回 score == 0 表示这对不是候选。
func scoreCandidatePair(a, b models.GraphEntity, neighbors map[int]map[int]bool) (float64, []string, int, int) {
	nameA, nameB := normalizeEntityName(a.Name), normalizeEntityName(b.Name)
	if nameA == "" || nameB == "" {
		return 0, nil, 0, 0
	}
	sameType := a.Type == b.Type

	score := 0.0
	reasons := []string{}
	suggestPrimary, suggestDuplicate := a.ID, b.ID // 默认：id 小的（更旧的）当主词条

	// 名称层。与抽取侧的查重口径一致：只有「完全相同」才不看类型（人名全局唯一）
	if nameA == nameB {
		score = 1.0
		reasons = append(reasons, "名称完全相同")
		if a.ID > b.ID {
			suggestPrimary, suggestDuplicate = b.ID, a.ID
		}
	} else if sameType && isNameContainment(nameA, nameB) {
		score = 0.85
		reasons = append(reasons, "同类型且名称包含")
		// 更具体的那个（名字更长的）更适合当主词条
		if len([]rune(nameA)) < len([]rune(nameB)) {
			suggestPrimary, suggestDuplicate = b.ID, a.ID
		}
	}

	// 相似层：一律要求同类型，跨类型大多是不同东西
	if sameType {
		if j := bigramJaccard(charBigrams(a.Description), charBigrams(b.Description)); j >= 0.6 {
			if s := 0.4 + j*0.35; s > score {
				score = s
				reasons = append(reasons, fmt.Sprintf("描述相似度 %.0f%%", j*100))
			}
		}
		// 共同邻居要同时满足「数量 ≥ 2」和「占比 ≥ 50%」，否则两个高连接度节点很容易偶然重叠
		if inter := intSetIntersect(neighbors[a.ID], neighbors[b.ID]); inter >= 2 {
			if j := intSetJaccard(neighbors[a.ID], neighbors[b.ID]); j >= 0.5 {
				if s := 0.4 + j*0.3; s > score {
					score = s
					reasons = append(reasons, fmt.Sprintf("共同关联 %d 个实体", inter))
				}
			}
		}
	}

	if score < mergeCandidateThreshold {
		return 0, nil, 0, 0
	}
	return score, reasons, suggestPrimary, suggestDuplicate
}

// resolvePrimaryRoot 顺链找到真正的主词条：传入的 id 若本身是副词条，就沿 primary_id 往上走。
// 带步数上限，防止脏数据成环时死循环。
func (h *Handler) resolvePrimaryRoot(id int) (int, error) {
	cur := id
	for step := 0; step < 16; step++ {
		var next *int
		if err := h.db.QueryRow("SELECT primary_id FROM graph_entities WHERE id = ?", cur).Scan(&next); err != nil {
			return 0, err
		}
		if next == nil || *next <= 0 || *next == cur {
			return cur, nil
		}
		cur = *next
	}
	return cur, nil
}

// MergeEntities 把若干实体合并到主词条下，被合并者变成副词条。
// POST /api/knowledge-graph/entities/merge
// body: {"bookId":16,"versionId":10,"primaryId":233,"memberIds":[261,262]}
func (h *Handler) MergeEntities(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID    int   `json:"bookId"`
		VersionID *int  `json:"versionId"`
		PrimaryID int   `json:"primaryId"`
		MemberIDs []int `json:"memberIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if req.PrimaryID <= 0 {
		respondFail(w, 400, "缺少主词条 primaryId")
		return
	}
	if len(req.MemberIDs) == 0 {
		respondFail(w, 400, "memberIds 不能为空")
		return
	}

	// 主词条允许传副词条 id，这里顺链归到根，避免出现「副词条指向副词条」的链
	if err := h.db.QueryRow("SELECT id FROM graph_entities WHERE id = ?", req.PrimaryID).Scan(new(int)); err != nil {
		if err == sql.ErrNoRows {
			respondFail(w, 404, "主词条不存在")
			return
		}
		respondFail(w, 500, err.Error())
		return
	}
	rootID, err := h.resolvePrimaryRoot(req.PrimaryID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	var bookID int
	var rootVersion *int
	if err := h.db.QueryRow("SELECT book_id, version_id FROM graph_entities WHERE id = ?", rootID).
		Scan(&bookID, &rootVersion); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	merged, reparented, skipped := 0, 0, 0
	for _, memberID := range req.MemberIDs {
		if memberID <= 0 || memberID == rootID {
			skipped++
			continue
		}
		var mBookID int
		var mVersion, mPrimary *int
		err := tx.QueryRow("SELECT book_id, version_id, primary_id FROM graph_entities WHERE id = ?", memberID).
			Scan(&mBookID, &mVersion, &mPrimary)
		if err == sql.ErrNoRows {
			skipped++
			continue
		}
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		// 只允许同书同版本内合并：跨版本合并会把版本快照语义搞坏
		if mBookID != bookID || !sameVersion(mVersion, rootVersion) {
			skipped++
			continue
		}
		if mPrimary != nil && *mPrimary == rootID {
			skipped++ // 已经在同一组里了，幂等
			continue
		}

		// 它自己名下的副词条改挂到新的根上
		res, err := tx.Exec("UPDATE graph_entities SET primary_id = ?, updated_at = CURRENT_TIMESTAMP WHERE primary_id = ?", rootID, memberID)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if n, _ := res.RowsAffected(); n > 0 {
			reparented += int(n)
		}

		if _, err = tx.Exec("UPDATE graph_entities SET primary_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", rootID, memberID); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		merged++
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.markGraphDirty(req.BookID, req.VersionID)
	respondOK(w, map[string]interface{}{
		"primaryId":  rootID,
		"merged":     merged,
		"reparented": reparented,
		"skipped":    skipped,
	})
}

// UnmergeEntities 解除合并：传主词条则解散整组，传副词条则只解绑它自己。
// POST /api/knowledge-graph/entities/unmerge
// body: {"entityIds":[233]}
func (h *Handler) UnmergeEntities(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EntityIDs []int `json:"entityIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if len(req.EntityIDs) == 0 {
		respondFail(w, 400, "entityIds 不能为空")
		return
	}

	// 解绑会改 primary_id，而 primary 写在文件里 —— 先把受影响的版本记下来
	scopes := make(map[string][2]int)
	for _, eid := range req.EntityIDs {
		if bookID, vid := h.graphRowScope("graph_entities", eid); bookID > 0 {
			v := 0
			if vid != nil {
				v = *vid
			}
			scopes[bookstore.GraphDirtyKey(v, bookID)] = [2]int{bookID, v}
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	released := 0
	seen := make(map[int]bool, len(req.EntityIDs))
	for _, id := range req.EntityIDs {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true

		// 自己名下的副词条全部解绑（等于解散这一组）
		res, err := tx.Exec("UPDATE graph_entities SET primary_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE primary_id = ?", id)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if n, _ := res.RowsAffected(); n > 0 {
			released += int(n)
		}

		// 自己若本身是别人的副词条，也一并解绑
		res, err = tx.Exec("UPDATE graph_entities SET primary_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND primary_id IS NOT NULL", id)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if n, _ := res.RowsAffected(); n > 0 {
			released += int(n)
		}
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	for _, sc := range scopes {
		bookID, v := sc[0], sc[1]
		if v > 0 {
			h.markGraphDirty(bookID, &v)
		} else {
			h.markGraphDirty(bookID, nil)
		}
	}
	respondOK(w, map[string]interface{}{"released": released})
}

// GetMergeCandidates 找出该范围内疑似重复的主词条对，供人工合并检测使用。
// GET /api/knowledge-graph/merge-candidates/{bookId}?versionId=10
// 已合并的副词条不参与（它们已经在某人名下了）。
func (h *Handler) GetMergeCandidates(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	entityWhere := "book_id = ? AND primary_id IS NULL"
	relationWhere := "book_id = ?"
	entityArgs := []interface{}{bookID}
	relationArgs := []interface{}{bookID}
	if raw := r.URL.Query().Get("versionId"); raw != "" {
		if versionID, err := strconv.Atoi(raw); err == nil && versionID > 0 {
			entityWhere += " AND version_id = ?"
			relationWhere += " AND version_id = ?"
			entityArgs = append(entityArgs, versionID)
			relationArgs = append(relationArgs, versionID)
		}
	}

	rows, err := h.db.Query("SELECT "+graphEntityColumns+" FROM graph_entities WHERE "+entityWhere+" ORDER BY id", entityArgs...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var entities []models.GraphEntity
	inScope := make(map[int]bool)
	for rows.Next() {
		var e models.GraphEntity
		if err := scanGraphEntity(rows, &e); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		entities = append(entities, e)
		inScope[e.ID] = true
	}
	if err := rows.Err(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// 共同邻居用的邻接表（只统计两端都在候选集合内的边）
	neighbors := make(map[int]map[int]bool, len(entities))
	relRows, err := h.db.Query("SELECT source_id, target_id FROM graph_relations WHERE "+relationWhere, relationArgs...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer relRows.Close()
	for relRows.Next() {
		var sourceID, targetID int
		if err := relRows.Scan(&sourceID, &targetID); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if !inScope[sourceID] || !inScope[targetID] {
			continue
		}
		if neighbors[sourceID] == nil {
			neighbors[sourceID] = make(map[int]bool)
		}
		if neighbors[targetID] == nil {
			neighbors[targetID] = make(map[int]bool)
		}
		neighbors[sourceID][targetID] = true
		neighbors[targetID][sourceID] = true
	}

	candidates := make([]models.EntityMergeCandidate, 0)
	for i := 0; i < len(entities); i++ {
		for j := i + 1; j < len(entities); j++ {
			score, reasons, suggestPrimary, suggestDuplicate := scoreCandidatePair(entities[i], entities[j], neighbors)
			if score <= 0 {
				continue
			}
			candidates = append(candidates, models.EntityMergeCandidate{
				A:                  entities[i],
				B:                  entities[j],
				Score:              score,
				Reasons:            reasons,
				SuggestPrimaryID:   suggestPrimary,
				SuggestDuplicateID: suggestDuplicate,
			})
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })
	respondOK(w, candidates)
}
