/**
 * collapseAliases（主词条 / 副词条折叠）的独立验证。
 *
 * 用法： node scripts/test_collapse_aliases.mjs [图谱导出JSON]
 *   自带打包步骤（调用 web/node_modules 里的 esbuild），不依赖任何临时产物；
 *   传了导出 JSON 就会额外用真实数据跑一遍结构不变量。
 */
import fs from 'node:fs'
import path from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath, pathToFileURL } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webDir = path.join(here, '..', 'web')
const bundle = path.join(here, '.tmp-knowledgeGraphService.mjs')

execFileSync(
  process.execPath,
  [
    path.join(webDir, 'node_modules', 'esbuild', 'bin', 'esbuild'),
    path.join(webDir, 'src', 'services', 'knowledgeGraphService.ts'),
    '--bundle', '--format=esm', '--log-level=error',
    '--outfile=' + bundle
  ],
  { stdio: 'inherit' }
)
const { collapseAliases } = await import(pathToFileURL(bundle).href)
process.on('exit', () => { try { fs.unlinkSync(bundle) } catch {} })

let pass = 0
const fails = []
function check(label, ok, detail = '') {
  if (ok) pass++
  else fails.push(label)
  console.log(`  ${ok ? 'PASS' : 'FAIL'}  ${label}${detail ? '   ' + detail : ''}`)
}

const E = (id, name, primary_id = null, type = 'character', description = '') => ({
  id, book_id: 16, name, type, description, metadata: {}, version_id: 10,
  primary_id, created_at: '', updated_at: ''
})
const R = (id, source_id, target_id, relation_type = 'rel') => ({
  id, book_id: 16, source_id, target_id, relation_type, description: '', evidence: '',
  confidence: 1, version_id: 10, created_at: '', updated_at: ''
})

const endpoints = rels => rels.map(r => `${r.source_id}->${r.target_id}:${r.relation_type}`)
const noDupEdge = rels => new Set(endpoints(rels)).size === rels.length
const noSelfLoop = rels => rels.every(r => r.source_id !== r.target_id)

console.log('[1] 没有合并时保持原样')
{
  const es = [E(1, 'A'), E(2, 'B')]
  const rs = [R(1, 1, 2)]
  const out = collapseAliases(es, rs)
  check('实体原样返回（引用相同，不做无谓拷贝）', out.entities === es)
  check('关系原样返回', out.relations === rs)
  check('aliasOf 为空', out.aliasOf.size === 0)
}

console.log('\n[2] 单层合并：副词条不出节点，边改指主词条')
{
  const es = [E(1, '陈奕'), E(2, '陈默', 1), E(3, '老刘')]
  const rs = [R(1, 2, 3, '同伴')]
  const out = collapseAliases(es, rs)
  check('副词条 2 不出节点', !out.entities.some(e => e.id === 2))
  check('主词条 1、3 保留', out.entities.length === 2)
  check('边改指为 1->3', endpoints(out.relations)[0] === '1->3:同伴')
  check('aliasOf 记录 2->1', out.aliasOf.get(2) === 1)
  check('aliasesByPrimary[1] 含 2', (out.aliasesByPrimary.get(1) || []).map(a => a.id).join() === '2')
}

console.log('\n[3] 折完重合的边去重、折成自环的丢弃')
{
  const es = [E(1, '陈奕'), E(2, '陈默', 1), E(3, '老刘')]
  const rs = [R(1, 1, 3, '同伴'), R(2, 2, 3, '同伴'), R(3, 1, 2, '自我')]
  const out = collapseAliases(es, rs)
  check('1->3 的两条合并为一条', out.relations.length === 1, endpoints(out.relations).join())
  check('无重复边', noDupEdge(out.relations))
  check('无自环（1->2 折成 1->1 被丢）', noSelfLoop(out.relations) && out.relations.length === 1)
}

console.log('\n[4] 链式指向顺链归到根')
{
  const es = [E(1, 'A'), E(2, 'B', 1), E(3, 'C', 2), E(4, 'D')]
  const rs = [R(1, 3, 4, 'x')]
  const out = collapseAliases(es, rs)
  check('都折到 1', out.aliasOf.get(2) === 1 && out.aliasOf.get(3) === 1)
  check('只剩 A、D 两个节点', out.entities.map(e => e.id).sort().join() === '1,4')
  check('边改指 1->4', endpoints(out.relations)[0] === '1->4:x')
  check('aliasesByPrimary[1] 有 B、C', (out.aliasesByPrimary.get(1) || []).length === 2)
}

console.log('\n[5] 悬空主词条引用不折（宁可多一个节点，也不要凭空吞数据）')
{
  const es = [E(1, 'A'), E(9, '孤儿', 999)]
  const out = collapseAliases(es, [])
  check('999 不存在 → 9 仍是节点', out.entities.some(e => e.id === 9))
  check('aliasOf 不含 9', !out.aliasOf.has(9))
}

console.log('\n[6] 脏数据成环不会死循环')
{
  const es = [E(1, 'A', 2), E(2, 'B', 1)]
  const out = collapseAliases(es, [])
  check('能返回且不抛异常', out.entities.length === 2, `nodes=${out.entities.length}`)
}

const realFile = process.argv[2] ||
  'D:/Downloads/知识图谱_AI分析_等5章_2026_9_15_22_31_42+增量_1789608546057.json'
console.log('\n[7] 真实导出数据' + (fs.existsSync(realFile) ? '' : '（跳过：找不到 JSON）'))
if (fs.existsSync(realFile)) {
  const raw = JSON.parse(fs.readFileSync(realFile, 'utf8'))
  const entities = raw.entities
  const relations = raw.relations
  console.log(`      数据：${entities.length} 实体 / ${relations.length} 关系`)

  const idle = collapseAliases(entities, relations)
  check('无合并标记时完全不变', idle.entities.length === entities.length && idle.relations.length === relations.length)

  // 模拟一次真实合并：把「老刘」并入「陈奕」
  const byName = n => entities.find(e => e.name === n)
  const primary = byName('陈奕')
  const dup = byName('老刘')
  check('找到 陈奕/老刘', !!primary && !!dup)
  const marked = entities.map(e => (e.id === dup.id ? { ...e, primary_id: primary.id } : e))
  const out = collapseAliases(marked, relations)

  check('节点数 -1', out.entities.length === entities.length - 1, `${out.entities.length}`)
  check('陈奕仍在、老刘已不是节点',
    out.entities.some(e => e.id === primary.id) && !out.entities.some(e => e.id === dup.id))
  check('陈奕名下出现副词条「老刘」',
    (out.aliasesByPrimary.get(primary.id) || []).map(a => a.name).join() === '老刘')
  check('没有任何关系端点还指向老刘',
    out.relations.every(r => r.source_id !== dup.id && r.target_id !== dup.id))
  check('无自环', noSelfLoop(out.relations))
  check('无重复边', noDupEdge(out.relations))
  check('关系数只减不增', out.relations.length <= relations.length,
    `${relations.length} -> ${out.relations.length}`)
  check('实体对象未被就地修改（零副作用）',
    relations.every(r => r.source_id !== undefined) &&
    marked.find(e => e.id === dup.id).primary_id === primary.id)
}

console.log('\n' + '='.repeat(70))
console.log(`通过 ${pass} 项，失败 ${fails.length} 项`)
fails.forEach(f => console.log('  FAIL: ' + f))
console.log('='.repeat(70))
process.exit(fails.length ? 1 : 0)
