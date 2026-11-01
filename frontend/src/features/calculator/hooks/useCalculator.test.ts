import { act, renderHook } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '../../../tests/mocks/server'
import { MAX_HISTORY, useCalculator } from './useCalculator'

function setup() {
  const { result } = renderHook(() => useCalculator())
  const type = (keys: string) =>
    act(() => {
      for (const key of keys) {
        if (key === '.') result.current.inputDecimal()
        else result.current.inputDigit(key)
      }
    })
  return { result, type }
}

describe('useCalculator', () => {
  it('starts at zero with no pending operation', () => {
    const { result } = setup()
    expect(result.current.display).toBe('0')
    expect(result.current.expression).toBe('')
    expect(result.current.activeOperation).toBeNull()
  })

  it('edits the entry: digits, decimal, sign and backspace', () => {
    const { result, type } = setup()
    type('12.50')
    expect(result.current.display).toBe('12.50')

    act(() => result.current.toggleSign())
    expect(result.current.display).toBe('-12.50')

    act(() => result.current.backspace())
    expect(result.current.display).toBe('-12.5')

    act(() => result.current.clear())
    expect(result.current.display).toBe('0')
  })

  it('calculates through the API and records history and the last call', async () => {
    const { result, type } = setup()
    type('12')
    await act(() => result.current.chooseOperation('divide'))
    expect(result.current.expression).toBe('12 ÷')
    expect(result.current.activeOperation).toBe('divide')

    type('4')
    await act(() => result.current.evaluate())

    expect(result.current.display).toBe('3')
    expect(result.current.expression).toBe('12 ÷ 4 =')
    expect(result.current.history[0]).toMatchObject({ expression: '12 ÷ 4', result: 3 })
    expect(result.current.lastCall).toEqual({
      request: { operation: 'divide', a: 12, b: 4 },
      status: 200,
      response: { result: 3 },
    })
  })

  it('chains operations by evaluating the pending one first', async () => {
    const { result, type } = setup()
    type('2')
    await act(() => result.current.chooseOperation('add'))
    type('3')
    await act(() => result.current.chooseOperation('multiply'))

    expect(result.current.display).toBe('5')
    expect(result.current.expression).toBe('5 ×')

    type('4')
    await act(() => result.current.evaluate())
    expect(result.current.display).toBe('20')
  })

  it('switches the operator when no second operand was typed', async () => {
    const { result, type } = setup()
    type('5')
    await act(() => result.current.chooseOperation('add'))
    await act(() => result.current.chooseOperation('subtract'))

    expect(result.current.expression).toBe('5 −')
    expect(result.current.history).toHaveLength(0)
  })

  it('applies unary operations to the current entry', async () => {
    const { result, type } = setup()
    type('81')
    await act(() => result.current.chooseOperation('sqrt'))

    expect(result.current.display).toBe('9')
    expect(result.current.expression).toBe('√81 =')
    expect(result.current.lastCall?.request).toEqual({ operation: 'sqrt', a: 81 })
  })

  it('validates that a second number was entered', async () => {
    const { result, type } = setup()
    type('5')
    await act(() => result.current.chooseOperation('add'))
    await act(() => result.current.evaluate())

    expect(result.current.error).toBe('Enter a second number')
    expect(result.current.lastCall).toBeNull()
  })

  it('does nothing on equals without a pending operation', async () => {
    const { result, type } = setup()
    type('5')
    await act(() => result.current.evaluate())
    expect(result.current.display).toBe('5')
    expect(result.current.lastCall).toBeNull()
  })

  it('shows API errors and keeps the input so it can be corrected', async () => {
    const { result, type } = setup()
    type('8')
    await act(() => result.current.chooseOperation('divide'))
    type('0')
    await act(() => result.current.evaluate())

    expect(result.current.error).toBe('Division by zero')
    expect(result.current.lastCall).toMatchObject({ status: 400, response: { error: 'division by zero' } })
    expect(result.current.display).toBe('0')
    expect(result.current.expression).toBe('8 ÷')

    type('2')
    expect(result.current.error).toBeNull()
    await act(() => result.current.evaluate())
    expect(result.current.display).toBe('4')
  })

  it('keeps state when a chained or unary call fails', async () => {
    server.use(http.post('/api/calculate', () => HttpResponse.json({ error: 'boom' }, { status: 400 })))
    const { result, type } = setup()
    type('4')
    await act(() => result.current.chooseOperation('sqrt'))
    expect(result.current.display).toBe('4')

    await act(() => result.current.chooseOperation('add'))
    type('1')
    await act(() => result.current.chooseOperation('add'))
    expect(result.current.error).toBe('Boom')
    expect(result.current.expression).toBe('4 +')
  })

  it(`keeps the latest ${MAX_HISTORY} history entries`, async () => {
    const { result, type } = setup()
    for (let i = 1; i <= MAX_HISTORY + 1; i++) {
      type(String(i))
      await act(() => result.current.chooseOperation('add'))
      type('1')
      await act(() => result.current.evaluate())
    }

    expect(result.current.history).toHaveLength(MAX_HISTORY)
    expect(result.current.history[0]?.expression).toBe(`${MAX_HISTORY + 1} + 1`)

    act(() => result.current.clearHistory())
    expect(result.current.history).toHaveLength(0)
  })
})
