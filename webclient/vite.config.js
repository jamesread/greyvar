import { createReadStream, existsSync, statSync } from 'node:fs'
import { extname, join, normalize } from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import basicSsl from '@vitejs/plugin-basic-ssl'

const rootDir = fileURLToPath(new URL('.', import.meta.url))
const resDir = normalize(process.env.GREYVAR_RES ?? join(rootDir, '../res'))
const serverTarget = process.env.GREYVAR_SERVER ?? 'https://localhost:8443'

const contentTypes = {
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.json': 'application/json',
  '.ttf': 'font/ttf',
  '.ogg': 'audio/ogg',
  '.wav': 'audio/wav',
  '.mp3': 'audio/mpeg',
}

function greyvarResPlugin() {
  const root = `${resDir}/`

  return {
    name: 'greyvar-res',
    configureServer(server) {
      server.middlewares.use('/res', (req, res, next) => {
        const rel = decodeURIComponent((req.url ?? '/').split('?')[0])
        const file = normalize(join(root, rel.replace(/^\//, '')))

        if (!file.startsWith(root)) {
          res.statusCode = 403
          res.end('Forbidden')
          return
        }

        if (!existsSync(file) || statSync(file).isDirectory()) {
          next()
          return
        }

        res.setHeader('Content-Type', contentTypes[extname(file)] ?? 'application/octet-stream')
        createReadStream(file).pipe(res)
      })
    },
  }
}

// The dev server only speaks HTTPS. A browser given the bare LAN address
// (192.168.x.y:5173) tries plain HTTP and got an empty response (Chrome:
// ERR_EMPTY_RESPONSE / ERR_TOO_MANY_RETRIES). Answer those with a redirect to
// https:// on the same address instead.
function httpToHttpsRedirectPlugin() {
  return {
    name: 'greyvar-http-to-https',
    configureServer(server) {
      server.httpServer?.on('tlsClientError', (err, tlsSocket) => {
        if (err?.code !== 'ERR_SSL_HTTP_REQUEST') {
          return
        }
        const raw = tlsSocket?._parent ?? tlsSocket
        const host = String(raw?.localAddress ?? 'localhost').replace(/^::ffff:/, '')
        const hostPart = host.includes(':') ? `[${host}]` : host
        const location = `https://${hostPart}:${raw?.localPort ?? 5173}/`
        try {
          raw.end(`HTTP/1.1 301 Moved Permanently\r\nLocation: ${location}\r\nContent-Length: 0\r\nConnection: close\r\n\r\n`)
        } catch {
          // the socket may already be gone
        }
      })
    },
  }
}

export default defineConfig({
  plugins: [
    greyvarResPlugin(),
    httpToHttpsRedirectPlugin(),
    basicSsl(),
    Components({
      dirs: ['resources/vue/'],
      extensions: ['vue'],
      deep: true,
      dts: false,
    }),
    vue(),
  ],
  server: {
    https: true,
    host: true,
    fs: {
      allow: [join(rootDir, '..')],
    },
    proxy: {
      '/api': {
        target: serverTarget,
        changeOrigin: true,
        secure: false,
        ws: true,
      },
    },
  },
})
