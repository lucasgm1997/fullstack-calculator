/**
 * Pure helpers for editing the number being typed on the keypad.
 * They only handle input formatting; all arithmetic happens in the API.
 */

/** float64 represents up to 15 significant decimal digits exactly. */
export const MAX_DIGITS = 15

const DIGIT = /^[0-9]$/

function countDigits(entry: string): number {
  return entry.replace(/[^0-9]/g, '').length
}

export function appendDigit(entry: string, digit: string): string {
  if (!DIGIT.test(digit) || countDigits(entry) >= MAX_DIGITS) return entry
  if (entry === '0') return digit
  if (entry === '-0') return `-${digit}`
  return entry + digit
}

export function appendDecimal(entry: string): string {
  return entry.includes('.') ? entry : `${entry}.`
}

export function toggleSign(entry: string): string {
  if (entry.startsWith('-')) return entry.slice(1)
  return entry === '0' ? entry : `-${entry}`
}

export function backspace(entry: string): string {
  const next = entry.slice(0, -1)
  return next === '' || next === '-' ? '0' : next
}

/** Formats an API result for display, hiding binary floating-point noise. */
export function formatNumber(value: number): string {
  if (Object.is(value, -0)) value = 0
  const abs = Math.abs(value)

  if (abs !== 0 && (abs >= 1e15 || abs < 1e-9)) {
    return value.toExponential(6).replace(/\.?0+e/, 'e')
  }
  if (Number.isInteger(value)) {
    return value.toLocaleString('en-US')
  }
  // 0.1 + 0.2 = 0.30000000000000004 → "0.3"
  return Number(value.toPrecision(12)).toLocaleString('en-US', { maximumFractionDigits: 10 })
}
