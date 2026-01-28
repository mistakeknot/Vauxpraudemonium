# Plan: Autarch Coordination Infrastructure (Gas Town Response)

> **Strategy**: Position Autarch as the shared services layer that makes any agent colony smarter. Not an orchestrator — a platform.
>
> **Context**: Steve Yegge's Gas Town (Jan 2026) popularized multi-agent coding orchestration. Community debate suggests stronger models may make thick orchestrators obsolete (like LangChain). Autarch's durable value is *domain intelligence* (research, specs, signals) — not agent spawning logic.
>
> **Brainstorm**: [docs/brainstorms/2026-01-27-multi-agent-strategy-brainstorm.md](../brainstorms/2026-01-27-multi-agent-strategy-brainstorm.md)

---

## Phase 0: Documentation Updates

Update existing docs to reflect the coordination infrastructure strategy and planned API surfaces.

### CLAUDE.md

Add to Quick Commands:
```bash
# API servers (local-only by default)
go run ./cmd/pollard serve --addr 127.0.0.1:8090   # Pollard research API
go run ./cmd/gurgeh serve --addr 127.0.0.1:8091    # Gurgeh spec API (read-only)
go run ./cmd/signals serve --addr 127.0.0.1:8092   # Signals WS server
```

Add to Key Paths:
| `internal/pollard/server/` | Pollard HTTP API server |
| `internal/gurgeh/server/` | Gurgeh Spec API server |
| `internal/bigend/colony/` | Agent colony detection |
| `pkg/signals/broker.go` | Signal broadcast (WebSocket) |
| `internal/signals/` | Signals CLI + server wiring |

Add to Design Decisions:
- Coordination infrastructure strategy: Autarch provides shared services (research, specs, signals) to local agent colonies — not an orchestrator
- Local-only by default: bind to loopback; remote/multi-host support deferred
- Pollard API defaults to 127.0.0.1:8090 (local-only; non-loopback requires explicit opt-in + auth)
- Gurgeh Spec API is read-only (local-only)
- Signals broadcast via WebSocket to local subscribers (standalone server; remote subscribers deferred)

### AGENTS.md

Add to Project Status → TODO:
- Pollard HTTP API server (expose research to external agents)
- Gurgeh read-only Spec API (agents query acceptance criteria, CUJs)
- Signal broadcast via WebSocket (external agent colony subscriptions)
- Bigend colony detection (Git worktrees, agent processes)
- Intermute request/response pattern (async agent-to-Pollard queries)

Add to Documentation Map:
| [docs/VISION.md](docs/VISION.md) | Strategic vision and coordination infrastructure |
| [docs/brainstorms/](docs/brainstorms/) | Design brainstorms |

### docs/FLOWS.md

Add new **Section 17: Coordination Infrastructure (API Surfaces)** after Section 16. Add colony detection note to Section 5 and signal broadcast note to Section 12. See plan file for full content.

---

## Phase 1: Pollard HTTP API Server (highest value)

Wrap existing `Scanner` and `ResearchOrchestrator` in a thin `net/http` server.
Local-only by default (bind to loopback unless explicitly overridden).
Use async jobs for scans/research; responses use a standard envelope with typed errors.

**Create:**
- `internal/pollard/server/server.go` — HTTP server with routes:
  - `GET /health`
  - `POST /api/scan` → enqueue `Scanner.Scan()` job
  - `POST /api/scan/targeted` → enqueue `Scanner.RunTargetedScan()` job
  - `POST /api/research` → enqueue `ResearchOrchestrator.Research()` job
  - `GET /api/jobs/{id}` → job status
  - `GET /api/jobs/{id}/result` → job result (if complete)
  - `GET /api/insights` — cached insights from `.pollard/insights/`
  - `GET /api/hunters` — list available hunters
- `internal/pollard/server/cache.go` — TTL + size-bounded LRU + singleflight de-dup. Key = hash of scan options (5m quick, 15m balanced, 30m deep).
- `internal/pollard/server/jobs.go` — in-memory job store (status, timestamps, result/error)
- `internal/pollard/server/responses.go` — standard API envelope + typed errors
- `internal/pollard/cli/serve.go` — `pollard serve --addr :8090` cobra subcommand

**Modify:**
- `internal/pollard/cli/root.go` — add `serveCmd` to `init()`

**Reference pattern:** `internal/bigend/daemon/server.go` (existing net/http setup)

## Phase 2: Gurgeh Spec API (high value)

Read-only HTTP endpoints over `.gurgeh/specs/*.yaml` files.
Local-only by default (bind to loopback unless explicitly overridden).
Use standard response envelope + typed errors; add basic pagination to list endpoints.

**Create:**
- `internal/gurgeh/server/server.go` — HTTP server with routes:
  - `GET /api/specs` — list specs (summary, paginated)
  - `GET /api/specs/{id}` — full spec as JSON
  - `GET /api/specs/{id}/requirements`
  - `GET /api/specs/{id}/cujs`
  - `GET /api/specs/{id}/hypotheses`
  - `GET /api/specs/{id}/history` — version history
- `internal/gurgeh/server/responses.go` — standard API envelope + typed errors
- `internal/gurgeh/cli/serve.go` — `gurgeh serve --addr :8091`

**Modify:**
- `internal/gurgeh/cli/root.go` — add `serveCmd`

Uses existing `specs.LoadSummaries` and `specs.LoadSpec`. No caching needed (small YAML files).

## Phase 3: Bigend Colony Detection (medium value, parallelizable)

Detect external agent colonies and display in dashboard.
Cross-platform baseline: git worktrees + convention markers; Linux-only `/proc` scan behind OS guard.

**Create:**
- `internal/bigend/colony/detector.go` — scans for:
  - Git worktrees (`git worktree list --porcelain`)
  - Agent processes (`/proc/PID/cwd` mapping)
  - Convention markers (`.colony/`, `.agents/` dirs)
- `internal/bigend/colony/types.go` — `Colony`, `ColonyMember` types

**Modify:**
- `internal/bigend/aggregator/aggregator.go` — add `Colonies []Colony` to `State`, call `loadColonies()` from `Refresh()`

## Phase 4: Signals Broadcast (medium-high value, standalone)

WebSocket broadcast of signals to local subscribers (remote deferred).
Run as a standalone server (`signals serve`) to avoid coupling to Pollard/Gurgeh lifecycles.

**Create:**
- `pkg/signals/broker.go` — fan-out broker with `Subscribe(types)`, `Publish(signal)`, `ServeWS()`
- `pkg/signals/server.go` — HTTP+WS server wrapping the broker
- `internal/signals/cli/root.go` — cobra root for `signals`
- `internal/signals/cli/serve.go` — `signals serve --addr :8092`
- `cmd/signals/main.go` — entry point for `signals` CLI

**Modify:**
- `internal/gurgeh/arbiter/orchestrator.go` — optional `broker.Publish()` when signals raised
- `internal/pollard/watch/` — publish `SignalCompetitorShipped` to broker

## Phase 5: Intermute Request/Response (lowest priority)

Async agent-to-Pollard research queries via Intermute messaging.

**Create:**
- `internal/pollard/inbox/handler.go` — inbox polling loop, parses `research:` messages, replies with results
- `internal/pollard/inbox/protocol.go` — `ResearchRequest`/`ResearchResponse` types

**Modify:**
- `internal/pollard/cli/serve.go` — start inbox handler alongside HTTP server when Intermute available

---

## Build Order

```
Phase 0 (Docs)           ── do first
Phase 1 (Pollard API)  ─┐
Phase 2 (Gurgeh API)   ─┼── can build in parallel
Phase 3 (Colony detect) ─┘
Phase 4 (Signals)       ── independent (standalone server)
Phase 5 (Intermute)     ── after Phase 1 (reuses Scanner)
```

## Verification

1. `go build ./cmd/pollard && go run ./cmd/pollard serve --addr 127.0.0.1:8090` — server starts on 127.0.0.1:8090
2. `curl localhost:8090/health` — returns 200
3. `curl -X POST localhost:8090/api/scan -d '{"hunters":["github-scout"],"mode":"quick"}'` — returns job id
4. `curl localhost:8090/api/jobs/<id>` — returns status
5. `curl localhost:8090/api/jobs/<id>/result` — returns results when complete
6. `go build ./cmd/gurgeh && go run ./cmd/gurgeh serve --addr 127.0.0.1:8091` — server starts on 127.0.0.1:8091
7. `curl localhost:8091/api/specs` — returns spec list
8. `go build ./cmd/signals && go run ./cmd/signals serve --addr 127.0.0.1:8092` — server starts on 127.0.0.1:8092
9. `go test ./internal/pollard/server/... ./internal/gurgeh/server/... ./pkg/signals/...`
10. Colony detection: run `./dev bigend` with Git worktrees present, verify they appear in dashboard

## Key Files (read before implementing)

- `internal/pollard/api/scanner.go` — Scanner type the HTTP server wraps
- `internal/pollard/api/orchestrator.go` — ResearchOrchestrator to expose
- `internal/pollard/api/targeted.go` — TargetedScanOpts/Result types
- `internal/gurgeh/specs/schema.go` — Spec struct for JSON serialization
- `internal/bigend/aggregator/aggregator.go` — State struct to extend
- `pkg/signals/signal.go` — Signal types for broker
- `pkg/signals/server.go` — HTTP/WS server for signals
- `internal/signals/cli/root.go` — Signals CLI root
- `internal/bigend/daemon/server.go` — Reference net/http pattern
