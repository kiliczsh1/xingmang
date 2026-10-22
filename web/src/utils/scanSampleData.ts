// 扫榜样例数据（从 ai扫榜.md 参考快照提取，UI 先行占位，后续由后端接口替代）

export interface ScanTab { key: string; label: string }
export interface ScanCategory { id: string; name: string; count: number }
export interface ScanBook {
  rank: number; title: string; author: string; category: string; status: string;
  reading: string; words: string; coverUrl: string; link: string
}

export const scanTabs: ScanTab[] = [
  { key: "short", label: "短篇榜单" },
  { key: "male_read", label: "男频阅读榜" },
  { key: "male_new", label: "男频新书榜" },
  { key: "female_read", label: "女频阅读榜" },
  { key: "female_new", label: "女频新书榜" }
]

export const scanCategories: ScanCategory[] = [
  { id: "total", name: "总榜", count: 16393 },
  { id: "都市高武", name: "都市高武", count: 2146 },
  { id: "玄幻脑洞", name: "玄幻脑洞", count: 1477 },
  { id: "悬疑脑洞", name: "悬疑脑洞", count: 1296 },
  { id: "都市日常", name: "都市日常", count: 1238 },
  { id: "都市脑洞", name: "都市脑洞", count: 1235 },
  { id: "历史古代", name: "历史古代", count: 1008 },
  { id: "传统玄幻", name: "传统玄幻", count: 937 },
  { id: "东方仙侠", name: "东方仙侠", count: 782 },
  { id: "都市修真", name: "都市修真", count: 699 },
  { id: "历史脑洞", name: "历史脑洞", count: 690 },
  { id: "悬疑灵异", name: "悬疑灵异", count: 654 },
  { id: "科幻末世", name: "科幻末世", count: 639 },
  { id: "男频衍生", name: "男频衍生", count: 634 },
  { id: "动漫衍生", name: "动漫衍生", count: 584 },
  { id: "都市种田", name: "都市种田", count: 580 },
  { id: "游戏体育", name: "游戏体育", count: 577 },
  { id: "抗战谍战", name: "抗战谍战", count: 508 },
  { id: "西方奇幻", name: "西方奇幻", count: 428 },
  { id: "战神赘婿", name: "战神赘婿", count: 279 },
]

export const scanBooks: ScanBook[] = [
  { rank: 1, title: "诸神愚戏", author: "一月九十秋", category: "都市高武", status: "已完结", reading: "276.02万人", words: "368.78万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/a8816f21-338b-4624-867a-1a4632a2e206.png", link: "https://fanqienovel.com/page/7338034838672247833" },
  { rank: 2, title: "十日终焉", author: "杀虫队队员", category: "悬疑脑洞", status: "已完结", reading: "259.03万人", words: "320.13万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/fbb4fa52-53b9-4710-962e-d00ba9ed3bd5.jpg", link: "https://fanqienovel.com/page/7143038691944959011" },
  { rank: 3, title: "我不是戏神", author: "三九音域", category: "都市高武", status: "已完结", reading: "255.06万人", words: "400.36万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/29528f3c-aaaa-4827-983e-795eae93427b.jpg", link: "https://fanqienovel.com/page/7276384138653862966" },
  { rank: 4, title: "这个游戏不对劲，我挖矿成神！", author: "道小易", category: "玄幻脑洞", status: "连载中", reading: "253.66万人", words: "159.12万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/27/0a201141-b206-4507-8270-f877c505509b.jpg", link: "https://fanqienovel.com/page/7590221243043826712" },
  { rank: 5, title: "我在精神病院学斩神", author: "三九音域", category: "都市高武", status: "已完结", reading: "190.20万人", words: "426.37万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/62151ccc-40f5-436b-bd8a-6a6a22c5a63a.jpg", link: "https://fanqienovel.com/page/6982529841564224526" },
  { rank: 6, title: "时停起手，邪神也得给我跪下！", author: "六个葫芦", category: "都市高武", status: "连载中", reading: "163.78万人", words: "219.87万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/33e9317c-5627-40b6-8d6d-ec9d22ed853d.jpg", link: "https://fanqienovel.com/page/7504849932138859545" },
  { rank: 7, title: "全球死考：刚下刑场你让我去考试", author: "我爱吃猫片哇", category: "悬疑脑洞", status: "连载中", reading: "152.85万人", words: "58.22万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/05/05/65f51d82-8796-4336-a682-ca40cd428bdd.jpg", link: "https://fanqienovel.com/page/7623686597409508376" },
  { rank: 8, title: "惊鸿", author: "一夕烟雨", category: "传统玄幻", status: "连载中", reading: "134.56万人", words: "904.80万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/1c1e4164-c6dd-4128-b6d5-9399625765e1.jpg", link: "https://fanqienovel.com/page/6883748331202284558" },
  { rank: 9, title: "规则怪谈：无所谓，系统会出手", author: "巡山主神", category: "悬疑脑洞", status: "连载中", reading: "121.17万人", words: "104.33万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/0b1b226e-fe15-4cb1-9929-0bb161dd5aff.png", link: "https://fanqienovel.com/page/7552067105378536473" },
  { rank: 10, title: "我操作的少女正在成为旧日之主", author: "陷阵营营长", category: "都市高武", status: "连载中", reading: "120.43万人", words: "58.64万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/05/12/2f97cbe5-1153-450a-88ff-c16c477ebc06.jpg", link: "https://fanqienovel.com/page/7592279900040465433" },
  { rank: 11, title: "天渊", author: "沐潇三生", category: "传统玄幻", status: "连载中", reading: "101.40万人", words: "596.38万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/14da7dae-98f8-4e13-aa12-9cb7ec06b546.png", link: "https://fanqienovel.com/page/7138249318909086727" },
  { rank: 12, title: "归国留洋水货？叫我芯片之父！", author: "这肉有毒", category: "都市脑洞", status: "连载中", reading: "95.29万人", words: "75.39万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/04/25/7ccf7118-3dab-40ff-86fd-e5b86c315379.png", link: "https://fanqienovel.com/page/7623581376586976281" },
  { rank: 13, title: "穿到女频，整顿满城霸总", author: "嘎嘣蹦", category: "都市日常", status: "连载中", reading: "92.01万人", words: "62.34万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/06/04/a01bf7ee-48e9-47e1-9747-67f5f87e8623.jpg", link: "https://fanqienovel.com/page/7637508604198652990" },
  { rank: 14, title: "诡舍", author: "夜来风雨声丶", category: "悬疑脑洞", status: "已完结", reading: "91.42万人", words: "213.31万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/2a5b62c1-c118-4b3d-971e-17d9e4910520.jpg", link: "https://fanqienovel.com/page/7256784068786785336" },
  { rank: 15, title: "天眼风水师", author: "道之光", category: "都市日常", status: "已完结", reading: "88.97万人", words: "130.69万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/0c9bd2d8-b7bd-4c6c-8bc4-16485277bfb1.png", link: "https://fanqienovel.com/page/7326876174989134910" },
  { rank: 16, title: "从变身少女开始斩妖除魔", author: "八爪触手怪", category: "玄幻脑洞", status: "连载中", reading: "87.90万人", words: "207.96万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/d89addb8-04b6-4e45-8a53-0789bc43fe1c.jpg", link: "https://fanqienovel.com/page/7544766518324644889" },
  { rank: 17, title: "嘉豪就变强？操场雨中舞剑开天门", author: "落叶之刃", category: "都市高武", status: "连载中", reading: "87.23万人", words: "58.55万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/05/25/8f2082b8-81a9-4b0e-b951-606102ebeda5.jpg", link: "https://fanqienovel.com/page/7638538561230228504" },
  { rank: 18, title: "我以狐仙镇百鬼", author: "紫梦游龙", category: "悬疑灵异", status: "连载中", reading: "85.62万人", words: "963.18万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/c9428d0e-efd0-420f-aa8a-d7165fc4e849.jpg", link: "https://fanqienovel.com/page/7087872099557051400" },
  { rank: 19, title: "系统赋我长生，我熬死了所有人", author: "一只榴莲3号", category: "玄幻脑洞", status: "连载中", reading: "85.37万人", words: "514.55万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/7f5a6c54-47be-4487-b822-c87743ee78ff.jpg", link: "https://fanqienovel.com/page/7228940788351110182" },
  { rank: 20, title: "开局长生万古，苟到天荒地老", author: "紫灵风雪", category: "玄幻脑洞", status: "已完结", reading: "82.31万人", words: "608.07万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/ecaa9b51-9f47-4024-a615-52feeeae020c.jpg", link: "https://fanqienovel.com/page/7147596144942517284" },
  { rank: 21, title: "大一实习，你跑去749收容怪物", author: "我爱罗的沙", category: "都市高武", status: "连载中", reading: "81.06万人", words: "355.25万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/e2746473-dabe-4a47-af2b-001ce3be4415.png", link: "https://fanqienovel.com/page/7421167583522458648" },
  { rank: 22, title: "冒姓琅琊", author: "东周公子南", category: "历史古代", status: "连载中", reading: "80.90万人", words: "141.06万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/92ca5a66-cc4b-4c15-bccc-b08bb30cdf49.png", link: "https://fanqienovel.com/page/7414011485757639704" },
  { rank: 23, title: "末日生存方案供应商", author: "板面王仔", category: "科幻末世", status: "已完结", reading: "77.11万人", words: "272.47万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/4d667628-f486-4524-a0a2-23aab01a6ffa.png", link: "https://fanqienovel.com/page/7496026299845053465" },
  { rank: 24, title: "我，最强毒士，女帝直呼活阎王", author: "星星子", category: "历史古代", status: "连载中", reading: "77.02万人", words: "382.69万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/7fb6077d-39d6-4b7e-8af0-e1e9003f01f5.png", link: "https://fanqienovel.com/page/7363901972950305854" },
  { rank: 25, title: "皇家金牌县令", author: "板面王仔", category: "历史古代", status: "已完结", reading: "72.70万人", words: "272.25万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/b9d20084-c714-4285-98c8-b25093f3508e.jpg", link: "https://fanqienovel.com/page/7046842344074644484" },
  { rank: 26, title: "我家世代提刀，到我这儿提笔了！", author: "牧野横舟", category: "历史古代", status: "连载中", reading: "69.79万人", words: "73.24万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/05/10/44584075-f442-4f8c-bb29-cad63bc03227.jpg", link: "https://fanqienovel.com/page/7623459019432152089" },
  { rank: 27, title: "你们刷怪啊，刷我干嘛！", author: "喜欢吃橘子", category: "都市高武", status: "连载中", reading: "68.73万人", words: "285.81万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/05da19dc-c132-4488-8443-5af70cc22c5d.png", link: "https://fanqienovel.com/page/7462405945180572697" },
  { rank: 28, title: "序列：吃神者", author: "不要大脑要小脑", category: "都市高武", status: "连载中", reading: "68.06万人", words: "373.44万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/2dd8b5d1-ed36-415e-baa6-460484e7c78b.jpg", link: "https://fanqienovel.com/page/7346057588347259929" },
  { rank: 29, title: "凡骨", author: "壹更大师", category: "东方仙侠", status: "已完结", reading: "67.96万人", words: "830.43万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/01/12/9e9b83f2-f1da-44f3-bd6d-26f847a8bc3e.png", link: "https://fanqienovel.com/page/7077516958534470656" },
  { rank: 30, title: "拥有八大面具的我，打造无上禁城", author: "笔起烟火", category: "都市高武", status: "连载中", reading: "67.81万人", words: "264.75万字", coverUrl: "https://cdn.zhiyuxiezuo.com/api/profile/download//2026/07/23/abfaf61f-4049-4b2e-8f43-d5c83948eb3d.jpg", link: "https://fanqienovel.com/page/7439362788444605465" },
]
