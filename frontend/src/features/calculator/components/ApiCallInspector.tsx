import type { ApiCall } from '../types/calculator.types'

interface ApiCallInspectorProps {
  call: ApiCall | null
}

/** Shows the last request sent to the Go API and what it answered. */
export function ApiCallInspector({ call }: ApiCallInspectorProps) {
  const isSuccess = call?.status === 200

  return (
    <details className="card group p-4">
      <summary className="flex cursor-pointer list-none items-center justify-between text-label text-fg-muted uppercase">
        Last API call
        <span aria-hidden="true" className="text-fg-subtle transition-transform group-open:rotate-180">
          ⌄
        </span>
      </summary>
      <div className="mt-3 space-y-2">
        <p className="text-caption text-fg-subtle">Request</p>
        <pre className="overflow-x-auto rounded-lg bg-bg-deep p-3 text-caption text-fg">
          {call ? `POST /api/calculate\n${JSON.stringify(call.request, null, 2)}` : '—'}
        </pre>
        <p className="flex items-center gap-2 text-caption text-fg-subtle">
          Response
          {call && (
            <span
              className={`rounded-xs px-1.5 py-0.5 text-[11px] font-semibold ${
                isSuccess ? 'bg-mint/15 text-mint' : 'bg-danger/15 text-danger'
              }`}
            >
              {call.status ?? 'network error'}
            </span>
          )}
        </p>
        <pre className="overflow-x-auto rounded-lg bg-bg-deep p-3 text-caption text-fg">
          {call?.response != null ? JSON.stringify(call.response, null, 2) : '—'}
        </pre>
      </div>
    </details>
  )
}
