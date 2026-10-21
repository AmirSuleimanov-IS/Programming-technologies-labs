// Скриншоты фронтенда для отчёта 1: каталог, ошибки валидации, создание, изменение (PATCH)
import { chromium } from 'playwright'

const [base = 'http://localhost:5173', outDir = '../reports/lab_1/img'] = process.argv.slice(2)
const browser = await chromium.launch({ channel: 'msedge' })
const page = await browser.newPage({ viewport: { width: 1366, height: 820 } })
const shot = async (name, full = false) => {
  await page.waitForTimeout(400)
  await page.screenshot({ path: `${outDir}/${name}.png`, fullPage: full })
  console.log('saved', name)
}

await page.goto(base + '/competitions')
await page.waitForSelector('table')
await shot('ui_list')

// Поиск по названию
await page.fill('input[placeholder="Поиск по названию"]', 'турнир')
await page.waitForTimeout(600)
await shot('ui_search')
await page.fill('input[placeholder="Поиск по названию"]', '')

// Ошибки валидации от бэкенда (422) под полями формы
await page.click('text=Добавить соревнование')
await page.fill('form input >> nth=0', 'Ку')
await page.fill('form input[type=number]', '0')
await page.click('button[type=submit]')
await page.waitForSelector('.field__error')
await shot('ui_validation')

// Создание
await page.fill('form input >> nth=0', 'Кубок ЗГУ по баскетболу')
await page.fill('form input[list=sport-suggestions]', 'Баскетбол')
await page.fill('form input[type=date]', '2026-12-12')
await page.fill('form input[type=number]', '16')
await page.fill('form input >> nth=4', 'Норильск, ФОК «Арктика»')
await page.fill('form textarea', 'Команды факультетов, 5 игроков + 2 запасных.')
await shot('ui_form_filled')
await page.click('button[type=submit]')
await page.waitForSelector('text=Кубок ЗГУ по баскетболу')
await shot('ui_created')

// Изменение: форма отправит PATCH только с изменённым полем
const row = page.locator('tr', { hasText: 'Кубок ЗГУ по баскетболу' })
await row.locator('text=Изменить').click()
await page.fill('form input[type=number]', '24')
const patch = page.waitForRequest((r) => r.method() === 'PATCH')
await page.click('button[type=submit]')
const req = await patch
console.log('PATCH', req.url(), req.postData())
await page.waitForSelector('form', { state: 'detached' })
await shot('ui_updated')

await browser.close()
