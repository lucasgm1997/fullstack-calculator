import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react(), tailwindcss()],
    server: {
      // In development the browser calls the same origin (/api/...) and Vite
      // forwards the request to the Go API, so no CORS setup is needed.
      proxy: {
        '/api': env.VITE_API_PROXY_TARGET ?? 'http://localhost:8080',
      },
    },
  }
})
