title: feat: Unified TUI non-printable shortcuts
name: Unified TUI non-printable shortcuts
date: 2026-01-29
status: proposed
owner: core-tui

# feat: Unified TUI non-printable shortcuts

## Enhancement Summary

**Deepened on:** 2026-01-29  
**Sections enhanced:** Overview, Proposed Solution, Technical Considerations, SpecFlow Analysis, Implementation Plan, Dependencies & Risks, Documentation + Tests  
**Research sources used:** repo conventions (`docs/tui/SHORTCUTS.md`, `docs/ARCHITECTURE.md`), keymap locations (`pkg/tui/keys.go`, `internal/tui/unified_app.go`, `internal/tui/views/*`), institutional learnings (`docs/solutions/ui-bugs/*`)

### Key Improvements
1. Added explicit non‑printable binding guidelines and reserved key caveats (ctrl/alt/terminal signals).
2. Added fallback strategy for tab switching and help discoverability without printable keys.
3. Added audit checklist for view‑specific keys and help overlays to prevent regressions.

### New Considerations Discovered
- Terminal control keys (ctrl+z, ctrl+d, ctrl+s/ctrl+q, ctrl+l) should remain unbound or carefully guarded.
- Alt/meta combos are sent as Esc prefixes; avoid for global bindings to prevent accidental Esc handling.
- Some terminals may not emit F-keys or ctrl+arrow consistently; plan for fallbacks and visible help.

## Overview
Standardize keyboard shortcuts in the global three‑pane UI so they never conflict with text input. Printable keys (letters, digits, symbols, space) are forbidden everywhere in the unified UI. Pane focus remains on Tab/Shift+Tab, and dashboard tab switching moves to modifier navigation. This makes typing safe in the chat panel and eliminates ambiguous or view‑specific printable bindings.

### Research Insights
- `docs/tui/SHORTCUTS.md` already frames the goal: avoid collisions with terminal-reserved keys and never capture printable characters in text inputs. This plan extends that policy to the entire unified UI (not just inputs).
- Consistency is enforced through shared keymap usage (`pkg/tui/keys.go`) and help overlays in `internal/tui/unified_app.go` plus view help methods.

## Problem Statement / Motivation
Typing in the chat pane currently triggers global and view‑specific printable shortcuts (e.g., j/k, n/p, g/G, /, space). This causes accidental navigation, toggles, and mode switches. The dashboard’s Tab behavior also intercepts pane focus, making the UI inconsistent and confusing. A non‑printable shortcut policy removes these conflicts and improves predictability.

## Proposed Solution
1. **Global keymap becomes non‑printable only.** Update `pkg/tui/keys.go` to remove all printable bindings (letters, digits, space, symbols) and replace with modifier/non‑printable keys.
2. **Dashboard tabs move off Tab.** Replace dashboard tab switching with Ctrl+Left/Ctrl+Right (and non‑printable fallback), leaving Tab/Shift+Tab for pane focus.
3. **View‑specific shortcuts become non‑printable.** Replace printable bindings in all unified views (help overlays and handlers) with non‑printable/modifier equivalents.
4. **Documentation and help overlay updated.** Update `docs/tui/SHORTCUTS.md` and in‑app help to match the new policy.

### Canonical Binding Table (Proposed)
**Global (applies everywhere):**
- `ctrl+c` → Quit
- `F1` → Help overlay toggle
- `esc` → Back/close (non-destructive)
- `tab` / `shift+tab` → Cycle pane focus
- `enter` → Select/activate primary action
- `up` / `down` → Navigate lists
- `home` / `end` → Top/bottom
- `pgup` / `pgdn` → Prev/next
- `ctrl+f` → Search (where supported)
- `ctrl+r` → Refresh

**Policy note:** Printable keys (letters, digits, symbols, space) are ignored everywhere in the unified UI, including view‑specific actions.

**Dashboard tabs:**
- `ctrl+left` / `ctrl+right` → Previous/next tab
- `ctrl+pgup` / `ctrl+pgdn` → Fallback tab switching

**View‑specific actions (non‑printable only):**
- `F2` → Model selector toggle (consistent with existing hints)
- `F3–F6` → View‑specific actions (define per‑view; document in help)
- `F7–F12` → Destructive or high‑impact view actions (define per‑view; document in help)

**Reserved / Do‑Not‑Bind (terminal collisions):**
- `ctrl+z`, `ctrl+d`, `ctrl+s`, `ctrl+q`, `ctrl+l`, `ctrl+v`, `shift+insert`, any `alt/meta` combo (Esc prefix)

### Research Insights
- Favor function keys, arrows, PageUp/PageDown, Home/End, and Enter as low‑collision defaults (aligned with `docs/tui/SHORTCUTS.md`).
- Avoid binding to terminal control signals (`ctrl+z`, `ctrl+d`, `ctrl+s`, `ctrl+q`, `ctrl+l`) and alt/meta combos (Esc prefix collisions).
- Preserve `ctrl+c` as a universal quit (already policy in `docs/tui/SHORTCUTS.md` and tests).

## Technical Considerations
- **Common keys live in** `pkg/tui/keys.go` and are used across `internal/tui/unified_app.go` and most views (`internal/tui/views/*`).
- **Help overlay content** is assembled in `internal/tui/unified_app.go` (global bindings) and view help methods (ShortHelp/FullHelp).
- **Tab focus** is owned by `pkg/tui/shelllayout.go` and should not be intercepted for dashboard tab switching.
- **Terminal compatibility:** Ctrl+Left/Ctrl+Right may not be emitted consistently by all terminals. Provide a non‑printable fallback (e.g., Ctrl+PgUp/Ctrl+PgDn).

### Research Insights
- Some terminals do not emit F-keys or ctrl+arrow consistently; provide at least one alternate navigation path visible in help.
- Alt/meta shortcuts arrive as `esc` prefixes in Bubble Tea; avoid using alt/meta for global shortcuts so Esc remains a reliable “back/close”.
- Keep `enter`/`esc` semantics consistent across views to reduce help overlay dependency.

## SpecFlow Analysis (User Flows & Gaps)
**Primary flows impacted:**
- Switch pane focus (sidebar → doc → chat)
- Switch dashboard tabs
- Navigate lists and toggle items in doc pane
- Open/close help overlay
- Refresh data and search inside views

**Gaps to close:**
- Ensure every action has a non‑printable binding and is discoverable in help.
- Ensure chat input never receives intercepted printable keys.
- Provide a fallback for tab switching in terminals without Ctrl+Left/Right support.

### Research Insights
- Chat input focus must bypass all printable handlers in unified views; auditing handlers for `commonKeys.*` usage is required.
- Help overlay should list non‑printable keys for each view so removal of printables does not reduce discoverability.

## Acceptance Criteria
- [ ] No printable keys are bound in unified UI (global or view‑specific).
- [ ] Tab/Shift+Tab always cycles pane focus (never switches dashboard tabs).
- [ ] Dashboard tab switching uses Ctrl+Left/Ctrl+Right with fallback to Ctrl+PgUp/Ctrl+PgDn.
- [ ] Help overlay and `docs/tui/SHORTCUTS.md` reflect the new policy and bindings.
- [ ] `ctrl+c` still quits from any screen.
- [ ] Tests updated to reflect new bindings and focus behavior.

## Implementation Plan

### Phase 1: Keymap Policy + Common Keys
- Update `pkg/tui/keys.go`:
  - Remove printable key bindings: `?`, `/`, `j/k`, `g/G`, `n/p`, `r`, `space`, `1-9`.
  - Replace with non‑printable/modifier keys, e.g.:
    - Help: `F1`
    - Search: `ctrl+f`
    - Refresh: `ctrl+r`
    - Nav up/down: `up`/`down` only
    - Top/bottom: `home`/`end` only
    - Next/Prev: `pgdown`/`pgup` (or `ctrl+right`/`ctrl+left` if available)
    - Toggle: `enter` only (no `space`)
    - Sections: remove numeric keys entirely
- Update `pkg/tui/help.go` / help helpers if they reference removed printable keys.

### Research Insights (Phase 1)
- Consider a small helper to document “reserved keys” in code comments to prevent reintroduction (`ctrl+z`, `ctrl+d`, `ctrl+s`, `ctrl+q`, `ctrl+l`, alt/meta).
- Prefer `home/end` and `pgup/pgdn` as replacements for `g/G` and `n/p` since they are non‑printable and already conventional in TUIs.

### Phase 2: Unified App Behavior
- Update `internal/tui/unified_app.go`:
  - Remove dashboard tab switching on `TabCycle` and numeric sections.
  - Add tab switching bindings for `ctrl+left`/`ctrl+right` and fallback `ctrl+pgup`/`ctrl+pgdn`.
  - Ensure help overlay shows new bindings.

### Research Insights (Phase 2)
- If ctrl+arrow isn’t emitted in a terminal, ctrl+pgup/pgdn should still work. Help overlay should present both to reduce support friction.
- Ensure help overlay itself does not require printable keys to close (keep Esc and F1 toggles consistent).

### Phase 3: View‑Specific Shortcuts
- Audit and replace printable view keys in `internal/tui/views/*`:
  - `task_review`, `epic_review`, `spec_summary`, `task_detail`, `research_overlay`, `kickoff`, `bigend`, `pollard`, `coldwine`, `gurgeh`.
  - Replace toggle/group/view‑specific actions with non‑printable/modifier equivalents (e.g., `enter`, `ctrl+g`, `ctrl+t`, function keys).
- Update each view’s `ShortHelp()` and `FullHelp()` to reflect new bindings.
- Update any user‑facing shortcut hints embedded in chat/system messages (e.g., kickoff hints referencing `Ctrl+G`, `Ctrl+S`, or `n`).

### Research Insights (Phase 3)
- Prefer function keys for view‑specific actions that don’t map to standard navigation to avoid collisions and ensure clarity in help overlays.
- When multiple actions compete, prioritize single‑action keys for destructive or irreversible operations (e.g., F7/F8) and keep primary action on Enter.

### Phase 4: Documentation + Tests
- Update `docs/tui/SHORTCUTS.md`:
  - Remove “Common Letter Bindings” section or replace with “Non‑Printable Only” guidance.
  - Document new global binding set and fallback keys.
- Update tests referencing old key help:
  - `internal/tui/views/views_test.go`
  - Any view‑specific tests expecting printable help strings.

### Test Strategy (Additions)
- **Common keymap test**: validate `pkg/tui/keys.go` exposes no printable keys in `CommonKeys` bindings.
- **Help overlay test**: ensure `internal/tui/unified_app.go` help output contains only non‑printable keys.
- **View help test**: per‑view `ShortHelp`/`FullHelp` should not expose printable keys.
- **Tab behavior test**: ensure `TabCycle` no longer switches dashboard tabs (tabs only via ctrl+left/right or ctrl+pgup/pgdn).

### Research Insights (Phase 4)
- Ensure help overlays for each view mention Tab/Shift+Tab for focus so users discover pane navigation without printable keys.
- Verify tests cover the new help text (F1, ctrl+f, ctrl+r, ctrl+pgup/pgdn) to prevent regressions.

## Dependencies & Risks
- **Terminal compatibility** for Ctrl+Left/Ctrl+Right varies; fallback is required.
- **Discoverability**: removing printable keys increases reliance on help overlay.
- **Scope creep**: some views may rely on printable actions for quick toggles; must ensure replacements are ergonomic.

### Research Insights
- Avoid alt/meta bindings due to Esc prefix collisions; it can interfere with existing Esc/back behavior.
- Consider an explicit “show help” hint in the footer when no printable keys exist for quick discovery.

## Success Metrics
- No accidental navigation or toggles when typing in chat input.
- Help overlay accurately reflects all available actions.
- Users can switch tabs and panes without collisions.

## References & Research
- Brainstorm: `docs/brainstorms/2026-01-29-unified-tui-shortcuts-brainstorm.md`
- Common keys: `pkg/tui/keys.go`
- Unified app help bindings: `internal/tui/unified_app.go`
- Shortcut guide: `docs/tui/SHORTCUTS.md`
- Views using common keys: `internal/tui/views/*`
- Prior UI layout note: `docs/solutions/ui-bugs/tui-breadcrumb-hidden-by-oversized-child-view-20260127.md`

### Institutional Learnings Review
- **Relevant matches:** none directly about keybindings.
- **Nearby lessons:** layout sizing bugs in `internal/tui` and `pkg/tui` suggest being careful when help overlays or header/footer height calculations change (`docs/solutions/ui-bugs/tui-breadcrumb-hidden-by-oversized-child-view-20260127.md`, `docs/solutions/ui-bugs/tui-dimension-mismatch-splitlayout-20260126.md`).
