interface CalculatorDisplayProps {
  expression: string
  value: string
  error: string | null
  isLoading: boolean
}

/** Shrinks long numbers so they fit the card on small screens. */
function sizeClass(value: string): string {
  if (value.length > 14) return 'text-h2'
  if (value.length > 10) return 'text-h1'
  return 'text-h1 sm:text-display'
}

export function CalculatorDisplay({ expression, value, error, isLoading }: CalculatorDisplayProps) {
  return (
    <section
      aria-label="Calculator display"
      aria-busy={isLoading}
      className={`card flex min-h-40 flex-col justify-end gap-1 p-5 text-right transition-colors ${
        error ? 'border-danger' : ''
      }`}
    >
      <p className="min-h-6 truncate text-body text-fg-muted tabular-nums">
        {expression}
      </p>
      <output
        aria-label="Result"
        aria-live="polite"
        title={value}
        className={`block truncate text-fg-strong tabular-nums transition-opacity ${sizeClass(value)} ${
          isLoading ? 'opacity-60' : ''
        }`}
      >
        {value}
      </output>
      {error && (
        <p role="alert" className="text-sm font-medium text-danger">
          {error}
        </p>
      )}
    </section>
  )
}
