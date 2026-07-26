/**
 * 插件系统 API 路由 V3
 *
 * 端点：
 * - GET  /api/plugins/list              → 列出 data/user-plugins/ 下所有插件
 * - GET  /api/plugins/frontend-scripts  → 获取插件前端 JS 代码
 * - POST /api/plugins/install           → 安装插件（接收 zip base64，解压到用户目录）
 * - DELETE /api/plugins/:name           → 卸载插件（删除整个目录）
 */

const express = require('express')
const router = express.Router()
const fs = require('fs')
const path = require('path')
const PluginManager = require('../plugins/PluginManager')

// 辅助：获取用户插件目录
function getPluginsRoot() {
  return PluginManager.getPluginsRoot()
}

// ─── GET /api/plugins/list ───
router.get('/list', (req, res) => {
  try {
    const pluginsRoot = getPluginsRoot()
    const plugins = []

    if (fs.existsSync(pluginsRoot)) {
      const entries = fs.readdirSync(pluginsRoot, { withFileTypes: true })
      for (const entry of entries) {
        if (!entry.isDirectory()) continue
        const dir = path.join(pluginsRoot, entry.name)
        const manifestPath = path.join(dir, 'plugin.json')

        if (fs.existsSync(manifestPath)) {
          const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
          const hasFrontend = fs.existsSync(path.join(dir, 'frontend.js'))
          const hasBackend = fs.existsSync(path.join(dir, 'index.js'))

          plugins.push({
            dirName: entry.name,
            name: manifest.name,
            version: manifest.version,
            description: manifest.description || '',
            author: manifest.author?.name || manifest.author || '未知',
            permissions: manifest.permissions || [],
            interface: manifest.interface || null,
            hasFrontend,
            hasBackend
          })
        }
      }
    }

    res.json({ success: true, data: plugins })
  } catch (error) {
    res.status(500).json({ success: false, message: error.message })
  }
})

// ─── GET /api/plugins/frontend-scripts ───
router.get('/frontend-scripts', (req, res) => {
  try {
    const pluginsRoot = getPluginsRoot()
    const plugins = []

    if (fs.existsSync(pluginsRoot)) {
      const entries = fs.readdirSync(pluginsRoot, { withFileTypes: true })
      for (const entry of entries) {
        if (!entry.isDirectory()) continue
        const dir = path.join(pluginsRoot, entry.name)
        const manifestPath = path.join(dir, 'plugin.json')
        const frontendPath = path.join(dir, 'frontend.js')

        if (fs.existsSync(manifestPath) && fs.existsSync(frontendPath)) {
          const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
          const frontendCode = fs.readFileSync(frontendPath, 'utf-8')
          plugins.push({ manifest, frontendCode })
        }
      }
    }

    res.json({ success: true, plugins })
  } catch (error) {
    res.status(500).json({ success: false, message: error.message })
  }
})

// ─── POST /api/plugins/install ───
// 接收 { fileName, data_base64 } → 解压到 data/user-plugins/
router.post('/install', (req, res) => {
  try {
    const { fileName, data_base64 } = req.body || {}
    if (!fileName || !data_base64) {
      return res.status(400).json({ success: false, message: '缺少 fileName 或 data_base64' })
    }

    const pluginsRoot = getPluginsRoot()
    const zipPath = path.join(pluginsRoot, fileName)

    // 写入 zip 文件
    const buffer = Buffer.from(data_base64, 'base64')
    fs.writeFileSync(zipPath, buffer)

    // 确定解压目录名（去掉 .zip 后缀）
    const dirName = fileName.replace(/\.zip$/i, '')
    const targetDir = path.join(pluginsRoot, dirName)

    // 如果已存在同名插件，先删后装
    if (fs.existsSync(targetDir)) {
      fs.rmSync(targetDir, { recursive: true, force: true })
    }

    // 使用纯 JS 解压（零依赖）
    const { unpackZip } = require('../utils/zip')
    unpackZip(zipPath, targetDir)
    let extracted = true

    // 清理 zip 文件
    const nestedDirs = fs.readdirSync(targetDir, { withFileTypes: true })
      .filter(e => e.isDirectory())
    const manifestPath = fs.existsSync(path.join(targetDir, 'plugin.json'))
      ? path.join(targetDir, 'plugin.json')
      : (nestedDirs.length === 1 && fs.existsSync(path.join(targetDir, nestedDirs[0].name, 'plugin.json')))
        ? path.join(targetDir, nestedDirs[0].name, 'plugin.json')
        : null

    // 如果是嵌套了一层目录，提出来
    if (manifestPath && nestedDirs.length === 1 && !fs.existsSync(path.join(targetDir, 'plugin.json'))) {
      const nestedDir = path.join(targetDir, nestedDirs[0].name)
      const files = fs.readdirSync(nestedDir)
      for (const f of files) {
        fs.renameSync(path.join(nestedDir, f), path.join(targetDir, f))
      }
      fs.rmdirSync(nestedDir)
    }

    // 重新扫描插件
    PluginManager.rescan()

    res.json({ success: true, message: `插件已安装到 ${dirName}` })
  } catch (error) {
    res.status(500).json({ success: false, message: error.message })
  }
})

// ─── GET /api/plugins/:name/export ───
// 将插件文件夹打包为 zip 并返回 base64（供下载分享）
router.get('/:name/export', (req, res) => {
  try {
    const pluginsRoot = getPluginsRoot()
    const pluginDir = path.join(pluginsRoot, req.params.name)

    if (!pluginDir.startsWith(pluginsRoot)) {
      return res.status(403).json({ success: false, message: '非法路径' })
    }
    if (!fs.existsSync(pluginDir)) {
      return res.status(404).json({ success: false, message: '插件不存在' })
    }

    const manifestPath = path.join(pluginDir, 'plugin.json')
    const manifest = fs.existsSync(manifestPath)
      ? JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
      : { name: req.params.name }

    // 使用纯 JS zip 工具（零依赖）
    const { packDir } = require('../utils/zip')
    const tmpZip = path.join(pluginsRoot, `_export_${req.params.name}.zip`)
    packDir(pluginDir, tmpZip)

    const zipBuffer = fs.readFileSync(tmpZip)
    const base64 = zipBuffer.toString('base64')
    try { fs.unlinkSync(tmpZip) } catch (_) {}

    const displayName = manifest.interface?.displayName || manifest.name
    res.json({
      success: true,
      data: {
        fileName: `${manifest.name}-v${manifest.version}.zip`,
        dataBase64: base64,
        pluginName: displayName
      }
    })
  } catch (error) {
    res.status(500).json({ success: false, message: error.message })
  }
})

// ─── DELETE /api/plugins/:name ───
// 删除 data/user-plugins/:name/ 整个目录
router.delete('/:name', (req, res) => {
  try {
    const pluginsRoot = getPluginsRoot()
    const pluginDir = path.join(pluginsRoot, req.params.name)

    // 安全检查：确保要删除的目录在 user-plugins 下面
    if (!pluginDir.startsWith(pluginsRoot)) {
      return res.status(403).json({ success: false, message: '非法路径' })
    }

    if (!fs.existsSync(pluginDir)) {
      return res.status(404).json({ success: false, message: '插件不存在' })
    }

    // 卸载插件钩子
    const manifestPath = path.join(pluginDir, 'plugin.json')
    if (fs.existsSync(manifestPath)) {
      const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
      PluginManager.unloadPlugin(manifest.name)
    }

    // 删除目录
    fs.rmSync(pluginDir, { recursive: true, force: true })

    res.json({ success: true, message: '插件已卸载' })
  } catch (error) {
    res.status(500).json({ success: false, message: error.message })
  }
})

module.exports = router
