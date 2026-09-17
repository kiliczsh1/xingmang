const { chromium } = require('playwright-core');
const fs = require('fs');

(async () => {
  const context = await chromium.launchPersistentContext('C:/Users/25134/AppData/Local/Temp/pw-repro-theme4', {
    executablePath: 'C:/Users/25134/AppData/Local/ms-playwright/chromium-1181/chrome-win/chrome.exe',
    headless: true,
    viewport: { width: 1440, height: 900 },
    args: ['--no-sandbox', '--disable-gpu']
  });
  const page = context.pages()[0] || await context.newPage();
  const errors = [];
  page.on('console', m => { if (m.type() === 'error') errors.push(m.text().slice(0, 300)); });
  page.on('pageerror', e => errors.push('PAGEERROR: ' + String(e).slice(0, 300)));

  await page.goto('http://localhost:5173/books', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(1500);

  const pluginInfo = await page.evaluate(() => {
    const raw = localStorage.getItem('plugin-local');
    let records = [];
    try { records = raw ? JSON.parse(raw) : []; } catch {}
    return {
      hasPlugin: records.some(r => r.manifest && r.manifest.name === 'theme-quick-bar'),
      records: records.map(r => r.manifest && r.manifest.name),
      theme: localStorage.getItem('main-theme'),
      hasDataTheme: document.documentElement.hasAttribute('data-theme')
    };
  });
  console.log('PLUGIN_INFO:', JSON.stringify(pluginInfo));

  if (!pluginInfo.hasPlugin) {
    const entryCode = fs.readFileSync('d:/xingguang/xingmang-desktop/theme-quick-bar-extracted/index.js', 'utf8');
    const manifest = JSON.parse(fs.readFileSync('d:/xingguang/xingmang-desktop/theme-quick-bar-extracted/manifest.json', 'utf8'));
    await page.evaluate(({ entryCode, manifest }) => {
      const raw = localStorage.getItem('plugin-local');
      let records = [];
      try { records = raw ? JSON.parse(raw) : []; } catch {}
      records = records.filter(r => r.manifest && r.manifest.name !== manifest.name);
      records.push({ manifest, entryCode, installedAt: new Date().toISOString() });
      localStorage.setItem('plugin-local', JSON.stringify(records));
    }, { entryCode, manifest });
    console.log('Plugin installed, reloading...');
    await page.reload({ waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
  }

  await page.goto('http://localhost:5173/write/15', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(3000);

  const editorState = await page.evaluate(() => {
    const vd = document.querySelector('.vditor');
    if (!vd) return { hasEditor: false };
    const cs = (sel) => {
      const el = vd.querySelector(sel);
      if (!el) return null;
      const s = getComputedStyle(el);
      return { bg: s.backgroundColor, color: s.color, border: s.borderColor };
    };
    return {
      hasEditor: true,
      isDarkClass: vd.classList.contains('vditor--dark'),
      htmlTheme: document.documentElement.hasAttribute('data-theme'),
      vditor: cs('.vditor'),
      content: cs('.vditor-content'),
      wysiwyg: cs('.vditor-wysiwyg'),
      reset: cs('.vditor-reset'),
      resetInline: (() => { const el = vd.querySelector('.vditor-reset'); return el ? el.style.backgroundColor : null; })(),
      wysiwygInline: (() => { const el = vd.querySelector('.vditor-wysiwyg'); return el ? el.style.backgroundColor : null; })(),
      hasPluginBar: !!document.getElementById('theme-quick-bar')
    };
  });
  console.log('EDITOR_STATE:', JSON.stringify(editorState, null, 1));

  const barState = await page.evaluate(() => {
    const bar = document.getElementById('theme-quick-bar');
    if (!bar) return { hasBar: false };
    const darkBtn = bar.querySelector('[data-theme="dark"]');
    if (darkBtn) darkBtn.click();
    return { hasBar: true, clicked: true };
  });
  console.log('BAR_CLICK:', JSON.stringify(barState));
  await page.waitForTimeout(1500);

  const darkState = await page.evaluate(() => {
    const vd = document.querySelector('.vditor');
    if (!vd) return { hasEditor: false };
    const cs = (sel) => {
      const el = vd.querySelector(sel);
      if (!el) return null;
      const s = getComputedStyle(el);
      return { bg: s.backgroundColor, color: s.color, border: s.borderColor };
    };
    const reset = vd.querySelector('.vditor-reset');
    const blocks = reset ? Array.from(reset.children).slice(0, 5).map(el => ({
      tag: el.tagName,
      bg: getComputedStyle(el).backgroundColor,
      inlineBg: el.style.backgroundColor
    })) : [];
    return {
      htmlTheme: document.documentElement.hasAttribute('data-theme'),
      isDarkClass: vd.classList.contains('vditor--dark'),
      vditor: cs('.vditor'),
      content: cs('.vditor-content'),
      wysiwyg: cs('.vditor-wysiwyg'),
      reset: cs('.vditor-reset'),
      resetInline: reset ? reset.style.backgroundColor : null,
      wysiwygInline: (() => { const el = vd.querySelector('.vditor-wysiwyg'); return el ? el.style.backgroundColor : null; })(),
      blocks
    };
  });
  console.log('DARK_STATE:', JSON.stringify(darkState, null, 1));

  await page.screenshot({ path: '_repro_dark.png', fullPage: false });
  console.log('CONSOLE_ERRORS:', errors.length ? errors.join('\n') : 'none');
  await context.close();
})();
