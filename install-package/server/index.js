const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');
const path = require('path');
const fs = require('fs');
const { exec } = require('child_process');

const isPortablePackage = () => {
  if (process.pkg) return true;
  return process.execPath.includes('XingNovel-node');
};

const getRootDir = () => {
  if (isPortablePackage()) {
    return path.dirname(process.execPath);
  }
  return path.join(__dirname, '..');
};

const getDbPath = () => {
  if (isPortablePackage()) {
    return path.join(getRootDir(), 'server', 'database', 'novel.db');
  }

  return path.join(__dirname, 'database', 'novel.db');
};

process.env.DB_PATH = getDbPath();

require('./database/db');
const bookRoutes = require('./routes/books');
const chapterRoutes = require('./routes/chapters');
const promptRoutes = require('./routes/prompts');
const memoRoutes = require('./routes/memos');
const providerRoutes = require('./routes/providers');
const apiConfigRoutes = require('./routes/apiConfig');
const aiRoutes = require('./routes/ai');
const conversationRoutes = require('./routes/conversations');
const generatorRoutes = require('./routes/generators');
const characterRoutes = require('./routes/characters');
const statsRoutes = require('./routes/stats');
const volumeRoutes = require('./routes/volumes');
const experienceShareRoutes = require('./routes/experienceShares');
const workflowRoutes = require('./routes/workflows');
const knowledgeGraphRoutes = require('./routes/knowledgeGraph');

const app = express();
const PORT = Number(process.env.PORT) || (isPortablePackage() ? 3014 : 3000);
const distDir = isPortablePackage()
  ? path.join(getRootDir(), 'novel', 'dist')
  : path.join(__dirname, '..', 'novel', 'dist');
const indexHtmlPath = path.join(distDir, 'index.html');
const uploadsDir = isPortablePackage()
  ? path.join(getRootDir(), 'server', 'uploads')
  : path.join(__dirname, 'uploads');

fs.mkdirSync(uploadsDir, { recursive: true });

app.use(cors());
app.use(bodyParser.json({ limit: '50mb' }));
app.use(bodyParser.urlencoded({ extended: true, limit: '50mb' }));
app.use('/api/uploads', express.static(uploadsDir));

app.use('/api/books', bookRoutes);
app.use('/api/chapters', chapterRoutes);
app.use('/api/prompts', promptRoutes);
app.use('/api/memos', memoRoutes);
app.use('/api/providers', providerRoutes);
app.use('/api/config', apiConfigRoutes);
app.use('/api/ai', aiRoutes);
app.use('/api/conversations', conversationRoutes);
app.use('/api/generators', generatorRoutes);
app.use('/api/characters', characterRoutes);
app.use('/api/stats', statsRoutes);
app.use('/api/volumes', volumeRoutes);
app.use('/api/experience-shares', experienceShareRoutes);
app.use('/api/workflows', workflowRoutes);
app.use('/api/knowledge-graph', knowledgeGraphRoutes);

app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', message: '服务运行正常' });
});

if (fs.existsSync(indexHtmlPath)) {
  app.use(express.static(distDir));
  app.get(/^\/(?!api\/).*/, (req, res) => {
    res.sendFile(indexHtmlPath);
  });
}

const startServer = (port) => {
  const server = app.listen(port, () => {
    console.log(`服务已启动：http://localhost:${port}`);

    if (process.pkg) {
      exec(`start "" "http://localhost:${port}"`);
    }
  });

  server.on('error', (error) => {
    if (error && error.code === 'EADDRINUSE' && process.pkg && !process.env.PORT && port < 3024) {
      console.warn(`端口 ${port} 已被占用，尝试使用端口 ${port + 1}...`);
      startServer(port + 1);
      return;
    }

    if (error && error.code === 'EADDRINUSE') {
      console.error(`端口 ${port} 已被占用，请关闭占用进程后重试。`);
      process.exit(1);
    }

    console.error('服务启动失败:', error);
    process.exit(1);
  });
};

startServer(PORT);
