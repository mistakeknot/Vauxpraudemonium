# Shared Run-Target Registry Design (Praude + Tandemonium + Vauxhall)

## Goal
Introduce a shared run-target registry and resolver so Praude, Tandemonium, and Vauxhall can use the same agent definitions, with global defaults and per-project overrides.

## Context
We want users to use any or all of the three tools independently or together. Today:
- Praude has its own agent profile map in `.praude/config.toml`.
- Vauxhall has its own `agents` config and a resolver with fallbacks.
- Tandemonium does not have a shared target registry and relies on tool-specific wiring.

Schmux’s model shows a clear separation of run targets (detected tools + user targets) plus variants and presets. We can adopt the core idea without introducing a daemon.

## Decision
Create a shared package (proposed `pkg/agenttargets`) that defines:
- A **target registry** (detected tools, user-defined promptable/command targets)
- Optional **variants** (env overlays on detected tools)
- **Presets** (target + prompt)
- A **resolver** that merges global config with per-project overrides and returns a runnable command

## Architecture

### Config Sources
- **Global**: `~/.config/autarch/agents.toml`
- **Project override**: `.praude/agents.toml`
- **Compat**: allow `[agents]` from `.praude/config.toml` to be merged as a project source

### Merge Rules
- Project overrides global by name.
- Detected tools are always available unless explicitly disabled in project config.
- If a variant exists, it resolves before the base tool (variant → detected tool → user target).

### Target Types
- **Detected**: built-in tools (claude, codex, gemini) with interactive and oneshot modes
- **Promptable**: user command that accepts prompt as final arg
- **Command**: user command with no prompt

### Resolver Output
`ResolveTarget(context, projectPath, name)` returns:
- `Command`, `Args`, `Env`
- `PromptMode` (interactive/oneshot/none)
- `Source` (project/global/detected/variant)

### Context Rules
- **Vauxhall (global)**: use global config only unless project path supplied for spawn
- **Praude (project)**: project overrides + global fallback
- **Tandemonium (per Praude project)**: resolve via project path, with global fallback

## Integration Plan (High-Level)
1. Introduce `pkg/agenttargets` with config parsing, merge logic, and resolver.
2. Adapt Praude to resolve targets through shared package (keep compat for existing config).
3. Adapt Vauxhall resolver to use shared package; allow projectPath to influence resolution when spawning.
4. Update Tandemonium to use shared resolver based on project root.

## Testing
- Unit tests for config parsing and merge precedence.
- Resolver tests per context (global-only, project override, variant resolution).
- Small integration tests for Praude and Vauxhall (ensuring previous defaults still resolve).

## Risks
- Config migration confusion if we replace `.praude/config.toml` directly. Mitigate by supporting both formats.
- Divergent defaults if detected tools list isn’t centralized. Mitigate by keeping built-ins in shared package.

## Success Criteria
- One registry drives all three tools.
- Vauxhall (global) and Praude/Tandemonium (project) resolve consistently.
- Per-project overrides work without breaking existing configs.

