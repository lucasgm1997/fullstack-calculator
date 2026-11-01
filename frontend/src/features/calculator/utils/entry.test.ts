import { describe, expect, it } from 'vitest'
import { MAX_DIGITS, appendDecimal, appendDigit, backspace, formatNumber, toggleSign } from './entry'

describe('appendDigit', () => {
  it('replaces a leading zero', () => {
    expect(appendDigit('0', '7')).toBe('7')
    expect(appendDigit('-0', '7')).toBe('-7')
  })

  it('appends to existing input', () => {
    expect(appendDigit('12', '3')).toBe('123')
    expect(appendDigit('0.', '0')).toBe('0.0')
  })

  it('ignores non-digits', () => {
    expect(appendDigit('12', 'a')).toBe('12')
  })

  it(`limits input to ${MAX_DIGITS} digits`, () => {
    const full = '9'.repeat(MAX_DIGITS)
    expect(appendDigit(full, '1')).toBe(full)
    expect(appendDigit(`-${full}.`, '1')).toBe(`-${full}.`)
  })
})

describe('appendDecimal', () => {
  it('adds a single decimal point', () => {
    expect(appendDecimal('3')).toBe('3.')
    expect(appendDecimal('3.1')).toBe('3.1')
  })
})

describe('toggleSign', () => {
  it('negates and restores', () => {
    expect(toggleSign('5')).toBe('-5')
    expect(toggleSign('-5')).toBe('5')
  })

  it('leaves zero unsigned', () => {
    expect(toggleSign('0')).toBe('0')
  })
})

describe('backspace', () => {
  it('removes the last character', () => {
    expect(backspace('123')).toBe('12')
    expect(backspace('1.')).toBe('1')
  })

  it('falls back to zero', () => {
    expect(backspace('7')).toBe('0')
    expect(backspace('-7')).toBe('0')
  })
})

describe('formatNumber', () => {
  it.each([
    [1234567, '1,234,567'],
    [-0, '0'],
    [0.1 + 0.2, '0.3'],
    [2.5, '2.5'],
    [1 / 3, '0.3333333333'],
    [123456789012345, '123,456,789,012,345'],
    [1e21, '1e+21'],
    [1.5e-12, '1.5e-12'],
  ])('formats %s as %s', (value, expected) => {
    expect(formatNumber(value)).toBe(expected)
  })
})
