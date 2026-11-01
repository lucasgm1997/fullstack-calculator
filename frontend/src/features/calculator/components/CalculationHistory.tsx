import type { HistoryEntry } from '../types/calculator.types'
import { formatNumber } from '../utils/entry'

interface CalculationHistoryProps {
  entries: HistoryEntry[]
  onClear: () => void
}

export function CalculationHistory({ entries, onClear }: CalculationHistoryProps) {
  return (
    <section aria-labelledby="history-heading" className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 id="history-heading" className="text-lead text-fg-strong">
          History
        </h2>
        {entries.length > 0 && (
          <button type="button" className="text-sm text-fg-muted hover:text-fg" onClick={onClear}>
            Clear
          </button>
        )}
      </div>

      {entries.length === 0 ? (
        <p className="rounded-xl border border-dashed border-line p-4 text-center text-sm text-fg-subtle">
          No calculations yet
        </p>
      ) : (
        <ul className="space-y-2">
          {entries.map((entry) => (
            <li key={entry.id} className="flex items-center justify-between gap-4 rounded-xl bg-surface px-4 py-3">
              <span className="truncate text-sm text-fg-muted tabular-nums">{entry.expression}</span>
              <span className="shrink-0 font-medium text-fg-strong tabular-nums">= {formatNumber(entry.result)}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
