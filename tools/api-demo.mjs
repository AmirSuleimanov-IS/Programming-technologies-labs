// Прогон запросов к API с выводом в стиле терминала (для скриншотов отчёта).
// Использование: node api-demo.mjs <steps.json> [baseUrl]
// Шаг: { "title", "method", "path", "body" | "raw", "token", "headers": ["X-Cache"],
//        "saveAs": "имя" (запомнить data.id), "saveToken": "имя" (запомнить data.token),
//        "short": N (показать N строк тела), "pick": ["data.free_places", "meta"] (показать только эти поля),
//        "quiet": true (не печатать шаг) }
// В path и token можно подставлять запомненное: "/competitions/{c1}", "token": "admin";
// в теле значение "{c1}" заменяется числом: {"competition_id": "{c1}"}.
// Тело уходит в UTF-8 (curl в Windows портит кириллицу в аргументах).
import fs from 'node:fs'

const [stepsFile, base = 'http://127.0.0.1:8090/api'] = process.argv.slice(2)
const steps = JSON.parse(fs.readFileSync(stepsFile, 'utf8'))
const saved = {}
const fill = (s) => s.replace(/\{(\w+)\}/g, (m, k) => (k in saved ? saved[k] : m))

for (const step of steps) {
  const path = fill(step.path)
  const headers = {}
  let body
  if (step.raw !== undefined) body = step.raw
  else if (step.body !== undefined)
    body = JSON.stringify(step.body).replace(/"\{(\w+)\}"/g, (m, k) => (k in saved ? String(saved[k]) : m))
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (step.token) headers.Authorization = `Bearer ${saved[step.token] ?? step.token}`

  const started = performance.now()
  const res = await fetch(base + path, { method: step.method, headers, body })
  const ms = Math.round(performance.now() - started)
  const text = await res.text()
  let json = null
  try {
    json = JSON.parse(text)
  } catch {}
  if (step.saveAs && json?.data?.id) saved[step.saveAs] = json.data.id
  if (step.saveToken && json?.data?.token) saved[step.saveToken] = json.data.token
  if (step.quiet) continue

  if (step.title) console.log(`# ${step.title}`)
  const auth = step.token ? `  (Authorization: Bearer <${step.token}>)` : ''
  console.log(`$ ${step.method} ${path}${body !== undefined ? ' ' + body : ''}${auth}`)
  const shown = (step.headers ?? []).map((h) => `${h}: ${res.headers.get(h)}`)
  console.log(`HTTP ${res.status} ${res.statusText}${step.time ? `  (${ms} мс)` : ''}`)
  if (shown.length) console.log(shown.join('\n'))
  let out = json ? JSON.stringify(json, null, 2) : text
  if (step.pick && json) {
    const get = (p) => p.split('.').reduce((v, k) => (v == null ? v : v[k]), json)
    out = step.pick.map((p) => `${p} = ${JSON.stringify(get(p))}`).join('\n')
  }
  if (step.short && out) {
    const lines = out.split('\n')
    if (lines.length > step.short) out = lines.slice(0, step.short).join('\n') + '\n  …'
  }
  if (out) console.log(out)
  console.log('')
}
