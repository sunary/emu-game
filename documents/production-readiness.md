# Production Readiness & Publishing Guide

## 1. Prerequisites
- Go 1.25+, Docker 24+, docker compose plugin, Redis 7.
- Environment variables (.env based on `.env.example`).
- CircleCI project hooked to GitHub repo (config in `.circleci/config.yml`).

## 2. Build & Release Pipeline
1. **CI**: CircleCI job installs Redis, runs `go test ./...`.
2. **Docker Image**:
   ```bash
   docker build -t emu-game-server:latest .
   docker run --rm -p 8080:8080 --env-file .env emu-game-server:latest
   ```
3. **Compose (local/staging)**:
   ```bash
   docker compose up --build
   ```
4. **Tag & push**:
   ```bash
   docker tag emu-game-server:latest ghcr.io/<org>/emu-game-server:v1
   docker push ghcr.io/<org>/emu-game-server:v1
   ```

## 3. Deployment Checklist
- [ ] TLS termination / reverse proxy (e.g., Traefik, nginx) configured.
- [ ] Secrets (JWT keys, Redis credentials) stored in secret manager.
- [ ] Health endpoint (`/health`) integrated with load balancer.
- [ ] Redis persistence (backup, monitoring) enabled.
- [ ] Logging/metrics shipping (CloudWatch, Loki, etc.).
- [ ] Websocket scaling plan (sticky sessions or shared pub/sub bus).

## 4. Security & Quality Gates
- JWT secrets rotated per environment.
- `go test ./...` and `docker compose up --build` on staging before release.
- Optional: `golangci-lint run`, `gosec ./...` for static analysis.
- Stress test before prod cutover:
  ```bash
  TOKEN=$(go run ./cmd/gen-token | awk -F"'" 'NR==2 {print $2}')
  TOTAL_REQUESTS=1000 CONCURRENCY=150 TOKEN="$TOKEN" ./scripts/stress-test.sh
  ```

## 5. Rollback Strategy
- Maintain previous Docker image tags.
- Redis snapshot before deployment.
- Roll back by re-deploying prior tag and restoring snapshot if necessary.

## 6. Reference Docs
- `README.md` (setup + integration)
- `documents/architecture.md`
- `documents/presentation.md`

## 7. Maintenance & Extension Recommendations
- **Ownership & Knowledge Transfer**: Ensure the next maintainer reviews `documents/architecture.md`, `documents/presentation.md`, and `documents/production-readiness.md` for full context.
- **Backlog Ideas**:
  1. **Quiz lifecycle**: add start/end times, multi-round support, and ability to rejoin finished quizzes.
  2. **Websocket payloads**: switch from plaintext to structured JSON with event types and metadata.
  3. **Leaderboard filtering**: support per-quiz leaderboards, time-bound rankings (daily/weekly), and pagination controls.
  4. **Analytics**: track submission/latency metrics, integrate with Prometheus/Grafana.
  5. **Auth enhancements**: plug into centralized identity provider, add role-based permissions (admin/moderator).
  6. **Infrastructure**: add automated deploy pipeline (Docker image publish, IaC templates for Redis + server), enable horizontal scaling with sticky sessions.
  7. **Testing**: expand integration tests (e.g., `docker-compose` based), add fuzz testing for join/submit payloads.
- **Onboarding tips**:
  - Run `go test ./...`, `docker compose up --build`, and the stress script before pushing changes.
  - Keep documentation synchronized—update README/docs when behavior changes.
  - Follow `.cursor/rules.md` to ensure AI/code assistants respect project conventions.
