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
	{
		board:   BoardTutorial,
		title:   "如何拆书：五步走完，按钮在哪、为什么点不动",
		summary: "导入或创建作品 → 点卡片进正文 → 右上角第四个按钮进拆书库 → 左侧配好章节、提示词、模型和拆分方式 → 点右下角「开始拆书」。",
		author:  "星芒用户",
		pinned:  1,
		daysAgo: 0,
		content: strings.TrimSpace(`
<h2>一共五步</h2>
<p>导入或创建作品 → 点卡片进入正文界面 → 点右上角第四个按钮进拆书库 → 左侧配好章节、提示词、模型和拆分方式 → 点右下角「开始拆书」。</p>
<h2>第一步：让库里先有一本书</h2>
<p>在「书本管理」页顶部选<strong>导入作品</strong>（支持 TXT / JSON，可一次多选）或<strong>创建作品</strong>。</p>
<p>说明：作品卡片右下角的进度条是<strong>创作进度</strong>，不是导入进度。每本书都有一个目标字数（默认 10 万），进度 = 已写字数 ÷ 目标字数，最高 100%。导入一本 30 万字的书，进度会直接顶满；导入几千字的短篇，进度可能只有个位数。两种都跟导入是否完成无关——文件落盘那一刻导入就结束了。</p>
<h2>第二步：进入正文界面</h2>
<p>点作品卡片本身，不要点卡片底部那排小图标按钮。</p>
<h2>第三步：打开拆书库</h2>
<p>正文界面右上角有一排图标按钮，从左到右依次是 AI 对话、AI 写作、角色库、<strong>拆书库</strong>，再往后是知识图谱。<strong>第四个</strong>就是拆书库。</p>
<p>说明：鼠标悬停在按钮上会显示名称，拿不准时停一下即可，不必逐个点开试。</p>
<h2>第四步：左侧配好四项</h2>
<p>拆书库左侧的配置区从上到下四块，按顺序配完：<strong>1 选择章节</strong> → <strong>2 选择提示词</strong> → <strong>3 选择 AI 模型</strong> → <strong>4 选择拆分方式</strong>，最后一块里的「补充信息」可以不填。</p>
<p>说明：拆分方式有两种。<strong>分章拆</strong>是逐章分开分析，每章结论互不干扰；<strong>合并拆</strong>是把选中的章节并成一份整体内容来看，适合分析整体结构和跨章伏笔。</p>
<p>说明：这四项缺任何一项，右下角的「开始拆书」都是灰的、点不动。卡在这里通常不是软件出问题，而是提示词或模型还没选。</p>
<h2>第五步：点开始拆书</h2>
<p>点右下角「开始拆书」，右侧会一段段输出结果；中途想停就点「停止拆书」，已输出的部分会保留。</p>
<p>说明：结果不会自动保存。拆完需要点「保存这次拆书」，这次结果才会记进这本书的历史；不点就关掉页面，结果不留。</p>
<h2>容易卡住的五个地方</h2>
<ol>
  <li>进度条是创作进度，不是导入进度。</li>
  <li>「开始拆书」灰着，先检查章节、提示词、模型是不是缺了一项。</li>
  <li>分章拆和合并拆的区别不是拆得细不细，而是一次看一章，还是一次全看完。</li>
  <li>合并拆不要一次勾太多章，容易超出上下文，结果会断在半路。</li>
  <li>拆完不点「保存这次拆书」，关掉页面就没了。</li>
</ol>
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

	// 种子的 id 从 1 开始按顺序编号（1、2、3），不跟随开发机上游数据库的 id 水位——
	// 打包版新装后看到的就是干净的 1/2/3。库里已有帖子时（老用户升级/标记丢失）
	// 接在 max(id) 之后，避免主键冲突。
	var maxID int
	if err := tx.QueryRow(`SELECT COALESCE(MAX(id), 0) FROM experience_shares`).Scan(&maxID); err != nil {
		return 0, err
	}
	nextID := maxID + 1

	inserted := 0
	for _, p := range seedPosts {
		// 该吧已有同标题的帖子就不种（老用户升级上来时别硬塞内容）。
		// ⚠ 不能用「该吧已有置顶帖就跳过」：同一个吧有多条种子时，
		// 种完第一条就会把后面几条全部挡掉。
		var existing int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM experience_shares WHERE COALESCE(board,'experience') = ? AND title = ?`, p.board, p.title).Scan(&existing); err != nil {
			return 0, err
		}
		if existing > 0 {
			continue
		}

		createdAt := time.Now().UTC().AddDate(0, 0, -p.daysAgo).Format("2006-01-02 15:04:05")
		if _, err := tx.Exec(`INSERT INTO experience_shares
			(id, title, summary, content, content_render_mode, create_type, author_name, status,
			 board, is_featured, is_pinned, like_count, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'html', 'manual', ?, 'published', ?, 0, ?, 0, ?, CURRENT_TIMESTAMP)`,
			nextID, p.title, p.summary, p.content, p.author, p.board, p.pinned, createdAt); err != nil {
			return inserted, err
		}
		nextID++
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
