# 知识图谱升级指导：七层嵌套布局 · 全曲线连线 · AI 累加合并

> 适用范围：`xingmang-desktop/web`（Vue3 + TS + sigma/graphology 前端）与可选的 Go 后端小改。
> 本文是实施指导书：先讲原理，再列每处代码改动的位置与验收要求。

---

## 一、背景与目标

当前效果（sigma 版已跑通 39 实体 / 36 关系）存在三个问题：

| # | 现状问题 | 目标 |
|---|---------|------|
| 1 | FA2 力导向直接从环形初始坐标收敛，孤立节点漂在画布左侧，主体挤在右侧，杂乱无章 | **七层嵌套布局**：7 类实体各占一层同心环，任何节点（含孤立节点）都落在所属类型的环上，整体规整 |
| 2 | 边虽已用 `arrow` program + controlPoints 渲染，但 BASE_BEND=0.08 弯曲幅度几乎看不出来 | **所有连线都是可见的曲线**，平行边扇形展开，对标马良观感 |
| 3 | 每次 AI 分析都是「只存本次抽取结果为新版本」，**覆盖式**，丢了旧图 | **累加式**：在当前图谱基础上合并查重后叠加新实体/新关系，新抽实体能自动与已有实体连边 |

---

## 二、方案原理

### 2.1 七层嵌套布局（按实体类型分环）

7 类实体（角色/地点/势力/事件/物件/功法/线索）恰好对应 7 层同心环：

```
             ┌─ 第7环 线索
          ┌─┐└─ 第6环 功法
       ┌─┐│└── 第5环 物件
    ┌─┐│└─── 第4环 事件
 ┌─┐│└────── 第3环 势力
 │├────────── 第2环 地点
 ●─────────── 第1环 角色（最内层，度数最高的主角贴近圆心）
```

布局步骤（纯函数，不改 DOM）：

1. **分环**：按固定类型顺序 `character → location → faction → event → item → skill → clue` 从内到外分配环；某类为空则跳过该环（层数收缩，仍是「嵌套」结构）。
2. **定半径**：第 k 环半径 `r_k = innerR + k × ringGap`。`ringGap` 按「该环节点数 × 节点最小间距 26px / 2π」自适应外扩，避免内环挤爆。
3. **环内排序**：度数降序 → 同度数按名称，沿环均布；**相邻环错开半个角步长**（蜂窝式错位），嵌套感和可读性都更好。
4. **孤立节点（度数=0）**：不特殊处理，直接落在所属类型的环尾——这是「不漂、不乱」的关键。后续 AI 累加分析补上关系后，自然从环尾获得连边。
5. **FA2 只做轻度修饰**：初始坐标已经是嵌套环，FA2 仅用少量迭代（80~120）+ 较强 gravity（3~5）解开局部打结，**不能**再像现在这样 300 次迭代把结构打散。

### 2.2 全曲线连线

sigma 3.x 内置 `arrow` program（`EdgeArrowProgram`）原生支持边属性 `controlPoints: [{x,y}]` 渲染二次贝塞尔曲线+箭头，`rebuild()` 已在用。需要调整的只是**弯曲参数**：

- 每条边都给控制点（现状已满足：单边也有 `BASE_BEND`，平行边按索引扇形展开 `BEND_STEP`）。
- `BASE_BEND` 0.08 → **0.18~0.22**，让单边弯曲肉眼可见。
- 可选增强：弯曲量随边长自适应（短边弯多一点，长边弯少一点），公式 `curve = bend / max(1, dist/200)`，防止长边弯成大弧。
- 已知限制：sigma 的边标签渲染在两端点直线的几何中点，不在曲线上；弯曲加大后标签会略偏离弧线，可接受，不在本次范围。

### 2.3 AI 分析累加合并（增量构建图谱）

**现状链路**（覆盖式，要改）：

```
analyzeChapters() → 本次抽取结果
GraphAnalyzeDialog.save() → store.importAsNewVersion(name, 本次结果)   ← 新版本里只有本次数据
```

**目标链路**（累加式）：

```
当前版本数据(已有实体/关系)
        ↓ 注入实体名
analyzeChapters(baseEntities)   ← AI 抽取时 knownNames 含已有实体名，
        ↓                         关系端点指向已有实体不再被丢弃
本次抽取结果（可能含与已有实体同名的重复项）
        ↓
mergeGraphData(当前版本, 本次结果)   ← 前端纯函数合并查重
        ↓
store.importAsNewVersion(name, 合并结果)   ← 新版本 = 旧图 ⊕ 增量（版本链，可回退）
```

查重规则（与 `aiGraphExtraction.ts` 现有规则保持一致）：

| 对象 | 去重键 | 命中时行为 |
|------|--------|-----------|
| 实体 | `normalizeName(name)`（trim + 小写 + 空白折叠），**不看类型**——人名全局唯一，同名异类型以旧实体为准 | 复用旧实体 id；chapter_ids 并集；description 择优保留（取更长的那份） |
| 关系 | `normalizeName(source) | normalizeName(target) | normalizeName(relation_type)`（方向敏感，A→B 与 B→A 视为两条） | 丢弃新关系；chapter_ids 并集到旧关系 |

**「加上某些东西并连接」的关键**：`normalizeExtractedGraph(payload, knownNames)` 第二个参数现在只传了本次运行内已抽到的名字（`accumulatorNames`）。把**当前版本的已有实体名也并入 knownNames**，AI 在新章节里写的「陈默 出现在 青州」这类指向旧实体的关系就会解析成功，而不是被当作端点丢失丢掉。

**可选增强（第二期，不阻塞主线）**：
- AI 别名归并 pass：把「旧实体名列表 + 新实体名列表」发给模型，让它输出 `{新名: 旧名}` 别名映射（如 默儿→陈默），merge 前先按映射改名再查重。
- 孤岛补边 pass：对度数=0 的实体，拿它们的 name+description 让模型提议关系（confidence 标低），把「几乎所有节点相连」做满。

**后端不需要改动**：`ImportGraph` 是「整批 INSERT 到指定 versionId」的语义，天然支持「合并结果写入新版本」——旧版本不动，新版本是完整快照，与现有 `saveAsNewVersion` 的复制逻辑同构。

---

## 三、代码改动清单

### 改动 1：七层嵌套布局

**文件**：`web/src/composables/useGraphLayout.ts`

| 位置 | 改动 | 要求 |
|------|------|------|
| 文件头注释 | 更新说明 | 描述从「环形初始坐标」改为「七层嵌套环初始坐标」 |
| 新增常量 | `TYPE_RING_ORDER: readonly GraphEntityType[] = ['character','location','faction','event','item','skill','clue']` | 顺序即环序，从内到外 |
| 新增函数 `computeTypeRingLayout(ctx)` | 替代现在的 `circularLayout`：分环 → 自适应半径 → 环内度数降序均布 → 相邻环错开半角 | ① 空类型跳过不占环；② 度数 0 的节点也必须落在所属类型环上（禁止漂在空白处）；③ 返回值类型与现 `computeLayout` 相同（`Map<number, NodePosition>`） |
| `computeLayout()` | 内部改调 `computeTypeRingLayout` | 对外签名不变（`GraphCanvas.rebuild()` 无需改调用） |
| 删除 | `circularLayout`、`sortByDegreeThenType` 若无引用则删 | 不留死代码 |

### 改动 2：FA2 退居修饰 + 曲线参数

**文件**：`web/src/components/graph/GraphCanvas.vue`

| 位置 | 现状 | 改动 | 要求 |
|------|------|------|------|
| L182 `BASE_BEND` | `0.08` | `0.2` | 单边弯曲肉眼可见 |
| L241 `curve` 计算 | 仅平行边展开 | 可选：按边长自适应 `curve / max(1, dist/200)` | 长边不弯成大弧，短边不重叠 |
| L258~268 `forceAtlas2.assign` | `iterations: 300, gravity: 1, scalingRatio: 2` | `iterations: 100, gravity: 4, scalingRatio: 1.2`（参数可微调） | FA2 只解局部打结；嵌套环结构在收敛后仍清晰可辨；孤立节点不飞出主团 |
| `fitView()` / `focusNode()` | 已修复（归一化空间） | 不动 | 保持现状 |

### 改动 3：合并查重纯函数

**文件**：新建 `web/src/services/graphMergeService.ts`（放 services 层，可 import 类型；去重键复用 utils 的 `normalizeName`/`relationKey`）

```ts
export interface MergeStats {
  baseEntities: number      // 旧实体数
  baseRelations: number     // 旧关系数
  newEntities: number       // 真正新增的实体数
  reusedEntities: number    // 同名复用旧实体的数量
  newRelations: number      // 真正新增的关系数
  duplicateRelations: number// 因查重被丢弃的新关系数
}
export function mergeGraphData(
  base: { entities: GraphEntity[]; relations: GraphRelation[] },
  incoming: { entities: GraphEntity[]; relations: GraphRelation[] }
): { data: GraphData; stats: MergeStats }
```

要求：
1. 实体按 `normalizeName(name)` 查重（不看 type）；命中 → 保留旧实体，`chapter_ids` 并集，`description` 取更长的一份。
2. 新实体分配临时 id：`max(旧实体id) + 1` 递增（`ImportGraph` 的 `entityIDMap` 会做 old→new 映射，id 只需在请求内唯一）。
3. 关系按 `relationKey(source, target, type)` 查重（方向敏感）；命中 → 丢弃并 `duplicateRelations++`，同时把新章节 id 并到旧关系 `chapter_ids`。
4. 新关系端点 id 解析顺序：先查「旧实体名→旧id」映射，再查「新实体名→新临时id」映射——这就是「新实体自动连到旧实体」的实现点。
5. **零副作用纯函数**：不发请求、不改 store，输入输出均为普通对象；便于用 `npx tsc` 独立验证（同 `aiGraphExtraction.ts` 的做法）。

### 改动 4：抽取时注入已有实体名

**文件**：`web/src/services/graphAnalyzeService.ts`

| 位置 | 改动 | 要求 |
|------|------|------|
| `AnalyzeOptions` | 新增可选字段 `baseEntities?: Array<{ name: string; type: string }>` | 调用方传当前版本实体 |
| `analyzeChapters()` L247 `normalizeExtractedGraph(extractJsonPayload(text), accumulatorNames(accumulator))` | knownNames 改为 `[...baseNames, ...accumulatorNames(accumulator)]` | baseNames 在循环外由 `baseEntities` 预计算（`normalizeName` 去重一次）；旧实体名参与关系端点白名单，指向旧实体的关系不再被丢 |
| `assemble()` | 签名增加 base 实体映射；`nameToId` 先装入 base（normalizeName → 旧 id），新实体临时 id 从 `maxBaseId+1` 起 | 抽取结果中与旧实体同名的条目**不再生成新实体行**（否则 merge 会重复）；统计里单列 `reusedWithBase` 或并入 duplicateEntities |

### 改动 5：store 提供合并入口

**文件**：`web/src/stores/knowledgeGraph.ts`

| 位置 | 改动 | 要求 |
|------|------|------|
| 新 action `analyzeMergePreview(data)`（或直接在 Dialog 中调用纯函数） | 无 | 合并逻辑在纯函数里，store 只负责取 `entities/relations` 当前值，保持薄 |
| `importAsNewVersion()` | 不动 | 复用现有链路写新版本 |
| 可选：保存成功后 `refresh()` | 已有 | 确保切到新版本 |

### 改动 6：分析弹窗支持「累加合并」

**文件**：`web/src/components/graph/GraphAnalyzeDialog.vue`

| 位置 | 现状 | 改动 | 要求 |
|------|------|------|------|
| `start()` L369 | 未传 base | `analyzeChapters({ ..., baseEntities: 当前版本实体快照 })` | 从 store 取 `{name, type}` 列表；当前版本为空则不传 |
| `save()` L401 | `importAsNewVersion(name, 本次结果)` | 改为：`mergeGraphData(当前版本数据, 本次结果)` 后导入合并结果 | ① 默认累加；② 弹窗内新增「保存内容」单选：`合并到当前图谱（推荐，生成新版本快照）` / `只存本次结果（旧行为）`；③ 保存前展示合并统计：`新增实体 N · 复用已有 M · 新增关系 X · 去重重复 Y`；④ 版本名默认值带「+增量」语义（如 `第1-10章增量`） |
| 结果预览区 | 只显示本次统计 | 增加一行与当前图谱合并后的预览计数 | 用户保存前就知道图会变成多大 |

### 改动 7（可选二期）：孤岛补边

**文件**：`web/src/utils/aiGraphExtraction.ts` + `web/src/services/graphAnalyzeService.ts`

- 新增 `buildOrphanLinkPrompt(orphans, candidates)`：输入度数 0 实体列表与全书实体名列表，输出同样的 JSON 三元组结构。
- 新增 `linkOrphanEntities()` 编排：找孤岛 → 一次 AI 请求 → `normalizeExtractedGraph`（knownNames=全书实体名）→ mergeGraphData。
- 要求：提议关系 `confidence` 置 0.5、`evidence` 留空并写入 `relation_type` 前缀「推断·」，UI 上可识别；失败静默跳过不阻塞主流程。

---

## 四、验收标准

1. **七层嵌套**：打开含 7 类实体的版本，画布呈同心环结构，每类一层；**任意孤立节点不出现在环外空白区**；相邻环节点不重叠、不错位到别的环。
2. **全曲线**：任意缩放级别下所有边均为可见弧线；同一对节点多条关系呈扇形分布，箭头方向正确；边标签可读。
3. **累加合并**：
   - 对已有图谱再次跑 AI 分析 → 新版本实体数 ≥ 旧版本（只增不减）；
   - 与旧实体同名的新抽取项不产生重复节点；
   - 新章节里「新实体—旧实体」的关系在图上真实成边；
   - 合并统计数字与实际落库一致；
   - 旧版本数据未被改动（版本链可回退验证）。
4. `npm run type-check` 通过；console 无 sigma/graphology 报错。

---

## 五、实施顺序（建议提交粒度）

1. 改动 1 + 2（布局与曲线，纯前端渲染层，一次提交，浏览器验收 1/2）
2. 改动 3（纯函数 + 独立验证脚本跑样例数据）
3. 改动 4 + 5 + 6（串起累加链路，浏览器验收 3）
4. 改动 7（二期，独立提交）
