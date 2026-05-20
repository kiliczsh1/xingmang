const net = require('net');
const path = require('path');
const fs = require('fs');
const { exec } = require('child_process');

const rootDir = __dirname;
const serverEntry = path.join(rootDir, 'server', 'index.js');
const logFile = path.join(rootDir, 'launcher.log');

const log = (msg) => {
  const timestamp = new Date().toISOString();
  const line = `[${timestamp}] ${msg}\n`;
  try { fs.appendFileSync(logFile, line); } catch (e) {}
  console.log(line.trim());
};

const canUsePort = (port) => new Promise((resolve) => {
  const tester = net.createServer();
  tester.once('error', (err) => {
    if (err.code === 'EADDRINUSE') {
      resolve(false);
    } else {
      resolve(false);
    }
  });
  tester.once('listening', () => {
    tester.close(() => resolve(true));
  });
  tester.listen(port, '127.0.0.1');
});

const findPort = async (start, end) => {
  for (let port = start; port <= end; port += 1) {
    const available = await canUsePort(port);
    if (available) {
      return port;
    }
    log(`Port ${port} is in use, trying next...`);
  }
  throw new Error(`No available port found between ${start} and ${end}`);
};

const openBrowser = (port) => {
  if (process.env.NO_OPEN_BROWSER === '1') {
    return;
  }
  setTimeout(() => {
    exec(`start "" "http://localhost:${port}"`);
  }, 2000);
};

const main = async () => {
  log('Starting XingNovel launcher...');
  log(`Root directory: ${rootDir}`);
  log(`Server entry: ${serverEntry}`);

  if (!fs.existsSync(serverEntry)) {
    throw new Error(`Server entry not found: ${serverEntry}`);
  }

  process.chdir(rootDir);
  log(`Working directory changed to: ${process.cwd()}`);

  const port = await findPort(3014, 3024);
  log(`Found available port: ${port}`);

  process.env.PORT = String(port);
  log(`Loading server entry...`);
  require(serverEntry);
  log('Server entry loaded successfully');
  openBrowser(port);
};

main().catch((error) => {
  log(`[ERROR] Failed to start app: ${error.message}`);
  log(`Stack: ${error.stack}`);
  console.error('[launcher] Failed to start app:', error);
  process.exit(1);
});
