/**
 * Base URL of the API. Empty by default so the browser calls the same origin
 * (`/api/...`), which the Vite dev proxy or nginx forwards to the Go service.
 */
export const API_BASE_URL: string = import.meta.env.VITE_API_BASE_URL ?? ''

/** Raised for any failed HTTP call. `status` is 0 when the network failed. */
export class HttpError extends Error {
  readonly status: number
  readonly body: unknown

  constructor(message: string, status: number, body: unknown = null) {
    super(message)
    this.name = 'HttpError'
    this.status = status
    this.body = body
  }
}

async function readJson(response: Response): Promise<unknown> {
  try {
    return await response.json()
  } catch {
    return null
  }
}

/** Sends a JSON request and returns the decoded JSON body on 2xx responses. */
export async function requestJson<T>(path: string, init: RequestInit = {}): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers: { 'Content-Type': 'application/json', Accept: 'application/json', ...init.headers },
    })
  } catch {
    throw new HttpError('Network request failed', 0)
  }

  const body = await readJson(response)
  if (!response.ok) {
    throw new HttpError(`Request failed with status ${response.status}`, response.status, body)
  }
  return body as T
}
