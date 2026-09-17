import { chromium } from 'playwright-core'

const CHROME = 'C:/Users/25134/AppData/Local/ms-playwright/chromium-1181/chrome-win/chrome.exe'
const URL = 'http://localhost:5173/experience-shares/3'

const browser = await chromium.launch({
  executablePath: CHROME,
  headless: true,
  args: ['--disable-gpu']
})
const page = await browser.newPage({ viewport: { width: 1280, height: 900 } })

const errors = []
page.on('console', m => { if (m.type() === 'error') errors.push(m.text()) })
page.on('pageerror', e => errors.push(String(e)))

await page.goto(URL, { waitUntil: 'domcontentloaded', timeout: 30000 })

// 1. 正文编辑器是否渲染
await page.waitForSelector('.article-body .tiptap', { timeout: 10000 })
console.log('[1] 正文 TipTap 编辑器已渲染')
await page.screenshot({ path: 'shot-1-detail.png', fullPage: false })

// 2. 点击正文，工具栏应浮现
await page.click('.article-body .tiptap p')
await page.waitForTimeout(500)
const toolbarVisible = await page.isVisible('.article-body .editor-toolbar')
console.log('[2] 聚焦后工具栏可见:', toolbarVisible)
await page.screenshot({ path: 'shot-2-focus.png', fullPage: false })

// 3. 目录是否生成
const tocCount = await page.locator('.toc-link').count()
console.log('[3] 目录条目数:', tocCount)

// 4. 打开元信息弹窗
await page.click('button:has-text("编辑信息")')
await page.waitForSelector('.el-dialog', { timeout: 5000 })
await page.waitForTimeout(400)
const dialogTitle = await page.locator('.el-dialog__title').textContent()
console.log('[4] 弹窗标题:', dialogTitle)
await page.screenshot({ path: 'shot-3-dialog.png', fullPage: false })

// 5. 关闭弹窗
await page.click('.el-dialog button:has-text("取消")')
await page.waitForTimeout(300)

console.log('--- console errors ---')
console.log(errors.length ? errors.join('\n') : '(none)')

await browser.close()
