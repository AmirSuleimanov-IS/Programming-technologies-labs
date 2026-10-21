// Рендер кода или вывода терминала в PNG (как окно редактора/терминала).
// Использование:
//   node codeshot.mjs <file|-> <out.png> --title "server.go" [--lang go|bash|json|plaintext]
//        [--from N --to M] [--width 1100] [--terminal]
// Если файл "-", текст читается из stdin.
import { chromium } from 'playwright';
import fs from 'node:fs';
import hljs from 'highlight.js';

const args = process.argv.slice(2);
const src = args[0];
const out = args[1];
const opt = (name, def) => {
  const i = args.indexOf('--' + name);
  return i >= 0 ? args[i + 1] : def;
};
const flag = (name) => args.includes('--' + name);

let text = src === '-' ? fs.readFileSync(0, 'utf8') : fs.readFileSync(src, 'utf8');
text = text.replace(/\r\n/g, '\n').replace(/\t/g, '    ');
let lines = text.split('\n');
const from = +opt('from', 1);
const to = +opt('to', lines.length);
lines = lines.slice(from - 1, to);
while (lines.length && lines[lines.length - 1] === '') lines.pop();
const code = lines.join('\n');

const terminal = flag('terminal');
const lang = opt('lang', terminal ? 'plaintext' : 'go');
const title = opt('title', '');
const width = +opt('width', 1100);
const wrap = flag('wrap');
const lineNumbers = !terminal && !flag('nonum') && !wrap;

const highlighted = lang === 'plaintext'
  ? code.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  : hljs.highlight(code, { language: lang }).value;
const htmlLines = highlighted.split('\n');
const gutter = lineNumbers
  ? `<div class="gutter">${htmlLines.map((_, i) => `<div>${from + i}</div>`).join('')}</div>`
  : '';

const css = fs.readFileSync(new URL('./node_modules/highlight.js/styles/' + (terminal ? 'github-dark.css' : 'vs2015.css'), import.meta.url), 'utf8');
const html = `<!doctype html><html><head><meta charset="utf-8"><style>
${css}
body { margin: 0; background: transparent; }
.win { display: inline-block; ${wrap ? 'width' : 'min-width'}: ${width}px; border-radius: 10px; overflow: hidden; background: ${terminal ? '#0c0c0c' : '#1e1e1e'};
       box-shadow: 0 6px 24px rgba(0,0,0,.25); font-family: Consolas, 'Cascadia Mono', monospace; }
.bar { height: 34px; display: flex; align-items: center; gap: 8px; padding: 0 14px; background: ${terminal ? '#2b2b2b' : '#323233'}; color: #ccc; font: 13px 'Segoe UI', sans-serif; }
.dot { width: 12px; height: 12px; border-radius: 50%; }
.title { margin-left: 10px; }
.body { display: flex; padding: 14px 0; font-size: 14px; line-height: 1.5; }
.gutter { color: #6e7681; text-align: right; padding: 0 14px 0 16px; user-select: none; }
pre { margin: 0; padding: 0 18px 0 ${lineNumbers ? 0 : 18}px; white-space: ${wrap ? 'pre-wrap' : 'pre'}; color: ${terminal ? '#cccccc' : '#d4d4d4'}; flex: 1; }
pre code.hljs { padding: 0; background: transparent; }
</style></head><body><div class="win" id="win">
<div class="bar"><span class="dot" style="background:#ff5f57"></span><span class="dot" style="background:#febc2e"></span><span class="dot" style="background:#28c840"></span><span class="title">${title}</span></div>
<div class="body">${gutter}<pre><code class="hljs">${htmlLines.join('\n')}</code></pre></div>
</div></body></html>`;

const browser = await chromium.launch({ channel: 'msedge' });
const page = await browser.newPage({ viewport: { width: 2600, height: 600 }, deviceScaleFactor: 1.5 });
await page.setContent(html);
await page.locator('#win').screenshot({ path: out });
await browser.close();
console.log('saved', out);
