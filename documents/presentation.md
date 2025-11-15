# Presentation Playbook

## 1. Assignment Overview
- Goal: Build a Go-based game server that manages quiz participation, maintains a leaderboard, and broadcasts quiz submissions in real time.
- Scope: REST endpoints for join/submit/leaderboard, Redis persistence, websocket events, stress tooling, documentation.

## 2. Solution Overview
- Architecture: Go HTTP + websocket server (`cmd/server`), repository layer wrapping Redis sorted sets, JWT authentication, stress/load tooling, comprehensive README and docs.
- Key highlights:
  - Join/submit REST flow with single-quiz enforcement.
  - Sorted-set leaderboard with websocket fan-out and heartbeat pings.
  - Load script simulating 1000 random users + metrics (duration, events, leaderboard).
  - Architecture document (`documents/architecture.md`) and README guidance.

## 3. AI Collaboration Story (Required)
### Partners & Tasks
- Collaborated with GenAI (e.g., ChatGPT) for brainstorming architecture doc structure, websocket heartbeat patterns, and README wording.
- Used AI suggestions when drafting stress-test enhancements (latency logging, websocket counting) and clarifying documentation language.

### How AI Helped
- Accelerated ideation: quick mermaid diagram layout and component descriptions.
- Provided code snippets for websocket ping/pong patterns and bash scripting hints.
- Assisted with wording for integration notes and justification for Redis usage.

### Challenges / Limitations
- AI occasionally suggested non-compilable Go snippets (e.g., incorrect imports). Required manual verification.
- Some bash suggestions needed adaptation to local env (permissions, path issues).

### Quality & Verification Process
1. **Code review**: Inspected AI-proposed code for style/logic before applying.
2. **Formatting**: Ran `gofmt` after Go changes.
3. **Testing**: `go test ./...` after each major change; stress-test script manually run to validate behavior.
4. **Runtime verification**: For websocket changes, ensured heartbeat logic builds and updated README to educate clients.
5. **Security**: Reviewed auth middleware to ensure JWT context is set and validated.

## 4. Demo Script
1. **Start Redis & server**:
   ```bash
   redis-server --daemonize yes
   go run ./cmd/server
   ```
2. **Generate token & sample commands**:
   ```bash
   TOKEN=$(go run ./cmd/gen-token | awk -F"'" 'NR==2 {print $2}')
   ```
3. **Join & submit via curl**:
   ```bash
   curl -i -X POST http://localhost:8080/user/quiz/demo/join \
     -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{}'

   curl -i -X POST http://localhost:8080/user/quiz/demo/submit \
     -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"score":123}'
   ```
4. **Watch websocket**:
   ```bash
   wscat -c ws://localhost:8080/ws
   ```
5. **Fetch leaderboard**:
   ```bash
   curl -s -X POST http://localhost:8080/leaderboard -H "Content-Type: application/json" -d '{"from":0,"limit":10}' | jq
   ```
6. **Run stress test**:
   ```bash
   TOTAL_REQUESTS=1000 CONCURRENCY=150 ./scripts/stress-test.sh
   ```

## 5. Conclusion
- **Learnings**: Websocket heartbeat importance, Redis sorted-set strengths, value of clear integration docs.
- **Challenges**: Managing JWT context propagation, keeping stress tooling deterministic, documenting architecture succinctly.
- **Future ideas**: Add quiz lifecycle (start/end), user dashboards, automated leaderboard snapshots, richer websocket payloads (JSON instead of plaintext).

## 6. Command & AI Collaboration Summary
The challenge states: *"This challenge assesses your core technical skills and ability to strategically leverage AI to enhance productivity and solution quality. You must demonstrate and document your AI usage throughout the design and implementation process; this is a key evaluation criterion."* Below is the summarized command/request log that drove the build, showing how each user directive was fulfilled with responsible AI assistance:

1. **Project bootstrap** – init Go module, HTTP + websocket server, restructure to standard layout.
2. **Redis integration** – add repository layer, quiz join/submit APIs, config via Viper.
3. **Websocket authentication & JWT validation** – external package, middleware wiring, docs updates.
4. **Leaderboard & quiz logic changes** – single-quiz enforcement, path updates, stress script evolutions.
5. **Documentation suite** – README expansions, architecture doc, presentation script, production-readiness guide.
6. **Tooling & automation** – stress-test script, Dockerfile, docker-compose, CircleCI config, .env template, .gitignore.

Each major request triggered AI-guided iteration (see Section 3 for specifics). After receiving AI suggestions, I reviewed, formatted, and tested (`go test ./...`, manual curl/ws checks, stress runs) to ensure correctness and security before moving on.
