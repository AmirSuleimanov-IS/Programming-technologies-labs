// Сравнение времени ответа каталога из Redis (HIT) и из БД (MISS).
// Запросы с уникальным фильтром всегда дают промах, с одинаковым — попадание.
// Время берётся из журнала доступа бэкенда (duration_ms) по X-Request-ID.
// Использование: node cache-bench.mjs [N=200] > raw/l3_bench.txt
import { execSync } from 'node:child_process'

const N = Number(process.argv[2] ?? 200)
const base = 'http://127.0.0.1:8090/api'
const login = await fetch(base + '/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: 'ivan.petrov@mail.ru', password: 'Sport2026' }),
}).then((r) => r.json())
const headers = { Authorization: `Bearer ${login.data.token}` }

async function run(path) {
  const res = await fetch(base + path, { headers })
  await res.text()
  return { id: res.headers.get('X-Request-ID'), cache: res.headers.get('X-Cache') }
}

const ids = { HIT: [], MISS: [] }
await run('/competitions?q=') // прогрев ключа для попаданий
for (let i = 0; i < N; i++) {
  const miss = await run(`/competitions?q=&sport=bench-${Date.now()}-${i}`)
  const hit = await run('/competitions?q=')
  ids[miss.cache]?.push(miss.id)
  ids[hit.cache]?.push(hit.id)
}

await new Promise((r) => setTimeout(r, 1000))
const logs = execSync('docker compose -f ../docker-compose.yml logs backend --no-log-prefix --since 10m', {
  encoding: 'utf8',
  maxBuffer: 256 * 1024 * 1024,
})
const duration = new Map()
for (const line of logs.split('\n')) {
  if (!line.includes('"msg":"http request"')) continue
  const rec = JSON.parse(line)
  duration.set(rec.request_id, rec.duration_ms)
}

const stats = (list) => {
  const v = list.map((id) => duration.get(id)).filter((x) => x !== undefined).sort((a, b) => a - b)
  const q = (p) => v[Math.min(v.length - 1, Math.floor(p * v.length))]
  const avg = v.reduce((s, x) => s + x, 0) / v.length
  return { n: v.length, avg: avg.toFixed(3), p50: q(0.5).toFixed(3), p95: q(0.95).toFixed(3) }
}
console.log(`$ node cache-bench.mjs ${N}`)
console.log('GET /api/competitions, время обработки на сервере (duration_ms из журнала доступа)')
console.log('источник      запросов   среднее, мс   p50, мс   p95, мс')
for (const kind of ['MISS', 'HIT']) {
  const s = stats(ids[kind])
  const name = kind === 'MISS' ? 'MISS (БД)' : 'HIT (Redis)'
  console.log(
    `${name.padEnd(13)} ${String(s.n).padStart(8)} ${s.avg.padStart(13)} ${s.p50.padStart(9)} ${s.p95.padStart(9)}`,
  )
}
