// Скриншоты Prometheus и Grafana для отчёта 2
import { chromium } from 'playwright'

const [outDir = '../reports/lab_2/img'] = process.argv.slice(2)
const browser = await chromium.launch({ channel: 'msedge' })
const page = await browser.newPage({ viewport: { width: 1600, height: 1000 } })

// Prometheus: цель опроса и запрос к метрике
await page.goto('http://127.0.0.1:9091/targets')
await page.waitForSelector('text=sportreg-backend')
await page.waitForTimeout(1500)
await page.screenshot({ path: `${outDir}/prom_targets.png` })
const query = encodeURIComponent('sum by (route, status) (http_requests_total)')
await page.goto(`http://127.0.0.1:9091/query?g0.expr=${query}&g0.tab=table`)
await page.waitForTimeout(3000)
await page.screenshot({ path: `${outDir}/prom_query.png` })
console.log('prometheus saved')

// Grafana: вход и дашборд из provisioning
await page.goto('http://127.0.0.1:3001/login')
await page.fill('input[name=user]', 'admin')
await page.fill('input[name=password]', 'admin')
await page.click('button[type=submit]')
await page.waitForTimeout(2500)
if (page.url().includes('change-password')) {
  await page.click('text=Skip').catch(() => {})
  await page.waitForTimeout(1500)
}
await page.setViewportSize({ width: 1600, height: 1450 })
await page.goto('http://127.0.0.1:3001/d/sportreg-overview?orgId=1&from=now-5m&to=now&kiosk')
await page.waitForTimeout(8000)
await page.screenshot({ path: `${outDir}/grafana_dashboard.png` })
await page.goto('http://127.0.0.1:3001/connections/datasources')
await page.waitForTimeout(2500)
await page.setViewportSize({ width: 1600, height: 700 })
await page.screenshot({ path: `${outDir}/grafana_datasource.png` })
console.log('grafana saved')
await browser.close()
