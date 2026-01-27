---
name: spec-sprint
description: This skill guides 10-minute PRD creation using Arbiter's propose-first workflow.
---

# Spec Sprint Skill

## When to Use

- Starting a new feature or project
- Converting a rough idea into a validated PRD
- Onboarding to an existing codebase that needs product direction

## The Sprint Flow

### Opening (1 min)
- For existing projects: Arbiter reads context and proposes a Problem statement
- For blank slate: Ask "Describe your idea" and draft from response

### Sections (6-8 min)

For each section, Arbiter proposes a draft. User can:
- **Accept** (press 'a') - Use draft as-is
- **Edit** (press 'e') - Modify directly
- **Alternative** (press 1-3) - Apply suggested rephrasing

| # | Section | What Arbiter Drafts |
|---|---------|---------------------|
| 1 | Problem | Pain point with context |
| 2 | Users | Target personas with characteristics |
| 3 | Features + Goals | Capabilities with measurable outcomes |
| 4 | Scope + Assumptions | Boundaries and foundational beliefs |
| 5 | CUJs | Critical User Journeys with steps |
| 6 | Acceptance Criteria | Testable success conditions |

### Quick Scan (after Problem)

Ranger runs github-scout + hackernews (~30 sec) to find:
- Similar OSS projects
- HN discussions about the problem space

Findings inform the Features section.

### Consistency Checking

After each section, check for conflicts:
- 🔴 **Blockers** - Must resolve (e.g., enterprise features for solo users)
- 🟡 **Warnings** - Can dismiss (e.g., goal without supporting feature)

### Handoff (1 min)

After all sections:
1. **Research & iterate** (Recommended first time) - Deep dive with Ranger
2. **Generate tasks** - Create epics/stories with Forger
3. **Export** - YAML/Markdown for coding agents

## Commands

```bash
# Start a sprint
/autarch:prd

# Start with initial idea
/autarch:prd "Build a reading tracker for developers"

# Resume a sprint
/autarch:prd SPRINT-1234567890
```

## Output

PRD saved to `.gurgeh/specs/PRD-{id}.yaml`
Sprint state saved to `.gurgeh/sprints/SPRINT-{id}.yaml`
