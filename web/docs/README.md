# 星芒 AI 写作 · 前端开发文档

> 本目录为前端开发参考文档。渲染版（可经应用访问）：`web/public/docs/`（随构建进入 dist）。

## 文档索引

| 文档 | 内容 | 渲染版 |
|---|---|---|
| [api-reference.md](./api-reference.md) | **全项目 API 参考**：17 个模块 142 个方法 + backupAPI，含页面→模块映射、全局约定、端点表 | [plugins-api.html](../public/docs/plugins-api.html)（合并版，Part 1） |
| [plugins-api.md](./plugins-api.md) | **插件系统 API**：ChatPlugin 运行时契约、注册表、插件中心 Hub、持久化、pluginAPI（26 方法） | [plugins-api.html](../public/docs/plugins-api.html)（合并版，Part 2） |

## 维护约定

- `src/api/index.ts` 是 API 客户端唯一来源；新增/修改端点后同步更新 `api-reference.md` 对应章节。
- 插件相关变更同步更新 `plugins-api.md` 对应章节。
- 渲染版 [plugins-api.html](../public/docs/plugins-api.html) 为**两份 Markdown 的合并 HTML**（Part 1 全项目 + Part 2 插件），更新任一 .md 后需同步其对应 Part。
- 渲染版为自包含文件（含样式），应用内「查看 API 文档」按钮即指向它。
