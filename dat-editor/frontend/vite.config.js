import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const backendPort = process.env.PORT || '8080'
const backendTarget = `http://localhost:${backendPort}`

export default defineConfig({
  resolve: {
    alias: {
      '../../gen/deditor/v1/deditor_pb.js': resolve(__dirname, 'gen/deditor/v1/deditor_pb.ts'),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
  },
  plugins: [vue()],
  server: {
    port: 5175,
    proxy: {
      '/deditor.v1.DatEditorService': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/res': {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
})
