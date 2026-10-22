/**
 * 更新日志数据源。
 *
 * ================== 口径 ==================
 *
 * 版本由两类节点构成：
 *  1. **release**（有依据）—— 能在 `dist/` 里找到归档包，或在 GitHub releases / tags 里找到发布记录。
 *     版本号、日期、说明都取自实际产物，不作推断。
 *  2. **derived**（时间线细分）—— 两次 release 之间隔了很多天、期间的改动没另打包，
 *     按构建/提交时间拆成若干补号版本。**版本号是补的，改动内容是真实的**。
 *
 * 补号版本的编号规则：
 *   - 落在 `0.7.6.1` 与 `0.7.7.0` 之间 → 依次编 `0.7.6.2 ... 0.7.6.9`（09-29 至 10-15）。
 *     注意 `0.7.6.2` 是**真实发布的**版本号（09-28），只有 `0.7.6.3` 起才是补号。
 *   - 落在 `0.735` 与 `0.7.6.0` 之间 → 依次编 `0.736 ... 0.740`（08-23 至 09-11）。
 *   - 落在 `0.7.7.0` 之后（尚未打包）→ 依次编 `0.771`、`0.772`（10-16 至 10-19）。
 *     `0.771` 的内容已进 10-16 / 10-17 / 10-18 三次日快照（`edf6f4c · a9916eb · 41f4a36`），
 *     `0.772` 的内容截至 10-19 仍在工作区未提交。
 *
 * 依据来源：
 *  - 归档包：`dist/V-*.zip` 内 README.txt 的「构建时间 / Git Commit」行。
 *  - 仓库提交：`D:\xingguang`（桌面版）+ GitHub `kiliczsh1/xingmang`（Web 版，0.7.6.0 时并入桌面版）。
 *  - GitHub releases 公告原文（0.6.4.2 / 0.6.5.1 / 0.6.5.7 的更新说明直接引用）。
 *  - 项目工作记录（`.workbuddy/memory/YYYY-MM-DD.md`），用于把未打包的改动定位到具体日期。
 *
 * ⚠ 三条必须记住的事实：
 *  1. `0.7.6.1` / `0.7.6.2` / `0.7.7.0` 三个包的 README **都写 `ef12d9e`**：
 *     09-22 到 10-15 的工作区改动一直没提交，直到 `6d0ce1f` 才一次性收口。
 *     所以 commit 号区分不了这三版的内容，只能靠构建时间划界。
 *  2. `0.733` / `0.734` / `0.735`（08-14/08-15）的版本号比 `0.6.8.1`（06-22）大，
 *     但时间在后 —— 版本号本来就不是单调的，**排序按数组顺序（时间倒序），不要按版本号排**。
 *  3. 判定某功能属于哪版，最硬的办法是在该包的 `xingmang.exe` 里搜界面文案（命中即有）。
 *     未命中不能反证：内部叫法（如「贴吧版式」）不会出现在界面里。
 *
 * ⚠ 新增版本时：只追加到数组头部，不要改写历史条目。
 */

export interface ChangelogEntry {
  /** 分类标签：功能 / 界面 / 重构 / 修复 / 工具链 / 基线 */
  tag: string
  /** 条目正文 */
  text: string
}

export interface ChangelogRelease {
  /** 版本号 */
  version: string
  /** 构建 / 发布日期 */
  date: string
  /** 版本依据：release = 有归档或发布记录；derived = 按时间细分的补号 */
  basis: 'release' | 'derived'
  /** 对应提交（短哈希）；工作区未提交期间多项工作会共用同一提交号 */
  commit: string
  /** 一句话概括 */
  summary: string
  /** 简略版要点 */
  brief: string[]
  /** 详略版完整条目 */
  details: ChangelogEntry[]
}

export const changelogReleases: ChangelogRelease[] = [
  {
    version: '0.772',
    date: '2026-10-18 ~ 10-19',
    basis: 'derived',
    commit: '（工作区未提交）',
    summary: '作品全面文件夹化：正文、人物、世界观与全部记忆落成 Markdown，数据库退回索引；品牌图标体系落地',
    brief: [
      '作品变成真实文件夹：正文 / 人物 / 世界观 / 大纲 / 灵感 / 备忘录都落成文件，数据库只做索引',
      '五类记忆（对话 / Agent / 图谱 / 拆书 / 生成历史）落盘为单文件 Markdown，改文件即改数据',
      '新增「接管目录」：把已有作品目录指过来就能认回全部数据，不搬移、不复制',
      '封面分「当前 + 画廊」，历史封面可回切；AI 生成的封面直接落进作品文件夹',
      '品牌图标体系落地：全套尺寸图标 + exe 图标 + 启动页 / 顶栏换标',
      '设置页新增「作品库位置」；API 设置页不再自动塞四条空 Key 服务商'
    ],
    details: [
      {
        tag: '功能',
        text: '作品文件夹化（F0–F4 全期落地）：正文、人物、世界观、大纲、灵感、备忘录全部写成磁盘文件（Markdown / TXT），`books/<书名>/` 就是作品本体；数据库只保留索引（路径、字数、哈希、mtime）。新增 `book_docs`（大纲 / 灵感）与 `app_settings`（应用设置）两张表，`books.folder_path`、`chapters.file_path / word_count / content_hash / file_mtime / missing` 等增量列走 `ensureColumns`，老库启动即补。'
      },
      {
        tag: '功能',
        text: '作品记忆落盘（M0–M5）：对话、Agent 会话、知识图谱、拆书、生成历史五类记忆全部落成 `memory/` 下的单文件 Markdown —— 文件人可读、机器可解析。正文只存在于文件，数据库不再保存副本（`conversation_messages` 等停止写入），新增 `memory_records` 只做索引。写操作只标脏，由 2 秒合并写入器统一落盘，退出前先 flush 再关库。'
      },
      {
        tag: '功能',
        text: '新增「接管目录」：`/api/settings/books-root/preview` 先预检目标目录里有哪些作品，再按 `mode=link` 登记入库 —— 不搬移、不复制、不删除任何文件；随后从目录重建索引（书 / 章节 / 人物 / 世界观 / 记忆 / 封面）。把作品目录拷到任何地方再指过来，数据就回来了。'
      },
      {
        tag: '功能',
        text: '封面分「当前 + 画廊」：`covers/` 存当前封面、`gallery/` 存历史版本并可回切，单张删除走系统回收站。AI 生成的封面带作品上下文时直接落进该作品的 `gallery/`，与「被顶掉的封面」同一个池子；不带上下文才落共享缓存目录。'
      },
      {
        tag: '功能',
        text: '设置页新增「作品库位置」：可改库根目录、可打开文件夹；迁移提供「接管 / 搬移 / 取消」三选一。导入导出同步补齐 —— 导入已有目录、导入旧版导出的作品 JSON、导出作品 ZIP（含 `memory/` 与 `covers/`）。'
      },
      {
        tag: '修复',
        text: '一次性治愈「封面住在缓存目录」：早期换封面只认 data URL，AI 生成封面返回的站内 URL 被当字符串写进 `books.cover`，文件其实一直停在 `data/generated-images` 里、作品自己的 `gallery/` 永远是空的。启动时先把这类封面搬成真文件，再按内容指纹把缓存目录里剩下的生成图归位到各作品的 `gallery/`（认不出的进 `_orphan-images/`）。幂等，搬完即不再命中。'
      },
      {
        tag: '界面',
        text: '品牌图标体系落地：新增 `tools/logogen`（Go 图标生成器，从原图整幅等比例烘出 16–512 全套 + `favicon.ico` / `app.ico`），exe 图标（`rsrc_windows_*.syso`）、顶栏 brand-mark、启动页、Agent 空态、创作仪表盘页眉统一换标。'
      },
      {
        tag: '界面',
        text: '书架与写作页对齐文件夹口径：删作品改「归档 / 回收站」二选一，卡片加「打开文件夹」；编辑器摘要为空时不显示摘要栏。拆书页改为手动「保存这次拆书」（此前生成即自动存历史），未保存离开时二次确认；历史中心与生成历史的读写从浏览器 localStorage 改为后端落盘。'
      },
      {
        tag: '修复',
        text: 'API 设置页不再在首次进入时自动插入四条空 Key 服务商（DeepSeek / SiliconFlow / bltcy.ai / ModelScope）—— 它们既不能直接使用，又会被当成「我的渠道」参与分组与计数，用户还得手动删掉。空状态本来就该是空的。'
      },
      {
        tag: '工具链',
        text: '新增 `docs/books-folder-storage.md`、`docs/books-memory-storage.md` 两份实施文档；`internal/bookstore` 新增约 30 个文件与配套单测（路径清洗与越界防护、原子写、front-matter 保真、Markdown 标记契约、链接重建、回收站）。'
      }
    ]
  },
  {
    version: '0.771',
    date: '2026-10-16 ~ 10-18',
    basis: 'derived',
    commit: 'edf6f4c · a9916eb · 41f4a36',
    summary: 'max_tokens 改为请求时探测并清除全部假值兜底；退出程序时一并收走伴生进程',
    brief: [
      '模型上限不再靠猜：请求前按递减档位实测一次，命中即写缓存并复用',
      '清除各处互相矛盾且无依据的 max_tokens 兜底值，改成「不传优于假值」',
      '模型能力弹窗显示取值来源徽标，空值如实写「未设置 / 来源未知」',
      '缓存加 7 天过期 + 定期重探，修正「缓存只降不升」',
      '退出程序时按清单结束 vite / 番茄桥，不再残留孤儿进程',
      '「AI 对话」弹窗换成「星芒 Agent」弹窗，与侧栏共用一份数据'
    ],
    details: [
      {
        tag: '功能',
        text: 'max_tokens 改为「请求时探测 + 缓存」：新增 `internal/aitokens`，不再靠静态表猜值 —— 请求前按递减档位试一次，命中即写 `model_max_tokens` 并复用；只读接口 `GET /api/ai/max-tokens-cache` 可查。接在 AI 对话与 Agent 循环的 pre-stream 状态码检查点（开始转发 SSE 之后不能重试，否则会重复输出、工具重复执行）。'
      },
      {
        tag: '功能',
        text: '「不传优于假值」：`omitTokens = 0` 作哨兵，表示**不发送** `max_tokens` 字段。清除五处互相矛盾且全无依据的兜底值（`ai/helpers.go` 2000、`agent/loop.go` 4096、`aitokens` 4000、`modelMeta.ts` 8192、`media.go` 2000），只有 Anthropic 格式必填时才用家族起步值兜底。'
      },
      {
        tag: '功能',
        text: '取值来源可溯源：新增 `api_models.max_tokens_source`（`upstream` / `catalog` / `table` / `manual` / `probed` / 空），模型能力弹窗显示来源徽标；来源为空时如实显示「未设置 / 来源未知」，不伪造一个确定值。'
      },
      {
        tag: '功能',
        text: '缓存衰减与失败分类：缓存加 7 天 TTL 与「每 20 次重探一次」，修正「缓存只降不升」；超限失败分三类 —— 输出超限降档**并**写缓存、上下文超限只降档**不**写缓存（那是本次输入太长，缓存它等于把偶发当永久）、其余不降档。'
      },
      {
        tag: '功能',
        text: '退出程序时收走伴生进程：启动脚本把 vite / 番茄桥的 PID 写进 `logs/.dev-companions.json`，后端收尾时按清单逐个结束（校验 PID 存活 + 映像名匹配防误杀，清单超 12 小时整份丢弃），不依赖脚本存活，覆盖「用户直接关掉黑窗口」。'
      },
      {
        tag: '界面',
        text: '顶栏「AI 对话」弹窗改为「星芒 Agent」弹窗，与侧栏「星芒 Agent」共用同一份数据；顶栏「论坛 / 教程」换成「日志」入口。'
      },
      {
        tag: '修复',
        text: '纠正 `KNOWN_MODEL_MAX_OUTPUT`：此前是凭印象填的「保守估值」，系统性偏低（deepseek 16K→393216、glm 4K→131072、claude-opus-4 32K→128000、gemini-3 32768→65536）。改为只放核实过的官方值，查不到的家族不猜、走兜底让人工改。同时确认「上游 `/models` 拿不到模型上限」是代码没接，不是抓不到。'
      },
      {
        tag: '修复',
        text: '超限文案识别补全：中文用「主体词（token / 上下文 / 输出 / 长度）× 超限词（超出 / 过长 / 最大）」复合判定，英文补 `maxOutputTokens` / `INVALID_ARGUMENT` / `output token` / `reduce the length`。⚠ OpenAI 的 `maximum context length is N tokens` 在 max_tokens 过大时也会出现，必须靠 `in the completion` / `messages resulted in` 区分输出侧与输入侧。'
      }
    ]
  },
  {
    version: '0.7.7.0',
    date: '2026-10-16',
    basis: 'release',
    commit: 'f4c1a85 · 6d0ce1f',
    summary: '版本握手、退出入口与数据落点修正；编辑器内核精简，exe 体积减少 8.9 MB',
    brief: [
      '启动时校验后端版本，不再出现「更新了、跑的还是上一版」',
      '新增退出程序入口，退出即合并数据库，不再残留 -wal',
      '数据文件固定落在程序所在目录',
      '编辑器内核精简，exe 从 33.5 MB 降到 24.6 MB',
      '打包脚本补前置检查，不再白等一次前端构建'
    ],
    details: [
      {
        tag: '功能',
        text: '单实例复用改为版本握手：/api/health 返回 version / commit / buildTime，探测到旧版本实例时拒绝启动并弹窗提示。旧行为会直接复用旧后端，叠加 exe 被占用导致解压失败，表现为「更新了但用的还是上一版」。'
      },
      {
        tag: '功能',
        text: '数据落点固定到 exe 所在目录（新增 internal/appdir），修正四处依赖当前工作目录的相对路径。exe 目录不可写时回退 %AppData%\\xingmang，启动日志写明实际使用路径。'
      },
      {
        tag: '功能',
        text: '新增退出入口 POST /api/shutdown，仅放行本机来源并校验 Origin；服务生命周期改由 ctx 驱动，先处理完在途请求再停止。'
      },
      {
        tag: '修复',
        text: '退出时补 PRAGMA wal_checkpoint(TRUNCATE)。此前退出后最新数据可能只留在 -wal 文件里，单独复制 xingmang.db 做备份会缺内容。'
      },
      {
        tag: '界面',
        text: 'API 设置页重构：Config.vue 拆出 CapDialog / ChannelDialog / ParamGroupEditor，新增 services/modelMeta.ts。'
      },
      {
        tag: '界面',
        text: '侧边栏底部与个人中心显示当前版本号，悬停可见 commit 与构建时间。'
      },
      {
        tag: '界面',
        text: '首屏增加主题预执行与 vite:preloadError 自愈重载，避免换版本后浏览器拿着旧 chunk 白屏。'
      },
      {
        tag: '界面',
        text: '顶栏导航调整：移除「论坛 / 教程」两个入口，新增「日志」查看更新记录。'
      },
      {
        tag: '工具链',
        text: '编辑器内核瘦身：新增 monaco-lite.ts 替代官方入口，剔除 css / html / typescript 语言服务、81 个基础语言与未使用的 lsp client。前端产物 21.52 MB → 12.65 MB，exe 33.51 MB → 24.64 MB。'
      },
      {
        tag: '工具链',
        text: '修复 start-dev.ps1 中 robocopy 传入不存在的 /NBT 参数、导致内嵌前端目录从未同步的问题；改用退出码判断并加 index.html 指纹比对兜底。'
      },
      {
        tag: '工具链',
        text: '打包脚本 build.ps1 补前置检查：缺少 go、缺少前端依赖、检测到运行中的 xingmang 时给出提示并中止。'
      },
      {
        tag: '工具链',
        text: '移除 86 个已失效的临时调试脚本；fanqie_* 与 _check_theme.js 从项目根移入 plugins/ 对应子目录；清理星月（旧项目）残留痕迹。'
      }
    ]
  },
  {
    version: '0.7.6.9',
    date: '2026-10-12 ~ 10-15',
    basis: 'derived',
    commit: '（工作区未提交，计入 6d0ce1f）',
    summary: '经验分享详情页布局重排，正文排版与阅读参数细化',
    brief: [
      '详情页布局重排：作者卡片进左栏、题图置顶、左栏可折叠',
      '评论区改版 + 楼中楼落地',
      '正文默认字体与字号可调，中栏加宽',
      '卡片「管理」菜单误触改为仅单击触发'
    ],
    details: [
      {
        tag: '界面',
        text: '详情页布局连续重排：作者卡片移入左栏、题图置顶；左栏改可折叠 rail 形态；目录树改为仿贴吧样式；删掉整条顶栏与右栏分区块。'
      },
      {
        tag: '界面',
        text: '评论区改版，楼中楼落地；主帖底部精简、回复区无痕化、主帖卡版式按参考图重做。'
      },
      {
        tag: '界面',
        text: '帖内正文默认字体改为微软雅黑（此前华文宋体在部分设备上发虚），右上角加字体选择器与字号下拉。'
      },
      {
        tag: '界面',
        text: '正文中栏宽度加宽 1.2 倍，并按截图宽度设了宽度下限，避免窄屏把正文压成细条。'
      },
      {
        tag: '修复',
        text: '经验分享一级主页卡片「管理」菜单误触：原为悬停即展开，改为仅单击触发。'
      },
      {
        tag: '修复',
        text: '正文中文字体发虚的排查与修法落地（换字体 + 字号可控）。'
      }
    ]
  },
  {
    version: '0.7.6.8',
    date: '2026-10-10 ~ 10-11',
    basis: 'derived',
    commit: '（工作区未提交，计入 6d0ce1f）',
    summary: '经验分享图片代理缓存、沉浸式发帖/编辑页与正文插图',
    brief: [
      '新增图片代理磁盘缓存，解决外链图床防盗链',
      '发帖与编辑改为沉浸式整页，版式对齐贴吧',
      '正文支持插图：点击 / 拖拽 / 粘贴',
      '帖子改为整帖文件化存储，旧帖保持双读'
    ],
    details: [
      {
        tag: '功能',
        text: '新增图片代理磁盘缓存（data/image-cache/），从根上解决外链图床防盗链导致的封面与插图不显示。'
      },
      {
        tag: '功能',
        text: '正文支持插图，点击、拖拽、粘贴三种方式都可插入。'
      },
      {
        tag: '功能',
        text: '帖子改为整帖文件化存储，同时保留旧帖双读兼容，不做破坏性迁移。'
      },
      {
        tag: '界面',
        text: '发帖弹窗按贴吧风格重做；随后升级为沉浸式发帖页（方案 C+D 合成）。'
      },
      {
        tag: '界面',
        text: '沉浸式编辑页与发帖页保持一致；详情页内的「编辑」也统一为复刻发帖弹窗，最终定稿为原地编辑。'
      },
      {
        tag: '修复',
        text: '相关推荐文字错位；卡片图片不显示（即上述防盗链问题）。'
      }
    ]
  },
  {
    version: '0.7.6.7',
    date: '2026-10-09',
    basis: 'derived',
    commit: '（工作区未提交，计入 6d0ce1f）',
    summary: '经验分享主页改为贴吧复刻版，并补齐封面相关接口',
    brief: [
      '经验分享主页按贴吧复刻落地，含分区、吧数据、置顶与精华',
      '复刻版三项缺陷修复',
      '封面上传后端接口补全'
    ],
    details: [
      {
        tag: '界面',
        text: '经验分享主页改版并落地为贴吧复刻版：吧首页含吧数据、最新发布、置顶（红）与精华（绿）标识，徽标按 PDF / 心得区分颜色。'
      },
      {
        tag: '修复',
        text: '复刻版三项缺陷修复（列表与详情页的样式与交互对齐）。'
      },
      {
        tag: '修复',
        text: '封面上传后端接口补全，此前只有前端入口。'
      },
      {
        tag: '修复',
        text: '写作页智能排版两处结果不符预期；智能排版 / 标点「生效一秒后被复原」（状态被回写覆盖）。'
      }
    ]
  },
  {
    version: '0.7.6.6',
    date: '2026-10-08',
    basis: 'derived',
    commit: '（工作区未提交，计入 6d0ce1f）',
    summary: '抽卡生成器与 AI 对话的若干修复',
    brief: [
      '生成器弹窗思考内容默认折叠',
      '修复 Gemini 空 content 导致的上游 400',
      '固定提示词不再跟随提示词库刷新而变化'
    ],
    details: [
      {
        tag: '界面',
        text: '变量抽卡页生成器弹窗的思考内容默认折叠，不再一上来就铺满整屏。'
      },
      {
        tag: '修复',
        text: 'AI 对话空 content 会被中转转成 Gemini 未初始化 Part、导致上游报「parts[0].data required oneof」400；后端改为丢弃空 content，思考内容以独立字段下发。'
      },
      {
        tag: '修复',
        text: '固定提示词会跟随提示词库刷新而变化，改为固定不跟随（纯前端）。'
      }
    ]
  },
  {
    version: '0.7.6.5',
    date: '2026-10-07',
    basis: 'derived',
    commit: '（工作区未提交，计入 6d0ce1f）',
    summary: '工作流模块重建：IDE 三栏工作台复刻，并改为两级结构',
    brief: [
      '工作流从零重建，复刻源版 IDE 三栏工作台',
      '7 种节点类型，代码 / 可视化双视图',
      '改为两级结构：卡片墙 → 全屏画布',
      '后端补齐 workflows / runs / outputs 三表与接口'
    ],
    details: [
      {
        tag: '功能',
        text: '工作流模块重建（该模块在 0.7.6.1 曾被整体移除）。按源版复刻为 IDE 三栏工作台：左侧文件树 + 中栏编辑器 + 右侧面板，含多标签页与欢迎屏。'
      },
      {
        tag: '功能',
        text: '7 种节点类型（开始 / AI 处理 / 结束 / 文本 / 合并 / 循环 / 上下文），每种带独立配色与状态标签；画布支持代码（Monaco）与可视化（Vue Flow）双视图切换。'
      },
      {
        tag: '功能',
        text: '改为两级结构：/workflow 是工作流卡片墙，点卡片进入 /workflow/run/:id 全屏画布。'
      },
      {
        tag: '功能',
        text: '后端补齐 workflows / workflow_runs / workflow_outputs 三表与 8 组接口；产物按「名称_日期_时间_runId」落盘，路径拼接拒绝绝对路径与 .. 穿越。'
      },
      {
        tag: '修复',
        text: '画布状态色不刷新的两个根因：画布深拷贝了节点导致状态永远是初始值；自定义节点的 data 不跟随 store 重渲染。改为引用 store 节点数组 + 运行时状态叠加层。'
      }
    ]
  },
  {
    version: '0.7.6.4',
    date: '2026-10-01 ~ 10-03',
    basis: 'derived',
    commit: '（工作区未提交，计入 6d0ce1f）',
    summary: '新增「星芒 Agent」板块，提示预览顶部栏改版',
    brief: [
      '新增「星芒 Agent」对话式写作助手板块',
      'Agent 支持思考链折叠、工具调用与三档写权限',
      '历史对话改为右上角下拉，输入区为居中悬浮卡',
      '提示预览页顶部布局改版'
    ],
    details: [
      {
        tag: '功能',
        text: '新增「星芒 Agent」独立板块：对话式写作助手，含对话界面 AB 结构、居中悬浮卡片式输入区、可折叠思考卡、工具调用后的「任务容器」排版、助手回答可编辑且编辑内容进入上下文。'
      },
      {
        tag: '功能',
        text: 'Agent 写能力按三档权限落地；暗色适配补齐。'
      },
      {
        tag: '修复',
        text: '思考链相关三处：`<think>` 内联推理分离、思考动画即时化、思考链开关与折叠可发现性；另修正「重新生成」的语义。'
      },
      {
        tag: '界面',
        text: '输入框改为「居中悬浮卡」，历史对话收进右上角时钟下拉，回到 Agent 页不再总是空状态。'
      },
      {
        tag: '界面',
        text: '提示预览页顶部布局改版（方案 C + 排序下移）。'
      },
      {
        tag: '修复',
        text: 'books.cover 为 NULL 导致书列表接口 500。'
      }
    ]
  },
  {
    version: '0.7.6.3',
    date: '2026-09-29',
    basis: 'derived',
    commit: '（工作区未提交，计入 6d0ce1f）',
    summary: 'AI 分析弹窗支持关联章节，章节正文截断修复与在线缓存',
    brief: [
      'AI 分析弹窗支持关联章节并带完整提示词',
      '修复章节正文截断：2000 字只剩 1300 字',
      '七猫接入关联章节',
      '新增在线章节 24 小时本地缓存'
    ],
    details: [
      {
        tag: '功能',
        text: 'AI 分析弹窗支持关联章节并展示完整提示词，分析结果可直接对着原文核对。'
      },
      {
        tag: '功能',
        text: '七猫接入「关联章节」，成为 ChapterLister 的第二个实现。'
      },
      {
        tag: '功能',
        text: '新增在线章节 24 小时本地缓存，避免同一章反复拉取。'
      },
      {
        tag: '修复',
        text: '章节正文截断：2000 字的章节抓回来只剩 1300 字。'
      }
    ]
  },
  {
    version: '0.7.6.2',
    date: '2026-09-28',
    basis: 'release',
    commit: 'ef12d9e',
    summary: '经验分享改贴吧版式，修复提示词导入导出失效与打包版白屏',
    brief: [
      '经验分享改为贴吧式版式，用分区替代原分类',
      '知识图谱节点拖拽发散的严重 bug 已修',
      '提示词库六个导入 / 导出入口点击无反应的问题已修',
      '打包版首次打开白屏已修',
      '新增 5 张星空风格宣传海报'
    ],
    details: [
      {
        tag: '界面',
        text: '经验分享改为贴吧式版式：新增分区（上限 15 个）替代原分类，精华与分区并行；详情页复刻贴吧楼层结构，含楼中楼。'
      },
      {
        tag: '修复',
        text: '知识图谱节点拖拽时严重发散（「轻拖飞出十万八千里、其余节点压成一点」），根因是拖拽期间未冻结包围盒导致正反馈；同步修掉拖拽期间仍弹悬停详情。'
      },
      {
        tag: '修复',
        text: '提示词库「恢复整体备份 / 导入卡包 / 导入独立卡片 / txt·md 转换导入 / 导出卡包 / 导出独立卡片」六项点击无反应。根因是瘦身时抽出的 PromptIODialogs 只加了 import、没在模板里挂载，调用被可选链静默吞掉。'
      },
      {
        tag: '修复',
        text: '打包 exe 启动后白屏。根因是 index.html 里阻塞加载 Google Fonts，外网不通时首屏一直卡住、JS 不执行；改为非阻塞加载并补 noscript 兜底。'
      },
      {
        tag: '修复',
        text: '写作页 AI 工具收藏上限从 5 提到 10。'
      },
      {
        tag: '界面',
        text: '新增 5 张星空风格宣传海报（docs/posters/），含星尘、星轨、极简、星云、星座五种视觉。'
      },
      {
        tag: '工具链',
        text: '尝试把版本号写进打包脚本，因不接受该做法已回退；build.bat 保持纯 ASCII + CRLF 的无版本号形态。'
      }
    ]
  },
  {
    version: '0.7.6.1',
    date: '2026-09-23',
    basis: 'release',
    commit: 'ef12d9e',
    summary: 'AI 扫榜扩到四源并新增跨天快照；提示词库与书籍库重构，新增心得与锁定',
    brief: [
      'AI 扫榜接入七猫 / 起点 / 飞卢，共四源',
      '新增跨天快照：涨跌从「只有方向」升级为「有幅度」',
      '新增提示词心得（多条、带评分与标签）与卡片锁定',
      '移除工作流模块',
      '知识图谱改用 sigma + d3-force 重写'
    ],
    details: [
      {
        tag: '功能',
        text: 'AI 扫榜从单一番茄扩到四源：番茄 / 七猫 / 起点 / 飞卢，改为注册表驱动，后续新增数据源不必改前端。'
      },
      {
        tag: '功能',
        text: '新增 rank_snapshots 跨天快照，把「只有涨跌方向」补成「有涨跌幅度」，并给出在榜天数与 NEW 标记；起点支持整榜抓取与「掉榜」列表。'
      },
      {
        tag: '功能',
        text: '七猫换到分类接口，题材维度到位（含「题材清单跨频道」问题的零额外请求修法）。'
      },
      {
        tag: '功能',
        text: '新增提示词心得（可写多条，带评分与标签），列表接口用子查询直接带出心得条数与平均分。'
      },
      {
        tag: '功能',
        text: '新增提示词卡片锁定（prompts.is_locked），锁定后只读——禁编辑、禁删除、禁写心得，导出仍可用。'
      },
      {
        tag: '界面',
        text: '提示词卡片操作收进右下角「⋮」（编辑 / 导出 / 写心得 / 锁定·解锁 / 删除）；提示词预览顶部栏重构。'
      },
      {
        tag: '界面',
        text: '知识图谱改用 sigma + d3-force 重写，拆成画布 / 工具栏 / 力度面板 / 详情浮层 / 图例 / 侧栏等独立组件，新增实体合并与版本快照。'
      },
      {
        tag: '界面',
        text: '创作页拆分，抽出 6 个弹窗组件与独立样式文件；主界面按皮肤令牌对齐主题。'
      },
      {
        tag: '界面',
        text: 'AI 扫榜默认数据源固定为番茄（去掉「记忆上次选择」）；页头改三栏栅格使切换器真居中，右侧加设置弹窗设默认常用榜单。'
      },
      {
        tag: '重构',
        text: '移除工作流模块：后端 handler、前端视图与路由入口全部摘除，无残留引用（该模块在 0.7.6.5 重建）。'
      },
      {
        tag: '重构',
        text: 'Books.vue 从 4642 行降到 2326 行，Prompts.vue 同步大幅瘦身，抽出弹窗组件、composables 与工具模块。'
      },
      {
        tag: '工具链',
        text: '新增样式自检闭环（编译真实 scoped CSS → 亮 / 暗两套截图 → 对比度探针），以及两篇大视图瘦身工程规程文档。'
      }
    ]
  },
  {
    version: '0.7.6.0',
    date: '2026-09-17',
    basis: 'release',
    commit: '5e5d425c · 35e4de85',
    summary: 'Web 版升级为桌面版源码，并入 Go 单文件分发',
    brief: [
      '源码快照发布，Web 版正式升级为桌面版',
      '改为 Go 单文件 exe，内嵌前端产物',
      '不再依赖 Node 服务端'
    ],
    details: [
      {
        tag: '重构',
        text: 'GitHub 上发布「v0.7.6.0 源码快照」并合并至 main，Web 版（server/ + novel.db）正式升级为桌面版源码：前端产物内嵌进 Go 单文件 exe，不再需要单独跑 Node 服务端。'
      },
      {
        tag: '基线',
        text: '此节点之后进入桌面版独立演进；此前 0.6.x 的 Web 版历史见下方各条。'
      }
    ]
  },
  {
    version: '0.740',
    date: '2026-09-11',
    basis: 'derived',
    commit: 'b71d050',
    summary: '手机预览阅读器重构，打包脚本支持版本号',
    brief: [
      '手机预览重做为带翻页动画的阅读器',
      '新增阅读设置、跟随滚动、预排版',
      '打包脚本改造，支持按版本号归档'
    ],
    details: [
      {
        tag: '功能',
        text: '写作页「手机预览」重构为完整阅读器：抽出 ReaderSheets 渲染层，新增 useReaderPager（翻页）、useReaderSettings（阅读设置）、useFollowScroll（跟随滚动）与 reader-theme.css，含多种阅读配色。'
      },
      {
        tag: '功能',
        text: '新增 utils/readerText.ts 处理阅读态文本切分与预排版。'
      },
      {
        tag: '工具链',
        text: '打包脚本改造（build.bat / build.ps1 / tools/release）：支持指定版本号归档、明确不打归档开关、参数大小写兼容，统一了产物命名规则。'
      }
    ]
  },
  {
    version: '0.739',
    date: '2026-09-10',
    basis: 'derived',
    commit: '2315ea6',
    summary: '写作页新增审稿、插画与敏感词工具',
    brief: [
      '新增 AI 审稿面板',
      '新增插画生成面板',
      '新增敏感词检测面板',
      '独立出写作页主题样式文件'
    ],
    details: [
      {
        tag: '功能',
        text: '新增 AI 审稿面板（AiReviewDialog，942 行）。'
      },
      {
        tag: '功能',
        text: '新增插画面板（IllustrationDialog）与敏感词检测面板（SensitiveWordPanel）。'
      },
      {
        tag: '界面',
        text: '抽出 write-theme.css 承载写作页主题；写作页与其侧栏、AI 工具常量表（aiTools.ts）同步调整。'
      },
      {
        tag: '重构',
        text: '删除 SceneOutlineMap.vue（场景大纲图），功能并入 AI 工具面板。'
      }
    ]
  },
  {
    version: '0.738',
    date: '2026-08-31',
    basis: 'derived',
    commit: '2232d1f',
    summary: '写作页 AI 工具面板群落地',
    brief: [
      '写作页一次性接入 11 个 AI 工具面板',
      '侧栏 AI 面板、续写、角色库、台词库、词频、人名生成等',
      '写作页暗色适配补齐'
    ],
    details: [
      {
        tag: '功能',
        text: '写作页接入一批 AI 工具面板：侧栏 AI 面板（AiSidePanel）、AI 续写（AiContinueDialog）、角色库（CharacterLibraryDialog）、悖论箱 / 草稿箱（DraftBoxDialog）、人名生成（NameGeneratorDialog）、台词库（TermLibraryDialog）、词频统计（WordFrequencyPanel）、参考面板（ReferencePanel）、划词工具条（SelectionToolbar）等。'
      },
      {
        tag: '功能',
        text: '新增 constants/aiTools.ts 统一登记 AI 工具清单，后续增删工具只改一处。'
      },
      {
        tag: '界面',
        text: '写作页与目录树（WriteCatalogTree）配套调整，补 135 行暗色样式。'
      }
    ]
  },
  {
    version: '0.737',
    date: '2026-08-29',
    basis: 'derived',
    commit: '0a14579',
    summary: '新增拆书 AI 分析弹窗，扫榜界面继续打磨',
    brief: ['新增 AI 分析弹窗（拆书）', '扫榜页交互与展示打磨'],
    details: [
      {
        tag: '功能',
        text: '新增 AI 分析弹窗（AiAnalyzeDialog，921 行），用于对拆书结果做 AI 解读。'
      },
      {
        tag: '界面',
        text: 'AI 扫榜页继续打磨：列表展示与交互细节调整。'
      }
    ]
  },
  {
    version: '0.736',
    date: '2026-08-23',
    basis: 'derived',
    commit: 'a1eef2c',
    summary: '主题皮肤体系成型，提示词库与扫榜界面大改',
    brief: [
      '主界面主题 / 皮肤体系成型',
      '提示词库界面大幅扩充',
      'AI 扫榜界面重做',
      '消息中心改版'
    ],
    details: [
      {
        tag: '界面',
        text: '主界面主题与皮肤体系成型：主题模式、皮肤强调色、护眼模式统一在 MainLayout 收敛，暗色变量补齐。'
      },
      {
        tag: '界面',
        text: '提示词库界面大幅扩充（+990 行），卡片与分类交互重做。'
      },
      {
        tag: '界面',
        text: 'AI 扫榜界面重做（+344 行）；消息中心改版；API 配置页同步调整。'
      }
    ]
  },
  {
    version: '0.735',
    date: '2026-08-15',
    basis: 'release',
    commit: '0542c4b · acad04f · 9fd453e',
    summary: '打包支持版本号，内置番茄榜单',
    brief: [
      '内置番茄小说榜单数据源，界面开箱即用',
      '单实例复用：重复启动只在浏览器新开标签',
      '打包支持 -Version，产物按版本号归档'
    ],
    details: [
      {
        tag: '功能',
        text: 'AI 扫榜内置番茄小说数据源，Go 原生实现（含 a_bogus 签名与 sm3），不再依赖 Python / Node 环境。'
      },
      {
        tag: '功能',
        text: '单实例复用：服务已在运行时再次双击 exe，只在浏览器里新开一个标签页，不会重复启动服务、也不会因端口被占用而报错。'
      },
      {
        tag: '功能',
        text: '提示词支持版本信息与元数据字段。'
      },
      {
        tag: '界面',
        text: 'PackDetail 提示词弹窗功能已同步到独立页面。'
      },
      {
        tag: '工具链',
        text: '打包工具新增 -Version 参数，归档可按版本号命名；build.bat 重构为直接委托 Go 打包工具，与 build.ps1 行为对齐。'
      },
      {
        tag: '工具链',
        text: 'MarkCardDialog 重构为可辨识联合类型，按 type 安全收窄；修正 SplitRichTextEditor 的 Vditor i18n 类型声明不匹配。'
      }
    ]
  },
  {
    version: '0.733 / 0.734',
    date: '2026-08-14',
    basis: 'release',
    commit: '4b5f51f',
    summary: '桌面版首个可分发版本',
    brief: [
      '桌面版首个打包分发的版本',
      '我的小说 / 提示词库 / 创作 / AI 扫榜等基础模块成形',
      '番茄榜单爬虫初版'
    ],
    details: [
      {
        tag: '基线',
        text: '桌面版首个打包分发的版本，功能模块：我的小说 · 提示词库 · 提示预览 · 变量抽卡 · 创作 · AI 扫榜 · wiki 图谱 · 经验分享 · 插件系统 · API 配置 · 个人中心。'
      },
      {
        tag: '基线',
        text: '番茄榜单爬虫初版落地，为后续四源扫榜与跨天快照打基础。'
      },
      {
        tag: '工具链',
        text: '桌面窗口依赖 Microsoft Edge / WebView2 运行时（后续版本改为在系统默认浏览器中打开界面）。'
      }
    ]
  },
  {
    version: '0.6.8.1',
    date: '2026-06-22',
    basis: 'release',
    commit: 'c97819e4（GitHub）',
    summary: 'Web 版阶段性更新',
    brief: ['Web 版打包发布（V-0.6.8.1）', '发布说明仅标注「更新」'],
    details: [
      {
        tag: '基线',
        text: 'GitHub 发布星芒写作 V-0.6.8.1（.7z 分发包）。该次发布说明只写了「更新」两字，没有逐条列出改动 —— 作为 Web 版演进的节点保留。'
      }
    ]
  },
  {
    version: '0.6.5.7',
    date: '2026-05-20',
    basis: 'release',
    commit: '926bcf18（GitHub）',
    summary: '世界书分组、卡包卡片独立操作与官方社区入口',
    brief: [
      '世界书支持建分组；增加酒馆世界书兼容转换',
      '提示词卡包内卡片可单独导出 / 转包 / 删除',
      '增加功能反馈文档与官方社区按钮',
      '删除正则表达式功能；修复模型切换与追问不显示'
    ],
    details: [
      {
        tag: '功能',
        text: '增加酒馆（SillyTavern）世界书兼容转换：可导入世界书创建新的，也可直接导入其他世界书。'
      },
      {
        tag: '功能',
        text: '世界书支持创建分组。'
      },
      {
        tag: '功能',
        text: '提示词管理卡包弹窗内，卡片支持单独导出、转入其他卡包、单独删除。'
      },
      {
        tag: '功能',
        text: '创建生成器时可自动创建卡包；模型选择增加更多调试按钮（仅 AI 对话板块）。'
      },
      {
        tag: '界面',
        text: '调整部分 UI；增加功能反馈文档与官方社区按钮。'
      },
      {
        tag: '修复',
        text: '修复 AI 对话模型切换问题；修复变量抽卡区追问不显示。'
      },
      {
        tag: '重构',
        text: '删除正则表达式功能。'
      }
    ]
  },
  {
    version: '0.6.5.1',
    date: '2026-05-20',
    basis: 'release',
    commit: '926bcf18（GitHub）',
    summary: '世界书注入系统与拓展中心',
    brief: [
      'AI 对话增加世界书注入系统',
      '知识图谱与世界书打通，实体节点右键可导入世界书',
      '新增拓展中心，可在此关闭世界书'
    ],
    details: [
      {
        tag: '功能',
        text: 'AI 对话增加世界书注入系统。'
      },
      {
        tag: '功能',
        text: '知识图谱与世界书对接打通：实体节点可通过右键一键导入世界书。'
      },
      {
        tag: '功能',
        text: '新增拓展中心板块：不需要世界书时可在其中取消；默认完全复刻酒馆世界书的默认状态且不连接，需要手动按按钮连接。'
      }
    ]
  },
  {
    version: '0.6.4.4',
    date: '2026-04-25',
    basis: 'release',
    commit: '050c2708（GitHub）',
    summary: 'Web 版迭代（发布说明未列明细）',
    brief: ['Web 版打包发布（V-0.6.4.4）', '发布说明仅含 compare 链接'],
    details: [
      {
        tag: '基线',
        text: 'GitHub 发布 V-0.6.4.4（.7z 分发包）。发布说明只有一个版本对比链接、未列改动明细 —— 作为 Web 版演进的节点保留。'
      }
    ]
  },
  {
    version: '0.6.4.2',
    date: '2026-04-19',
    basis: 'release',
    commit: 'ddf5a8aa（GitHub，tag v1）',
    summary: 'Web 版首个公开版本：知识图谱、内置提示词与 API 预设',
    brief: [
      'Web 版首个公开发布',
      '知识图谱：先建图谱文件，再选章节分析生成',
      '内置 5 个提示词与 4 个 API 设置（key 自备）'
    ],
    details: [
      {
        tag: '基线',
        text: '星芒 Web 版首个公开发布（V-0.6.4.2）。形态为本地 Node 服务 + 浏览器界面，数据存 server/database/novel.db。'
      },
      {
        tag: '功能',
        text: '知识图谱：需先创建图谱文件，再选择章节进行分析，即可生成图谱内容。'
      },
      {
        tag: '功能',
        text: '初步内置五个提示词与四个 API 设置（仅基础配置，key 需自行准备）。'
      },
      {
        tag: '修复',
        text: '修复部分 bug。'
      }
    ]
  }
]
