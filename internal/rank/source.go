package rank

import (
	"os"
	"sort"
	"strings"
)

// ─── 数据模型 ─────────────────────────────────────────────────────────

// FilterOption 是筛选器的一个取值。
type FilterOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FilterDim 是榜单内部的一层筛选维度。
// 起点用它承载题材（catId）、月份（yearmonth，仅男频月票榜）、周期（rankPeriod，部分更新榜）；
// 番茄与七猫没有这一层，它们的 GroupItem.Filters 为空。
//
// ⚠ 维度内容由各数据源的接口动态给出，**不要在数据源里硬编码**：
// 起点的 catId 是性别作用域内的，跨性别传会静默返回空列表。
type FilterDim struct {
	Key     string         `json:"key"`
	Name    string         `json:"name"`
	Options []FilterOption `json:"options"`
	// DefaultID 是缺省选中的取值，取自接口返回的 selectedIndex。
	DefaultID string `json:"defaultId"`
}

// BookQuery 是一次榜单书籍请求。
type BookQuery struct {
	// Tab 是选中的榜单（GroupItem.ID）。
	Tab string
	// Filters 是该榜的筛选取值，key 为 FilterDim.Key。没有第三层的源忽略它。
	Filters map[string]string
	Offset  int
	Limit   int
}

// ScopeKey 是快照存储与对比用的榜标识：Tab 加排序后的筛选条件。
// 必须带上筛选条件，否则同一榜单下不同题材/月份的快照会互相覆盖，
// 算出来的名次变动全是错的。
func (q BookQuery) ScopeKey() string {
	if len(q.Filters) == 0 {
		return q.Tab
	}
	keys := make([]string, 0, len(q.Filters))
	for k := range q.Filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys)+1)
	parts = append(parts, q.Tab)
	for _, k := range keys {
		parts = append(parts, k+"="+q.Filters[k])
	}
	return strings.Join(parts, "|")
}

// ─── 数据源接口 ───────────────────────────────────────────────────────

// Source 是一个榜单数据源。新增站点只需实现本接口并在 init 中注册，
// 前端右上角的切换器与页脚来源说明都由注册表驱动，不需要改前端。
type Source interface {
	// ID 数据源标识，对应接口的 source 参数。
	ID() string
	// Name 展示名，用于前端切换器。
	Name() string
	// Desc 数据来源说明，用于前端页脚与空态提示。
	Desc() string
	// Groups 返回榜单分组（第一级），每组 Items 是该组下的榜单（第二级）。
	// Items 可以自带 Filters，交给前端渲染第三层筛选。
	Groups() ([]Group, error)
	// Books 返回某榜第 Offset 条起的 Limit 本书。
	// Tab / Filters 的格式由各数据源自定。
	Books(q BookQuery) ([]Book, error)
}

// FullCapturer 是可选能力：能把整榜一次抓完的数据源。
//
// 有它才能做出**可信的**「新上榜」和真正的「掉榜」——否则快照只覆盖用户翻过的那几页，
// 把没翻过的页当成"新上榜"就是假信号。实现方必须自己限速，并遵守 maxPages 上限。
type FullCapturer interface {
	BooksFull(q BookQuery, maxPages int) ([]Book, error)
	// WorthFullCapture 报告这次查询值不值得花一次整榜抓取。
	// 整榜要翻几十页，而筛选变体可能上百种，不能每个变体都抓：
	// 起点只在「全站 + 各维度都是默认值」时返回 true。
	WorthFullCapture(q BookQuery) bool
}

// SourceInfo 是数据源的元信息，前端据此渲染切换器。
type SourceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	// FullCapture 表示该源能把整榜一次抓完，前端据此决定掉榜视图是否可用。
	FullCapture bool `json:"fullCapture"`
}

// DefaultSourceID 是缺省数据源（请求未带 source 或 source 非法时使用）。
const DefaultSourceID = "fanqie"

var (
	sourceOrder []string
	sourceByID  = map[string]Source{}
)

// register 注册数据源，保持注册顺序（List 的输出顺序）。
func register(s Source) {
	if _, exists := sourceByID[s.ID()]; !exists {
		sourceOrder = append(sourceOrder, s.ID())
	}
	sourceByID[s.ID()] = s
}

func init() {
	register(NewBridgeClient(os.Getenv("FANQIE_BRIDGE_URL")))
	register(newQimaoSource())
	register(newQidianSource())
	register(newFeiluSource())
}

// Get 按 id 取数据源。id 未知或为空时返回默认数据源，保证调用方永不为 nil。
func Get(id string) Source {
	if s, ok := sourceByID[id]; ok {
		return s
	}
	return sourceByID[DefaultSourceID]
}

// List 返回全部数据源的元信息。
func List() []SourceInfo {
	out := make([]SourceInfo, 0, len(sourceOrder))
	for _, id := range sourceOrder {
		s := sourceByID[id]
		_, full := s.(FullCapturer)
		out = append(out, SourceInfo{ID: s.ID(), Name: s.Name(), Desc: s.Desc(), FullCapture: full})
	}
	return out
}
