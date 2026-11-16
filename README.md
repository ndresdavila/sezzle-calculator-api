# Sezzle calculator API

REST service for exact decimal arithmetic. The React client lives in [sezzle-calculator-web](https://github.com/ndresdavila/sezzle-calculator-web).

The sample from the design reference:

```text
346 + 74 × 24 = 2122
```

## Quick start

```bash
go run ./cmd/api
```

The server listens on `:8080`. Change it with `PORT`. `CORS_ORIGINS` defaults to `*`.

```bash
curl http://localhost:8080/health
curl -s http://localhost:8080/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"346+74*24"}'
```

```json
{"expression":"346+74*24","result":"2122"}
```

## Tests

```bash
make test
make cover
```

`make cover` writes `coverage/summary.txt` and prints each function. The latest local run is recorded in [docs/testing.md](docs/testing.md).

## Docker

```bash
docker compose up --build
```

To run the UI and this API together, use the compose file in the web repository.

## Layout

| Path | Role |
| --- | --- |
| `internal/domain` | Arithmetic, expression grammar, domain errors |
| `internal/application` | Use cases and the inbound port |
| `internal/adapter/http` | JSON translation and HTTP status codes |
| `internal/platform/config` | Environment configuration |
| `cmd/api` | Composition root |
| `docs/` | Architecture, API, setup, tests, decisions |
| `prompts/` | The prompt used to build this service |

## Design in one page

- Hexagonal: the domain does not import HTTP, and a new operation is one registry entry.
- Money-safe decimals (`shopspring/decimal`), so `0.1 + 0.2` is `0.3`.
- Results are JSON strings. JSON numbers are accepted on input, but they are read from the raw literal so they are not parsed as floats first.
- `200+10%` means 220. A bare `50%` means 0.5. Parentheses force the percent to resolve immediately.
- Division by zero, a negative square root, and an undefined power return `422` with a stable `error.code`.

Details are in [docs/design-decisions.md](docs/design-decisions.md) and [docs/api.md](docs/api.md).
