import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/admin': 'http://localhost:8080',
      '/auth': 'http://localhost:8080',
      '/v1': 'http://localhost:8080',
    }
  }
})
