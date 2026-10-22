import type { ApiConfigType } from '@/types'

/**
 * 预置平台库（API 广场的数据源）。
 *
 * 数据来源：xingnovel 0.3.0 的广场预设（Ta 数组），共 14 个平台。
 * 搬运时排除 3 个：
 *   - vertex    ：无默认端点，预设不完整
 *   - xingpan   ：蓝本自有服务（xingpan.site）
 *   - aistartlab：蓝本生态的聚合平台，来源不明
 * 因此实际搬入 11 个平台。
 *
 * 纯本地内置，离线可用。模型 ID 只是常用值，各平台会随版本调整，创建后可在模型工作台修改。
 */

/** 参数组：图片 / 视频模型的可变生成参数（如 duration / resolution / ratio） */
export interface ParamGroup {
  name: string
  type: 'number' | 'text'
  options: { label: string; value: string | number }[]
}

/** 参数组默认值，如 { duration: 5, resolution: '720p' } */
export type DefaultParams = Record<string, string | number>

/**
 * 视频端点（蓝本原始字段）。
 *
 * 注意：这里**不预生成** request_rules —— 各家 response 里 task_id / video_url / status
 * 的路径不同，臆造会让轮询永远取不到值。S7 翻译时只填 submit / poll，
 * response 映射由用户在弹窗里按平台文档确认。
 */
export interface VideoEndpoints {
  create: string
  query: string
  queryMethod?: string
}

export interface PresetModel {
  /** 展示名 */
  name: string
  /** 模型标识（发给上游的 model 字段） */
  model: string
  config_type: ApiConfigType
  description?: string
  /** 平台侧标签，如「官方直连」「多参考图」 */
  tags?: string[]
  max_tokens?: number
  /** 参数组：仅 image / video 类型有意义 */
  paramGroups?: ParamGroup[]
  defaultParams?: DefaultParams
  /** 视频类模型的提交 / 查询端点 */
  videoEndpoints?: VideoEndpoints
}

export interface ProviderPreset {
  key: string
  name: string
  /** 卡片首字母/简写 */
  initial: string
  /** 平台主题色（来自蓝本） */
  color?: string
  provider_type: string
  api_format: 'openai' | 'claude' | 'gemini' | 'responses' | 'custom_json'
  api_url: string
  auth_type: 'api_key' | 'ak_sk' | 'group_id' | 'endpoint_id'
  use_full_url?: boolean
  /** 一句话说明 */
  desc: string
  /** 能力标签，用于广场筛选 */
  tags: ApiConfigType[]
  /** 本地服务：无需 API Key */
  noKey?: boolean
  /** 官方文档地址 */
  docsUrl?: string
  models: PresetModel[]
}

export const PROVIDER_PRESETS: ProviderPreset[] = [
  {
    key: "volcengine",
    name: "火山引擎（方舟）",
    initial: "火",
    color: "#4E6BFB",
    provider_type: "volcengine",
    api_format: "openai",
    api_url: "https://ark.cn-beijing.volces.com/api/v3",
    auth_type: 'api_key',
    desc: "字节跳动云服务平台，豆包、Seedance、Seedream 系列模型的官方直连接口。",
    tags: ['chat','image','video'],
    docsUrl: "https://www.volcengine.com/docs/82379",
    models: [
      {
        name: "豆包 Seed 1.6",
        model: "doubao-seed-1-6",
        config_type: "chat",
        description: "豆包新一代深度思考模型，兼顾推理速度与效果。",
        tags: ["官方直连"],
      },
      {
        name: "豆包 Pro 32K",
        model: "doubao-1-5-pro-32k-68k",
        config_type: "chat",
        description: "通用对话模型，长上下文稳定，性价比高。",
      },
      {
        name: "Seedance 1.0 Pro",
        model: "seedance-1-0-pro",
        config_type: "video",
        description: "专业级多模态视频创作模型，支持图生视频与首尾帧。",
        tags: ["多参考图"],
        paramGroups: [{"name":"duration","type":"number","options":[{"label":"5s","value":5},{"label":"10s","value":10}]},{"name":"resolution","type":"text","options":[{"label":"480p","value":"480p"},{"label":"720p","value":"720p"},{"label":"1080p","value":"1080p"}]},{"name":"ratio","type":"text","options":[{"label":"16:9","value":"16:9"},{"label":"9:16","value":"9:16"},{"label":"1:1","value":"1:1"},{"label":"4:3","value":"4:3"},{"label":"3:4","value":"3:4"}]}],
        defaultParams: {"duration":5,"resolution":"720p","ratio":"16:9"},
        videoEndpoints: { create: "/api/v3/contents/generations/tasks", query: "/api/v3/contents/generations/tasks/{id}", queryMethod: "GET" },
      },
      {
        name: "Seedance 1.0 Lite",
        model: "seedance-1-0-lite-t2v",
        config_type: "video",
        description: "轻量视频生成模型，速度快、成本低，适合批量出片。",
        paramGroups: [{"name":"duration","type":"number","options":[{"label":"5s","value":5},{"label":"10s","value":10}]},{"name":"ratio","type":"text","options":[{"label":"16:9","value":"16:9"},{"label":"9:16","value":"9:16"},{"label":"1:1","value":"1:1"},{"label":"4:3","value":"4:3"},{"label":"3:4","value":"3:4"}]}],
        defaultParams: {"duration":5,"ratio":"16:9"},
        videoEndpoints: { create: "/api/v3/contents/generations/tasks", query: "/api/v3/contents/generations/tasks/{id}", queryMethod: "GET" },
      },
      {
        name: "Seedream 4.0",
        model: "seedream-4-0",
        config_type: "image",
        description: "新一代图片生成模型，支持参考图与 2K 分辨率。",
        tags: ["多参考图"],
      },
    ]
  },
  {
    key: "dashscope",
    name: "阿里云百炼",
    initial: "阿",
    color: "#615CEF",
    provider_type: "openai",
    api_format: "openai",
    api_url: "https://dashscope.aliyuncs.com/compatible-mode/v1",
    auth_type: 'api_key',
    desc: "阿里云一站式大模型开发与应用平台，提供通义千问、通义万相等系列模型。",
    tags: ['chat','video'],
    docsUrl: "https://help.aliyun.com/zh/model-studio",
    models: [
      {
        name: "通义千问 Max",
        model: "qwen-max",
        config_type: "chat",
        description: "千亿级参数对话模型，复杂推理与创作能力强。",
      },
      {
        name: "通义千问 Plus",
        model: "qwen-plus",
        config_type: "chat",
        description: "均衡效果与成本的对话模型，适合日常写作。",
      },
      {
        name: "通义万相 文生视频",
        model: "wan2.2-t2v-plus",
        config_type: "video",
        description: "通义万相视频生成，支持多分辨率与时长选择。",
        paramGroups: [{"name":"duration","type":"number","options":[{"label":"5s","value":5}]},{"name":"ratio","type":"text","options":[{"label":"16:9","value":"16:9"},{"label":"9:16","value":"9:16"},{"label":"1:1","value":"1:1"},{"label":"4:3","value":"4:3"},{"label":"3:4","value":"3:4"}]}],
        defaultParams: {"duration":5,"ratio":"16:9"},
        videoEndpoints: { create: "/api/v1/services/aigc/video-generation/video-synthesis", query: "/api/v1/tasks/{id}", queryMethod: "GET" },
      },
    ]
  },
  {
    key: "minimax",
    name: "MiniMax",
    initial: "M",
    color: "#E85348",
    provider_type: "openai",
    api_format: "openai",
    api_url: "https://api.minimax.chat/v1",
    auth_type: 'api_key',
    desc: "MiniMax 大模型开放平台，提供 abab 对话与海螺视频生成系列。",
    tags: ['chat','video'],
    docsUrl: "https://platform.minimaxi.com/document",
    models: [
      {
        name: "abab 6.5s",
        model: "abab6.5s-chat",
        config_type: "chat",
        description: "MiniMax 通用对话模型，中文语境表现优秀。",
      },
      {
        name: "MiniMax Text 01",
        model: "MiniMax-Text-01",
        config_type: "chat",
        description: "长文本对话模型，支持百万级上下文。",
      },
      {
        name: "海螺 Director",
        model: "T2V-01-Director",
        config_type: "video",
        description: "海螺视频生成，支持导演运镜指令与多参考图。",
        tags: ["多参考图"],
        paramGroups: [{"name":"ratio","type":"text","options":[{"label":"16:9","value":"16:9"},{"label":"9:16","value":"9:16"},{"label":"1:1","value":"1:1"},{"label":"4:3","value":"4:3"},{"label":"3:4","value":"3:4"}]}],
        defaultParams: {"ratio":"16:9"},
        videoEndpoints: { create: "/v1/video_generation", query: "/v1/query/video_generation?task_id={id}", queryMethod: "GET" },
      },
    ]
  },
  {
    key: "kling",
    name: "可灵 Kling",
    initial: "K",
    color: "#2DB4F2",
    provider_type: "kling",
    api_format: "openai",
    api_url: "https://api.klingai.com",
    auth_type: 'api_key',
    desc: "快手可灵大模型，视频生成效果逼真，支持文生视频与图生视频。",
    tags: ['video'],
    docsUrl: "https://app.klingai.com/global/dev/document-api",
    models: [
      {
        name: "可灵 1.6",
        model: "kling-v1-6",
        config_type: "video",
        description: "可灵视频标准版，画面质感与运动自然度高。",
        paramGroups: [{"name":"duration","type":"number","options":[{"label":"5s","value":5},{"label":"10s","value":10}]},{"name":"ratio","type":"text","options":[{"label":"16:9","value":"16:9"},{"label":"9:16","value":"9:16"},{"label":"1:1","value":"1:1"},{"label":"4:3","value":"4:3"},{"label":"3:4","value":"3:4"}]}],
        defaultParams: {"duration":5,"ratio":"16:9"},
        videoEndpoints: { create: "/v1/videos/text2video", query: "/v1/videos/text2video/{id}", queryMethod: "GET" },
      },
      {
        name: "可灵 2.0 Master",
        model: "kling-v2-master",
        config_type: "video",
        description: "可灵 2.0 专业版，语义理解与运镜控制更强。",
        paramGroups: [{"name":"duration","type":"number","options":[{"label":"5s","value":5}]},{"name":"ratio","type":"text","options":[{"label":"16:9","value":"16:9"},{"label":"9:16","value":"9:16"},{"label":"1:1","value":"1:1"},{"label":"4:3","value":"4:3"},{"label":"3:4","value":"3:4"}]}],
        defaultParams: {"duration":5,"ratio":"16:9"},
        videoEndpoints: { create: "/v1/videos/text2video", query: "/v1/videos/text2video/{id}", queryMethod: "GET" },
      },
    ]
  },
  {
    key: "gemini",
    name: "Google Gemini",
    initial: "G",
    color: "#3E8EF7",
    provider_type: "gemini",
    api_format: "gemini",
    api_url: "https://generativelanguage.googleapis.com",
    auth_type: 'api_key',
    desc: "Google Gemini 原生接口，支持对话、图片生成与 Veo 视频生成。",
    tags: ['chat','video'],
    docsUrl: "https://ai.google.dev/gemini-api/docs",
    models: [
      {
        name: "Gemini 2.5 Flash",
        model: "gemini-2.5-flash",
        config_type: "chat",
        description: "速度快、性价比高的多模态对话模型。",
      },
      {
        name: "Gemini 2.5 Pro",
        model: "gemini-2.5-pro",
        config_type: "chat",
        description: "旗舰思考模型，复杂推理与长上下文能力强。",
      },
      {
        name: "Veo 3",
        model: "veo-3.0-generate-001",
        config_type: "video",
        description: "Google Veo 视频生成，支持高分辨率与原生音轨。",
        videoEndpoints: { create: "/v1beta/models/{model}:predictLongRunning", query: "/v1beta/{operation_name}", queryMethod: "GET" },
      },
    ]
  },
  {
    key: "openai",
    name: "OpenAI",
    initial: "O",
    color: "#10A37F",
    provider_type: "openai",
    api_format: "openai",
    api_url: "https://api.openai.com/v1",
    auth_type: 'api_key',
    desc: "OpenAI 官方接口，提供 GPT 对话、图像与 Sora 视频生成。",
    tags: ['chat','image','video'],
    docsUrl: "https://platform.openai.com/docs",
    models: [
      {
        name: "GPT-4o",
        model: "gpt-4o",
        config_type: "chat",
        description: "多模态旗舰对话模型，创作与理解能力均衡。",
      },
      {
        name: "GPT-4.1",
        model: "gpt-4.1",
        config_type: "chat",
        description: "长上下文增强版，适合长篇写作任务。",
      },
      {
        name: "GPT Image 1",
        model: "gpt-image-1",
        config_type: "image",
        description: "OpenAI 图片生成模型，指令遵循能力强。",
      },
      {
        name: "Sora 2",
        model: "sora-2",
        config_type: "video",
        description: "OpenAI 视频生成模型，物理感与运镜表现优秀。",
        videoEndpoints: { create: "/v1/videos", query: "/v1/videos/{id}", queryMethod: "GET" },
      },
    ]
  },
  {
    key: "deepseek",
    name: "DeepSeek",
    initial: "D",
    color: "#4D6BFE",
    provider_type: "openai",
    api_format: "openai",
    api_url: "https://api.deepseek.com/v1",
    auth_type: 'api_key',
    desc: "深度求索系列模型，中文创作能力强，价格实惠。",
    tags: ['chat'],
    docsUrl: "https://api-docs.deepseek.com",
    models: [
      {
        name: "DeepSeek Chat",
        model: "deepseek-chat",
        config_type: "chat",
        description: "V3 通用对话模型，写作与改写表现出色。",
      },
      {
        name: "DeepSeek Reasoner",
        model: "deepseek-reasoner",
        config_type: "chat",
        description: "R1 推理模型，适合大纲构思与逻辑推演。",
        tags: ["思维链"],
      },
    ]
  },
  {
    key: "siliconflow",
    name: "硅基流动",
    initial: "硅",
    color: "#7C3AED",
    provider_type: "openai",
    api_format: "openai",
    api_url: "https://api.siliconflow.cn/v1",
    auth_type: 'api_key',
    desc: "多模型聚合平台，一站调用 Qwen、DeepSeek、Kolors 等开源模型。",
    tags: ['chat','image'],
    docsUrl: "https://docs.siliconflow.cn",
    models: [
      {
        name: "DeepSeek V3",
        model: "deepseek-ai/DeepSeek-V3",
        config_type: "chat",
        description: "硅基托管版 DeepSeek V3 对话模型。",
      },
      {
        name: "Qwen2.5 72B",
        model: "Qwen/Qwen2.5-72B-Instruct",
        config_type: "chat",
        description: "通义千问开源旗舰，中文写作能力优秀。",
      },
      {
        name: "Kolors",
        model: "Kwai-Kolors/Kolors",
        config_type: "image",
        description: "快手可图开源文生图模型，中文语义理解好。",
      },
    ]
  },
  {
    key: "zhipu",
    name: "智谱AI",
    initial: "智",
    color: "#3859FF",
    provider_type: "openai",
    api_format: "openai",
    api_url: "https://open.bigmodel.cn/api/paas/v4",
    auth_type: 'api_key',
    desc: "智谱 GLM 系列大模型开放平台。",
    tags: ['chat'],
    docsUrl: "https://docs.bigmodel.cn",
    models: [
      {
        name: "GLM-4.5",
        model: "glm-4.5",
        config_type: "chat",
        description: "智谱旗舰对话模型，创作与工具调用能力均衡。",
      },
      {
        name: "GLM-4.5 Air",
        model: "glm-4.5-air",
        config_type: "chat",
        description: "轻量高速版，适合高频调用场景。",
      },
    ]
  },
  {
    key: "moonshot",
    name: "月之暗面",
    initial: "月",
    color: "#16B0A1",
    provider_type: "openai",
    api_format: "openai",
    api_url: "https://api.moonshot.cn/v1",
    auth_type: 'api_key',
    desc: "Kimi 系列模型官方接口，长文本处理能力突出。",
    tags: ['chat'],
    docsUrl: "https://platform.moonshot.cn/docs",
    models: [
      {
        name: "Kimi K2",
        model: "kimi-k2-0711-preview",
        config_type: "chat",
        description: "K2 万亿参数模型，写作风格自然流畅。",
      },
      {
        name: "Moonshot 128K",
        model: "moonshot-v1-128k",
        config_type: "chat",
        description: "128K 长上下文对话模型。",
      },
    ]
  },
  {
    key: "local",
    name: "本地模型服务",
    initial: "L",
    color: "#64748B",
    provider_type: "ollama",
    api_format: "openai",
    api_url: "http://localhost:11434/api/chat",
    auth_type: 'api_key',
    desc: "Ollama / LM Studio 等本地推理服务，无需 API Key。",
    tags: ['chat'],
    noKey: true,
    models: [
      {
        name: "Qwen2.5 14B (Ollama)",
        model: "qwen2.5:14b",
        config_type: "chat",
        description: "本地部署的通义千问 14B，离线可用。",
      },
      {
        name: "LM Studio 本地模型",
        model: "lmstudio-model",
        config_type: "chat",
        description: "LM Studio 加载的本地模型，按实际名称填写模型 ID。",
      },
    ]
  }
]

/** 平台能力标签 → 中文名（广场筛选用） */
export const PRESET_TAG_LABELS: Record<ApiConfigType, string> = {
  chat: '对话',
  image: '图片',
  video: '视频',
  vision: '视觉',
  embed: '嵌入'
}
