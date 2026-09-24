import { defineConfig } from 'vite'
import basicSsl from '@vitejs/plugin-basic-ssl'

export default defineConfig({
  plugins: [basicSsl()],
  server: {
    https: true,
    host: true,
    port: 5175,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/res': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
