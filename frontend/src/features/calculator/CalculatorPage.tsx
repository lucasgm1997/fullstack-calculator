import { useEffect, useState } from 'react'
import { ApiCallInspector } from './components/ApiCallInspector'
import { CalculationHistory } from './components/CalculationHistory'
import { CalculatorDisplay } from './components/CalculatorDisplay'
import { CalculatorKeypad } from './components/CalculatorKeypad'
import { useApiStatus, type ApiStatus } from './hooks/useApiStatus'
import { useCalculator } from './hooks/useCalculator'
import type { Operation } from './types/calculator.types'

const KEY_TO_OPERATION: Record<string, Operation> = {
  '+': 'add',
  '-': 'subtract',
  '*': 'multiply',
  x: 'multiply',
  '/': 'divide',
  '^': 'power',
  '%': 'percentage',
}

const STATUS_TEXT: Record<ApiStatus, string> = {
  checking: 'Checking API…',
  online: 'API online',
  offline: 'API offline',
}

const STATUS_DOT: Record<ApiStatus, string> = {
  checking: 'bg-fg-subtle',
  online: 'bg-mint',
  offline: 'bg-danger',
}

export function CalculatorPage() {
  const calculator = useCalculator()
  const apiStatus = useApiStatus()
  const [showAdvanced, setShowAdvanced] = useState(false)

  // Physical keyboard support. Re-subscribed on each render so the handler
  // always sees the latest calculator state.
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.ctrlKey || event.metaKey || event.altKey) return
      // Let Enter/Space activate a focused button or <summary> natively.
      const target = event.target as HTMLElement | null
      if ((event.key === 'Enter' || event.key === ' ') && target?.closest('button, summary')) return

      const { key } = event
      if (/^[0-9]$/.test(key)) calculator.inputDigit(key)
      else if (key === '.' || key === ',') calculator.inputDecimal()
      else if (KEY_TO_OPERATION[key]) calculator.chooseOperation(KEY_TO_OPERATION[key])
      else if (key === 'Enter' || key === '=') calculator.evaluate()
      else if (key === 'Backspace') calculator.backspace()
      else if (key === 'Escape' || key === 'Delete') calculator.clear()
      else return

      event.preventDefault()
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  })

  return (
    <div className="relative min-h-dvh overflow-hidden bg-bg font-sans text-fg">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 top-[-20%] h-[70vh] bg-(image:--gradient-hero-glow) opacity-30 blur-2xl"
      />

      <main className="relative mx-auto flex w-full max-w-md flex-col gap-4 px-4 pt-6 pb-10">
        <header className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3 rounded-pill bg-surface py-1.5 pr-4 pl-1.5">
            <span aria-hidden="true" className="grid size-9 place-items-center rounded-full bg-brand-500 font-medium text-white">
              C
            </span>
            <div className="leading-tight">
              <h1 className="text-sm font-medium text-fg-strong">Calculator</h1>
              <p className="flex items-center gap-1 text-caption text-fg-muted" role="status">
                <span aria-hidden="true" className={`size-2 rounded-full ${STATUS_DOT[apiStatus]}`} />
                {STATUS_TEXT[apiStatus]}
              </p>
            </div>
          </div>
          <button
            type="button"
            className="btn-ghost h-9 px-4 text-caption aria-pressed:border-mint aria-pressed:bg-mint/10 aria-pressed:text-mint"
            aria-pressed={showAdvanced}
            onClick={() => setShowAdvanced((on) => !on)}
          >
            Advanced
          </button>
        </header>

        <CalculatorDisplay
          expression={calculator.expression}
          value={calculator.display}
          error={calculator.error}
          isLoading={calculator.isLoading}
        />

        <CalculatorKeypad
          activeOperation={calculator.activeOperation}
          showAdvanced={showAdvanced}
          isLoading={calculator.isLoading}
          onDigit={calculator.inputDigit}
          onDecimal={calculator.inputDecimal}
          onOperation={calculator.chooseOperation}
          onEquals={calculator.evaluate}
          onClear={calculator.clear}
          onBackspace={calculator.backspace}
          onToggleSign={calculator.toggleSign}
        />

        <ApiCallInspector call={calculator.lastCall} />

        <CalculationHistory entries={calculator.history} onClear={calculator.clearHistory} />
      </main>
    </div>
  )
}
