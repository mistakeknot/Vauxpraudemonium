# Brainstorm: Vision Document Lifecycle in Gurgeh/Autarch

Date: 2026-01-27

## What We're Building

A first-class vision document lifecycle within Gurgeh, where any project can maintain a living vision spec that serves as **canonical root context** for all downstream specs. The vision is versioned, confidence-scored, research-informed, and periodically self-reviewed — using the same Arbiter infrastructure that drives PRD creation. Crucially, every PRD created under a project with a vision spec is checked for alignment against it: the vision is not a peer document but the upstream source of truth that shapes and constrains all specifications beneath it.

## Why This Approach

Autarch already has the machinery for living documents: versioned snapshots, assumption confidence decay, typed signals, phase-driven Pollard research, and the Arbiter's propose/accept/revise sprint flow. But none of this has been applied to strategic documents like vision statements. The gap is conceptual, not technical — vision docs sit as static markdown files outside the spec system.

By treating vision as a regular Gurgeh spec with `type=vision`, we close the loop: the system that manages product specifications can manage its own strategic direction.

## Key Decisions

### 1. Vision is a Spec, not a new type
Vision documents reuse the existing `specs.Spec` schema with `type=vision`. Fields are reinterpreted:

| Spec field | Vision interpretation |
|---|---|
| Goals | Principles |
| Assumptions | Strategic bets |
| Hypotheses | Falsifiable predictions about the system's direction |
| CUJs | Key workflows the system enables |
| Features | Tool capabilities and integration points |

**Rationale**: Avoids duplicating Arbiter infrastructure. Assumption decay, confidence scoring, consistency checking, and research integration all apply naturally without new code paths.

### 2. Vision is canonical root context for all PRDs
The vision spec is not a peer to PRDs — it is upstream. The Arbiter's consistency engine gains a **vertical check** in addition to its existing horizontal (cross-section) checks:

| PRD section | Checked against vision field | Example violation |
|---|---|---|
| Problem | Goals (Principles) | Problem doesn't serve any stated principle |
| Users | Goals (Principles) | Target users outside the vision's scope |
| Features | Assumptions (Strategic bets) | Feature contradicts a strategic bet (e.g., "Bigend is read-only" but PRD proposes Bigend writes) |
| Scope | Assumptions (Strategic bets) | Scope exceeds what the vision considers in-bounds |
| CUJs | CUJs (Key workflows) | Journey doesn't connect to any vision-level workflow |
| AC | Hypotheses (Predictions) | Acceptance criteria that can't be evaluated against vision hypotheses |

**Mechanics**: When a project has a `type=vision` spec, the Arbiter loads it at sprint start and includes it in every consistency check pass. Conflicts surface as warnings (not blockers) — the human decides whether the PRD is intentionally diverging or needs adjustment.

**Reverse flow**: If multiple PRDs consistently diverge from the vision in the same direction, that's a signal that the vision may need updating. This feeds into the signal accumulation trigger for vision review (Decision 3).

### 3. Full loop: time + signal triggered review (including PRD divergence)
Reviews trigger in two ways:
- **Time-based cadence**: Every N days (configurable, default 30), the system schedules a vision review
- **Signal accumulation**: When N unaddressed signals accumulate against vision assumptions, trigger early review

Both mechanisms can fire. Time is the baseline heartbeat; signals accelerate when the world changes faster than expected.

### 3. Review uses full Arbiter sprint with auto-skip
When a vision review triggers, it runs the standard Arbiter sprint flow against the existing vision spec, but:
- Sections with no active signals/decay are pre-marked `auto_accept` and skip past quickly
- Sections with signals/decay pause for human review — Arbiter proposes changes, human accepts/revises
- User can manually stop on any section regardless of auto-accept status
- Cross-section consistency checking runs on all sections (including auto-accepted ones)

**Rationale**: Preserves the Arbiter's most valuable property (cross-section consistency) while not wasting time on healthy sections. Small implementation delta: add auto-accept flag logic.

### 4. Arbiter proposes draft revisions
When a flagged section is presented, the Arbiter:
1. Shows current section content
2. Lists accumulated signals/decayed assumptions
3. Summarizes relevant Pollard research since last review
4. Proposes specific text changes (or "no change recommended")
5. Human accepts, revises, or dismisses

This is the existing propose/accept/revise cycle, just seeded with prior content instead of blank.

### 5. Any project, no hierarchy
Any `.gurgeh/` project can have a `type=vision` spec. No special suite-level logic. No parent/child vision hierarchy. If hierarchy becomes needed later, a `parent_ref` field can be added — but cascade logic is not built until someone needs it.

### 6. Output is a new spec version
Accepted changes produce a new versioned snapshot via `evolution.SaveRevision()` with trigger=`scheduled_review` or `signal_triggered_review`. The full edit history is preserved.

## Open Questions

- **Default cadence**: 30 days feels right for a fast-moving project, but should this be configurable per-spec?
- **Signal threshold**: How many unaddressed signals trigger early review? 3? 5? Should severity matter (1 critical = immediate)?
- **Research scope**: During a vision review sprint, which Pollard hunters run? The phase-specific mapping in `research_phases.go` is designed for PRD sections — vision sections may need different hunter combinations.
- **Bigend integration**: Should Bigend surface "vision review overdue" as a dashboard-level alert? This seems natural but isn't designed yet.
- **Bootstrap**: How does the initial `docs/VISION.md` become a Gurgeh spec? Manual import? A `gurgeh import --type=vision` command?
- **Divergence tracking**: How many PRDs need to diverge in the same direction before it becomes a vision-level signal? Is this count-based, or does the Arbiter need semantic similarity detection?

## What We're NOT Building

- Vision hierarchy / cascading reviews (YAGNI)
- Autonomous vision changes without human approval (human always accepts/revises)
- New Spec schema fields for vision (reuse existing)
- Background daemons for review scheduling (check-on-load pattern, like assumption decay)

## Next Steps

Run `/workflows:plan` to design the implementation.
