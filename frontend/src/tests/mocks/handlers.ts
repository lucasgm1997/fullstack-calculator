import { http, HttpResponse } from 'msw'
import type { CalculateRequest } from '../../features/calculator/types/calculator.types'

/**
 * Minimal stand-in for the Go API so UI tests are deterministic and offline.
 * Only covers what the tests exercise; real arithmetic is tested in Go.
 */
export const handlers = [
  http.get('/api/health', () => HttpResponse.json({ status: 'ok' })),

  http.post('/api/calculate', async ({ request }) => {
    const { operation, a, b = 0 } = (await request.json()) as CalculateRequest
    switch (operation) {
      case 'add':
        return HttpResponse.json({ result: a + b })
      case 'subtract':
        return HttpResponse.json({ result: a - b })
      case 'multiply':
        return HttpResponse.json({ result: a * b })
      case 'divide':
        return b === 0
          ? HttpResponse.json({ error: 'division by zero' }, { status: 400 })
          : HttpResponse.json({ result: a / b })
      case 'sqrt':
        return HttpResponse.json({ result: Math.sqrt(a) })
      default:
        return HttpResponse.json({ error: `unsupported operation "${operation}"` }, { status: 400 })
    }
  }),
]
