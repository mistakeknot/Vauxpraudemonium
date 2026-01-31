# Gurgeh Roadmap

> The best tool for vibecoders to create specs that produce great agent output

## Vision

Vibecoders want to describe what they want and have agents build it well. Gurgeh's job is to minimize the time between "I have an idea" and "my agent has a clear, testable brief." Every feature should either reduce user effort or increase output quality — ideally both.

The north star: **a spec that makes agents succeed on the first try.**

---

## Done

### M0: Foundation ✅

**Goal:** Phase-based PRD sprint with propose-first UX

| Task | Status | Notes |
|------|--------|-------|
| 8-phase arbiter sprint | ✅ | Vision → Acceptance Criteria |
| Draft generation with alternatives | ✅ | 2-3 options per phase |
| Consistency checking | ✅ | 4 checkers: user-feature, goal-feature, scope-creep, assumption |
| Confidence scoring | ✅ | Two calculators: fast (arbiter) + content-aware (gurgeh) |
| Spec storage + YAML persistence | ✅ | .gurgeh/specs/, sprint save/load |
| TUI sprint view | ✅ | Bubble Tea, keyboard nav, research panel |
| Spec validation (hard/soft) | ✅ | Required fields, CUJ validation, status checks |

### M1: Research Integration ✅

**Goal:** Pollard findings inform spec generation

| Task | Status | Notes |
|------|--------|-------|
| Intermute research provider | ✅ | Create/link specs, fetch findings |
| Phase-specific research config | ✅ | Vision→github-scout, Problem→arxiv, Features→competitor-tracker |
| Deep scan async plumbing | ✅ | Start/check/import lifecycle |
| Research quality in confidence | ✅ | 0.3×count + 0.3×diversity + 0.4×relevance |
| Vision alignment | ✅ | Load vision spec, cross-check PRD sections |

### M2: Thinking Shapes ✅

**Goal:** Metacognitive preambles force quality-standard formulation before generation

| Task | Status | Notes |
|------|--------|-------|
| 5 shapes: Deductive, Inductive, Abductive, Contrapositive, DSL | ✅ | pkg/thinking/ |
| Phase-to-shape defaults | ✅ | Each phase gets the right thinking strategy |
| Per-sprint shape overrides | ✅ | SprintState.ShapeOverrides |
| Shape-aware confidence | ✅ | Deductive/DSL → +specificity, Contrapositive → +assumptions |
| Pollard agent hunter integration | ✅ | Contrapositive for competitive, Abductive for general |

---

## Planned

### M3: Agent-Ready Brief Decomposition

**Goal:** Transform monolithic specs into focused, context-window-sized briefs that agents actually consume well

**Why first:** This is the closest to the user's actual goal. A perfect spec that produces a bad agent brief is a perfect failure.

| Task | Status | Notes |
|------|--------|-------|
| Task-level brief extraction | ⬜ | Decompose spec into independent work units with focused context |
| Context budget system | ⬜ | Size briefs for agent context windows (2K-4K tokens per task) |
| Acceptance criteria → test stubs | ⬜ | Generate skeleton test files from Given/When/Then requirements |
| Dependency graph per brief | ⬜ | Each brief knows what it depends on and what depends on it |
| Brief quality scoring | ⬜ | Is this brief specific enough for an agent to act on without asking questions? |

### M4: Self-Critique Loop

**Goal:** Agents evaluate their own drafts against shape criteria before proposing to the user

**Why second:** Better first drafts = less user effort. The user should rarely need to revise.

| Task | Status | Notes |
|------|--------|-------|
| Generate → evaluate → revise pipeline | ⬜ | One self-critique pass before proposing |
| Shape-specific evaluation rubrics | ⬜ | Deductive: "did it state criteria first?" Contrapositive: "did it enumerate failures?" |
| Critique-to-revision mapping | ⬜ | Failed rubric items become revision instructions |
| Critique visibility in TUI | ⬜ | Show what was caught and fixed (builds trust) |
| Configurable critique depth | ⬜ | 0 = off (current behavior), 1 = single pass, 2 = thorough |

### M5: Shape Output Validation

**Goal:** Verify that thinking shape preambles actually improved output quality

**Why third:** Without validation, thinking shapes are cargo cult prompting.

| Task | Status | Notes |
|------|--------|-------|
| Per-shape output validators | ⬜ | Check structural compliance (e.g., DSL output has schema fields) |
| Validation score per section | ⬜ | 0-1 "did the shape help?" metric |
| Validator feedback → re-generation | ⬜ | If shape wasn't followed, re-prompt with explicit correction |
| Shape effectiveness tracking | ⬜ | Over time, learn which shapes work best for which project types |

### M6: Subagent Enrichment Passes

**Goal:** The 9 subagents automatically critique and enrich drafts during the sprint

**Why fourth:** Moves from "user catches problems" to "system catches problems."

| Task | Status | Notes |
|------|--------|-------|
| Phase → subagent mapping | ⬜ | Which subagents run on which phases |
| Enrichment pass pipeline | ⬜ | Draft → subagent critique → merge suggestions → propose to user |
| Strategist: architecture implications | ⬜ | Auto-flag technical risks in features |
| Navigator: journey completeness | ⬜ | Check CUJs cover all features |
| Sentinel: security surface analysis | ⬜ | Flag features that need security requirements |
| Configurable subagent depth | ⬜ | none / quick / thorough |

### M7: Research-Annotated Drafts

**Goal:** Proactive intelligence — research findings annotate drafts inline, not in a side panel

**Why fifth:** Vibecoders won't check a research panel. Findings must appear where they're relevant.

| Task | Status | Notes |
|------|--------|-------|
| Finding → section relevance matching | ⬜ | Link findings to the specific section they inform |
| Inline annotations in draft content | ⬜ | "⚡ Competitor X already ships this as..." |
| Conflict detection: assumption vs finding | ⬜ | "Your assumption A conflicts with finding F" |
| Auto-import research on phase advance | ⬜ | No manual polling; findings flow in as phases progress |

### M8: Hypothesis Lifecycle

**Goal:** Close the spec → implementation → validation feedback loop

**Why sixth:** Without this, specs are write-once documents that rot. Hypotheses must be trackable.

| Task | Status | Notes |
|------|--------|-------|
| Hypothesis status tracking | ⬜ | untested → testing → validated / invalidated |
| Link hypotheses to metrics | ⬜ | Connect "metric Y changes by Z" to actual measurement |
| Invalidation → spec revision trigger | ⬜ | Failed hypothesis flags affected sections for review |
| Assumption decay automation | ⬜ | Background check on DecayDays, emit signals, nudge user |

### M9: Spec Versioning & Diff

**Goal:** Multi-session iteration with full history

**Why seventh:** Vibecoders iterate across sessions. They need to see what changed and why.

| Task | Status | Notes |
|------|--------|-------|
| Version snapshots on save | ⬜ | .gurgeh/specs/history/{spec-id}/v{N}.yaml |
| Structured diff between versions | ⬜ | Per-section change summary, not raw text diff |
| Side-by-side TUI comparison | ⬜ | View two versions simultaneously |
| Change reason tracking | ⬜ | Why was this section revised? (user annotation or auto-detected) |

### M10: Agent-Native Export

**Goal:** Meet vibecoders where their agents live

**Why last:** Only valuable once the specs themselves are excellent.

| Task | Status | Notes |
|------|--------|-------|
| Export to Claude Projects format | ⬜ | Project instructions + task files |
| Export to Cursor rules | ⬜ | .cursorrules with spec context |
| Export to Codex instructions | ⬜ | AGENTS.md-compatible format |
| Export to CLAUDE.md | ⬜ | Project-level context for Claude Code |
| Spec templates by domain | ⬜ | SaaS, CLI tool, mobile app, API service starters |

---

## Principles

1. **Less user effort, better output.** Every feature must pass this test.
2. **Agents are the audience.** The spec exists to make agents succeed, not to satisfy a process.
3. **Proactive over passive.** Don't make users check panels — bring insights to where they're working.
4. **Validate, don't assume.** Thinking shapes, subagent passes, and self-critique must prove they help.
5. **Context windows are real.** Briefs must be sized for how agents actually consume instructions.
6. **Specs are living documents.** Versioning, decay, and feedback loops keep them honest.
