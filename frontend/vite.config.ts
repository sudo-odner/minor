import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: true, // Нужно для работы внутри Docker
    port: 5173,
    watch: {
      usePolling: true, // Помогает Hot Reload работать стабильно на Windows/Linux
    },
    // Настройка HMR (Hot Module Replacement) через прокси Traefik
    hmr: {
      clientPort: 80, 
    },
  },
})