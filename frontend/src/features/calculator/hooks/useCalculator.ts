import { useRef, useState } from 'react'
import { CalculatorError, calculate } from '../services/calculatorService'
import {
  OPERATIONS,
  type ApiCall,
  type CalculateRequest,
  type HistoryEntry,
  type Operation,
} from '../types/calculator.types'
import { appendDecimal, appendDigit, backspace, formatNumber, toggleSign } from '../utils/entry'

export const MAX_HISTORY = 6

interface KeypadState {
  /** The number currently shown, as typed (e.g. "0.50"). */
  entry: string
  /** Left operand of a pending binary operation. */
  accumulator: number | null
  pendingOperation: Operation | null
  /** The next digit replaces `entry` instead of appending to it. */
  isEntryFresh: boolean
  /** A binary operation was chosen but its second operand was not typed yet. */
  awaitingOperand: boolean
  /** Secondary display line, e.g. "12 ÷" or "12 ÷ 4 =". */
  expression: string
}

const INITIAL_KEYPAD: KeypadState = {
  entry: '0',
  accumulator: null,
  pendingOperation: null,
  isEntryFresh: true,
  awaitingOperand: false,
  expression: '',
}

function describe({ operation, a, b }: CalculateRequest): string {
  const { symbol, unary } = OPERATIONS[operation]
  return unary ? `${symbol}${formatNumber(a)}` : `${formatNumber(a)} ${symbol} ${formatNumber(b ?? 0)}`
}

/**
 * ViewModel of the calculator: owns keypad input, pending operation, loading,
 * error, history and the last API call. Views only render this state and call
 * the returned actions.
 */
export function useCalculator() {
  const [keypad, setKeypad] = useState<KeypadState>(INITIAL_KEYPAD)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [history, setHistory] = useState<HistoryEntry[]>([])
  const [lastCall, setLastCall] = useState<ApiCall | null>(null)
  // A ref (not state) so that two key presses within one render cannot both
  // start a request.
  const busy = useRef(false)
  const nextHistoryId = useRef(1)

  /** Sends a request; resolves with the result, or null after showing the error. */
  async function run(request: CalculateRequest): Promise<number | null> {
    busy.current = true
    setIsLoading(true)
    setError(null)
    try {
      const result = await calculate(request)
      setLastCall({ request, status: 200, response: { result } })
      setHistory((entries) =>
        [{ id: nextHistoryId.current++, expression: describe(request), result }, ...entries].slice(0, MAX_HISTORY),
      )
      return result
    } catch (caught) {
      const failure = caught instanceof CalculatorError ? caught : new CalculatorError('Something went wrong.')
      setError(failure.message)
      setLastCall({ request, status: failure.status, response: failure.body })
      return null
    } finally {
      busy.current = false
      setIsLoading(false)
    }
  }

  /** Applies a synchronous keypad edit unless a request is in flight. */
  function edit(update: (state: KeypadState) => KeypadState) {
    if (busy.current) return
    setError(null)
    setKeypad(update)
  }

  const inputDigit = (digit: string) =>
    edit((s) =>
      s.isEntryFresh
        ? { ...s, entry: appendDigit('0', digit), isEntryFresh: false, awaitingOperand: false, expression: s.pendingOperation ? s.expression : '' }
        : { ...s, entry: appendDigit(s.entry, digit) },
    )

  const inputDecimal = () =>
    edit((s) =>
      s.isEntryFresh
        ? { ...s, entry: '0.', isEntryFresh: false, awaitingOperand: false, expression: s.pendingOperation ? s.expression : '' }
        : { ...s, entry: appendDecimal(s.entry) },
    )

  const negate = () => edit((s) => (s.awaitingOperand ? s : { ...s, entry: toggleSign(s.entry) }))

  const deleteLast = () => edit((s) => (s.isEntryFresh ? s : { ...s, entry: backspace(s.entry) }))

  const clear = () => edit(() => INITIAL_KEYPAD)

  async function chooseOperation(operation: Operation) {
    if (busy.current) return
    const value = Number(keypad.entry)

    if (OPERATIONS[operation].unary) {
      const result = await run({ operation, a: value })
      if (result === null) return
      setKeypad((s) => ({
        ...s,
        entry: String(result),
        isEntryFresh: true,
        awaitingOperand: false,
        expression: s.pendingOperation ? s.expression : `${describe({ operation, a: value })} =`,
      }))
      return
    }

    let accumulator = value
    if (keypad.pendingOperation && keypad.accumulator !== null) {
      if (keypad.awaitingOperand) {
        // "5 + ×" just switches the operator.
        accumulator = keypad.accumulator
      } else {
        // "2 + 3 ×" evaluates 2 + 3 first, then continues with ×.
        const result = await run({ operation: keypad.pendingOperation, a: keypad.accumulator, b: value })
        if (result === null) return
        accumulator = result
      }
    } else {
      setError(null)
    }

    setKeypad({
      entry: String(accumulator),
      accumulator,
      pendingOperation: operation,
      isEntryFresh: true,
      awaitingOperand: true,
      expression: `${formatNumber(accumulator)} ${OPERATIONS[operation].symbol}`,
    })
  }

  async function evaluate() {
    if (busy.current) return
    const { pendingOperation, accumulator, awaitingOperand, entry } = keypad
    if (!pendingOperation || accumulator === null) return
    if (awaitingOperand) {
      setError('Enter a second number')
      return
    }

    const request = { operation: pendingOperation, a: accumulator, b: Number(entry) }
    const result = await run(request)
    if (result === null) return
    setKeypad({ ...INITIAL_KEYPAD, entry: String(result), expression: `${describe(request)} =` })
  }

  const clearHistory = () => setHistory([])

  return {
    /** Text for the main display line. */
    display: keypad.isEntryFresh ? formatNumber(Number(keypad.entry)) : keypad.entry,
    expression: keypad.expression,
    activeOperation: keypad.awaitingOperand ? keypad.pendingOperation : null,
    isLoading,
    error,
    history,
    lastCall,
    inputDigit,
    inputDecimal,
    toggleSign: negate,
    backspace: deleteLast,
    clear,
    chooseOperation,
    evaluate,
    clearHistory,
  }
}

export type CalculatorViewModel = ReturnType<typeof useCalculator>
