const net = require('net');
const path = require('path');
const { exec } = require('child_process');

const rootDir = __dirname;
const serverEntry = path.join(rootDir, 'server', 'index.js');

const canUsePort = port => new Promise(resolve => {
  const tester = net.createServer();
  tester.once('error', () => resolve(false));
  tester.once('listening', () => tester.close(() => resolve(true)));
  tester.listen(port);
});

const findPort = async (start, end) => {
  for (let port = start; port <= end; port += 1) {
    if (await canUsePort(port)) {
      return port;
    }
  }

  throw new Error(`No available port found between ${start} and ${end}`);
};

const openBrowser = port => {
  if (process.env.NO_OPEN_BROWSER === '1') {
    return;
  }

  setTimeout(() => {
    exec(`start "" "http://localhost:${port}"`);
  }, 1500);
};

const main = async () => {
  process.chdir(rootDir);
  const port = await findPort(3014, 3024);
  process.env.PORT = String(port);
  require(serverEntry);
  openBrowser(port);
};

main().catch(error => {
  console.error('[launcher] Failed to start app:', error);
  process.exit(1);
});
