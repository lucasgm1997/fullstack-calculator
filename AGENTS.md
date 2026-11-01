# AGENTS.md — Full-Stack Calculator Challenge

## Mission and priorities
Build a small, maintainable calculator with a **React + TypeScript + Vite** frontend and a **Go + Gin** REST API. Follow the assignment requirements first. Optimize for correctness, clarity, idiomatic code, and testability rather than unnecessary abstractions. The assignment suggests **2–4 hours** of work; prioritize the required functionality before optional enhancements.

**Required operations:** addition, subtraction, multiplication, division.
**Optional operations (only after core requirements):** exponentiation, square root, percentage.

The UI must validate inputs, display results and errors clearly, and support basic mobile layouts. The API must validate requests, handle division by zero and malformed inputs, and return JSON.

## Proposed repository layout

```text
calculator-app/
├── frontend/
│   ├── src/
│   │   ├── features/calculator/
│   │   │   ├── components/
│   │   │   │   ├── CalculatorForm.tsx
│   │   │   │   └── CalculatorResult.tsx
│   │   │   ├── hooks/useCalculator.ts
│   │   │   ├── services/calculatorService.ts
│   │   │   ├── types/calculator.types.ts
│   │   │   └── CalculatorPage.tsx
│   │   ├── lib/api.ts
│   │   ├── tests/
│   │   │   ├── setup.ts
│   │   │   └── mocks/{handlers.ts,server.ts}
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── e2e/calculator.spec.ts
│   ├── playwright.config.ts
│   ├── vitest.config.ts
│   ├── package.json
│   └── Dockerfile
├── backend/
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── router/router.go
│   │   ├── handler/calculator_handler.go
│   │   ├── usecase/calculate.go
│   │   ├── domain/calculator.go
│   │   └── dto/calculator_dto.go
│   ├── tests/calculator_integration_test.go
│   ├── go.mod
│   └── Dockerfile
├── .github/workflows/ci.yml
├── compose.yaml
├── .env.example
└── README.md
```

Place unit tests beside their implementation (`*_test.go`, `*.test.ts`, `*.test.tsx`). Create only files that are actually needed; the tree is a guide, not a mandate to add placeholders.

## Frontend architecture
Use **feature-based organization** with a lightweight **MVVM-inspired** separation:

- **View:** `CalculatorPage`, `CalculatorForm`, `CalculatorResult`; present state and handle user interaction, but do not duplicate backend math rules.
- **ViewModel-like hook:** `useCalculator`; manages input, selected operation, loading, result, and error state.
- **Service:** `calculatorService`; calls the REST API and maps transport errors to useful messages.
- **Types:** request/response contracts and operation types.
- **HTTP infrastructure:** `lib/api.ts`; base URL from environment configuration, shared fetch wrapper if useful.

Flow: **View → useCalculator → calculatorService → Go API**.

Prefer React `useState` and native `fetch`. Do **not** add Zustand, Redux, TanStack Query, or a full MVVM framework without a demonstrated need. Use accessible labels, keyboard-operable controls, and responsive styling.

## Backend architecture
Use a **pragmatic, simplified Clean Architecture**:

- **Gin Router:** registers endpoints and middleware.
- **Handler (Controller):** binds/validates HTTP input, calls the use case, maps errors to HTTP statuses and JSON responses.
- **Use Case:** orchestrates the requested calculation; stays independent of Gin.
- **Domain:** pure arithmetic logic and domain errors; stays independent of HTTP and JSON.
- **DTO:** request/response structs; a transport contract, not an extra execution layer.

Flow: **Gin Router → Handler → Use Case → Domain**.

Use constructor-based dependency injection when useful. Do not introduce a repository, database, or redundant service layer. Keep arithmetic errors explicit, particularly division by zero and invalid square roots if the optional operation is implemented.

## API contract
Start with one endpoint: `POST /api/calculate`.

Example request:
```json
{"operation":"add","a":10,"b":5}
```

Example success response:
```json
{"result":15}
```

Example error response:
```json
{"error":"division by zero"}
```

Use consistent operation identifiers (`add`, `subtract`, `multiply`, `divide`; extend only if needed). Document unary-operation behavior if square root is implemented. Return appropriate HTTP statuses (e.g., 400 for invalid input and division by zero, 200 for success). Ensure frontend and backend use the same contract. Treat the examples as the proposed design, not as an assignment-mandated API schema.

## Testing strategy
**Required: unit tests in both frontend and backend, plus a coverage report.**

- **Go unit:** standard `testing` plus optional `testify`; test arithmetic, edge cases, and use-case behavior.
- **Go integration:** `net/http/httptest` with Gin router and `testify` if helpful; test actual HTTP request/response behavior without a database.
- **React unit:** Vitest; test hook logic, utility behavior, and key component states.
- **React integration:** React Testing Library + MSW; test form → hook → HTTP service behavior with mocked network responses.
- **E2E (optional):** Playwright Test against the running React and Go applications; cover a successful calculation and an error path. Playwright MCP can assist interactive development but is not a replacement for committed, repeatable Playwright tests. Maestro is unnecessary for this web application.

Avoid duplicating every arithmetic case at every test level. Make tests deterministic. Generate coverage reports for both languages and document how to run them.

## Docker and CI
Docker is optional in the assignment but selected for this implementation. Use multi-stage Dockerfiles where appropriate and `compose.yaml` to run frontend and backend together. Configure the frontend API URL correctly for browser access; Docker service hostnames are not automatically resolvable from a user's browser. A Vite development proxy or reverse proxy is acceptable.

If time permits, add GitHub Actions to run Go and frontend tests, coverage generation, and builds. Do not let optional infrastructure delay the required deliverables.

## README and submission checklist
Document setup, prerequisites, running locally, running with Docker, tests/coverage, sample API requests, design choices, assumptions, and known limitations. Keep a record of AI prompts used: **the assignment explicitly requires sharing prompts**. Push the solution to a Git repository and submit its URL through the assignment page.

## Implementation rules for coding agents
1. Read the challenge requirements and this file before making architectural changes.
2. Implement the smallest working vertical slice first: React form → API → domain calculation → result.
3. Finish all four required operations and error handling before optional operations or tooling.
4. Keep code readable and idiomatic; avoid speculative abstractions and unused dependencies.
5. Write meaningful unit tests and produce coverage reports; add integration and E2E tests if time allows.
6. Run formatters, linters (if configured), tests, and builds; report what passed and what was not executed.
7. Update README and preserve the AI-prompt log for submission.
8. Do not claim tests pass unless they have actually been run.
