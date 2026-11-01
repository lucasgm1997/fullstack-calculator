import { useEffect, useState } from 'react'
import { checkHealth } from '../services/calculatorService'

export type ApiStatus = 'checking' | 'online' | 'offline'

/** Pings GET /api/health once on mount. */
export function useApiStatus(): ApiStatus {
  const [status, setStatus] = useState<ApiStatus>('checking')

  useEffect(() => {
    let active = true
    checkHealth().then((ok) => {
      if (active) setStatus(ok ? 'online' : 'offline')
    })
    return () => {
      active = false
    }
  }, [])

  return status
}
