# API

Base URL: `http://localhost:8080`

Results are decimal strings. That keeps `0.1 + 0.2` equal to `0.3` on the wire.

## Health

`GET /health`

```json
{"status":"ok"}
```

## Catalog

`GET /api/v1/operations`

```json
{
  "operations": [
    {"name":"add","arity":2,"description":"Adds two numbers"},
    {"name":"sqrt","arity":1,"description":"Returns the square root of a non-negative number"}
  ]
}
```

The catalog is the operation registry. A new operation shows up here without a handler change.

## Single operation

`POST /api/v1/calculate`

Operands may be JSON strings or JSON numbers. Strings are preferred.

```bash
curl -s http://localhost:8080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","operands":["0.1","0.2"]}'
```

```json
{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}
```

| Operation | Arity | Example result |
| --- | --- | --- |
| `add` | 2 | `0.1 + 0.2 = 0.3` |
| `subtract` | 2 | `5 - 2 = 3` |
| `multiply` | 2 | `74 * 24 = 1776` |
| `divide` | 2 | `7 / 2 = 3.5` |
| `power` | 2 | `2 ^ 10 = 1024` |
| `sqrt` | 1 | `sqrt(9) = 3` |
| `percent` | 1 | `50% = 0.5` (absolute; the relative form is only in expressions) |
| `negate` | 1 | `4` becomes `-4` |

## Expression

`POST /api/v1/evaluate`

```bash
curl -s http://localhost:8080/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"346+74*24"}'
```

```json
{"expression":"346+74*24","result":"2122"}
```

Supported forms:

- `+`, `-`, `*`, `/`, and the display symbols `−`, `×`, `÷`
- `^`, right associative, so `2^3^2` is `512`
- postfix `%`
- `sqrt(...)` and parentheses
- a leading `+` or `-`

Precedence, tightest first: percent and unary sign, then power, then multiply and divide, then add and subtract.

Percent rules:

| Expression | Result | Why |
| --- | --- | --- |
| `50%` | `0.5` | A bare percent divides by 100 |
| `200+10%` | `220` | 10% of the left side |
| `200-10%` | `180` | Same rule for subtraction |
| `200*10%` | `20` | Percent becomes 0.1 before multiply |
| `200/10%` | `2000` | Percent becomes 0.1 before divide |
| `200+(10%)` | `200.1` | Parentheses resolve the percent first |

Whitespace is ignored. The expression may be at most 500 characters.

## Errors

```json
{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero"}}
```

| HTTP | Code | When |
| --- | --- | --- |
| 400 | `INVALID_JSON` | Malformed JSON, unknown field, or a second JSON value |
| 400 | `INVALID_NUMBER` | An operand is not a decimal |
| 400 | `INVALID_OPERAND_COUNT` | Arity does not match the operation |
| 400 | `UNKNOWN_OPERATION` | Name is not in the registry |
| 400 | `INVALID_EXPRESSION` | The expression does not parse |
| 422 | `DIVISION_BY_ZERO` | Divisor is zero, including `10/0%` |
| 422 | `NEGATIVE_SQRT` | Square root of a negative number |
| 422 | `INVALID_POWER` | `0^0`, or a negative base with a fractional exponent |
| 422 | `OVERFLOW` | Exponent or result is outside the calculator range |
| 408 | `TIMEOUT` | The request context is already canceled |

Unknown fields are rejected so a typo such as `"expresion"` fails loudly.
