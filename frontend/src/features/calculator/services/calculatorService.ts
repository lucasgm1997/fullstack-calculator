import { HttpError, requestJson } from '../../../lib/api'
import type { ApiErrorResponse, CalculateRequest, CalculateResponse } from '../types/calculator.types'

/** A calculation failure with a message that is safe to show to the user. */
export class CalculatorError extends Error {
  readonly status: number | null
  readonly body: unknown

  constructor(message: string, status: number | null = null, body: unknown = null) {
    super(message)
    this.name = 'CalculatorError'
    this.status = status
    this.body = body
  }
}

function isApiError(body: unknown): body is ApiErrorResponse {
  return typeof body === 'object' && body !== null && typeof (body as ApiErrorResponse).error === 'string'
}

function capitalize(text: string): string {
  return text.charAt(0).toUpperCase() + text.slice(1)
}

/** Maps transport failures to user-facing messages. */
function toCalculatorError(error: unknown): CalculatorError {
  if (!(error instanceof HttpError)) {
    return new CalculatorError('Something went wrong. Please try again.')
  }
  if (error.status === 0) {
    return new CalculatorError('Cannot reach the calculator service. Check that the API is running.')
  }
  if (error.status >= 500) {
    return new CalculatorError('The calculator service failed. Please try again.', error.status, error.body)
  }
  if (isApiError(error.body)) {
    return new CalculatorError(capitalize(error.body.error), error.status, error.body)
  }
  return new CalculatorError('The request was rejected by the calculator service.', error.status, error.body)
}

/** Calls POST /api/calculate and resolves with the numeric result. */
export async function calculate(request: CalculateRequest): Promise<number> {
  try {
    const response = await requestJson<CalculateResponse>('/api/calculate', {
      method: 'POST',
      body: JSON.stringify(request),
    })
    return response.result
  } catch (error) {
    throw toCalculatorError(error)
  }
}

/** Resolves true when GET /api/health answers successfully. */
export async function checkHealth(): Promise<boolean> {
  try {
    await requestJson('/api/health')
    return true
  } catch {
    return false
  }
}
