# XingNovel 网文创作辅助系统

> 基于 [星盘写作](https://github.com/nichuanfang/xingpan-write) v0.10 版本基座魔改，现已迭代至 **v0.680**

专业级网文 AI 辅助创作工具，集成 AI 续写、世界观管理、知识图谱、工作流等核心创作能力。

## 快速开始

### 一键启动（推荐）

双击 `start.bat`，脚本自动完成：

- 检查 Node.js 环境（要求 v20.19.0+）
- 自动检测并安装前后端依赖
- 启动后端服务（端口 3000）和前端服务（端口 5173）
- 单窗口合并日志输出，关闭窗口即停止所有服务

> 首次运行需要几分钟下载依赖，之后可直接使用 `quick-start.bat` 快速启动。

### 手动启动

```bash
# 安装依赖
npm run install-all

# 启动后端（终端 1）
cd server && npm start

# 启动前端（终端 2）
cd novel && npm run dev
```

### 访问地址

| 服务    | 地址                  |
| ------- | --------------------- |
| 前端界面 | http://localhost:5173 |
| 后端 API | http://localhost:3000 |

## 功能概览

### 创作核心

| 功能       | 说明                                           |
| ---------- | ---------------------------------------------- |
| 书本管理   | 创建、编辑、删除小说项目，支持卷和章节管理     |
| AI 续写    | 智能续写，支持流式输出，兼容多家 AI 服务商     |
| 提示词管理 | 自定义提示词模板，支持卡包分类、预览和测试     |
| 创意工坊   | 固定模板生成（抽卡），可配置提示词与占位符字段 |
| 工作流     | 链状多步骤 AI 创作流程，支持工作台独立运行     |
| 备忘录     | 跨书本共享的创作素材库                         |

### 世界观体系

| 功能         | 说明                                       |
| ------------ | ------------------------------------------ |
| WorldBook    | 管理角色、地点、物品、势力等世界观设定      |
| 词条库       | 统一管理小说中的专有名词和概念             |
| 知识图谱     | 基于 ECharts 可视化展示元素之间的关系网络   |
| 角色库       | 专门的角色管理与维护                       |
| 拆书库       | 对已有书籍进行分析拆解                      |

### 辅助工具

| 功能       | 说明                           |
| ---------- | ------------------------------ |
| 经验分享   | 编写和分享个人创作经验         |
| 对话历史   | 完整的聊天记录管理和搜索回顾   |
| 消息中心   | 系统通知与消息聚合             |
| API 配置   | 支持 OpenAI / Claude / 通义千问等 |
| 暗色主题   | 亮色 / 暗色主题切换            |
| 个人中心   | 用户信息管理                   |

## 技术栈

```
前端:  Vue 3.5 + TypeScript 5.9 + Vite 7.1 + Element Plus 2.5 + Pinia 3.0 + Vue Router 4.5
编辑器: TipTap 3.23（富文本）+ Markdown-it（Markdown 渲染）
可视化: ECharts 6.0（知识图谱）
后端:  Node.js + Express 4.18 + better-sqlite3 12.8
解析:  pdf-parse 2.4 + pdfjs-dist 5.4（PDF 解析）
数据库: SQLite
```

## 项目结构

```
XingNovel/
├── novel/                    # 前端 (Vue 3 + Vite)
│   ├── src/
│   │   ├── views/            # 页面组件
│   │   ├── components/       # 通用组件（知识图谱、编辑器、WorldBook 等）
│   │   ├── router/           # 路由配置
│   │   ├── stores/           # Pinia 状态管理
│   │   └── api/              # API 请求封装
│   └── vite.config.ts
├── server/                   # 后端 (Express + SQLite)
│   ├── routes/               # API 路由（17 个业务模块）
│   ├── database/             # 数据库层
│   ├── services/             # 业务逻辑
│   └── index.js              # 入口文件
├── plugins/                  # 插件系统
│   └── experience-shares/    # 经验分享插件
├── start.bat                 # 一键启动脚本
├── quick-start.bat           # 快速启动（跳过依赖检查）
└── dev-launcher.js           # 单窗口开发模式启动器
```

## 使用指南

### 1. 配置 AI 服务

进入「API 配置」页面，添加 AI 服务商的 API Key。支持 OpenAI、Claude、通义千问等多种服务商。

### 2. 创建小说项目

在「书本管理」中新建小说，填写书名、作者等信息。

### 3. 开始创作

进入创作页面，使用对话方式与 AI 交互进行续写。

### 4. 管理设定

在 WorldBook 中维护角色、地点、物品等世界观元素，AI 创作时自动引用相关设定保持一致性。

### 5. 使用知识图谱

在「wiki 图谱」中查看元素间的关系网络，帮助梳理复杂设定。

## 常见问题

**Q: 启动失败？**
确保 Node.js 版本 >= v20.19.0，或尝试重新安装依赖：
好的，请提供您需要翻译的文本。
运行install-all
运行install-all
```

好的，请提供您需要翻译的文本。
使用淘宝镜像加速：
好的，请提供您需要翻译的文本。
npm config set registry https://registry.npmmirror.com
```

**Q: better-sqlite3 安装失败？**
需要安装 Python 3.x 和 Visual Studio Build Tools（选择"使用 C++ 的桌面开发"）。

**Q: 端口被占用？**
修改 `server/index.js` 中的 `PORT`（默认 3000）或 `novel/vite.config.ts` 中的 `port`（默认 5173）。

## 开发参与

欢迎提交 Issue 和 Pull Request。

好的，请提供您需要翻译的文本。
运行 npm start 命令
cd XingNovel
运行install-all
运行 npm start 命令
```

与条款

与条款
