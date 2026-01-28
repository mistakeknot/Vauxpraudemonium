title: feat: Coordination API foundation
type: feat
date: 2026-01-28

# feat: Coordination API foundation

## Enhancement Summary

**Deepened on:** 2026-01-28  
**Sections enhanced:** Overview, Problem Statement, Proposed Solution, Technical Considerations, Acceptance Criteria, Success Metrics, Dependencies & Risks  
**Research sources used:** RFC 9110 (HTTP status semantics), Go net/http server documentation, local repo patterns (Intermute, Bigend WebSocket usage)

### Key Improvements
1. Added explicit HTTP status semantics for async jobs (202 Accepted) and conflict states.
2. Added concrete timeout guidance for HTTP servers to avoid resource exhaustion.
3. Added guidance for WebSocket implementation to match existing `nhooyr.io/websocket` usage.
4. Locked a concrete response envelope schema and expanded job state model.

### New Considerations Discovered
- Default zero timeouts in `net/http.Server` mean no limits; set timeouts explicitly for safety.
- HTTP 202 is the correct status for accepted-but-not-completed async jobs.

## Overview

Implement the API foundation for coordination infrastructure: standard response envelope + typed errors, async jobs for Pollard scans/research, bounded cache with singleflight de-dup, local-only bind guard, paginated Gurgeh lists, and a standalone Signals server. This plan operationalizes the updated coordination infrastructure plan while enforcing the local-only policy.

### Research Insights

**Best Practices:**
- Use HTTP 202 for job creation to communicate “accepted for processing” rather than completion.
- Standardize error bodies and include retry guidance to help agents decide how to recover.

**Implementation Details:**
```json
// Success envelope
{ "ok": true, "data": { ... }, "meta": { "cursor": "next", "limit": 50 } }

// Error envelope
{ "ok": false, "error": { "code": "invalid_request", "message": "..." , "retryable": false } }
```

**Envelope Schema (v1):**
- `ok` (bool, required)
- `data` (object, required when ok=true)
- `meta` (object, optional)
  - `cursor` (string, optional)
  - `limit` (int, optional)
- `error` (object, required when ok=false)
  - `code` (string, required)
  - `message` (string, required)
  - `details` (object, optional)
  - `retryable` (bool, optional; default false)

**Error Codes + HTTP Mapping (v1):**
- `invalid_request` → 400
- `unauthorized` → 401
- `forbidden` → 403
- `not_found` → 404
- `conflict` → 409
- `rate_limited` → 429
- `job_pending` → 409
- `job_failed` → 409
- `internal_error` → 500

## Problem Statement / Motivation

The current plan describes multiple new HTTP/WS surfaces but lacks a consistent response contract, robust handling for long-running scans, and a hard local-only enforcement mechanism. Without these, APIs will be brittle (timeouts, duplicated work), inconsistent for external agents, and easy to expose accidentally.

### Research Insights

**Edge Cases:**
- Long scans that exceed client timeouts should not block HTTP handlers.
- Duplicate concurrent scans can overwhelm hunter APIs if in-flight dedup is not enforced.

## Proposed Solution

- **Standard API envelope + typed errors** used by Pollard and Gurgeh servers.
- **Async job model** for Pollard scan/research endpoints with status and result retrieval.
- **Cache policy**: TTL + size-bounded LRU + singleflight in-flight de-dup.
- **Local-only bind guard** shared helper; reject non-loopback addresses.
- **Gurgeh list pagination** using cursor + limit aligned with Intermute.
- **Pollard insights pagination** using cursor + limit aligned with Intermute.
- **Signals server** as standalone CLI (`signals serve`) with WebSocket broadcast.

### Research Insights

**Implementation Details:**
```go
// HTTP 202 for async job creation
w.WriteHeader(http.StatusAccepted)
json.NewEncoder(w).Encode(envelope{OK: true, Data: jobSummary})
```

## Technical Considerations

- **Consistency with Intermute:** Intermute handlers return plain JSON without an envelope but do use cursor + limit pagination. We will adopt cursor + limit and layer a consistent envelope for the new APIs.
- **Jobs:** Keep in-memory for v1 with TTL cleanup. Expose explicit job status and result endpoints; return 202 for job creation.
- **Job states (v1):** `queued`, `running`, `succeeded`, `failed`, `canceled`, `expired`, `stalled`, `retrying`, `paused`.
  - **Transitions (v1):**
    - `queued` → `running` (worker picks up)
    - `running` → `succeeded` | `failed`
    - `running` → `retrying` (on retriable failure; backoff + requeue)
    - `retrying` → `running`
    - `queued` | `running` → `paused` (operator/admin action)
    - `paused` → `queued` (resume)
    - `queued` | `running` | `paused` → `canceled` (operator/admin action)
    - `queued` | `paused` → `expired` (TTL cleanup)
    - `running` → `stalled` (heartbeat timeout; operator may cancel or requeue)
- **Job retention (v1 defaults):** TTL 24h, max 20,000 jobs (evict oldest first).
- **Job control (v1):** cancel only (`POST /api/jobs/{id}/cancel`).
- **Cache:** Use Go stdlib only; avoid external deps. Maintain bounded size and TTL to avoid memory growth.
- **Local-only policy:** Allow `127.0.0.1`, `localhost`, `[::1]`; hard-fail otherwise with standard error message.
- **Cross-platform colony detection:** Git worktrees + markers everywhere; `/proc` scanning only on Linux via OS guard.

### Research Insights

**Best Practices:**
- Set explicit HTTP server timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`) to avoid slowloris-style resource exhaustion.
- Reuse `nhooyr.io/websocket` for the signals server to align with existing Bigend/websocket implementations.

**Edge Cases:**
- `localhost` may resolve to IPv6; accept `[::1]` explicitly in bind guards.
- Job results should differentiate “pending” vs “failed” to avoid ambiguous retries.
- Pollard insights can grow large; pagination should be required, not optional.

## Acceptance Criteria

- [ ] Pollard endpoints return 202 with job id; status + result endpoints work.
- [ ] `POST /api/jobs/{id}/cancel` transitions jobs to `canceled`.
- [ ] Response envelope is consistent across Pollard/Gurgeh (`ok`, `data`, `error`, `meta`).
- [ ] Typed error codes map to HTTP status codes and include clear messages.
- [ ] Pollard cache is TTL + size-bounded LRU + singleflight de-dup.
- [ ] Gurgeh list endpoints support `cursor` + `limit`.
- [ ] Pollard insights list supports `cursor` + `limit`.
- [ ] Signals server runs as standalone CLI and uses local-only bind guard.
- [ ] Non-loopback binds fail fast with standard error message.
- [ ] Colony detection behaves cross-platform; `/proc` only on Linux.

## Success Metrics

- Long scans no longer time out HTTP requests.
- Duplicate concurrent scans are avoided in practice.
- Local-only policy is enforced by default and testable.

### Research Insights

**Performance Considerations:**
- Track job latency percentiles (p50/p95) to validate async model effectiveness.
- Track cache hit rate for scan requests to validate dedup wins.

## Dependencies & Risks

- **Risk:** API envelope diverges from Intermute conventions.
  - **Mitigation:** Keep envelope minimal and document it clearly; align pagination with Intermute.
- **Risk:** Job store memory growth.
  - **Mitigation:** TTL cleanup, size caps, and simple lifecycle states.
- **Risk:** Premature generalization of job states in Intermute.
  - **Mitigation:** Keep job state model local to Pollard until 2+ tools expose async jobs; then move shared types to `pkg/contract` before touching Intermute schema.

### Research Insights

**Risk Mitigations:**
- Add bounded job retention (e.g., max N jobs, TTL) to prevent memory leaks.
- Validate actual Pollard/Gurgeh types before wiring responses to avoid plan-to-implementation drift.

## References & Research

- `docs/plans/2026-01-27-coordination-infrastructure-plan.md`
- `docs/brainstorms/2026-01-28-local-only-bind-guard-brainstorm.md`
- Intermute HTTP handlers: `/root/projects/Intermute/internal/http/handlers_*`
- Intermute auth local handling: `/root/projects/Intermute/internal/auth/middleware.go`
- RFC 9110: HTTP status semantics (202 Accepted, 409 Conflict)
- Go net/http server timeouts: `net/http.Server` docs
