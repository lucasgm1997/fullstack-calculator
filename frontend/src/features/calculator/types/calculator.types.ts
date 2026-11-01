/** Operation identifiers shared with the Go API. */
export type Operation = 'add' | 'subtract' | 'multiply' | 'divide' | 'power' | 'sqrt' | 'percentage'

/** Request body of POST /api/calculate. `b` is omitted for unary operations. */
export interface CalculateRequest {
  operation: Operation
  a: number
  b?: number
}

/** Success body of POST /api/calculate. */
export interface CalculateResponse {
  result: number
}

/** Error body returned by every API endpoint. */
export interface ApiErrorResponse {
  error: string
}

interface OperationInfo {
  /** Symbol shown in the display and history. */
  symbol: string
  /** Accessible name of the key. */
  label: string
  unary: boolean
}

/** Presentation metadata only; arithmetic rules live in the backend. */
export const OPERATIONS: Record<Operation, OperationInfo> = {
  add: { symbol: '+', label: 'Add', unary: false },
  subtract: { symbol: '−', label: 'Subtract', unary: false },
  multiply: { symbol: '×', label: 'Multiply', unary: false },
  divide: { symbol: '÷', label: 'Divide', unary: false },
  power: { symbol: '^', label: 'Power', unary: false },
  sqrt: { symbol: '√', label: 'Square root', unary: true },
  percentage: { symbol: '% of', label: 'Percent of', unary: false },
}

export interface HistoryEntry {
  id: number
  expression: string
  result: number
}

/** Snapshot of the last request/response, shown in the API inspector. */
export interface ApiCall {
  request: CalculateRequest
  status: number | null
  response: unknown
}
