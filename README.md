# Full-Stack Calculator

A calculator with a **React + TypeScript + Vite** frontend and a **Go + Gin** REST API. The frontend does no arithmetic: every calculation is sent to the backend.

- **Operations:** add, subtract, multiply, divide, plus optional power, square root and percentage
- **Validation and errors:** malformed JSON, wrong types, missing operands, unsupported operations, division by zero, square root of a negative number, overflow
- **UI:** keypad with physical keyboard support, validation feedback, an API status indicator, a "Last API call" inspector and calculation history. Mobile-first layout built on the Sezzle-inspired design system in [`design-system/`](design-system/DESIGN.md)
- **Tests:** Go unit and HTTP integration tests (100% statement coverage of `internal/`), plus Vitest unit and React Testing Library + MSW integration tests
- **Docker:** multi-stage images and Docker Compose, with nginx proxying `/api`
- **CI:** GitHub Actions for formatting, lint, type check, tests with coverage, builds and the Docker build

## Repository layout

```text
backend/                      Go + Gin API
  cmd/api/main.go             entrypoint (PORT env, graceful shutdown)
  internal/router/            route + middleware registration
  internal/handler/           HTTP ↔ use case, error → status mapping
  internal/usecase/           input validation, orchestration (no Gin)
  internal/domain/            pure arithmetic + domain errors
  internal/dto/               JSON request/response contract
  tests/                      httptest integration tests (full stack)
frontend/                     React + TypeScript + Vite + Tailwind v4
  src/features/calculator/
    CalculatorPage.tsx        page composition + keyboard support
    components/               Display, Keypad, ApiCallInspector, CalculationHistory
    hooks/useCalculator.ts    ViewModel: keypad state, loading, errors, history
    services/                 calculatorService: API calls + error mapping
    types/                    API contract types + operation metadata
    utils/entry.ts            pure input-editing / number-formatting helpers
  src/lib/api.ts              fetch wrapper, base URL from env
  src/tests/                  Vitest setup + MSW mock server
design-system/                tokens.css (Tailwind v4 @theme), DESIGN.md, sketch
compose.yaml                  frontend (nginx) + backend
```

## Prerequisites

| Tool | Version used |
| --- | --- |
| Go | 1.27 |
| Node.js / npm | 24+ / 11+ (developed on Node 25) |
| Docker (optional) | Docker Compose v2 |

## Running locally

**1. Backend** (listens on `:8080`, override with `PORT`)

```bash
cd backend
go run ./cmd/api
```

**2. Frontend** (in another terminal)

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173
```

The Vite dev server proxies `/api/*` to `http://localhost:8080`, so the browser only talks to one origin and the backend needs no CORS configuration. Set `VITE_API_PROXY_TARGET` to point the proxy elsewhere, or `VITE_API_BASE_URL` to call an absolute URL (see [`.env.example`](.env.example)).

## Running with Docker

```bash
docker compose up --build
```

- App: http://localhost:3000 (set `FRONTEND_PORT` to change it)
- API: http://localhost:8080

The frontend image is a static build served by nginx, which also proxies `/api/` to the `backend` service. The browser cannot resolve Docker service hostnames, so the browser always calls its own origin and nginx forwards the request inside the Docker network. The frontend image is built from the repository root because it imports `design-system/tokens.css`.

## Tests and coverage

**Backend**

```bash
cd backend
go test ./...                                                     # unit + integration
go test -coverpkg=./internal/... -coverprofile=coverage.out ./... # coverage across packages
go tool cover -func=coverage.out                                  # summary (100.0%)
go tool cover -html=coverage.out -o coverage.html                 # HTML report
```

- `internal/domain`: table-driven tests of every operation and every error (overflow, NaN, unknown operation, and so on)
- `internal/usecase`: operand presence (`nil` vs `0`), unary vs binary handling, and error propagation
- `internal/handler`: handler tests against a fake use case, covering JSON decoding errors, error-to-status mapping and a hidden 500
- `tests/`: integration tests through the real Gin router with `httptest`, covering every operation, every error, the JSON content type, 404 and 405

**Frontend**

```bash
cd frontend
npm test               # Vitest, run once
npm run test:coverage  # text summary + HTML report in frontend/coverage/index.html
npm run lint && npm run typecheck
```

- `utils/entry.test.ts`: pure input editing and number formatting
- `services/calculatorService.test.ts`: request body and error mapping (4xx message, 5xx hidden, network failure) against MSW
- `hooks/useCalculator.test.ts`: the keypad state machine (chaining, switching operator, unary ops, validation, error recovery, history cap)
- `CalculatorPage.test.tsx`: integration from View to hook to service to mocked HTTP (keypad, keyboard, error display, API status, advanced mode)

Current results: frontend 44 tests, 94% statements and 98% lines. Backend 100% statements of `internal/`.

## API

### `POST /api/calculate`

| Field | Type | Notes |
| --- | --- | --- |
| `operation` | string | `add`, `subtract`, `multiply`, `divide`, `power`, `sqrt`, `percentage` |
| `a` | number | required |
| `b` | number | required for every operation except `sqrt` (where it is ignored) |

- `percentage` means **"a percent of b"**: `(a / 100) * b`. For example, `15` and `200` give `30`.
- `sqrt` is unary: only `a` is used.

Success returns `200 {"result": <number>}`. Every error returns `{"error": "<message>"}`:

| Status | When | Example message |
| --- | --- | --- |
| 400 | body is not valid JSON or is empty | `request body must be valid JSON` |
| 400 | operand has the wrong type or is out of float64 range | `field "a" must be a finite number` |
| 400 | operand missing or `null` | `missing operand: field "b" is required for add` |
| 400 | unknown or missing operation | `unsupported operation "modulo"` |
| 400 | division by zero | `division by zero` |
| 400 | square root of a negative number | `square root of a negative number` |
| 400 | result overflows or is undefined (for example `(-8)^0.5`) | `result is not a finite number` |
| 400 | body larger than 4 KB | `request body too large` |
| 404 / 405 | unknown route / wrong method | `not found` / `method not allowed` |
| 500 | unexpected error (details are only logged) | `internal server error` |

### `GET /api/health`

Returns `200 {"status":"ok"}`. The UI uses it for the "API online" indicator.

### Examples

```bash
curl -X POST localhost:8080/api/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":10,"b":5}'
# 200 {"result":15}

curl -X POST localhost:8080/api/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"divide","a":10,"b":0}'
# 400 {"error":"division by zero"}

curl -X POST localhost:8080/api/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","a":81}'
# 200 {"result":9}

curl -X POST localhost:8080/api/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":"ten","b":5}'
# 400 {"error":"field \"a\" must be a finite number"}
```

## Design decisions

**Backend: a simplified Clean Architecture.** Requests flow Router → Handler → Use case → Domain.
- The **domain** is pure Go with sentinel errors (`ErrDivisionByZero`, …) and knows nothing about HTTP or JSON.
- The **use case** validates what each operation needs, for example that `b` is present for binary operations.
- The **handler** decodes JSON and maps errors to statuses with `errors.Is`. It depends on a small `Calculator` interface (constructor injection), which makes it testable with a fake.
- **DTOs** are only the transport contract.
- There is no repository or extra service layer, because nothing is persisted.

**Operands are `*float64`.** This keeps `{"a":0}` (valid) apart from a missing `a` (invalid), which a plain `float64` cannot do.

**One endpoint with an `operation` field** instead of one route per operation. This gives a single contract to validate, test and extend: adding an operation touches the domain switch and the UI metadata only.

**HTTP statuses.** Division by zero and the other arithmetic errors return 400, as proposed in the project guidelines. They are client input problems, not server faults.

**Non-finite results are errors.** JSON cannot encode `Inf` or `NaN`, so overflow and undefined results are rejected explicitly instead of failing during encoding.

**Frontend: an MVVM-inspired structure.**
- The **Views** (`CalculatorPage`, `components/*`) only render state and forward events.
- **`useCalculator`** is the ViewModel. It owns the keypad state machine, loading, errors, history and the last API call.
- **`calculatorService`** calls the API and turns transport failures into user-friendly messages. 4xx responses show the API message, while 5xx and network errors show a generic message.
- **`lib/api.ts`** is a small `fetch` wrapper.
- State uses plain React `useState`, with no Redux, Zustand or React Query, since the state is local to one screen.

**The frontend has no math.** The keypad only edits the number being typed (digits, decimal point, sign, backspace). Every operation, including chained ones like `2 + 3 ×`, which first asks the API for `2 + 3`, is computed by the backend.

**Input validation in the UI:**
- one decimal point
- at most 15 significant digits (the float64 exact range)
- "Enter a second number" when `=` is pressed without a second operand
- no new input while a request is in flight

A `useRef` guard prevents double submits from fast key presses.

**Display formatting.** Results are rounded to 12 significant digits for display only, so `0.1 + 0.2` shows `0.3`. The full-precision value is kept for further calculations. Very large or very small numbers use exponential notation.

**Accessibility:**
- keys are real `<button>`s with accessible names ("Divide", "Square root", …)
- operators use `aria-pressed`
- the result is an `<output aria-live="polite">`
- errors use `role="alert"`
- full keyboard support: `0-9 . + - * / ^ % Enter = Backspace Escape`

**Design system.** `design-system/tokens.css` (Tailwind v4 `@theme` with `btn`/`card`/`input` utilities) is imported directly by `src/index.css`, so the app and the design reference share one source of truth. The UI follows `design-system/calculator-sketch.html`.

## Assumptions and known limitations

- The calculator works on two operands at a time. Chained input is evaluated left to right as you go, like a pocket calculator, so `2 + 3 × 4 = 20`. There is no operator precedence or expression parser.
- Numbers are IEEE-754 float64, so very large integers lose precision beyond 15–16 digits.
- History is kept in memory and is lost on reload.
- The "Last API call" panel shows the error body for HTTP errors. For network failures there is no body, so it shows "network error".
- Not implemented: Playwright E2E tests (optional), rate limiting and authentication.

## AI usage

This project was built with AI assistance (Claude Code). The prompts used are recorded in [`prompts_history.md`](prompts_history.md), as the assignment requests. The working agreements given to the agent are in [`AGENTS.md`](AGENTS.md) and [`IMPLEMENTATION_PRIORITIES.md`](IMPLEMENTATION_PRIORITIES.md).
