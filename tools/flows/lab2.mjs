// Скриншоты фронтенда для отчёта 2: вход, регистрация, страницы USER и ADMIN, 403
import { chromium } from 'playwright'

const [base = 'http://127.0.0.1:5173', outDir = '../reports/lab_2/img'] = process.argv.slice(2)
const browser = await chromium.launch({ channel: 'msedge' })
const page = await browser.newPage({ viewport: { width: 1366, height: 860 } })
const shot = async (name, full = false) => {
  await page.waitForTimeout(500)
  await page.screenshot({ path: `${outDir}/${name}.png`, fullPage: full })
  console.log('saved', name)
}

// Гость открывает защищённую страницу — переадресация на вход
await page.goto(base + '/competitions')
await page.waitForURL('**/login')
await shot('ui_login')

// Регистрация: ошибки валидации, затем успешная регистрация и автоматический вход
await page.click('text=Зарегистрируйтесь')
await page.fill('input[type=email]', 'maria@')
await page.fill('form input >> nth=1', 'М')
await page.fill('input[type=password]', 'qwerty')
await page.click('button[type=submit]')
await page.waitForSelector('.field__error')
await shot('ui_register_errors')
await page.fill('input[type=email]', 'maria.kuznetsova@mail.ru')
await page.fill('form input >> nth=1', 'Мария Кузнецова')
await page.fill('input[type=password]', 'Volley2026')
await page.click('button[type=submit]')
await page.waitForURL('**/competitions')
await page.waitForSelector('.card')
await shot('ui_user_page')

// Подача заявки на волейбол
const card = page.locator('.card', { hasText: 'Первенство ЗГУ по волейболу' })
await card.locator('text=Подать заявку').click()
await page.waitForSelector('.alert--info')
await shot('ui_user_applied', true)

// USER открывает страницу администратора — 403
await page.goto(base + '/admin')
await page.waitForURL('**/forbidden')
await shot('ui_forbidden')

// Выход и вход администратора
await page.click('text=Выйти')
await page.waitForURL('**/login')
await page.fill('input[type=email]', 'admin@sportreg.local')
await page.fill('input[type=password]', 'Admin12345')
await page.click('button[type=submit]')
await page.waitForURL('**/competitions')
await page.click('text=Администрирование')
await page.waitForSelector('.tabs')
await page.waitForSelector('table')
await shot('ui_admin_applications')
const row = page.locator('tr', { hasText: 'Мария Кузнецова' })
await row.locator('text=Одобрить').click()
await page.waitForTimeout(400)
await page.selectOption('select', '')
await page.waitForSelector('.badge--approved')
await shot('ui_admin_approved')

await page.click('.tab >> text=Соревнования')
await page.waitForSelector('table')
await shot('ui_admin_competitions')
await page.click('.tab >> text=Пользователи')
await page.waitForSelector('table')
await shot('ui_admin_users')

await browser.close()
