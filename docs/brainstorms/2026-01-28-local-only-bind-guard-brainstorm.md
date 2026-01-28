---
date: 2026-01-28
topic: local-only-bind-guard
---

# Local-Only Bind Guard

## What We're Building

A hard enforcement rule that forbids binding Autarch HTTP/WS servers to non-loopback addresses in v1. Any attempt to bind to a non-loopback address (e.g., 0.0.0.0, public IPs) should fail fast with a clear error message. This codifies the “local-only by default” policy as a technical guarantee rather than a doc-only convention.

## Why This Approach

We considered softer approaches (build tags or env-var escape hatches), but the explicit requirement is to forbid remote access entirely. A hard guard is simplest, least ambiguous, and prevents accidental exposure of endpoints that can trigger costly hunter calls or leak sensitive research/spec data.

## Key Decisions

- **Decision: Hard-fail on non-loopback binds.**
  Rationale: Guarantees local-only behavior and aligns with the policy decision.

- **Decision: Enforce consistently across all servers.**
  Rationale: Avoids policy drift between Pollard, Gurgeh, and any future HTTP/WS servers.

## Open Questions

- Which exact address patterns are allowed? (e.g., `127.0.0.1`, `localhost`, `[::1]`)
- Where should the guard live (shared helper vs per-server)?
- What error text should we standardize on for a clear user message?

## Next Steps

→ /workflows:plan to implement a shared bind guard and apply it to Pollard/Gurgeh/Signals servers.
