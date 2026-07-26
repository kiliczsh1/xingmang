const fs = require('fs')
const path = require('path')

const rootDir = __dirname
const outputFileName = 'experience-shares.template.js'
const excludeNames = new Set([
  '.git',
  'node_modules',
  outputFileName,
  path.basename(__filename)
])

const normalizePath = value => value.split(path.sep).join('/')

const readJsonIfExists = filePath => {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'))
  } catch {
    return null
  }
}

const readTextIfExists = filePath => {
  try {
    return fs.readFileSync(filePath, 'utf8')
  } catch {
    return ''
  }
}

const readPluginMeta = () => {
  const manifestPath = path.join(rootDir, '.codex-plugin', 'plugin.json')
  const manifest = readJsonIfExists(manifestPath)

  if (manifest) {
    return manifest
  }

  return {
    name: path.basename(rootDir),
    version: '0.0.0',
    description: 'XingNovel plugin bundle'
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

const toPluginId = value => {
  const safeName = String(value || 'experience-shares')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')

  return `custom.${safeName || 'experience-shares'}`
}

const renderBundle = ({ meta, files, generatedAt, readme, integrationDoc }) => {
  const interfaceMeta = meta.interface || {}
  const displayName = interfaceMeta.displayName || meta.name || 'Experience Shares'
  const description = interfaceMeta.shortDescription || meta.description || 'Experience share plugin bundle'
  const tags = Array.isArray(meta.keywords) && meta.keywords.length > 0 ? meta.keywords : ['experience-shares']

  const payload = {
    id: toPluginId(meta.name),
    name: displayName,
    version: meta.version || '0.0.0',
    description,
    tags,
    generatedAt,
    interface: interfaceMeta,
    bundle: {
      plugin: {
        name: meta.name || 'experience-shares',
        version: meta.version || '0.0.0',
        description: meta.description || description
      },
      fileCount: files.length,
      files: Object.fromEntries(files.map(file => [file.path, file.base64])),
      docs: {
        readme,
        integration: integrationDoc
      },
      integration: {
        frontendRoutes: [
          '/experience-shares',
          '/experience-shares/:id',
          '/experience-shares/:id/edit'
        ],
        backendMount: '/api/experience-shares',
        migrationFile: 'backend/database/experience-shares-migration.sql',
        routeFile: 'backend/routes/experienceShares.js'
      }
    }
  }

  const safePayload = JSON.stringify(payload, null, 2)

  return `(function () {
  const payload = ${safePayload};

  const register = window?.XingNovelApp?.registerPlugin;
  if (typeof register !== 'function') {
    console.warn('[experience-shares] XingNovelApp.registerPlugin is not available.');
    return;
  }

  const notify = (ctx, level, message) => {
    const fn = ctx?.notify?.[level] || ctx?.notify?.success || ctx?.notify?.info;
    if (typeof fn === 'function') {
      fn(message);
      return;
    }
    console.log(message);
  };

  register({
    id: payload.id,
    name: payload.name,
    version: payload.version,
    description: payload.description,
    tags: payload.tags,
    actions: [
      {
        id: 'open-experience-shares',
        label: '打开经验分享',
        variant: 'primary',
        async execute(ctx) {
          await ctx.navigate('/experience-shares');
        }
      },
      {
        id: 'show-bundle-summary',
        label: '显示插件信息',
        async execute(ctx) {
          notify(ctx, 'success', '经验分享插件已注册，内含 ' + payload.bundle.fileCount + ' 个打包文件。');
        }
      }
    ],
    routes: payload.bundle.integration.frontendRoutes,
    bundle: payload.bundle
  });
})();`
}

const meta = readPluginMeta()
const files = collectFiles(rootDir)
const generatedAt = new Date().toISOString()
const readme = readTextIfExists(path.join(rootDir, 'README.md'))
const integrationDoc = readTextIfExists(path.join(rootDir, 'docs', 'integration.md'))
const outputPath = path.join(rootDir, outputFileName)
const outputContent = renderBundle({
  meta,
  files,
  generatedAt,
  readme,
  integrationDoc
})

fs.writeFileSync(outputPath, outputContent, 'utf8')

console.log(`Created ${outputFileName} with ${files.length} files.`)
