package handlers

import (
	"database/sql"
	"log"
	"strings"
	"time"
)

// 系统引导帖。
//
// 以前这两条是**写死在模板里的假置顶**（`<div class="pin">`），点不进去、也改不了。
// 现在把它们种成真帖子：能点开看正文、能被管理、能被置顶/取消。
//
// ⚠ 「只种一次」用 app_meta 记标记，不是「不存在就插」——
// 否则用户删掉引导帖之后，下次启动又会被插回来（"反复复活"的老坑）。
const seedWelcomeKey = "seed:experience_welcome:v1"

type seedPost struct {
	board    string
	title    string
	summary  string
	content  string
	pinned   int
	author   string
	daysAgo  int
}

// seedPosts 两个吧各自的引导帖。正文用 HTML（与现有帖子的 content_render_mode 一致）。
var seedPosts = []seedPost{
	{
		board:   BoardExperience,
		title:   "欢迎把踩过的坑和方法论发成经验贴",
		summary: "精华帖由吧主人工挑选；发帖前先看一眼这里的规矩。",
		author:  "星芒用户",
		pinned:  1,
		daysAgo: 30,
		content: strings.TrimSpace(`
<h2>这个吧是干什么的</h2>
<p>把跑通过的方法、踩过的坑、可复用的流程写成帖子。不追求文采，追求<strong>别人照着能做</strong>。</p>
<h2>发帖建议</h2>
<ul>
  <li><strong>标题写清楚结论</strong>，别写「求助」「分享」这种没有信息量的词。</li>
  <li><strong>正文分段</strong>：先说背景（什么问题），再说做法（几步），最后给结果（效果 / 坑在哪）。</li>
  <li><strong>带上参数</strong>：模型名、温度、提示词片段，别人才能复现。</li>
  <li>能截图就别只写字，能贴提示词就别只说「调了一下」。</li>
</ul>
<h2>精华怎么来</h2>
<p>精华由吧主人工挑选，不看字数看<strong>可复现性</strong>。被加精的帖子会出现在「精华」分区，也更容易被搜到。</p>
<h2>分区怎么用</h2>
<p>点顶部「管理」可以新建分区（每个吧最多 15 个），发帖时选「版块分区」归类。不选就落在「未归类」。</p>
<h2>数据存在哪</h2>
<p>所有帖子都存在<strong>本机</strong>，不上传任何服务器。换机器前用「管理 → 导出全部」把内容导成 JSON 带走（导的是帖子内容，不含封面图片和 PDF 附件本体）。</p>
`),
	},
	{
		board:   BoardTutorial,
		title:   "教程吧怎么用：先看「功能速查表」",
		summary: "每个页面干什么、关键操作在哪，都整理在速查表里；这里放按模块拆开的详细教程。",
		author:  "星芒用户",
		pinned:  1,
		daysAgo: 30,
		content: strings.TrimSpace(`
<h2>先看速查表，再看教程</h2>
<p>左栏第二项「<strong>功能速查表</strong>」按「写作主线 / 分析与素材 / 工具与设置」三组，把 11 个页面各自干什么、关键操作在哪都列了出来，还带搜索。不知道某个功能在哪，先去那里找。</p>
<h2>教程吧放什么</h2>
<p>速查表只写「怎么用」，这里写<strong>「怎么用好」</strong>——一个功能展开讲一篇，比如：</p>
<ul>
  <li>「API 配置」里 max_tokens 到底怎么填，填大了会怎样、填小了会怎样；</li>
  <li>「拆书库」怎么拆才不是流水账；</li>
  <li>「wiki 图谱」实体太多看不清时怎么筛；</li>
  <li>「AI 扫榜」的筛选条件为什么每个源都不一样。</li>
</ul>
<h2>写教程的建议</h2>
<ul>
  <li><strong>一个帖子只讲一件事</strong>，标题就是那件事的结论。</li>
  <li>先写「你会遇到什么现象」，再写「原因是什么」，最后写「怎么解决」。</li>
  <li>涉及界面的，标清楚在哪个页面、哪个按钮（左上 / 右上 / 卡片右下角）。</li>
  <li>版本相关的写法要注明版本号，避免过期教程误导人。</li>
</ul>
<h2>和「经验分享吧」的区别</h2>
<p>教程吧 = <strong>说明书</strong>（相对客观、按模块）；经验分享吧 = <strong>交流区</strong>（个人踩坑、心得、方法论）。两边数据完全隔离，分区也各自独立。</p>
`),
	},
}

// SeedSystemPosts 首次启动时把引导帖种成真帖子（幂等：靠 app_meta 标记，只做一次）。
//
// 返回实际插入条数，0 表示之前已经种过。
func SeedSystemPosts(db *sql.DB) (int, error) {
	var marked string
	err := db.QueryRow(`SELECT value FROM app_meta WHERE key = ?`, seedWelcomeKey).Scan(&marked)
	if err == nil {
		log.Printf("[seed] 跳过：已种过（value=%s）", marked)
		return 0, nil // 种过了，哪怕用户把帖子删光也不再插
	}
	if err != sql.ErrNoRows {
		log.Printf("[seed] 查标记失败: %v", err)
		return 0, err
	}
	log.Printf("[seed] 未种过，开始种入 %d 条引导帖", len(seedPosts))

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inserted := 0
	for _, p := range seedPosts {
		// 该吧已经有帖子就不种（老用户升级上来时别硬塞内容）
		var existing int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM experience_shares WHERE COALESCE(board,'experience') = ? AND COALESCE(is_pinned,0) = 1`, p.board).Scan(&existing); err != nil {
			return 0, err
		}
		if existing > 0 {
			continue
		}

		createdAt := time.Now().UTC().AddDate(0, 0, -p.daysAgo).Format("2006-01-02 15:04:05")
		if _, err := tx.Exec(`INSERT INTO experience_shares
			(title, summary, content, content_render_mode, create_type, author_name, status,
			 board, is_featured, is_pinned, like_count, created_at, updated_at)
			VALUES (?, ?, ?, 'html', 'manual', ?, 'published', ?, 0, ?, 0, ?, CURRENT_TIMESTAMP)`,
			p.title, p.summary, p.content, p.author, p.board, p.pinned, createdAt); err != nil {
			return inserted, err
		}
		inserted++
	}

	if _, err := tx.Exec(`INSERT INTO app_meta (key, value, updated_at) VALUES (?, '1', CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value='1', updated_at=CURRENT_TIMESTAMP`, seedWelcomeKey); err != nil {
		return inserted, err
	}

	if err := tx.Commit(); err != nil {
		return inserted, err
	}
	if inserted > 0 {
		log.Printf("[seed] 已种入 %d 条系统引导帖", inserted)
	}
	return inserted, nil
}
