# Sezzle Full-Stack Calculator — Implementation Priorities

This checklist separates explicit assignment requirements from optional features and engineering enhancements. **Complete P1 before spending time on P2 or P3.** The assignment recommends approximately **2–4 hours** and prioritizes correctness, clarity, and maintainability.

## P1 — Required (assignment deliverables)

- [ ] **React frontend:** Build the calculator UI in React; TypeScript is preferred, not mandatory. Vite is an acceptable build tool.
- [ ] **Usability:** Provide intuitive inputs, operation selection, result display, input validation, error feedback, and basic mobile responsiveness.
- [ ] **Core operations:** Addition, subtraction, multiplication, and division.
- [ ] **Backend REST API:** Implement calculator endpoints returning JSON. Go is preferred, not mandatory; use Gin as the chosen router.
- [ ] **End-to-end functionality:** Have the React frontend call the actual backend API for calculations.
- [ ] **Edge cases:** Validate malformed or missing inputs, unsupported operations, and division by zero; return appropriate HTTP errors.
- [ ] **Unit tests:** Cover key frontend and backend functionality; generate a coverage report.
- [ ] **README:** Include setup instructions, how to run both services, example REST requests/responses, and design decisions or assumptions.
- [ ] **Submission:** Push the source code to a Git repository and share the repository link through the assignment page.
- [ ] **AI usage disclosure:** Share the prompts used during development, as explicitly requested.

## P2 — Explicitly optional in the assignment

- [ ] **Advanced operations:** Exponentiation, square root, and percentage.
- [ ] **Containerization:** Dockerfile(s) for full-stack deployment; Docker Compose is a practical additional choice.

## P3 — Engineering extras (not explicitly required)

- [ ] **Frontend organization:** Feature-based folders; lightweight MVVM-inspired separation using React components, `useCalculator` hook, typed contracts, and HTTP service. Prefer local `useState` over Zustand unless needed.
- [ ] **Backend organization:** Lightweight layered/Clean Architecture-inspired design: Gin router → handler → use case → domain, plus request/response DTOs. Constructor injection; no repository without persistence.
- [ ] **Go integration tests:** `net/http/httptest` with `testing` and optionally Testify to exercise the Gin API stack.
- [ ] **React integration tests:** Vitest + React Testing Library + MSW to test component/hook/service interactions with mocked HTTP.
- [ ] **Browser E2E tests:** Playwright Test against running frontend and backend; cover one successful calculation and one error scenario. Playwright MCP is an agent interaction tool, not a replacement for versioned Playwright Test cases.
- [ ] **CI:** GitHub Actions for unit tests, coverage, and builds.
- [ ] **Developer tooling:** Consistent linting, formatting, and error response conventions.

## Suggested implementation order

1. Implement core Go calculations, validation, REST endpoint, and backend unit tests.
2. Implement React calculator UI, API client, error handling, responsiveness, and frontend unit tests.
3. Connect both services; verify core operations and edge cases manually.
4. Produce test coverage reports and complete the README and AI prompt disclosure.
5. Push and verify the repository and submission materials.
6. Only if time remains: Docker, advanced operations, integration tests, Playwright E2E, and CI.

## Proposed technical stack (implementation decisions, not assignment requirements)

| Area | Choice |
| --- | --- |
| Frontend | React + TypeScript + Vite |
| Backend | Go + Gin |
| Frontend state | React `useState` / custom hook |
| HTTP | Native `fetch` |
| Backend unit tests | Go `testing` + optional Testify |
| Frontend unit tests | Vitest + React Testing Library |
| Integration | Go `httptest`; React MSW |
| E2E | Playwright Test |
| Deployment | Docker / Docker Compose |
| CI | GitHub Actions |

> **Guiding rule:** Keep the architecture proportional to a small calculator. Do not delay required deliverables to implement optional abstractions, tooling, or features.
