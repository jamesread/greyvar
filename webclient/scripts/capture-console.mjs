// Headless browser harness: loads the webclient and dumps browser console
// output to stdout so it can be inspected without a human at the keyboard.
//
// Usage: node scripts/capture-console.mjs [url] [seconds]

import puppeteer from 'puppeteer-core'

const url = process.argv[2] ?? 'https://localhost:5173/'
const seconds = Number(process.argv[3] ?? 12)

const executablePath = ['/usr/bin/chromium-browser', '/usr/bin/google-chrome']
  .find((p) => true) // both exist on this machine; prefer chromium

const browser = await puppeteer.launch({
  executablePath: '/usr/bin/chromium-browser',
  headless: 'new',
  args: [
    '--ignore-certificate-errors',
    '--no-sandbox',
    '--disable-gpu',
    '--use-gl=swiftshader',
    '--enable-unsafe-swiftshader',
    '--window-size=1280,800'
  ]
})

const page = await browser.newPage()
await page.setViewport({ width: 1280, height: 800 })

page.on('console', (msg) => {
  const loc = msg.location()
  const where = loc.url ? `${loc.url.split('/').pop()}:${loc.lineNumber}` : ''
  console.log(`[console.${msg.type()}] ${msg.text()} ${where ? `(${where})` : ''}`)
})

page.on('pageerror', (err) => {
  console.log(`[pageerror] ${err.message}`)
})

page.on('requestfailed', (req) => {
  console.log(`[requestfailed] ${req.url()} :: ${req.failure()?.errorText}`)
})

page.on('response', (res) => {
  if (res.status() >= 400) {
    console.log(`[http ${res.status()}] ${res.url()}`)
  }
})

console.log(`--- loading ${url}, capturing for ${seconds}s ---`)
await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 30000 })

await new Promise((r) => setTimeout(r, seconds * 1000))

// Final scene state snapshot from within the page
try {
  const snapshot = await page.evaluate(() => {
    const mgr = window.phaser?.scene
    if (!mgr) return 'phaser not booted'
    return mgr.scenes.map((s) => {
      const st = s.sys.settings
      return `${st.key}: status=${st.status} active=${st.active} visible=${st.visible}`
    }).join(' | ')
  })
  console.log(`--- final scene state: ${snapshot} ---`)

  const entities = await page.evaluate(() => {
    const grid = window.phaser?.scene?.getScene('grid')
    return {
      entdefs: window.gameState ? [...window.gameState.entdefs.keys()] : [],
      entities: grid ? Object.values(grid.entities).map((e) => ({
        id: e.entityId,
        def: e.definition,
        hasImg: e.img != null,
        x: e.x,
        y: e.y,
      })) : [],
    }
  })
  console.log('--- entity snapshot:', JSON.stringify(entities), '---')
} catch (e) {
  console.log(`--- could not snapshot scenes: ${e.message} ---`)
}

await page.screenshot({ path: '/tmp/greyvar-splash.png' })
console.log('--- screenshot saved to /tmp/greyvar-splash.png ---')

await browser.close()
