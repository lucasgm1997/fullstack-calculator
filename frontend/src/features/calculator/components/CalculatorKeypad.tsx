import type { ReactNode } from 'react'
import { OPERATIONS, type Operation } from '../types/calculator.types'

interface CalculatorKeypadProps {
  activeOperation: Operation | null
  showAdvanced: boolean
  isLoading: boolean
  onDigit: (digit: string) => void
  onDecimal: () => void
  onOperation: (operation: Operation) => void
  onEquals: () => void
  onClear: () => void
  onBackspace: () => void
  onToggleSign: () => void
}

const ADVANCED_KEYS: { operation: Operation; content: ReactNode }[] = [
  {
    operation: 'power',
    content: (
      <>
        x<sup>y</sup>
      </>
    ),
  },
  { operation: 'sqrt', content: '√x' },
  { operation: 'percentage', content: <span className="text-lg">x% of y</span> },
]

export function CalculatorKeypad({
  activeOperation,
  showAdvanced,
  isLoading,
  onDigit,
  onDecimal,
  onOperation,
  onEquals,
  onClear,
  onBackspace,
  onToggleSign,
}: CalculatorKeypadProps) {
  const operatorKey = (operation: Operation, content: ReactNode = OPERATIONS[operation].symbol) => {
    const isActive = activeOperation === operation
    return (
      <button
        key={operation}
        type="button"
        className={`key key-op ${isActive ? 'ring-2 ring-mint' : ''}`}
        aria-label={OPERATIONS[operation].label}
        aria-pressed={isActive}
        onClick={() => onOperation(operation)}
      >
        {content}
      </button>
    )
  }

  const digitKey = (digit: string, className = '') => (
    <button key={digit} type="button" className={`key ${className}`} onClick={() => onDigit(digit)}>
      {digit}
    </button>
  )

  return (
    <div className="flex flex-col gap-3">
      {showAdvanced && (
        <div className="grid grid-cols-3 gap-3" role="group" aria-label="Advanced operations">
          {ADVANCED_KEYS.map(({ operation, content }) => operatorKey(operation, content))}
        </div>
      )}

      <div className="grid grid-cols-4 gap-3" role="group" aria-label="Keypad">
        <button type="button" className="key key-fn" aria-label="All clear" onClick={onClear}>
          AC
        </button>
        <button type="button" className="key key-fn" aria-label="Backspace" onClick={onBackspace}>
          ⌫
        </button>
        <button type="button" className="key key-fn" aria-label="Toggle sign" onClick={onToggleSign}>
          ±
        </button>
        {operatorKey('divide')}

        {['7', '8', '9'].map((d) => digitKey(d))}
        {operatorKey('multiply')}

        {['4', '5', '6'].map((d) => digitKey(d))}
        {operatorKey('subtract')}

        {['1', '2', '3'].map((d) => digitKey(d))}
        {operatorKey('add')}

        {digitKey('0', 'col-span-2 justify-start pl-7')}
        <button type="button" className="key" aria-label="Decimal point" onClick={onDecimal}>
          .
        </button>
        <button
          type="button"
          className="key bg-mint text-mint-fg hover:bg-mint-strong"
          aria-label="Equals"
          onClick={onEquals}
        >
          {isLoading ? '…' : '='}
        </button>
      </div>
    </div>
  )
}
