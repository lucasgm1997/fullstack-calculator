import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react(), tailwindcss()],
    server: {
      // Listen on all interfaces (0.0.0.0) so phones on the same network can test the app.
      host: true,
      // In development the browser calls the same origin (/api/...) and Vite
      // forwards the request to the Go API, so no CORS setup is needed.
      proxy: {
        '/api': env.VITE_API_PROXY_TARGET ?? 'http://localhost:8080',
      },
    },
  }
})
