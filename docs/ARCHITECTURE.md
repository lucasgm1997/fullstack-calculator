# Architecture

Detailed diagrams of how the calculator works. For setup and the API reference, see the [README](../README.md). The live OpenAPI docs are served at `/api/docs`.

- [1. System overview](#1-system-overview)
- [2. Request lifecycle: successful calculation](#2-request-lifecycle-successful-calculation)
- [3. Request lifecycle: error paths](#3-request-lifecycle-error-paths)
- [4. Backend steps inside `POST /api/calculate`](#4-backend-steps-inside-post-apicalculate)
- [5. Backend package dependencies](#5-backend-package-dependencies)
- [6. Frontend layers (MVVM-inspired)](#6-frontend-layers-mvvm-inspired)
- [7. Keypad state machine](#7-keypad-state-machine)
- [8. Chained operations](#8-chained-operations)
- [9. Deployment topologies](#9-deployment-topologies)

---

## 1. System overview

The browser only ever talks to **its own origin**. A proxy (Vite in development, nginx in Docker) forwards `/api/*` to the Go service. As a result, there is no CORS setup, and the browser never needs to resolve Docker hostnames.

```mermaid
flowchart LR
    user([User<br/>desktop or phone])

    subgraph browser[Browser]
        spa[React SPA<br/>Vite + TypeScript + Tailwind]
    end

    subgraph edge[Same origin]
        proxy{{Vite dev proxy :5173<br/>or nginx :3000}}
    end

    subgraph api[Go API :8080]
        gin[Gin router] --> handler[Handler] --> usecase[Use case] --> domain[Domain]
        gin --> docs[OpenAPI spec<br/>+ Swagger UI]
    end

    user -- clicks / keys --> spa
    spa -- "fetch /api/calculate (JSON)" --> proxy
    proxy -- "/api/*" --> gin
    proxy -- "static assets" --> spa
```

## 2. Request lifecycle: successful calculation

What happens when the user presses `1` `2` `÷` `4` `=`. The digit and operator keys only change local state. **Only `=` makes a network call.**

```mermaid
sequenceDiagram
    autonumber
    actor U as User
    participant V as View<br/>(CalculatorPage / Keypad)
    participant VM as useCalculator<br/>(ViewModel)
    participant S as calculatorService
    participant P as Proxy<br/>(Vite / nginx)
    participant R as Gin router
    participant H as Handler
    participant UC as Use case
    participant D as Domain

    U->>V: press 1, 2
    V->>VM: inputDigit("1"), inputDigit("2")
    Note over VM: entry = "12" (local state, no request)
    U->>V: press ÷
    V->>VM: chooseOperation("divide")
    Note over VM: accumulator = 12<br/>pendingOperation = divide<br/>display "12 ÷"
    U->>V: press 4, =
    V->>VM: inputDigit("4"), evaluate()
    VM->>VM: busy = true, isLoading = true
    VM->>S: calculate({operation:"divide", a:12, b:4})
    S->>P: POST /api/calculate
    P->>R: forward /api/*
    R->>H: Calculate(c)
    H->>H: limit body to 4 KB, bind JSON → DTO
    H->>UC: Execute(CalculateInput)
    UC->>UC: validate operation + operands
    UC->>D: Calculate(divide, 12, 4)
    D-->>UC: 3, nil
    UC-->>H: 3, nil
    H-->>P: 200 {"result":3}
    P-->>S: 200 {"result":3}
    S-->>VM: 3
    VM->>VM: history.unshift("12 ÷ 4 = 3")<br/>lastCall = {request, 200, response}
    VM-->>V: display "3", expression "12 ÷ 4 ="
    V-->>U: shows 3
```

## 3. Request lifecycle: error paths

Each layer handles the errors it knows about. The **service** turns transport failures into messages the user can read. The **handler** turns domain and use-case errors into HTTP statuses.

```mermaid
sequenceDiagram
    autonumber
    actor U as User
    participant VM as useCalculator
    participant S as calculatorService
    participant API as Go API

    rect rgba(255, 86, 103, 0.08)
    Note over U,API: Client-side validation (no request sent)
    U->>VM: 9, +, =
    VM-->>U: "Enter a second number"
    end

    rect rgba(255, 86, 103, 0.08)
    Note over U,API: Domain error (4xx): show API message
    U->>VM: 9, ÷, 0, =
    VM->>S: calculate({divide, 9, 0})
    S->>API: POST /api/calculate
    API-->>S: 400 {"error":"division by zero"}
    S-->>VM: throw CalculatorError("Division by zero", 400)
    VM-->>U: red display + "Division by zero"<br/>input kept so the user can fix b
    end

    rect rgba(255, 86, 103, 0.08)
    Note over U,API: Server error (5xx): hide details
    VM->>S: calculate(...)
    S->>API: POST /api/calculate
    API-->>S: 500 {"error":"internal server error"}
    S-->>VM: throw CalculatorError("The calculator service failed…")
    end

    rect rgba(255, 86, 103, 0.08)
    Note over U,API: Network failure: API down
    VM->>S: calculate(...)
    S-xAPI: fetch rejects
    S-->>VM: throw CalculatorError("Cannot reach the calculator service…")
    end
```

## 4. Backend steps inside `POST /api/calculate`

Every exit from this flowchart is a JSON response.

```mermaid
flowchart TD
    start([POST /api/calculate]) --> mw[Gin middleware<br/>Logger + Recovery]
    mw --> limit[Limit body to 4 KB]
    limit --> bind{Bind JSON<br/>into DTO}

    bind -- "body > 4 KB" --> e413[400 request body too large]
    bind -- "wrong field type<br/>or out of float64 range" --> etype["400 field &quot;a&quot; must be a finite number"]
    bind -- "invalid / empty JSON" --> ejson[400 request body must be valid JSON]
    bind -- ok --> uc[[Use case: Execute]]

    uc --> opvalid{operation<br/>known?}
    opvalid -- no --> eop["400 unsupported operation &quot;x&quot;"]
    opvalid -- yes --> hasA{a present?}
    hasA -- "no / null" --> emissA["400 missing operand: field &quot;a&quot;"]
    hasA -- yes --> unary{unary op?<br/>sqrt}
    unary -- yes --> dom
    unary -- no --> hasB{b present?}
    hasB -- no --> emissB["400 missing operand: field &quot;b&quot;"]
    hasB -- yes --> dom[[Domain: Calculate]]

    dom --> div0{divide and b = 0?}
    div0 -- yes --> ediv[400 division by zero]
    div0 -- no --> neg{sqrt and a < 0?}
    neg -- yes --> esqrt[400 square root of a negative number]
    neg -- no --> compute[compute result]
    compute --> finite{finite?}
    finite -- "Inf / NaN" --> enf[400 result is not a finite number]
    finite -- yes --> ok([200 result])

    classDef err fill:#ff566722,stroke:#ff5667,color:#ff5667
    classDef good fill:#9ff4d922,stroke:#5eead4
    class e413,etype,ejson,eop,emissA,emissB,ediv,esqrt,enf err
    class ok good
```

The handler maps errors with `errors.Is`, so the use case and domain stay independent of HTTP:

```mermaid
flowchart LR
    subgraph domain[domain]
        d1[ErrUnsupportedOperation]
        d2[ErrDivisionByZero]
        d3[ErrNegativeSqrt]
        d4[ErrNonFiniteResult]
    end
    subgraph usecase[usecase]
        u1[ErrMissingOperand]
    end
    other[any other error]

    d1 & d2 & d3 & d4 & u1 --> s400[400 + err.Error]
    other --> s500[500 internal server error<br/>real error only logged]
```

## 5. Backend package dependencies

Dependencies point inward. `domain` imports nothing from the project, and only `handler` and `router` know about Gin.

```mermaid
flowchart TD
    main[cmd/api<br/>main.go] --> router
    router[internal/router] --> handler
    router --> docs[internal/docs<br/>embedded openapi.yaml]
    router --> usecase
    handler[internal/handler] --> usecase
    handler --> dto[internal/dto]
    handler --> domain
    usecase[internal/usecase] --> domain[internal/domain<br/>pure Go, no deps]

    gin((gin)) -.-> router
    gin -.-> handler
    gin -.-> docs
```

| Layer | Knows HTTP? | Knows JSON? | Tested with |
| --- | --- | --- | --- |
| `domain` | no | no | table-driven unit tests |
| `usecase` | no | no | unit tests |
| `handler` | yes | yes | `httptest` + fake use case |
| `router` (full stack) | yes | yes | `httptest` integration tests in `tests/` |
| `docs` | – | – | contract test: spec enum = `domain.Operations` |

## 6. Frontend layers (MVVM-inspired)

```mermaid
flowchart TD
    subgraph View
        page[CalculatorPage<br/>composition + keyboard]
        display[CalculatorDisplay]
        keypad[CalculatorKeypad]
        inspector[ApiCallInspector]
        history[CalculationHistory]
    end

    subgraph ViewModel
        vm[useCalculator<br/>entry · accumulator · pendingOperation<br/>isLoading · error · history · lastCall]
        status[useApiStatus]
        utils[utils/entry.ts<br/>pure input editing + formatting]
    end

    subgraph Model / Data
        service[calculatorService<br/>calculate · checkHealth · error mapping]
        api[lib/api.ts<br/>requestJson · HttpError]
        types[types/calculator.types.ts<br/>API contract]
    end

    page --> display & keypad & inspector & history
    page --> vm & status
    vm --> utils
    vm --> service
    status --> service
    service --> api
    api -- fetch --> backend[(Go API)]
    types -.-> vm & service & keypad
```

## 7. Keypad state machine

This is the local state managed by `useCalculator`. Network calls happen only on the `chooseOperation` (chained or unary) and `evaluate` transitions.

```mermaid
stateDiagram-v2
    [*] --> Idle: entry = "0"

    Idle --> TypingA: digit / "."
    TypingA --> TypingA: digit / "." / ± / ⌫
    TypingA --> AwaitingB: binary operator<br/>accumulator = entry
    TypingA --> ShowingResult: √ (API call)

    AwaitingB --> AwaitingB: other operator<br/>(just switches it)
    AwaitingB --> AwaitingB: "=" → "Enter a second number"
    AwaitingB --> TypingB: digit / "."

    TypingB --> TypingB: digit / "." / ± / ⌫
    TypingB --> Loading: "=" or operator (chain)
    Loading --> ShowingResult: 200 (after "=")
    Loading --> AwaitingB: 200 (after operator)<br/>accumulator = result
    Loading --> TypingB: 4xx / 5xx / network<br/>error shown, input kept

    ShowingResult --> TypingA: digit (starts a new number)
    ShowingResult --> AwaitingB: operator<br/>(continue from result)

    Idle --> Idle: AC
    TypingA --> Idle: AC
    AwaitingB --> Idle: AC
    TypingB --> Idle: AC
    ShowingResult --> Idle: AC
```

Keys pressed during **Loading** are not dropped. They are queued and replayed in order once the response arrives, so typing `2+3*4⏎` at full speed still gives `20`. This is covered by an E2E test and a unit regression test.

## 8. Chained operations

The calculator evaluates **left to right as you type**, like a pocket calculator, and always delegates the math to the API.

```mermaid
sequenceDiagram
    actor U as User
    participant VM as useCalculator
    participant API as Go API

    U->>VM: 2, +
    Note over VM: accumulator=2, pending=add
    U->>VM: 3, ×
    VM->>API: {add, 2, 3}
    API-->>VM: 5
    Note over VM: accumulator=5, pending=multiply<br/>display "5 ×"
    U->>VM: 4, =
    VM->>API: {multiply, 5, 4}
    API-->>VM: 20
    Note over VM: display "20", expression "5 × 4 ="
```

## 9. Deployment topologies

```mermaid
flowchart LR
    subgraph local["Local development"]
        direction LR
        b1([Browser / phone]) -- ":5173" --> vite[Vite dev server<br/>HMR + proxy]
        vite -- "/api/* → :8080" --> go1[go run ./cmd/api]
    end

    subgraph docker["docker compose up"]
        direction LR
        b2([Browser / phone]) -- ":3000" --> nginx[frontend container<br/>nginx + static build]
        nginx -- "/api/* → backend:8080<br/>(Docker network)" --> go2[backend container<br/>distroless, non-root]
        b2 -. ":8080 (optional, direct)" .-> go2
    end
```

Both setups listen on `0.0.0.0`, so a phone on the same Wi-Fi can open `http://<mac-ip>:5173` or `http://<mac-ip>:3000`.
