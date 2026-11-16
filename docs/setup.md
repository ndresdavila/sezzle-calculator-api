# Setup

Go 1.23 or newer. The module is `github.com/ndresdavila/sezzle-calculator-api`.

## Local

```bash
go run ./cmd/api
```

| Variable | Default | Meaning |
| --- | --- | --- |
| `PORT` | `8080` | Port to bind. A leading colon is stripped. |
| `CORS_ORIGINS` | `*` | Value of `Access-Control-Allow-Origin` |

The development UI proxies `/api` to this port, so the default is enough for local work.

## Binary

```bash
make build
./bin/api
```

## Docker

```bash
docker compose up --build
```

The image listens on 8080 and exposes `GET /health` as its health check. The process runs as an unprivileged user.

## Full stack

From a checkout of `sezzle-calculator-web` sitting next to this repository:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml up --build
```

Open `http://localhost:3000`. The UI container proxies `/api` to this service, so the browser stays on one origin.
