// Нагрузка для дашборда Grafana: смесь обычных запросов, ошибок и входов.
// Использование: node traffic.mjs [секунд=180] [baseUrl]
const [seconds = '180', base = 'http://127.0.0.1:8090/api'] = process.argv.slice(2)
const deadline = Date.now() + Number(seconds) * 1000

async function call(method, path, { token, body } = {}) {
  const headers = {}
  if (body) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(base + path, { method, headers, body: body && JSON.stringify(body) })
  const json = await res.json().catch(() => null)
  return { status: res.status, json }
}

const login = async (email, password) =>
  (await call('POST', '/auth/login', { body: { email, password } })).json?.data?.token

const admin = await login('admin@sportreg.local', 'Admin12345')
const users = []
for (let i = 1; i <= 5; i++) {
  const email = `runner${i}@mail.ru`
  await call('POST', '/auth/register', { body: { email, full_name: `Бегун ${i}`, password: 'Run2026pass' } })
  users.push(await login(email, 'Run2026pass'))
}

const pick = (list) => list[Math.floor(Math.random() * list.length)]
const scenarios = [
  [30, () => call('GET', '/competitions', { token: pick(users) })],
  [12, () => call('GET', `/competitions/${1 + Math.floor(Math.random() * 5)}`, { token: pick(users) })],
  [8, () => call('GET', '/competitions?sport=Футбол', { token: pick(users) })],
  [6, () => call('GET', '/applications/my', { token: pick(users) })],
  [5, () => login(`runner${1 + Math.floor(Math.random() * 5)}@mail.ru`, 'Run2026pass')],
  [3, () => login('runner1@mail.ru', 'wrong-password1')],
  [3, () => call('GET', '/competitions')], // без токена — 401
  [2, () => call('POST', '/competitions', { token: pick(users), body: { title: 'x' } })], // 403
  [2, () => call('GET', '/competitions/999', { token: admin })], // 404
  [2, () => call('GET', '/applications', { token: admin })],
  [2, () => call('POST', '/applications', { token: pick(users), body: { competition_id: 5 } })],
  [1, () => call('GET', '/users', { token: pick(users) })], // 403
]
const total = scenarios.reduce((s, [w]) => s + w, 0)

let count = 0
while (Date.now() < deadline) {
  let r = Math.random() * total
  const [, run] = scenarios.find(([w]) => (r -= w) < 0) ?? scenarios[0]
  await run().catch(() => {})
  count++
  await new Promise((resolve) => setTimeout(resolve, 40 + Math.random() * 120))
}
console.log('requests sent:', count)
