import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const publicGateway = process.env.VITE_CLIENT_PUBLIC_GATEWAY || 'http://127.0.0.1:8091'
const protectedGateway = process.env.VITE_CLIENT_PROTECTED_GATEWAY || 'http://127.0.0.1:8093'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3001,
    host: true,
    proxy: {
      '/api/login': { target: publicGateway, changeOrigin: true },
      '/api/refresh': { target: publicGateway, changeOrigin: true },
      '/api/logout': { target: publicGateway, changeOrigin: true },
      '/api/client/register': { target: publicGateway, changeOrigin: true },
      '/api': { target: protectedGateway, changeOrigin: true },
    },
  },
})
