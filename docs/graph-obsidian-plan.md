# 让知识图谱达到「Obsidian 观感」的实施计划

目标不是"抄参数"，而是把 Obsidian 的观感拆成**可验收的能力项**，逐项对齐。

---

## 一、把「Obsidian 效果」拆成 6 项能力

| # | 能力 | 项目现状 | 差距 |
|---|---|---|---|
| 1 | 力导向连续模拟（三力 + 碰撞），四个参数可调 | FA2 跑 100 次迭代后即静止，参数只有 gravity/scalingRatio | 缺 |
| 2 | 拖拽节点时"钉住 + 邻居实时弹性响应"，松手继续演化 | sigma 默认拖拽只挪动单个节点 | 缺 |
| 3 | 悬停/选中时压暗无关节点 | 已有（`nodeReducer` 压暗 + 邻接集合） | 已达标 |
| 4 | 视觉降噪：小节点、淡细边、标签按缩放渐显且无底 | 标签全量渲染 + 白底药丸 + 箭头 + 边标签默认开 | 差距最大 |
| 5 | 收敛后几乎完全静止（高阻尼） | 静态布局，天然满足 | 已达标 |
| 6 | 缩放平移不触发重排 | 静态布局，天然满足 | 已达标 |

**结论**：真正要补的是 **1 / 2 / 4**。其中 4 零风险、立刻见效；1 是核心；2 依赖 1。

---

## 二、现状盘点（核实过的位置）

| 文件 | 行 | 内容 |
|---|---|---|
| `web/src/composables/useGraphLayout.ts` | 全 | 七层类型环纯函数（角色→地点→势力→事件→物件→功法→线索），输出最终坐标 |
| `web/src/components/graph/GraphCanvas.vue` | 302 | `NODE_SIZE = 7`（sigma 的 size 是**半径**，直径 14） |
| 同上 | 313-316 | `computeCurvature`：单边也统一弯 `EDGE_CURVATURE = 0.2` |
| 同上 | 387-399 | `relaxLayout()`：FA2 100 迭代，`gravity 4 / scalingRatio 1.2 / slowDown 1 / adjustSizes` |
| 同上 | 210-238 | `drawNodeLabelPill`：标签带白底 + 边框（Obsidian 是纯文字） |
| 同上 | 561-580 | `initSigma`：`labelSize 13`、`minEdgeThickness 0.4`、**未设 `labelRenderedSizeThreshold`** |
| 同上 | 561-580 | `renderEdgeLabels: props.showEdgeLabels`，prop 默认 `true` |
| `web/src/components/graph/GraphToolbar.vue` | 全 | 已有：缩放/适配/重置/边标签开关/连线模式/AI 分析/刷新。**无力度参数入口** |
| `web/package.json` | 33-39 | `d3-force@^3.0.0`、`@types/d3-force` **已装但零引用**（迁移遗留） |

### 两个必须先纠正的认知

1. **sigma 会归一化图坐标**。`autoRescale` 把位置压到 `[0,1]` 空间，相机也基于该空间。
   所以 Obsidian 的 `linkDistance = 250` **不能照抄**——绝对值会被归一化掉。
   **守恒的是比例**：Obsidian 节点半径 4–8px、边长 250px → `linkDistance ÷ 半径 ≈ 40`。
   换算到本项目 `NODE_SIZE = 7` → **`linkDistance ≈ 280`**。
2. **FA2 没有 `linkDistance` 概念**（边长由 `scalingRatio` + 度数隐式决定），
   所以想复刻那四个参数**只能用 d3-force**。好消息是不用加依赖。

---

## 三、三条路线

### 路线 A：只做视觉降噪（不动布局）

保留类型环 + FA2，仅对齐能力 4。

- 优点：**零风险**，不动任何语义，改 4~5 处常量，当天可验收
- 缺点：不解决"环布局必然规整、密集环必然挤"这一根本观感差异
- 适合：只想先"看起来干净一点"，不确定要不要动布局

### 路线 B：d3-force + 类型锚点力 ★推荐

换成 d3-force 连续模拟，但把**类型编码成一组"弱锚点力"**：

- 7 个类型各有一个锚点，**扇形分布**（不是现在的同心嵌套）
- 每个节点受 `forceX/forceY` 拉向自己类型的锚点，强度 `anchorStrength` 可调
- `anchorStrength = 0` → 退化成**纯 Obsidian 效果**
- `anchorStrength` 调高 → 逐步回到"类型可辨识"的分层感

- 优点：**两条路都不堵死**。一个滑块就能在"有机松弛"和"语义分层"之间滑动，
  不用在"好看"和"有用"之间二选一
- 缺点：多一层实现，调参需要耐心
- 初始位置仍用现有 `computeLayout`（类型环）做**种子** → 收敛快、结果可复现

### 路线 C：纯 d3-force，放弃类型环

最接近 Obsidian，但位置完全由力决定，**丢掉"类型一眼可辨"**——
那是 `useGraphLayout.ts` 注释里明确写的设计意图（"不漂、不乱"）。

- 适合：确定类型分层不是核心诉求，编辑场景以"看清结构关系"为主

---

## 四、分阶段实施（按路线 B 展开）

### P0 — 视觉降噪（独立可交付，与后续阶段无耦合）

| 改动 | 现值 | 目标值 | 理由 |
|---|---|---|---|
| `labelRenderedSizeThreshold` | 未设（默认 6） | **11** | `NODE_SIZE = 7` 越线导致标签全量渲染；调到 11 后标签只在放大时出现 |
| 标签样式 `drawNodeLabelPill` | 白底 + 边框（`0.92` 不透明） | 纯文字 + 轻微描边，无底 | Obsidian 是裸文字；药丸在 60+ 节点时形成"白点阵列" |
| 边透明度 | `#9aa4b2` / `#3f4a5c` | 加 alpha → `rgba(154,164,178,0.45)` | 边是背景结构，不该抢节点 |
| 边粗细 | `1.4` / `1.8`（人工）；`minEdgeThickness 0.4` | `1.0` / `1.6`；`minEdgeThickness 0.3` | 同上 |
| 箭头 | 常显（`EdgeCurvedArrowProgram`） | 默认不画，hover/选中时才显示 | Obsidian 默认 `showArrow: false` |
| 边标签 | prop 默认 `true` | prop 默认 `false` | 关系文字是最大的噪声源 |
| `computeCurvature` | 单边也弯 `0.2` | 单边**取直**（0），仅平行边扇形展开 | 全图统一弯曲会让密集区更缠绕 |

**验收**：同一份数据（book 16 / version 10，61 实体 / 63 关系）前后截图对比，
标签数量显著下降、边不再抢视线。

### P1 — 引入 d3-force（核心）

新建 `web/src/composables/useForceLayout.ts`：

```
graphology 图  ──构建──▶  d3 nodes/links 数组
                            │
                     forceSimulation
                       .force('link',   forceLink().id().distance(linkDistance).strength(linkForce))
                       .force('charge', forceManyBody().strength(-repelForce * K).theta(0.9))
                       .force('x',      forceX(cx).strength(centerForce))
                       .force('y',      forceY(cy).strength(centerForce))
                       .force('collide',forceCollide(NODE_SIZE * 1.6))
                            │ tick
                            ▼
             graph.setNodeAttribute(id, 'x'/'y') → sigma.refresh()
```

要点：

1. **节点 id 映射**：graphology 的 key 是 string（`String(e.id)`），d3 的 `forceLink().id()` 要用同一套 key
2. **初始位置**：直接用现有 `computeLayout()` 的类型环坐标做种子（收敛快、可复现）
3. **参数刻度沿用 Obsidian**：UI 里就是 `centerForce 0–1`、`repelForce 0–20`、`linkForce 0–1`、
   `linkDistance 30–500`，默认 `0.52 / 10 / 1 / 280`（280 = 40 × NODE_SIZE，见前文换算）。
   内部再做 `K` 系数映射到 d3 的实际量纲，**别把 Obsidian 的数字和 d3 的数字混着调**
4. **阻尼**：`velocityDecay 0.6`（d3 默认 0.4，调高才有 Obsidian 那种"稳")、`alphaDecay 0.0228`
5. **收敛后停机**：`on('end') → simulation.stop()`，避免空转烧 CPU
6. **保留 FA2 路径**：先用配置项切换，验收通过后再决定是否删掉 `forceatlas2` 依赖

### P2 — 交互对齐

- **拖拽钉住**：`dragNode` 期间设 `node.fx/fy`，拖完释放并 `alpha(0.3).restart()` 让图重新松弛
  （这一步是"Obsidian 手感"的关键，也是现在完全缺失的）
- 细化：拖拽期间禁用 sigma 默认的单点位移，改由 simulation 驱动
- 「重置布局与视图」按钮语义改为"回到类型环种子 → 重跑 simulation"
- 切换筛选/版本时**保留已有坐标**，只增量加入新节点并短暂加热（避免每次筛选都重排）

### P3 — 类型锚点力（路线 B 的语义层）

- 锚点布局：7 类在半径 `R` 的圆上**扇形均布**（不是同心嵌套），空类型不占位
- `forceX(anchor.x).strength(anchorStrength)` / `forceY(anchor.y).strength(anchorStrength)`
- 工具栏加一个「类型聚类」滑块：`0` = 纯 Obsidian，`1` = 强分层
- **验收**：滑块拉到 0 与拉到 1 的形态要有肉眼可辨的差异，且拉回 0 不残留

### P4 — 稳定性与性能

- 当前规模（最大 176 实体 / 166 关系）单线程绰绰有余，**不需要 web worker**
- 需要做的：`forceManyBody.theta` 用 0.9（Barnes-Hut 近似，等价于原 `barnesHutOptimize`）
- 若将来破千节点，再考虑降级到 FA2 或 worker

---

## 五、风险与取舍

| 风险 | 说明 | 缓解 |
|---|---|---|
| **每次打开形态不同** | 力导向的天然不确定性会让人觉得"图在乱动" | 初始位置来自类型环（确定性）+ 固定 tick 数/alphaDecay → 结果可复现 |
| **丢掉类型分层** | 这是 `useGraphLayout.ts` 刻意的设计 | 路线 B 的锚点力让它在"可调"而非"丢失" |
| **调参比现在难** | 4 个参数相互耦合，用户容易调糊 | 严格按 `distance → repel → link → center` 顺序提供预设，别让用户从零调 |
| **参数语义被归一化抵消** | sigma `autoRescale` 让绝对像素无意义 | 全部按"相对节点半径的倍数"定义，UI 文案写清倍数而非像素 |
| **FA2 与 d3-force 并存混乱** | 两套布局代码同时存在 | 用配置开关隔离，验收后删除一套 |

---

## 六、验收清单

- [ ] 同一份数据（v10，61/63）在默认参数下的截图，留白比例与簇分离度接近 Obsidian
- [ ] 标签数量随缩放动态变化；低缩放下画布上几乎无文字
- [ ] 拖拽枢纽节点，邻居实时跟随、松手后回弹并重新收敛
- [ ] 四个参数改动后 <100ms 内有可见反馈
- [ ] 切换版本/筛选后位置不重排
- [ ] 收敛后 CPU 占用回落到静息水平（simulation 已 stop）
- [ ] `vue-tsc --build` 与 `go build` 均通过

---

## 七、建议的执行顺序

1. **先做 P0**（独立、零风险、当天可见效果）——先确认"降噪方向"是你要的
2. 再做 P1+P2（核心观感 + 手感）
3. P3 视你"要不要保留类型可辨识"的决定
4. P4 随 P1 一起做掉，不用单独排期

---

## 八、实施记录（2026-09-17）

按用户决定执行：**路线 C（纯力导向，放弃类型环）+ P0 + P1 + P2**。

### 改动的文件

| 文件 | 改动 |
|---|---|
| `web/src/composables/useForceLayout.ts` | **新增**。d3-force 模拟器 `ForceSimulator`、四个力参数的刻度与默认值、内部量纲换算系数 |
| `web/src/composables/useGraphLayout.ts` | 新增 `computeSeedLayout`（黄金角螺旋、无类型偏置）；`computeLayout` 改为转调它。**原 `computeTypeRingLayout` 完整保留**，要退回旧观感只需改这一行 |
| `web/src/components/graph/GraphCanvas.vue` | P0 视觉降噪 + P1 接 d3-force 替换 FA2 + P2 节点拖拽 |
| `web/src/components/graph/GraphToolbar.vue` | ~~新增「力度」参数面板~~ → 后来按要求移出，见文末调整 |
| `web/src/components/graph/GraphForcePanel.vue` | **新增**。右侧常驻的力度设置面板（可折叠、带参数行为说明） |
| `web/src/components/KnowledgeGraph.vue` | 接线力度参数（受控状态放页面层，画布重建不丢）；右侧改为 `.graph-right` 容器承载力度面板 + 详情面板 |
| `scripts/calibrate_force_params.mjs` | **新增**。力参数标定脚本 |
| `scripts/_export_graph_sample.py` | **新增**。从 `xingmang.db` 导出标定样本 |

备份（项目无 git）：`GraphCanvas.vue.bak`、`useGraphLayout.ts.bak`。

### 标定结论（默认值的依据）

样本 book 16 v10（60 实体 / 63 关系）：

- `REPEL_TO_STRENGTH = 30`（UI 的 repel 10 → d3 charge −300）→ gapRatio 4.61、零重叠
- `CENTER_TO_STRENGTH = 0.2`（UI 的 center 0.52 → strength 0.104）→ 图半径从无约束的 953 收到 400，且没压扁
- `LINK_DISTANCE_RATIO = 20`（linkDistance = 节点半径 × 20 = 140）→ **边交叉 22 次**。
  若照搬 Obsidian 的 40 倍比例（280），交叉暴涨到 **198 次**，而 gapRatio 完全一样（4.6）
  ——"白得一样多，但线不打架"
- `velocityDecay = 0.6`（d3 默认 0.4，调高才有 Obsidian 那种"静")

### P0 视觉降噪的实际改动

| 项 | 原值 | 现值 |
|---|---|---|
| `labelRenderedSizeThreshold` | 未设（默认 6，而 NODE_SIZE=7 → 标签全量渲染） | 11 |
| 节点标签 | 白底药丸（圆角块 + 边框） | 裸文字 + 3px 描边 |
| 边色 | `#9aa4b2` | `rgba(154,164,178,0.8)` |
| 边粗细 | 1.4 / 1.8（人工） | 1.0 / 1.6 |
| `minEdgeThickness` | 0.4 | 0.3 |
| 箭头 | 常显 | 默认无箭头；**选中节点的边**才切到带箭头 program |
| 边标签 | 默认开 | 默认关 |
| 单边曲率 | 0.2（统一朝一侧弯） | 0（直线）；只有平行边扇形展开 |
| 节点描边 | 固定白（暗色主题下发亮圈） | 随主题 |

### 踩坑记录（都是实测出来的，别再踩）

**1. `setCustomBBox(getBBox())` 会让整张图消失。**

`getBBox()` 返回 `nodeExtent`，而它在 `clearNodeIndices()` 里被重置为 `{x:[0,1], y:[0,1]}`，
要等下一次 `process` 才重算。`onTick` 用的是 `refresh({schedule:true})`，process **推迟到下一帧**，
而 d3 的 `'end'` 事件在最后一帧 tick 之后**同步**触发 —— 正好落进"已清空、未重算"的窗口，
于是把 `[0,1]` 当成全图范围冻住。节点实际坐标在 ±600 量级 → 整张图被映射到视野之外，
表现为「切换图谱几秒后节点与连线一起消失」。
**结论：不要冻结 bbox，交给 autoRescale。**

**2. 边色 alpha 压太低 = "关系线不见了"。**

`rgba(154,164,178,0.45)` 在白色画布上混合后约 `rgb(209,214,221)`，几乎与背景同色。
线没丢，是看不见了。**降噪必须留出可见度底线。**

**3. `defaultEdgeType` 必须显式指定。**

sigma 默认是 `'line'`，而 `edgeProgramClasses` 只注册了 `curve` / `curvedArrow`。
任何没有 `type` 属性的边都会落到 `'line'`，sigma 找不到对应 program 会抛错并中断边渲染
（表现为只看得见节点、看不见连线）。

**4. sigma 核心不含节点拖拽。**

dist 里 `draggedNode` 零命中。拖拽必须自己用 `viewportToGraph()` + 直接写坐标实现；
松手要 `fx=null` 并 `alphaTarget(0)` 才能回弹收敛。

**5. Node 里无法加载 sigma 做单测。**

`@sigma/edge-curve` → `sigma/rendering`，顶层就用了 `WebGL2RenderingContext`，
Node 环境直接 `ReferenceError`。sigma 相关的东西只能用静态阅读源码或浏览器验证。

### 验证状态

- `vue-tsc --build` exit=0 ✓
- `vite build` 通过（2249 模块 / 16.6s）✓
- 力参数标定：零重叠、gapRatio 4.61、边交叉 22 ✓
- **待用户在浏览器确认**：模拟生长过程、拖拽回弹、参数面板实时反馈、收敛后 CPU 静息

### P3（类型锚点力）未实施

按路线 C 放弃类型环，故 P3 不做。`computeTypeRingLayout` 保留作为退路。
日后若想找回"类型可辨识"，做法仍是在 `useForceLayout` 里追加指向各类型扇形锚点的
`forceX/forceY`（强度做成滑块，0 = 纯 Obsidian 观感）。

### 调整记录：力度面板移到右侧（11:29）

按要求把力度设置从工具栏弹层改为**右侧常驻面板**：

- **新增** `GraphForcePanel.vue`：标题可折叠、四个滑块、每个参数配一句行为说明
  （说明的排列顺序就是建议的调参顺序）、恢复默认按钮。
  `max-height: 62%` + 内部滚动，防止内容把下面的实体详情面板挤没。
- `KnowledgeGraph.vue` 右侧改为 `<aside class="graph-right">`，内含力度面板（常驻）
  与详情面板（选中时）。**左边框放在容器上，并覆盖掉详情面板自带的 `border-left`** ——
  否则详情面板隐藏时，力度面板左侧会缺一段边框，看起来像"半截线"。
- 工具栏的力度 popover 已移除（同一组参数不留两处入口），相关 prop / emit / 样式一并清掉。
- 验证：`vue-tsc --build` exit=0、`vite build` 通过（20.9s）。
