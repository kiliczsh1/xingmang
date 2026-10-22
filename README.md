# 星芒 (xingmang)

> ⚠️ **本分支 `v-0.7.7.3` 是测试版，尚未达到稳定版。**
> 功能、API、数据库结构、界面都可能随时变更，不保证向后兼容。
> 请勿用于正式交付或承载不可替代的数据。稳定版请见 `main` 分支。

## 这是什么形态的应用

**星芒现在是 Web 端应用**，界面在浏览器里运行。

仓库里的 Go 程序负责提供 HTTP 服务、读写本地数据、代理模型请求；前端是一个标准的 Vite 单页应用，
构建产物由 Go 用 `//go:embed` 内嵌进二进制。打包出来的 `xingmang.exe` 只是「让网页端能在一台机器上
凭一个可执行文件跑起来」的载体，**不是桌面客户端**。

启动后浏览器访问 `http://127.0.0.1:19825` 使用。桌面客户端形态属于早期版本，已不再作为主要形态维护。

> 目录名 `xingmang-desktop` 和 `go.mod` 的模块路径 `github.com/xingmang/xingmang-desktop`
> 是历史遗留命名，代码结构仍沿用，未做重命名。

## 功能

本地运行的小说创作工具：书籍、章节、卷、角色卡、知识图谱、提示词库、工作流，
以及一个可以读写这些数据的 AI Agent。所有数据保存在本地，不上传云端。

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

- 仓库不含 `web/dist`，首次构建必须先 `npm install`。
- 本分支与 `main` 没有共同祖先（独立根提交），无法直接发起 Pull Request。

图标资源 `cmd/xingmang/rsrc_windows_amd64.syso` 与 `rsrc_windows_arm64.syso` 已随源码提供，
构建出的 exe 带程序图标；它们是 COFF 资源对象，只是二进制形式，不参与编译逻辑。

## 许可证

本项目采用 **Apache License 2.0**，全文见 [`LICENSE`](LICENSE)。

你可以自由使用、修改、分发、商用，无需支付费用，也不需要公开自己的修改。分发时需要做到：

- 附上一份 `LICENSE` 副本
- 保留原有的版权、专利、商标与免责声明
- 修改过的文件，在文件里注明「已修改」

许可证明确授予专利许可；如果使用方就本项目发起专利诉讼，其获得的专利许可自动终止。
不授予商标权——「星芒」的名称与图标不随代码一起授权。免责条款明确：软件按「原样」提供，
作者不对使用产生的任何损失负责。

> 本项目没有 `NOTICE` 文件，因此分发时不产生额外的 NOTICE 附带义务。

许可证原文：http://www.apache.org/licenses/LICENSE-2.0

Copyright 2026 kiliczsh1
