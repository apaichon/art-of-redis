import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    proxy: {
      '/ws': {
        target: 'ws://localhost:9002',
        ws: true
      },
      '/api': {
        target: 'http://localhost:9002'
      }
    }
  }
})
