const http = require('http');
const net = require('net');
const path = require('path');
const { spawn } = require('child_process');

const rootDir = __dirname;
const isWindows = process.platform === 'win32';
const children = [];
let shuttingDown = false;

const services = [
  {
    label: 'server',
    colorCode: '36',
    cwd: path.join(rootDir, 'server'),
    port: 3000,
    args: ['start'],
    healthCheckPath: '/api/health'
  },
  {
    label: 'web',
    colorCode: '35',
    cwd: path.join(rootDir, 'novel'),
    port: 5173,
    args: ['run', 'dev'],
    healthCheckPath: '/'
  }
];

const formatLabel = (label, code) => `\u001b[${code}m[${label}]\u001b[0m`;

const pipeOutput = (stream, label, colorCode) => {
  stream.on('data', chunk => {
    const text = chunk.toString();
    const lines = text.split(/\r?\n/);

    for (const line of lines) {
      if (!line) {
        continue;
      }

      process.stdout.write(`${formatLabel(label, colorCode)} ${line}\n`);
    }
  });
};

const canUsePort = port => new Promise(resolve => {
  const tester = net.createServer();

  tester.once('error', () => resolve(false));
  tester.once('listening', () => tester.close(() => resolve(true)));
  tester.listen(port);
});

const requestOk = (port, pathName) => new Promise(resolve => {
  const request = http.get(
    {
      hostname: '127.0.0.1',
      port,
      path: pathName,
      timeout: 2000
    },
    response => {
      response.resume();
      resolve(response.statusCode && response.statusCode < 500);
    }
  );

  request.on('timeout', () => {
    request.destroy();
    resolve(false);
  });

  request.on('error', () => resolve(false));
});

const getListeningPid = port => new Promise(resolve => {
  const tester = net.createServer();

  tester.once('error', error => {
    if (error && error.code === 'EADDRINUSE') {
      const finder = spawn('cmd.exe', ['/d', '/s', '/c', `netstat -ano | findstr :${port} | findstr LISTENING`], {
        cwd: rootDir,
        env: process.env,
        stdio: ['ignore', 'pipe', 'ignore'],
        shell: false,
        windowsHide: true
      });

      let output = '';
      finder.stdout.on('data', chunk => {
        output += chunk.toString();
      });

      finder.on('exit', () => {
        const match = output.match(/LISTENING\s+(\d+)/);
        resolve(match ? Number(match[1]) : null);
      });

      finder.on('error', () => resolve(null));
      return;
    }

    resolve(null);
  });

  tester.once('listening', () => {
    tester.close(() => resolve(null));
  });

  tester.listen(port);
});

const killPid = pid => new Promise(resolve => {
  if (!pid) {
    resolve(false);
    return;
  }

  const killer = spawn('cmd.exe', ['/d', '/s', '/c', `taskkill /F /PID ${pid}`], {
    cwd: rootDir,
    env: process.env,
    stdio: ['ignore', 'ignore', 'ignore'],
    shell: false,
    windowsHide: true
  });

  killer.on('exit', code => resolve(code === 0));
  killer.on('error', () => resolve(false));
});

const stopAll = exitCode => {
  if (shuttingDown) {
    return;
  }

  shuttingDown = true;

  for (const child of children) {
    if (!child.killed) {
      child.kill('SIGINT');
    }
  }

  setTimeout(() => process.exit(exitCode), 400);
};

const startProcess = ({ label, colorCode, cwd, args }) => {
  const command = isWindows ? 'cmd.exe' : 'npm';
  const commandArgs = isWindows ? ['/d', '/s', '/c', `npm ${args.join(' ')}`] : args;

  const child = spawn(command, commandArgs, {
    cwd,
    env: process.env,
    stdio: ['inherit', 'pipe', 'pipe'],
    shell: false,
    windowsHide: false
  });

  children.push(child);
  pipeOutput(child.stdout, label, colorCode);
  pipeOutput(child.stderr, label, colorCode);

  child.on('error', error => {
    console.error(`${formatLabel(label, colorCode)} failed to start`, error);
    stopAll(1);
  });

  child.on('exit', (code, signal) => {
    if (shuttingDown) {
      return;
    }

    const reason = signal ? `signal ${signal}` : `code ${code}`;
    console.log(`${formatLabel(label, colorCode)} exited: ${reason}`);
    stopAll(code ?? 0);
  });
};

const main = async () => {
  console.log('Single-window dev mode started.');
  console.log('Closing this window will stop services started by this launcher.');

  let startedCount = 0;

  for (const service of services) {
    const available = await canUsePort(service.port);

    if (!available) {
      const healthy = await requestOk(service.port, service.healthCheckPath);

      if (healthy) {
        console.log(
          `${formatLabel(service.label, service.colorCode)} port ${service.port} is already in use and healthy, skipping start`
        );
        continue;
      }

      const existingPid = await getListeningPid(service.port);
      console.log(
        `${formatLabel(service.label, service.colorCode)} port ${service.port} is occupied by an unhealthy process${existingPid ? ` (PID ${existingPid})` : ''}, restarting it`
      );

      if (existingPid) {
        await killPid(existingPid);
      }
    }

    startProcess(service);
    startedCount += 1;
  }

  if (startedCount === 0) {
    console.log('Nothing was started because both dev ports are already in use.');
    process.exit(0);
  }
};

process.on('SIGINT', () => stopAll(0));
process.on('SIGTERM', () => stopAll(0));

main().catch(error => {
  console.error('Launcher failed:', error);
  process.exit(1);
});
