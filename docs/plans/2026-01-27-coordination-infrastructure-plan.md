# Plan: Autarch Coordination Infrastructure (Gas Town Response)

> **Status: ✅ COMPLETE** (implemented Jan 2026)
>
> **Strategy**: Position Autarch as the shared services layer that makes any agent colony smarter. Not an orchestrator — a platform.
>
> **Context**: Steve Yegge's Gas Town (Jan 2026) popularized multi-agent coding orchestration. Community debate suggests stronger models may make thick orchestrators obsolete (like LangChain). Autarch's durable value is *domain intelligence* (research, specs, signals) — not agent spawning logic.
>
> **Brainstorm**: [docs/brainstorms/2026-01-27-multi-agent-strategy-brainstorm.md](../brainstorms/2026-01-27-multi-agent-strategy-brainstorm.md)

---

## Implementation Summary

All 5 phases implemented. Two planned `responses.go` files (Pollard, Gurgeh) were folded inline into `server.go` handlers — a reasonable simplification for small APIs.

| Phase | Status | Key Files |
|-------|--------|-----------|
| **0: Docs** | ✅ | CLAUDE.md, AGENTS.md, FLOWS.md §17 |
| **1: Pollard HTTP API** | ✅ | `internal/pollard/server/` (server, cache, jobs), `internal/pollard/cli/serve.go` |
| **2: Gurgeh Spec API** | ✅ | `internal/gurgeh/server/server.go`, `internal/gurgeh/cli/commands/serve.go` |
| **3: Bigend Colony Detection** | ✅ | `internal/bigend/colony/` (detector, types), aggregator wired |
| **4: Signals Broadcast** | ✅ | `pkg/signals/` (broker, server), `cmd/signals/`, `internal/signals/cli/` |
| **5: Intermute Req/Resp** | ✅ | `internal/pollard/inbox/` (handler, protocol), auto-starts with `INTERMUTE_URL` |

### Deviations from Plan

- `responses.go` files not created as separate files — response envelope types integrated into `server.go` handlers directly. Acceptable: avoids premature abstraction until a third server needs shared response types.
- Gurgeh serve command lives at `internal/gurgeh/cli/commands/serve.go` (not `internal/gurgeh/cli/serve.go`) — follows Gurgeh's existing `commands/` subpackage pattern.

---

## Original Plan (for reference)

### Phase 0: Documentation Updates ✅

Updated CLAUDE.md (serve commands, key paths, design decisions), AGENTS.md (TODO → Done), FLOWS.md (Section 17: Coordination Infrastructure, colony detection note, signal broadcast note).

### Phase 1: Pollard HTTP API Server ✅

Thin `net/http` server wrapping existing `Scanner` and `ResearchOrchestrator`. Local-only by default (loopback). Async jobs for scans/research.

**Created:**
- `internal/pollard/server/server.go` — routes: `/health`, `/api/scan`, `/api/scan/targeted`, `/api/research`, `/api/jobs/{id}`, `/api/jobs/{id}/result`, `/api/insights`, `/api/hunters`
- `internal/pollard/server/cache.go` — TTL + size-bounded LRU + singleflight de-dup
- `internal/pollard/server/jobs.go` — in-memory job store
- `internal/pollard/cli/serve.go` — `pollard serve --addr 127.0.0.1:8090`

### Phase 2: Gurgeh Spec API ✅

Read-only HTTP endpoints over `.gurgeh/specs/*.yaml`. Local-only.

**Created:**
- `internal/gurgeh/server/server.go` — routes: `/health`, `/api/specs`, `/api/specs/{id}`, `/api/specs/{id}/requirements`, `/api/specs/{id}/cujs`, `/api/specs/{id}/hypotheses`, `/api/specs/{id}/history`
- `internal/gurgeh/cli/commands/serve.go` — `gurgeh serve --addr 127.0.0.1:8091`

### Phase 3: Bigend Colony Detection ✅

Git worktree + convention marker detection, integrated into aggregator State.

**Created:**
- `internal/bigend/colony/detector.go`, `types.go` (+ tests)
- `aggregator.go` — `Colonies []colony.Colony` in State

### Phase 4: Signals Broadcast ✅

Standalone WebSocket server for signal fan-out to local subscribers.

**Created:**
- `pkg/signals/broker.go` — Subscribe/Publish with type filtering
- `pkg/signals/server.go` — HTTP+WS server
- `cmd/signals/main.go`, `internal/signals/cli/` — `signals serve --addr 127.0.0.1:8092`

### Phase 5: Intermute Request/Response ✅

Async agent-to-Pollard research queries via Intermute messaging.

**Created:**
- `internal/pollard/inbox/handler.go` — inbox polling, auto-starts when `INTERMUTE_URL` set
- `internal/pollard/inbox/protocol.go` — message types

---

## Verification Commands

```bash
# Pollard API
go run ./cmd/pollard serve --addr 127.0.0.1:8090
curl localhost:8090/health
curl -X POST localhost:8090/api/scan -d '{"hunters":["github-scout"],"mode":"quick"}'

# Gurgeh API
go run ./cmd/gurgeh serve --addr 127.0.0.1:8091
curl localhost:8091/api/specs

# Signals WS
go run ./cmd/signals serve --addr 127.0.0.1:8092

# Tests
go test ./internal/pollard/server/... ./internal/gurgeh/server/... ./pkg/signals/...
```
