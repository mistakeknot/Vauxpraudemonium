# Spec Quality + Signal-Driven Replanning

## Status: Implementation Complete

See the full plan in the git history. Implementation covers all 7 capabilities:

1. **Deep Research Grounding** — `internal/pollard/api/targeted.go`, `internal/gurgeh/arbiter/research_phases.go`
2. **Spec Evolution** — `internal/gurgeh/specs/evolution.go`, `internal/gurgeh/specs/diff.go`
3. **Outcome Hypotheses** — `Hypothesis` type in `specs/schema.go`, generator prompts updated
4. **Structured Requirements** — `Requirement` type in `specs/schema.go`, Given/When/Then prompts
5. **Signal System** — `pkg/signals/`, emitters in `internal/{pollard,gurgeh,coldwine}/signals/`
6. **Competitor Watch Mode** — `internal/pollard/watch/`, `pollard watch` CLI command
7. **Agent-Powered Ranking** — `internal/gurgeh/prioritize/`, `gurgeh prioritize` CLI command

## New CLI Commands

- `gurgeh history <spec-id>` — show spec revision changelog
- `gurgeh diff <spec-id> v1 v2` — structured diff between versions
- `gurgeh prioritize <spec-id>` — agent-powered feature ranking
- `pollard watch [--once]` — continuous competitor monitoring
