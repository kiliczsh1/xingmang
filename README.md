# 网文创作辅助系统

> 🎉 经过 60+ 次迭代，专业级网文 AI 辅助创作工具

## 🚀 快速启动

### 方式一：一键启动（推荐 ⭐）
**从 GitHub 下载后，直接双击运行 `start.bat`**

脚本会自动完成：
- ✅ 检查 Node.js 环境
- ✅ 自动检测并安装所有依赖（首次运行需要几分钟）
- ✅ 启动后端服务（端口 3000）
- ✅ 启动前端服务（端口 5173）

> 💡 **说明**：GitHub 仓库不包含 `node_modules` 文件夹（体积过大），首次运行时会自动下载安装所有依赖包。

### 方式二：快速启动
如果已经安装过依赖，双击运行 `quick-start.bat` 即可快速启动。

### 方式三：手动安装依赖并启动
```bash
# 1. 安装所有依赖（首次使用必须执行）
npm run install-all

# 2. 启动后端服务
cd server
npm start

# 3. 启动前端（新开终端）
cd novel
npm run dev
```

### 方式四：分别安装依赖（如遇网络问题）
```bash
# 安装根目录依赖
npm install

# 安装后端依赖
cd server
npm install

# 安装前端依赖
cd novel
npm install
```

## 📱 访问地址

启动成功后，在浏览器访问：
- **前端界面**: http://localhost:5173
- **后端 API**: http://localhost:3000

## 🛠️ 功能特性

- ✅ **书本管理**: 创建、编辑、删除小说项目
- ✅ **AI 续写**: 智能续写功能，支持流式输出
- ✅ **提示词管理**: 自定义 AI 提示词模板
- ✅ **备忘录**: 跨书本共享的创作素材库
- ✅ **API 配置**: 支持多种 AI 服务商（OpenAI、Claude、通义千问等）
- ✅ **对话历史**: 完整的聊天记录管理

## 🔧 技术栈

### 前端
- Vue 3.5.22
- Element Plus 2.5.1
- TypeScript 5.9
- Vite 7.1.7
- Pinia 3.0.3
- Vue Router 4.5.1

### 后端
- Node.js
- Express 4.18.2
- better-sqlite3 12.8.0
- PDF 解析：pdf-parse 2.4.5 + pdfjs-dist 5.4.296

### 数据库
- SQLite

## 📝 使用说明

### 1. 配置 API 密钥
首次使用需要配置 AI 服务商的 API 密钥：
- 进入"API 配置"页面
- 填写你的 API Key（支持 OpenAI、Claude、通义千问等）

### 2. 创建小说
- 点击"新建书本"创建你的小说项目
- 填写书名、作者等基本信息

### 3. 开始创作
- 在对话界面使用 `user: 你的指令` 格式进行 AI 续写
- AI 会以 `-> 内容 <-` 格式回复

### 示例
```
user: 主角走进山洞，发现了一个古老的宝箱

-> 宝箱表面布满了神秘的符文，散发着微弱的蓝光。当你靠近时，符文突然亮起，一股温暖的力量涌入你的体内... <-
```

## ⚠️ 重要提示

### 系统要求
- **Node.js**: v20.19.0 或更高版本（推荐 v22.x LTS）
- **npm**: v10.x 或更高版本
- **操作系统**: Windows 10/11

### 关于依赖安装
本项目经过 60+ 次迭代，依赖包已更新到最新版本。

**为什么 GitHub 没有 node_modules？**
- `node_modules` 文件夹体积过大（通常几百 MB）
- 包含大量二进制文件，不适合版本控制
- 不同操作系统需要不同的二进制包
- **最佳实践**：只上传源代码，用户本地安装依赖

**首次运行会自动安装依赖吗？**
是的！双击 `start.bat` 后会自动：
1. 检查是否已安装依赖
2. 如果未安装，自动执行 `npm install`
3. 安装完成后启动服务

**安装依赖很慢怎么办？**
使用淘宝镜像加速：
```bash
npm config set registry https://registry.npmmirror.com
```

## 🆘 常见问题

### Q1: 启动失败怎么办？
**解决方案**：
1. 检查 Node.js 版本：运行 `node --version`，需要 v20.19.0 或更高
2. 重新安装依赖：
   ```bash
   # 删除 node_modules
   rmdir /s /q node_modules
   rmdir /s /q server\node_modules
   rmdir /s /q novel\node_modules
   
   # 重新安装
   npm run install-all
   ```
3. 查看错误信息，根据具体错误解决

### Q2: 如何停止服务？
直接关闭启动窗口即可，所有服务会自动停止。

### Q3: 端口被占用怎么办？
**方案一**：修改端口
- 后端端口：修改 `server/index.js` 中的端口号（默认 3000）
- 前端端口：修改 `novel/vite.config.ts` 中的端口号（默认 5173）

**方案二**：关闭占用端口的程序

### Q4: 依赖安装失败/卡住？
**解决方案**：
1. **使用淘宝镜像**（推荐）：
   ```bash
   npm config set registry https://registry.npmmirror.com
   npm run install-all
   ```

2. **清除 npm 缓存**：
   ```bash
   npm cache clean --force
   npm run install-all
   ```

3. **分别安装**：
   ```bash
   npm install
   cd server && npm install
   cd ../novel && npm install
   ```

4. **检查网络**：确保网络连接正常

### Q5: better-sqlite3 安装失败？
better-sqlite3 需要编译原生模块，需要：
1. 安装 Python（推荐 3.x）
2. 安装 Visual Studio Build Tools
   - 下载地址：https://visualstudio.microsoft.com/visual-cpp-build-tools/
   - 选择"使用 C++ 的桌面开发"

或者使用预编译包：
```bash
npm install --build-from-source
```

### Q6: 前端页面打不开？
1. 检查后端服务是否启动（访问 http://localhost:3000/api/health）
2. 检查浏览器控制台是否有错误
3. 尝试清除浏览器缓存

### Q7: AI 续写没有反应？
1. 检查是否已配置 API Key
2. 检查 API Key 是否有效
3. 查看后端服务日志，确认是否有错误信息

## 📄 许可证

MIT License

## 🌟 开源贡献

欢迎提交 Issue 和 Pull Request！

### 贡献指南
1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发环境搭建
```bash
# 克隆仓库
git clone https://github.com/yourusername/XingNovel.git

# 进入目录
cd XingNovel

# 安装所有依赖
npm run install-all

# 启动开发服务
npm start
```

## 📞 联系方式

如有问题或建议，欢迎通过以下方式联系：
- 提交 Issue
- 发送邮件至：[你的邮箱]

---

**Made with ❤️ for novel writers**
