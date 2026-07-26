const fs = require('fs')
const path = require('path')

const rootDir = __dirname
const outputFileName = 'experience-shares.single.js'
const excludeNames = new Set([
  '.git',
  'node_modules',
  outputFileName,
  path.basename(__filename)
])

const normalizePath = value => value.split(path.sep).join('/')

const readPluginMeta = () => {
  const manifestPath = path.join(rootDir, '.codex-plugin', 'plugin.json')

  try {
    return JSON.parse(fs.readFileSync(manifestPath, 'utf8'))
  } catch {
    return {
      name: path.basename(rootDir),
      version: '0.0.0',
      description: 'Single-file plugin bundle'
    }
  }
}

const collectFiles = dir => {
  const entries = fs.readdirSync(dir, { withFileTypes: true })
  const files = []

  for (const entry of entries) {
    if (excludeNames.has(entry.name)) {
      continue
    }

    const absolutePath = path.join(dir, entry.name)
    const relativePath = normalizePath(path.relative(rootDir, absolutePath))

    if (entry.isDirectory()) {
      files.push(...collectFiles(absolutePath))
      continue
    }

    files.push({
      path: relativePath,
      base64: fs.readFileSync(absolutePath).toString('base64')
    })
  }

  return files.sort((a, b) => a.path.localeCompare(b.path))
}

const renderBundle = ({ meta, files }) => {
  const safeMeta = JSON.stringify(
    {
      name: meta.name,
      version: meta.version,
      description: meta.description,
      generatedAt: new Date().toISOString()
    },
    null,
    2
  )

  const safeFiles = JSON.stringify(
    Object.fromEntries(files.map(file => [file.path, file.base64])),
    null,
    2
  )

  return `'use strict';

const fs = require('fs');
const path = require('path');

const plugin = ${safeMeta};
const files = ${safeFiles};

function ensureDir(targetPath) {
  fs.mkdirSync(path.dirname(targetPath), { recursive: true });
}

function extract(targetDir, options = {}) {
  const overwrite = options.overwrite !== false;
  const writtenFiles = [];

  for (const [relativePath, base64] of Object.entries(files)) {
    const targetPath = path.join(targetDir, relativePath);

    if (!overwrite && fs.existsSync(targetPath)) {
      continue;
    }

    ensureDir(targetPath);
    fs.writeFileSync(targetPath, Buffer.from(base64, 'base64'));
    writtenFiles.push(targetPath);
  }

  return {
    plugin,
    targetDir,
    fileCount: writtenFiles.length,
    writtenFiles
  };
}

function listFiles() {
  return Object.keys(files);
}

module.exports = {
  plugin,
  files,
  listFiles,
  extract
};

if (require.main === module) {
  const targetDir = path.resolve(process.argv[2] || path.join(process.cwd(), plugin.name));
  const result = extract(targetDir);
  console.log(\`Extracted \${result.fileCount} files to \${targetDir}\`);
}
`
}

const meta = readPluginMeta()
const files = collectFiles(rootDir)
const outputPath = path.join(rootDir, outputFileName)
const outputContent = renderBundle({ meta, files })

fs.writeFileSync(outputPath, outputContent, 'utf8')

console.log(`Created ${outputFileName} with ${files.length} files.`)
