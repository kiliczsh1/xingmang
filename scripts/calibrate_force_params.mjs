/**
 * 力参数标定：用真实图谱数据跑 d3-force，量测布局质量，用来确定
 * useForceLayout.ts 里的 CENTER_TO_STRENGTH / REPEL_TO_STRENGTH 系数。
 *
 * 为什么必须标定：
 *   Obsidian 的滑块值（repel 10 / distance 250）不能直接喂给 d3 —— 两者"单位力度"
 *   差一个量级；而 sigma 的 autoRescale 又让绝对像素失去意义。所以只能按
 *   「节点半径的倍数」定标，再用统计量验证底线：节点不重叠 + 留白足够。
 *
 * 核心指标 gapRatio = 中位最近邻距离 ÷ 节点直径：
 *   ≥3 视觉上明显分开（清爽）· 1~2 偏挤 · <1 已经重叠。
 * 次核心指标 crossings = 边交叉数（清爽与否的直接量度）。
 *
 * 数据来源：同目录的 _graph_sample.json，由 _export_graph_sample.py 从 xingmang.db 导出
 * （当前样本：book 16 / version 10，60 实体 / 63 关系）。换样本前先跑那个脚本。
 *
 * 用法：node scripts/calibrate_force_params.mjs
 */
import {
  forceCollide,
  forceLink,
  forceManyBody,
  forceSimulation,
  forceX,
  forceY
} from '../web/node_modules/d3-force/src/index.js'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'

const here = dirname(fileURLToPath(import.meta.url))
const SAMPLE = join(here, '_graph_sample.json')

const NODE_RADIUS = 7
const NODE_DIAMETER = NODE_RADIUS * 2
const CANVAS = { width: 900, height: 640 }

const data = JSON.parse(readFileSync(SAMPLE, 'utf8'))
const entities = data.entities
const relations = data.relations

console.log(
  `样本：book ${data.book_id} v${data.version_id} — ${entities.length} 实体 / ${relations.length} 关系\n`
)

/** 与前端 computeSeedLayout 完全一致的黄金角螺旋（否则标定结果对不上真实运行） */
function seedLayout(list, width, height) {
  const radius = Math.max(60, Math.min(width, height) / 2 - 48)
  const goldenAngle = Math.PI * (3 - Math.sqrt(5))
  const ordered = [...list].sort((a, b) => a.id - b.id)
  return ordered.map((entity, i) => ({
    id: String(entity.id),
    x: radius * Math.sqrt((i + 0.5) / ordered.length) * Math.cos(i * goldenAngle),
    y: radius * Math.sqrt((i + 0.5) / ordered.length) * Math.sin(i * goldenAngle)
  }))
}

function run(nodesSeed, { repelStrength, centerStrength, linkDistance, linkForce = 1 }) {
  const nodes = nodesSeed.map(n => ({ ...n }))
  const byId = new Map(nodes.map(n => [n.id, n]))
  const degree = new Map(nodes.map(n => [n.id, 0]))
  const links = []

  for (const rel of relations) {
    const s = String(rel.source_id)
    const t = String(rel.target_id)
    if (!byId.has(s) || !byId.has(t) || s === t) continue
    links.push({ source: s, target: t })
    degree.set(s, degree.get(s) + 1)
    degree.set(t, degree.get(t) + 1)
  }

  const sim = forceSimulation(nodes)
    .force(
      'link',
      forceLink(links)
        .id(d => d.id)
        .distance(linkDistance)
        .strength(l => linkForce / Math.max(1, Math.min(degree.get(l.source.id), degree.get(l.target.id))))
    )
    .force('charge', forceManyBody().strength(-repelStrength).theta(0.9))
    .force('x', forceX(0).strength(centerStrength))
    .force('y', forceY(0).strength(centerStrength))
    .force('collide', forceCollide(NODE_RADIUS * 1.8).strength(0.9))
    .velocityDecay(0.6)
    .stop()

  let ticks = 0
  while (sim.alpha() > 0.001 && ticks < 1200) {
    sim.tick()
    ticks++
  }

  return { nodes, links, ticks }
}

const quantile = (sorted, p) =>
  sorted[Math.min(sorted.length - 1, Math.max(0, Math.floor(sorted.length * p)))]

function measure({ nodes, links, ticks }) {
  const n = nodes.length
  const nearest = []
  for (let i = 0; i < n; i++) {
    let best = Infinity
    for (let j = 0; j < n; j++) {
      if (i === j) continue
      const d = Math.hypot(nodes[i].x - nodes[j].x, nodes[i].y - nodes[j].y)
      if (d < best) best = d
    }
    nearest.push(best)
  }
  nearest.sort((a, b) => a - b)

  const cx = nodes.reduce((s, d) => s + d.x, 0) / n
  const cy = nodes.reduce((s, d) => s + d.y, 0) / n
  const radii = nodes.map(d => Math.hypot(d.x - cx, d.y - cy)).sort((a, b) => a - b)

  const edgeLengths = links
    .map(l => Math.hypot(l.source.x - l.target.x, l.source.y - l.target.y))
    .sort((a, b) => a - b)

  // 边交叉数：清爽与否的直接量度（交叉多 = 乱），O(E²) 对 63 条边无压力
  const crossed = (a, b, c, d) => {
    const cross = (o, p, q) => (p.x - o.x) * (q.y - o.y) - (p.y - o.y) * (q.x - o.x)
    const d1 = cross(a, b, c)
    const d2 = cross(a, b, d)
    const d3 = cross(c, d, a)
    const d4 = cross(c, d, b)
    return ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) && ((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0))
  }
  let crossings = 0
  for (let i = 0; i < links.length; i++) {
    for (let j = i + 1; j < links.length; j++) {
      const a = links[i].source
      const b = links[i].target
      const c = links[j].source
      const d = links[j].target
      // 共享端点的边不算交叉
      if (a === c || a === d || b === c || b === d) continue
      if (crossed(a, b, c, d)) crossings++
    }
  }

  return {
    ticks,
    radius: quantile(radii, 0.95),
    nnMin: nearest[0],
    nnP10: quantile(nearest, 0.1),
    nnMedian: quantile(nearest, 0.5),
    edgeMedian: quantile(edgeLengths, 0.5),
    overlaps: nearest.filter(d => d < NODE_DIAMETER).length,
    gapRatio: quantile(nearest, 0.5) / NODE_DIAMETER,
    crossings
  }
}

const seed = seedLayout(entities, CANVAS.width, CANVAS.height)
// 与 useForceLayout.ts 的 LINK_DISTANCE_RATIO 保持一致
const LINK_DISTANCE_RATIO = 20
const linkDistance = NODE_RADIUS * LINK_DISTANCE_RATIO

console.log(`固定 linkDistance = ${linkDistance}（= 节点半径 × ${LINK_DISTANCE_RATIO}）`)
console.log(`centerStrength = 0.52 × 0.2 = 0.104\n`)
console.log('repelStrength |  收敛tick | 图半径 | 最近邻 min | p10 | median | 重叠数 | gapRatio | 交叉数')
console.log('--------------|-----------|--------|------------|-----|--------|--------|----------|-------')

const rows = []
for (const repelStrength of [30, 60, 120, 200, 300, 450, 600, 900]) {
  const result = measure(run(seed, { repelStrength, centerStrength: 0.104, linkDistance }))
  rows.push({ repelStrength, ...result })
  const r = result
  console.log(
    `${String(repelStrength).padStart(13)} |` +
      `${String(r.ticks).padStart(10)} |` +
      `${r.radius.toFixed(0).padStart(7)} |` +
      `${r.nnMin.toFixed(0).padStart(11)} |` +
      `${r.nnP10.toFixed(0).padStart(4)} |` +
      `${r.nnMedian.toFixed(0).padStart(7)} |` +
      `${String(r.overlaps).padStart(7)} |` +
      `${r.gapRatio.toFixed(2).padStart(9)} |` +
      `${String(r.crossings).padStart(6)}`
  )
}

console.log('\n反馈到 UI 刻度（repelForce 滑块 0–20，默认 10）：')
console.log('REPEL_TO_STRENGTH = repelStrength ÷ 10')
console.log(
  rows
    .map(r => `  repelForce 10 → charge ${r.repelStrength}（gapRatio ${r.gapRatio.toFixed(2)}）`)
    .join('\n')
)

console.log(`\n—— centerForce 扫描（固定 charge 300 / linkDistance ${linkDistance}）——`)
console.log('centerStrength | 图半径 | gapRatio | 说明')
for (const centerStrength of [0, 0.052, 0.104, 0.156, 0.2]) {
  const r = measure(run(seed, { repelStrength: 300, centerStrength, linkDistance }))
  const note =
    centerStrength === 0.104 ? '← 默认（0.52 × 0.2）' : centerStrength === 0.2 ? '压成圆饼的风险区' : ''
  console.log(
    `${centerStrength.toFixed(3).padStart(14)} | ${r.radius.toFixed(0).padStart(7)} | ` +
      `${r.gapRatio.toFixed(2).padStart(8)} | ${note}`
  )
}

console.log('\n—— linkDistance 扫描（固定 charge 300 / centerStrength 0.104）——')
console.log('linkDistance | 图半径 | gapRatio | 边长中位 | 边交叉数 | 边长/半径')
for (const dist of [70, 140, 210, 280, 400, 500]) {
  const r = measure(run(seed, { repelStrength: 300, centerStrength: 0.104, linkDistance: dist }))
  console.log(
    `${String(dist).padStart(12)} | ${r.radius.toFixed(0).padStart(7)} | ` +
      `${r.gapRatio.toFixed(2).padStart(8)} | ${r.edgeMedian.toFixed(0).padStart(9)} | ` +
      `${String(r.crossings).padStart(8)} | ${(r.edgeMedian / r.radius).toFixed(2).padStart(9)}` +
      (dist === linkDistance ? '  ← 默认' : '')
  )
}
