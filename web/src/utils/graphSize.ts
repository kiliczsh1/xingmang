/**
 * 节点半径：按度数缩放。
 *
 * ── 为什么必须按度数缩放 ──
 *
 * 固定大小（所有节点一样大）的图**没有层次感**：读者看不出谁是主角、
 * 谁是配角，只能靠位置和颜色猜。而"谁是主角"在这类图谱里就是
 * "谁的连线最多"—— 枢纽节点。
 *
 * ── 为什么用 √ 而不是线性或对数 ──
 *
 * 知识图谱的度数分布极度右偏（实测 200 实体：p50 = 2、p75 = 4、
 * 只有 5 个超过 20、max 75）。三种曲线在这里的表现：
 *
 * - 线性 `d / dmax`：degree 2 的节点只有 max 的 3%，几乎和 degree 0 一样大，
 *   而 5 个枢纽占满整个量程 → 99% 的节点挤在最小端，层次全丢
 * - 对数 `log2(1+d)`：低度数段拉太开（0→1 就跳掉 1/4 量程），
 *   反而显得小节点之间大小不一
 * - **√ `√(d/dmax)`**：低度数段温和（0、1、2、3 依次 0%、33%、47%、52%），
 *   枢纽明显跃出（100%）—— 既保住了"主角大配角小"，又不会让配角之间乱跳
 *
 * ── 为什么上下限要写死 ──
 *
 * 锚点是"本图最大度数"，所以不同图的大小分布会跟着数据变。
 * 但**绝对尺寸**必须固定，否则小图里的节点会大得离谱、挤成一团。
 * 上下限按参照图的实际观感标定。
 */

/** 最小半径（图坐标单位）—— degree 0 的孤立节点 */
export const NODE_SIZE_MIN = 9
/** 最大半径（图坐标单位）—— 本图度数最高的枢纽 */
export const NODE_SIZE_MAX = 18
/**
 * 平均半径的近似值。
 *
 * 力参数（连线长度、碰撞半径）需要"一个典型节点的尺寸"做基准，
 * 写死一个与上下限一致的值，比动态求平均更稳定（换书时参数不会跳）。
 */
export const NODE_RADIUS_TYPICAL = 13

/**
 * 把度数映射成节点半径。
 *
 * @param degree 实体 id → 度数
 * @returns 实体 id → 半径
 */
export function computeNodeSizes(degree: Map<number, number>): Map<number, number> {
  let dMax = 0
  for (const d of degree.values()) {
    if (d > dMax) dMax = d
  }

  const sizes = new Map<number, number>()
  // 整图无关系时（dMax = 0）全部给最小值，别让除零产生 NaN
  if (dMax <= 0) {
    for (const id of degree.keys()) sizes.set(id, NODE_SIZE_MIN)
    return sizes
  }

  const denom = Math.sqrt(dMax)
  for (const [id, d] of degree) {
    const ratio = Math.sqrt(Math.max(0, d)) / denom
    sizes.set(id, NODE_SIZE_MIN + (NODE_SIZE_MAX - NODE_SIZE_MIN) * ratio)
  }
  return sizes
}
