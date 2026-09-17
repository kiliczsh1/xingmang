"""
book 16 库内实体去重 / 别名归并（一次性数据修复）。

规则（对齐 docs/graph-upgrade-plan.md 的查重口径）：
  - 实体按 normalizeName(name) 查重，不看 type；同组保留 **最小 id**（最旧的那条，因其通常已有关系接入）
  - chapter_refs 取并集；description 取更长的一份
  - 关系：把指向被并入实体的端点改指到保留实体；改指后若与既有关系 (source,target,type) 撞车，
    则把章节关联并过去并丢弃这条；合并后成自环的直接丢弃
  - 同版本内的别名对（人工确认过）按 ALIASES 归并
  - 删除实体后按实际行数回写 graph_versions 计数

用法：python dedupe_graph_entities.py            # 只打印计划
      python dedupe_graph_entities.py --apply   # 执行（务必先备份 db）
"""
import collections
import re
import sqlite3
import sys

DB = r"D:\xingguang\xingmang-desktop\xingmang.db"
BOOK = 16

# 同版本内的别名归并：别名 -> 规范名。v9 / v10 都存在同一对（许玲珑的仇人即背叛飞虎寨的二当家）。
ALIASES = {"二当家": "飞虎寨二当家"}

# 类型对齐：飞虎寨在语义上是山贼势力，v4/v9/v10 的同一实体都是 faction，只有 v2 早期误标成 location
TYPE_FIX = {"飞虎寨": "faction"}

APPLY = "--apply" in sys.argv


def norm(s):
    return re.sub(r"\s+", " ", (s or "").strip().lower())


con = sqlite3.connect(DB)
con.row_factory = sqlite3.Row
cur = con.cursor()

rows = cur.execute(
    "SELECT id, name, type, description, book_id, version_id FROM graph_entities WHERE book_id = ? ORDER BY version_id, id",
    (BOOK,),
).fetchall()

groups = collections.defaultdict(list)
for r in rows:
    key = norm(r["name"])
    key = norm(ALIASES.get(r["name"], ALIASES.get(key, key)))
    groups[(r["version_id"], key)].append(r)

plan = []
for (vid, key), members in sorted(groups.items()):
    if len(members) < 2:
        continue
    members.sort(key=lambda m: m["id"])
    plan.append((vid, members[0], members[1:]))

print("=" * 74)
print("合并计划（book %d）" % BOOK)
print("=" * 74)
for vid, keep, drops in plan:
    print("版本 %s: 保留 id=%d %r (%s)" % (vid, keep["id"], keep["name"], keep["type"]))
    for d in drops:
        print("           并入 id=%d %r (%s)  desc=%r" % (d["id"], d["name"], d["type"], (d["description"] or "")[:30]))
    fix = TYPE_FIX.get(keep["name"])
    if fix and keep["type"] != fix:
        print("           类型对齐: %s -> %s" % (keep["type"], fix))

if not plan:
    print("没有需要合并的组。")
    sys.exit(0)

if not APPLY:
    print()
    print("（干跑结束，未写入。加 --apply 执行）")
    sys.exit(0)


def union_refs(kind, keep_id, drop_id):
    """把 drop 实体的章节关联并到 keep；返回新增行数。"""
    col = "entity_id" if kind == "entity" else "relation_id"
    have = {r[0] for r in cur.execute("SELECT chapter_id FROM graph_chapter_refs WHERE %s = ?" % col, (keep_id,))}
    src = cur.execute(
        "SELECT book_id, version_id, chapter_id FROM graph_chapter_refs WHERE %s = ?" % col, (drop_id,)
    ).fetchall()
    added = 0
    for s in src:
        if s["chapter_id"] in have:
            continue
        have.add(s["chapter_id"])
        cur.execute(
            "INSERT INTO graph_chapter_refs (book_id, version_id, chapter_id, entity_id, relation_id) VALUES (?, ?, ?, ?, ?)",
            (s["book_id"], s["version_id"], s["chapter_id"],
             keep_id if kind == "entity" else None,
             keep_id if kind == "relation" else None),
        )
        added += 1
    return added


def redirect_relations(keep_id, drop_id):
    """把指向 drop 的关系改指到 keep；返回 (改指数, 并掉数, 丢弃自环数)"""
    moved = merged = selfloop = 0
    rels = cur.execute(
        "SELECT * FROM graph_relations WHERE source_id = ? OR target_id = ?", (drop_id, drop_id)
    ).fetchall()
    for rel in rels:
        ns = keep_id if rel["source_id"] == drop_id else rel["source_id"]
        nt = keep_id if rel["target_id"] == drop_id else rel["target_id"]
        if ns == nt:
            cur.execute("DELETE FROM graph_chapter_refs WHERE relation_id = ?", (rel["id"],))
            cur.execute("DELETE FROM graph_relations WHERE id = ?", (rel["id"],))
            selfloop += 1
            continue
        dup = cur.execute(
            "SELECT id FROM graph_relations WHERE book_id = ? AND source_id = ? AND target_id = ? AND relation_type = ? AND id <> ?",
            (rel["book_id"], ns, nt, rel["relation_type"], rel["id"]),
        ).fetchone()
        if dup:
            union_refs("relation", dup["id"], rel["id"])
            cur.execute("DELETE FROM graph_chapter_refs WHERE relation_id = ?", (rel["id"],))
            cur.execute("DELETE FROM graph_relations WHERE id = ?", (rel["id"],))
            merged += 1
        else:
            cur.execute("UPDATE graph_relations SET source_id = ?, target_id = ? WHERE id = ?", (ns, nt, rel["id"]))
            moved += 1
    return moved, merged, selfloop


before = {
    "entities": cur.execute("SELECT COUNT(*) FROM graph_entities WHERE book_id = ?", (BOOK,)).fetchone()[0],
    "relations": cur.execute("SELECT COUNT(*) FROM graph_relations WHERE book_id = ?", (BOOK,)).fetchone()[0],
    "uniquenames": len({norm(r["name"]) for r in rows}),
}

affected = set()
print()
print("=" * 74)
print("执行")
print("=" * 74)
for vid, keep, drops in plan:
    k = keep["id"]
    best_desc = keep["description"] or ""
    for d in drops:
        added = union_refs("entity", k, d["id"])
        moved, merged, selfloop = redirect_relations(k, d["id"])
        if len(d["description"] or "") > len(best_desc):
            best_desc = d["description"]
        cur.execute("DELETE FROM graph_chapter_refs WHERE entity_id = ?", (d["id"],))
        cur.execute("DELETE FROM graph_entities WHERE id = ?", (d["id"],))
        affected.add(vid)
        print(
            "  v%s: %d(%s) -> %d(%s)  refs+%d  关系改指%d/并掉%d/自环%d"
            % (vid, d["id"], d["name"], k, keep["name"], added, moved, merged, selfloop)
        )
    cur.execute("UPDATE graph_entities SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", (best_desc, k))
    fix = TYPE_FIX.get(keep["name"])
    if fix and keep["type"] != fix:
        cur.execute("UPDATE graph_entities SET type = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", (fix, k))
        print("  类型对齐: id=%d %s -> %s" % (k, keep["type"], fix))

for vid in sorted(affected):
    cur.execute(
        """UPDATE graph_versions SET
             entity_count   = (SELECT COUNT(*) FROM graph_entities  WHERE version_id = ?),
             relation_count = (SELECT COUNT(*) FROM graph_relations WHERE version_id = ?),
             updated_at     = CURRENT_TIMESTAMP
           WHERE id = ?""",
        (vid, vid, vid),
    )

con.commit()

after = {
    "entities": cur.execute("SELECT COUNT(*) FROM graph_entities WHERE book_id = ?", (BOOK,)).fetchone()[0],
    "relations": cur.execute("SELECT COUNT(*) FROM graph_relations WHERE book_id = ?", (BOOK,)).fetchone()[0],
    "uniquenames": len({norm(r[0]) for r in cur.execute("SELECT name FROM graph_entities WHERE book_id = ?", (BOOK,))}),
}

print()
print("=" * 74)
print("结果")
print("=" * 74)
print("实体行数   : %d -> %d" % (before["entities"], after["entities"]))
print("关系行数   : %d -> %d" % (before["relations"], after["relations"]))
print("不同实体名 : %d -> %d" % (before["uniquenames"], after["uniquenames"]))
print()
print("各版本剩余实体/关系（与 graph_versions 计数对照）：")
for r in cur.execute(
    """SELECT v.id, v.name, v.entity_count, v.relation_count,
              (SELECT COUNT(*) FROM graph_entities  WHERE version_id = v.id) AS real_e,
              (SELECT COUNT(*) FROM graph_relations WHERE version_id = v.id) AS real_r,
              (SELECT COUNT(DISTINCT LOWER(TRIM(name))) FROM graph_entities WHERE version_id = v.id) AS uniq
         FROM graph_versions v WHERE v.book_id = ? ORDER BY v.id""",
    (BOOK,),
):
    print(
        "  v%-3s %-34s 计数 %s/%s  实际 %s/%s  唯一名 %s  %s"
        % (r["id"], (r["name"] or "")[:32], r["entity_count"], r["relation_count"],
           r["real_e"], r["real_r"], r["uniq"],
           "OK" if r["entity_count"] == r["real_e"] and r["relation_count"] == r["real_r"] else "!! 不一致")
    )
