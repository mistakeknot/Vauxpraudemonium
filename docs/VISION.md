# Autarch Vision

> The tools that build the tools that build the products.

## What Autarch Is

Autarch is an integrated system for **AI-assisted product development** — from research through specification through execution through observation. It treats the full product development lifecycle as a single, coherent workflow rather than a collection of disconnected steps.

The name comes from Gene Wolfe's *Book of the New Sun*, where the Autarch is the sovereign ruler of the Commonwealth — one who must synthesize all perspectives and memories into coherent governance. That's the aspiration: a system that synthesizes research, specification, execution, and observation into coherent product development.

## The Problem

Modern AI-assisted development is fragmented:

- **Research** happens in chat windows and browser tabs, then evaporates.
- **Specifications** are written once and immediately drift from reality.
- **Task breakdowns** are manual, inconsistent, and disconnected from the spec that motivated them.
- **Execution** happens in isolation — agents work without awareness of the broader context.
- **Observation** requires manually checking terminals, logs, and dashboards.

Each gap is a place where context is lost, decisions are forgotten, and work is duplicated.

## The Vision

Autarch closes these gaps with four specialized tools that share a common data contract and communicate through a unified coordination layer:

| Tool | Role | Analogy |
|------|------|---------|
| **Pollard** | Research intelligence | The analyst who reads everything |
| **Gurgeh** | Specification & planning | The architect who writes the blueprint |
| **Coldwine** | Task orchestration & execution | The foreman who coordinates the crew |
| **Bigend** | Observation & mission control | The dashboard that shows the whole picture |

These aren't four separate products bolted together. They're four perspectives on the same underlying reality: a project's journey from idea to implementation.

### Key Principles

**1. Research is a first-class input, not an afterthought.**
Pollard doesn't just run once — it continuously monitors the landscape, and its findings flow directly into specifications and task context. A competitor ships a new feature? That insight reaches the spec author and the implementing agent.

**2. Specifications are living documents.**
Gurgeh tracks spec evolution with versioned snapshots, assumption confidence decay, and consistency checking across sections. A spec isn't a static artifact — it's an evolving model of intent.

**3. Human-AI collaboration is the unit of work.**
Coldwine doesn't just dispatch tasks to agents. It manages the dialogue between human judgment and AI execution — knowing when to ask, when to proceed, and when to escalate.

**4. Observation without intervention.**
Bigend aggregates state from all tools but never writes. It's a pure observer — providing situational awareness without becoming a bottleneck or introducing side effects.

**5. Graceful degradation everywhere.**
Every integration is optional. Each tool works standalone. Intermute coordination enriches the experience but isn't required. You can use Gurgeh without Pollard, Coldwine without Gurgeh, Bigend without any of them.

**6. The system improves itself.**
Autarch's own development uses Autarch. Research findings, spec quality signals, and execution patterns feed back into the tools' evolution. This document itself should be maintained as part of that process.

## Where We Are

Autarch is a working system with all four tools building and running:

- **Pollard**: Multi-domain hunters (tech, academic, medical, legal), continuous watch mode, API for programmatic access
- **Gurgeh**: Arbiter-driven spec sprints with consistency checking, confidence scoring, research integration, spec evolution tracking
- **Coldwine**: Task orchestration with agent coordination, worktree management
- **Bigend**: Project discovery, agent state detection, web + TUI modes (TUI in progress)
- **Cross-cutting**: Signal system, event spine, Intermute coordination, shared TUI with Tokyo Night theme

## Where We're Going

The near-term focus areas (not a roadmap — priorities shift as research and usage reveal what matters):

- **Signal integration**: Surfacing typed alerts (competitor shipped, assumption decayed, execution drifted) inline across all TUIs
- **Deeper research loops**: Pollard findings triggering spec revisions and task re-prioritization automatically
- **Multi-repo orchestration**: Bigend and Coldwine operating across project boundaries
- **Agent runner abstraction**: Unified safety policies and pluggable backends for agent execution
- **Self-assessment**: Autarch evaluating and evolving its own specifications and workflows

## How This Document Is Maintained

This vision document is not frozen. It should be revisited when:

- A new tool or major capability is added
- The relationship between tools changes
- User feedback reveals a gap between stated vision and actual value
- Quarterly, as a forcing function for strategic clarity

The Gurgeh spec sprint process can be used to evaluate this document itself — treating the vision as a specification subject to consistency checking, confidence scoring, and research validation.

---

*The suite takes its name from Gene Wolfe's Book of the New Sun, where the Autarch carries the memories of all predecessors — much as this system carries context across every phase of development. The tools draw from Iain M. Banks (Gurgeh, from The Player of Games), China Miéville (Coldwine, from Bas-Lag), and William Gibson (Bigend and Pollard, from Pattern Recognition) — worlds where games, trade, and pattern-finding are forms of power.*
