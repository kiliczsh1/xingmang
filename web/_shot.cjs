const { chromium } = require('playwright-core');
(async () => {
  const browser = await chromium.launch({
    executablePath: 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
    headless: true
  });
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
  const errors = [];
  page.on('console', m => { if (m.type() === 'error') errors.push(m.text().slice(0, 200)); });
  page.on('pageerror', e => errors.push('PAGEERROR: ' + String(e).slice(0, 200)));
  await page.goto('http://localhost:5173/experience-shares', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: '_shot.png', fullPage: false });

  const diag = await page.evaluate(() => {
    const r = (sel) => {
      const el = document.querySelector(sel);
      if (!el) return null;
      const b = el.getBoundingClientRect();
      const cs = getComputedStyle(el);
      return { top: Math.round(b.top), bottom: Math.round(b.bottom), h: Math.round(b.height), w: Math.round(b.width), display: cs.display, overflowY: cs.overflowY, position: cs.position, left: Math.round(b.left), right: Math.round(b.right) };
    };
    return {
      viewport: { w: innerWidth, h: innerHeight },
      page: r('.experience-page'),
      wsBody: r('.ws-body'),
      wsSide: r('.ws-side'),
      wsMain: r('.ws-main'),
      wsToolbar: r('.ws-toolbar'),
      wsContent: r('.ws-content'),
      wsStats: r('.ws-stats'),
      wsGrid: r('.ws-grid'),
      cards: document.querySelectorAll('.ws-card').length,
      contentFrame: r('.content-frame'),
      docScrollH: document.documentElement.scrollHeight,
      bodyScrollH: document.body.scrollHeight
    };
  });
  console.log(JSON.stringify(diag, null, 1));
  console.log('CONSOLE_ERRORS:', errors.length ? errors.join('\n') : 'none');
  await browser.close();
})();
