---
date: 2026-01-27
topic: multi-agent-strategy
---

# Autarch Multi-Agent Strategy: Gas Town and the Orchestrator Trend

## Context

Steve Yegge's Gas Town (Jan 2026) popularized multi-agent coding orchestration — 20-30+ agents coordinated via Git worktrees, external memory, and a "Mayor" orchestrator. Community debate (notably @voooooogel) suggests stronger models may make thick orchestrators obsolete, similar to LangChain's decline.

## What We're Building

Autarch positions as **coordination infrastructure** — the shared services layer that makes any agent colony smarter. Not an orchestrator competitor, but the platform that orchestrators (and individual agents) consume for research, specifications, quality signals, and domain knowledge.

## Why This Approach

Four approaches were considered:

1. **Observability Layer** (Bigend as dashboard) — too passive, doesn't leverage unique assets
2. **Coordination Infrastructure** (platform) — ✅ chosen — model-proof, leverages all unique assets
3. **Full Orchestrator** (Gas Town competitor) — highest LangChain risk, splits focus
4. **Hybrid** (infra + light orchestration) — elements adopted for Coldwine's bounded 3-5 agent case

The core thesis: **data and domain knowledge are model-proof; orchestration logic is not.** Pollard's research, Gurgeh's structured specs, and the signals system are durable value. Agent spawning logic is exactly what smarter models will internalize.

## Key Decisions

- **Pollard becomes a shared research service** with HTTP API, caching, and async request/response via Intermute
- **Gurgeh exposes spec data as API** — agents query acceptance criteria, requirements, CUJs
- **Signals broadcast to external subscribers** — agent colonies subscribe to spec-health and research-invalidation
- **Bigend adds colony-aware monitoring** — detects external agent sessions (Gas Town, worktrees, etc.)
- **Coldwine keeps bounded orchestration (3-5 agents)** — doesn't scale to Gas Town's 30-agent swarms
- **Intermute adds agent-to-service request/response pattern** — agents ask for research, Pollard responds

## Open Questions

- What's the Pollard API authentication model for external consumers?
- Should Gurgeh spec API be read-only or allow agent-driven spec updates?
- How does Bigend discover external agent colonies (filesystem scan for worktrees? tmux pattern matching?)
- What's the research cache TTL strategy (per-query? per-domain? configurable?)

## Next Steps

→ Plan implementation starting with Pollard API server and research cache
