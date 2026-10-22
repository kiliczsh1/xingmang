# 星芒桌面版 (xingmang-desktop)

> ⚠️ **本分支 `v-0.7.7.3` 是测试版，尚未达到稳定版。**
> 功能、API、数据库结构、界面都可能随时变更，不保证向后兼容。
> 请勿用于正式交付或承载不可替代的数据。稳定版请见 `main` 分支。

本地运行的小说创作工具：Go 后端 + Vue 前端，内置可读写书籍、章节、角色卡与知识图谱的 AI Agent。所有数据保存在本地，不上传云端。

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.26（`net/http`，单二进制，前端通过 `//go:embed` 内嵌） |
| 前端 | Vue 3 + TypeScript + Vite + Element Plus |
| 数据库 | SQLite（本地文件） |

## 目录结构

```
cmd/xingmang/        后端主程序（含 //go:embed 的前端产物挂载点）
cmd/migrate/         数据库迁移工具
cmd/decodebackup/    备份解码工具
internal/            业务代码：handlers / models / server / agent / bookstore / fanqie …
tools/release/       打包工具（编译 + 前端构建 + 归档）
tools/logogen/       示例资源生成
tools/agentprobe/    Agent 探测工具
web/src/             前端源码
build.ps1 / build.bat    发布打包入口
start-dev.ps1 / .bat     开发模式启动（后端 + Vite + 番茄桥）
```

## 环境要求

- Go **1.26.10**（`go.mod` 的 `go` 指令会触发 toolchain 自动下载）
- Node.js **^20.19.0 || >=22.12.0**
- 构建脚本为 PowerShell（Windows）

## 构建

```powershell
cd web
npm install

cd ..
.\build.ps1 -Version 0.7.7.3
```

产物：`dist\windows-amd64\`（`xingmang.exe` + `README.txt`），加 `-Zip` 时额外产出 `dist\V-0.7.7.3.zip`。

⚠️ **不要直接 `go build ./cmd/xingmang`**：该包用 `//go:embed all:web/dist` 内嵌前端产物，
而仓库中不含 `cmd/xingmang/web/dist`（属于构建产物）。`tools/release` 会先构建前端再自动同步到该目录。
需要手动构建前端时：

```powershell
cd web && npm run build
# 然后把 web\dist 的内容同步到 cmd\xingmang\web\dist
```

## 开发模式

```bat
start-dev.bat
```

同时拉起三个进程：Go 后端 `:19825`、Vite 开发服务器 `:5173`、番茄小说桥 `:19900`。
关闭后端会一并结束另外两个伴生进程。

## 数据与隐私

本仓库**只包含源码**。以下内容属于个人数据或运行产物，不在版本库内，也请勿提交：

- 小说正文、章节、角色卡、大纲（`books/`）
- 提示词库、历史记录、会话（`xingmang.db` 及其 `-wal` / `-shm`）
- Agent 记忆与分析记录（`_memory/`）
- 运行数据（`data/`）、工作流输出（`workflow_outputs/`）
- 日志（`logs/`、`*.log`）、构建产物（`dist/`、`node_modules/`、`*.exe`）
- 密钥与配置（`.env`、`*.pem`、`*.key`）

## 已知限制

- 仓库不含 `cmd/xingmang/rsrc_windows_*.syso`（程序图标资源），直接构建出的 exe 没有自定义图标，
  不影响运行；`tools/release` 检测到缺失时会提示。
- 仓库不含 `web/dist`，首次构建必须先 `npm install`。
- 本分支与 `main` 没有共同祖先（独立根提交），无法直接发起 Pull Request。
