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

export default defineConfig({
  plugins: [
    greyvarResPlugin(),
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
