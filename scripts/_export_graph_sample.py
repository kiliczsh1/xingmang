"""一次性脚本：导出 book 16 最新版本的图谱到 JSON，供标定力参数使用。用完即删。"""
import json
import sqlite3

DB = r'D:\xingguang\xingmang-desktop\xingmang.db'
OUT = r'D:\xingguang\xingmang-desktop\scripts\_graph_sample.json'

con = sqlite3.connect(DB)
con.row_factory = sqlite3.Row
cur = con.cursor()

tables = [r[0] for r in cur.execute("select name from sqlite_master where type='table'")]
print('tables:', tables)

for table in ('graph_entities', 'graph_relations', 'graph_versions'):
    if table in tables:
        cols = [r[1] for r in cur.execute('PRAGMA table_info(%s)' % table)]
        print(table, '->', cols)

# 选实体最多的那个版本
row = cur.execute(
    """
    select book_id, version_id, count(*) as n
    from graph_entities
    where version_id is not null
    group by book_id, version_id
    order by n desc
    limit 1
    """
).fetchone()

if row is None:
    print('NO VERSIONED DATA')
    raise SystemExit(1)

book_id, version_id, n = row['book_id'], row['version_id'], row['n']
print('picked book=%s version=%s entities=%s' % (book_id, version_id, n))

entities = [
    {'id': r['id'], 'name': r['name'], 'type': r['type'], 'primary_id': r['primary_id']}
    for r in cur.execute(
        'select id, name, type, primary_id from graph_entities where book_id=? and version_id=?',
        (book_id, version_id),
    )
]
relations = [
    {'id': r['id'], 'source_id': r['source_id'], 'target_id': r['target_id']}
    for r in cur.execute(
        'select id, source_id, target_id from graph_relations where book_id=? and version_id=?',
        (book_id, version_id),
    )
]

with open(OUT, 'w', encoding='utf-8') as f:
    json.dump(
        {
            'book_id': book_id,
            'version_id': version_id,
            'entities': entities,
            'relations': relations,
        },
        f,
        ensure_ascii=False,
    )

print('wrote %s : %d entities, %d relations' % (OUT, len(entities), len(relations)))
con.close()
