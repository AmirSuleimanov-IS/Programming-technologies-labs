// Usage: node shot.mjs <url> <out.png> [width] [height] [fullPage:1|0] [waitMs]
import { chromium } from 'playwright';
const [url, out, w = '1440', h = '900', full = '0', wait = '800'] = process.argv.slice(2);
const browser = await chromium.launch({ channel: 'msedge' });
const page = await browser.newPage({ viewport: { width: +w, height: +h } });
await page.goto(url, { waitUntil: 'load', timeout: 60000 });
try { await page.waitForLoadState('networkidle', { timeout: 10000 }); } catch {}
await page.waitForTimeout(+wait);
await page.screenshot({ path: out, fullPage: full === '1' });
await browser.close();
console.log('saved', out);
