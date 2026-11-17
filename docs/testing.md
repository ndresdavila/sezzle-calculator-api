# Testing

Tests are table-driven and stay inside the package they cover. HTTP tests boot the real service behind `httptest`. There is no mock of the arithmetic: a wrong total should fail the test.

## Run

```bash
make test
```

That is `go test ./...`.

Coverage, including a function report:

```bash
make cover
```

The command writes:

- `coverage/coverage.out` for `go tool cover -html=coverage/coverage.out`
- `coverage/summary.txt` for the text report

Both files are gitignored. Regenerate them rather than trusting a committed snapshot.

## Latest local report

| Package | Coverage |
| --- | --- |
| `internal/domain` | 92.2% |
| `internal/application/service` | 100% |
| `internal/adapter/http` | 90.1% |
| `internal/platform/config` | 100% |

`cmd/api` is the composition root and is not unit tested. `go test ./... -coverprofile` reported 89.1% across the whole module, including that package.

## What is locked in

- `346+74*24 = 2122`, `0.1+0.2 = 0.3`, and `(1/3)*3 = 1`
- Relative percent, parentheses, power associativity, and square root
- Division by zero, negative roots, `0^0`, and rejected JSON
- `10^70` and `(-10)^70` are `OVERFLOW`. `10^-70` and `(-10)^-70` are `UNDERFLOW`. The HTTP adapter returns both as 422
- Every catalog operation can be executed with valid operands
- A fuzz seed calls `Evaluate` with arbitrary strings and expects no panic

`go test` runs the fuzz corpus once. A longer run is:

```bash
go test ./internal/domain -fuzz=FuzzEvaluate -fuzztime=10s
```
