import { defineConfig, devices } from '@playwright/test'

/**
 * End-to-end tests: the real React app (Vite) talking to the real Go API.
 * Playwright starts both servers on dedicated ports so a run never hits a
 * stale dev server or the Docker stack.
 */
const API_PORT = 18080
const WEB_PORT = 4173
const isCI = !!process.env.CI

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: isCI,
  retries: isCI ? 1 : 0,
  reporter: isCI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: `http://localhost:${WEB_PORT}`,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile', use: { ...devices['Pixel 7'] } },
  ],
  webServer: [
    {
      command: 'go run ./cmd/api',
      cwd: '../backend',
      env: { PORT: String(API_PORT), GIN_MODE: 'release', RATE_LIMIT_REQUESTS: '10000' },
      url: `http://localhost:${API_PORT}/api/health`,
      reuseExistingServer: false,
      timeout: 120_000,
    },
    {
      command: `npx vite --port ${WEB_PORT} --strictPort`,
      env: { VITE_API_PROXY_TARGET: `http://localhost:${API_PORT}` },
      url: `http://localhost:${WEB_PORT}`,
      reuseExistingServer: false,
      timeout: 60_000,
    },
  ],
})
