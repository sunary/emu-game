## Emu Game Server

Lightweight Go game server that exposes HTTP and WebSocket endpoints for joining quizzes, submitting scores, and broadcasting events to connected clients. Redis backs the leaderboard, and configuration is handled through embedded defaults plus environment overrides.

### Table of Contents
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [Key Endpoints](#key-endpoints)
- [Testing](#testing)
- [Running Locally](#running-locally)
- [Load / Stress Testing](#load--stress-testing)
- [Documents](#documents)
- [Acknowledgements](#acknowledgements)

### Prerequisites

- Go 1.25+
- Redis instance (local or remote). Tests use an in-memory Redis emulator and do not need a running server.

### Configuration

Default settings live in `configs/default.yaml`. Override any value via environment variables (see `configs/config.go`) or by editing the YAML file. Common options:

- `SERVER__ADDR` – HTTP listen address (default `:8080`)
- `REDIS__ADDR` – Redis address (default `localhost:6379`)

### Running the Server

```bash
go run ./cmd/server
```

To generate a JWT - `user-id: alice` (plus ready-to-run curl/websocket commands) use the helper CLI:

```bash
go run ./cmd/gen-token -sub alice

# for full curl
go run ./cmd/gen-token -sub alice -quiz quiz-42 -score 150 -curl
```

### Key Endpoints

| Method | Path                     | Description                            |
|--------|--------------------------|----------------------------------------|
| POST   | `/user/quiz/{id}/join`   | Join a quiz. Body: `{}` |
| POST   | `/user/quiz/{id}/submit` | Submit a quiz score. Body: `{"score":42}` |
| POST   | `/leaderboard`           | Fetch leaderboard segment. Body: `{"from":0,"limit":10}` |
| GET    | `/ws`                    | WebSocket for broadcast events         |

All `/user/*` routes require a valid `Authorization: Bearer <token>` header containing a signed JWT with the configured secret.

### Testing

```bash
go test ./...
```

Unit tests live under `internal/repositories` and cover join/submission/leaderboard behaviors using an in-memory Redis emulator.

### Running Locally
1. **Install dependencies**: Go ≥1.25 and Redis (e.g., `brew install redis && redis-server --daemonize yes`).
2. **Configure (optional)**: Override values via env vars, e.g. `export SERVER__ADDR=":9000"` or `export REDIS__ADDR="localhost:6379"`.
3. **Start the server**: `go run ./cmd/server`.
4. **Generate a JWT**: `go run ./cmd/gen-token -curl` (prints a token plus sample curl/websocket commands). Save `TOKEN='<value>'`.
5. **Exercise endpoints**: Use the provided curl snippets to join/submit or hit `/leaderboard`. Keep the websocket open with `wscat -c ws://localhost:8080/ws`.

### Load / Stress Testing

Use the bundled script to simulate up to 1,000 distinct users (defaults can be overridden via env vars such as `TOTAL_REQUESTS`, `CONCURRENCY`, `BASE_SCORE`, etc.). Every iteration generates a fresh random user id (`stress-<hex>`) and quiz id (`quiz-<hex>`), then performs a join and submit with a unique score (`BASE_SCORE + iteration`). Provide a valid JWT via `TOKEN`:

```bash
TOTAL_REQUESTS=1000 CONCURRENCY=100 ./scripts/stress-test.sh
```

The script issues `TOTAL_REQUESTS` independent join/submit flows using those randomized identifiers, then reports:
- Per-request latency samples (CSV in `/tmp/stress_latency.csv`)
- Websocket event counts (submissions seen on `/ws`)
- The final leaderboard slice (`LEADERBOARD_FROM`/`LEADERBOARD_LIMIT`)
- Start/end timestamps and total duration

**Stress script tips**:
```bash
# Ensure redis-cli can connect
redis-cli -h localhost -p 6379 ping

# Run the load
TOTAL_REQUESTS=1000 CONCURRENCY=150 ./scripts/stress-test.sh
```

### Documents
- [`documents/architecture.md`](documents/architecture.md): architecture diagram, components, data flow, tech stack.
- [`documents/presentation.md`](documents/presentation.md): presentation playbook (overview, AI collaboration, demo script).
- [`documents/production-readiness.md`](documents/production-readiness.md): deployment checklist, pipelines, security gates.

### Acknowledgements
Thanks to Cursor for accelerating ideation/coding during this project while maintaining tests and documentation along the way.

