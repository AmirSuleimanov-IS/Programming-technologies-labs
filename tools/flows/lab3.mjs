// Скриншоты для отчёта 3: источник данных каталога (кэш браузера, Redis, БД) и журнал кэша в консоли
import fs from 'node:fs'
import { chromium } from 'playwright'

const [base = 'http://127.0.0.1:5173', outDir = '../reports/lab_3/img'] = process.argv.slice(2)
const browser = await chromium.launch({ channel: 'msedge' })
const page = await browser.newPage({ viewport: { width: 1366, height: 560 } })
const consoleLines = []
const started = Date.now()
page.on('console', (msg) => {
  if (msg.text().startsWith('[cache]')) {
    const t = ((Date.now() - started) / 1000).toFixed(1).padStart(5)
    consoleLines.push(`${t} с  ${msg.text()}`)
  }
})
const note = (text) => consoleLines.push(`        ── ${text}`)
const shot = async (name) => {
  await page.waitForTimeout(500)
  await page.screenshot({ path: `${outDir}/${name}.png` })
  console.log('saved', name)
}
const sourceText = () => page.locator('.source').innerText()
const search = page.locator('input[placeholder="Поиск по названию"]')

note('вход пользователя')
await page.goto(base + '/login')
await page.fill('input[type=email]', 'ivan.petrov@mail.ru')
await page.fill('input[type=password]', 'Sport2026')
await page.click('button[type=submit]')
await page.waitForSelector('.card')
note('первая загрузка каталога')
await shot('ui_source_server')
console.log(await sourceText())

note('поиск «турнир», затем очистка поиска — прежний запрос')
await search.fill('турнир')
await page.waitForTimeout(800)
await search.fill('')
await page.waitForTimeout(800)
await shot('ui_source_browser')
console.log(await sourceText())

note('кнопка «Обновить с сервера»')
await page.click('text=Обновить с сервера')
await page.waitForTimeout(800)
await shot('ui_source_redis')
console.log(await sourceText())

note('ожидание 31 с: TTL записи в браузере — 30 с')
await page.waitForTimeout(31000)
await search.fill('турнир')
await page.waitForTimeout(800)
await search.fill('')
await page.waitForTimeout(800)

note('подача заявки — инвалидация кэша каталога')
const card = page.locator('.card', { hasText: 'Лыжная гонка' })
await card.locator('text=Подать заявку').click()
await page.waitForSelector('.alert--info')
await page.waitForTimeout(800)

note('выход — кэш очищается')
await page.click('text=Выйти')
await page.waitForURL('**/login')

fs.writeFileSync('raw/l3_console.txt', consoleLines.join('\n') + '\n')
console.log('console lines', consoleLines.length)
await browser.close()
