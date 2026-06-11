import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true,
    proxy: {
      // auth-service proxies /api/* to grpc-gateway (8090)
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: true },
    },
  },
})
