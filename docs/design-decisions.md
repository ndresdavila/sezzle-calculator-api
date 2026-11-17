# Design decisions

## Decimal strings, not floats

`0.1 + 0.2` is the bug a calculator cannot have. The domain uses `shopspring/decimal`, and JSON results are strings. When a client sends a JSON number, the handler parses the raw literal (`0.1`) instead of a decoded `float64`.

Finished results are rounded to 12 decimal places. A value within `1e-9` of an integer snaps to that integer, which is why `(1/3)*3` is `1` and `sqrt(4)` is `2`.

## One registry for operations

`Apply` does not have a special case in the HTTP layer. The registry holds the name, the arity, the description, and the function. `GET /operations` and `POST /calculate` both read it. Adding modulo later is a domain change with tests, not a new route.

The expression parser calls those same functions for add, subtract, multiply, and divide, so the two entry points cannot drift.

## Percent follows a hand calculator

The reference keypad has a percent key, so the expression language uses the rule people expect from that key:

- next to `+` or `-`, the percent is relative to the other side
- next to `*` or `/`, or on its own, it divides by 100
- parentheses resolve first, so `200+(10%)` is `200.1` and not `220`

`POST /calculate` with `"operation":"percent"` is the absolute form only. The UI sends the expression and lets `evaluate` apply the relative rule. That split is documented in the web repository.

## Power and square root

Integer powers use the decimal library's exact path. Fractional powers are rounded to 16 places and then normalized. A negative base with a fractional exponent is rejected rather than returning a complex number. Square root uses Newton iteration in decimal, so `sqrt(2)` is not a binary float.

The exponent is capped at 256 and the result magnitude at 60 digits. A value past that size, positive or negative, is `OVERFLOW`. A non-zero value closer to zero than that, positive or negative, is `UNDERFLOW`. A huge negative exponent such as `2^-999` is underflow, not a giant fraction. Those limits keep a public endpoint from building a gigantic coefficient.

## Errors are codes

Handlers never match on English sentences. The domain returns `*domain.Error` with a code, and the adapter picks 400 for bad input and 422 for a mathematically defined refusal. The message is safe to show in the UI.

## What was left out

There is no calculation history, no accounts, and no persisted tape. The port already accepts `context.Context`, so those features can be added beside the domain instead of inside the parser.
