import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api/customers': { target: 'http://127.0.0.1:8081', changeOrigin: true },
      '/api/vehicles': { target: 'http://127.0.0.1:8082', changeOrigin: true },
      '/api/deals': { target: 'http://127.0.0.1:8083', changeOrigin: true },
      '/api/parts': { target: 'http://127.0.0.1:8084', changeOrigin: true },
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: true },
    },
  },
})
