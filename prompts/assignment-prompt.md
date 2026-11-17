# Prompt: calculator API

## Role

You are a senior Go engineer building a small production-style service for a technical evaluation. Favor a design another engineer can extend in a year. Do not add features that the requirements do not ask for.

## Objective

Build a stateless REST calculator microservice in Go. It performs basic and advanced arithmetic, validates input, and returns JSON. The React client in the sibling repository is the only caller that must work on day one, but the API must also be usable with curl.

## Goal

A reviewer can clone the repository, run `make test`, start the server, and evaluate `346+74*24` to `2122`. The same reviewer can add a new arithmetic operation by editing the domain registry and a test, without touching the router.

## Context

- This is the backend half of a Sezzle full-stack calculator exercise.
- The UI shows a running expression and sends it when the user presses equals.
- The reference expression on the design is `346+74*24 = 2122`, which requires operator precedence.
- The sibling UI repository is `sezzle-calculator-web`.

## Requirements

### Functional

- Operations: add, subtract, multiply, divide.
- Also support power, square root, percent, and negate.
- `POST /api/v1/calculate` runs one named operation.
- `POST /api/v1/evaluate` runs an expression with precedence, parentheses, postfix `%`, and `sqrt()`.
- `GET /api/v1/operations` lists the registry.
- `GET /health` returns `{"status":"ok"}`.
- Reject division by zero, negative square roots, undefined powers, overflow, underflow, malformed JSON, unknown fields, unknown operations, and the wrong number of operands.
- Overflow means the result is too large, positive or negative. Underflow means it is too close to zero, positive or negative.
- Return every numeric result as a decimal string.

### Non-functional

- Hexagonal layout: domain, application port and service, HTTP adapter, composition root.
- The domain must not import `net/http`.
- Exact decimals. `0.1 + 0.2` must be `0.3`.
- Idiomatic Go, `gofmt`, standard-library HTTP.
- Unit tests for the domain, the service, and the HTTP adapter, plus a coverage command.
- Human-readable Markdown in `docs/`: setup, API, architecture, testing, design decisions.
- A Dockerfile and a compose file that run the API.

## Constraints

- Go 1.23 or newer.
- No database and no authentication.
- Expression length at most 500 characters.
- Cap exponent size and result magnitude so a request cannot allocate a huge coefficient.
- Percent next to `+` or `-` is relative (`200+10% = 220`). A bare percent divides by 100. Parentheses resolve first.
- Do not invent a user account, a history log, or a second transport.

## Architecture

```text
cmd/api                         wires config, service, and router
internal/domain                 operations, parser, errors, normalization
internal/application/port       Calculator interface
internal/application/service    use cases, context checks
internal/adapter/http           JSON, status codes, middleware
internal/platform/config        PORT, CORS_ORIGINS
```

A new binary operation is one registry entry with a name, an arity, a description, and a function. The expression parser must call those functions instead of reimplementing addition and division.

## Testing

- Table tests for operations and expressions, including the reference total 2122, overflow, and underflow for both signs.
- HTTP tests with `httptest` and the real service, not a fake calculator.
- Assert error codes, not only messages.
- Provide `make test` and `make cover`.
- Document the commands and the latest coverage in `docs/testing.md`.

## Acceptance criteria

- `go test ./...` passes.
- `curl` examples in the README return the documented JSON.
- Division by zero is HTTP 422 with code `DIVISION_BY_ZERO`.
- Invalid JSON is HTTP 400.
- `docker compose up --build` serves `/health`.
- README links to `docs/` and to the frontend repository.

## Deliverables

- Public GitHub repository `ndresdavila/sezzle-calculator-api`.
- README with setup, API examples, and design decisions.
- `docs/` and this prompt file.
- Dockerfile, compose file, Makefile, and a GitHub Actions test workflow.

## Out of scope

- GraphQL, gRPC, or a persistence adapter.
- User interface code.
- Rewriting percent to match every vendor's calculator. Document the chosen rule and test it.
