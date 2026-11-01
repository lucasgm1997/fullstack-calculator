import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '../../../tests/mocks/server'
import { CalculatorError, calculate, checkHealth } from './calculatorService'

describe('calculate', () => {
  it('posts the request and returns the result', async () => {
    let received: unknown
    server.use(
      http.post('/api/calculate', async ({ request }) => {
        received = await request.json()
        return HttpResponse.json({ result: 15 })
      }),
    )

    await expect(calculate({ operation: 'add', a: 10, b: 5 })).resolves.toBe(15)
    expect(received).toEqual({ operation: 'add', a: 10, b: 5 })
  })

  it('surfaces the API error message for 4xx responses', async () => {
    const error = await calculate({ operation: 'divide', a: 1, b: 0 }).catch((e: unknown) => e)

    expect(error).toBeInstanceOf(CalculatorError)
    expect(error).toMatchObject({ message: 'Division by zero', status: 400, body: { error: 'division by zero' } })
  })

  it('uses a generic message for 4xx responses without an error body', async () => {
    server.use(http.post('/api/calculate', () => new HttpResponse('nope', { status: 422 })))

    await expect(calculate({ operation: 'add', a: 1, b: 1 })).rejects.toThrow('rejected by the calculator service')
  })

  it('hides server error details', async () => {
    server.use(http.post('/api/calculate', () => HttpResponse.json({ error: 'stack trace' }, { status: 500 })))

    await expect(calculate({ operation: 'add', a: 1, b: 1 })).rejects.toThrow('The calculator service failed')
  })

  it('reports network failures', async () => {
    server.use(http.post('/api/calculate', () => HttpResponse.error()))

    const error = await calculate({ operation: 'add', a: 1, b: 1 }).catch((e: unknown) => e)
    expect(error).toMatchObject({ message: expect.stringContaining('Cannot reach'), status: null })
  })
})

describe('checkHealth', () => {
  it('is true when the API answers', async () => {
    await expect(checkHealth()).resolves.toBe(true)
  })

  it('is false when the API fails', async () => {
    server.use(http.get('/api/health', () => HttpResponse.error()))
    await expect(checkHealth()).resolves.toBe(false)
  })
})
