Below is a hard-nosed review with concrete fixes. Overall: the concept is strong (parallel research + spec, progressive disclosure, keyboard-first), but you’re currently underspecified in **(a)** how asynchronous Pollard results are mapped to interview questions, **(b)** how you prevent stale/out-of-order updates, and **(c)** how you actually end up in Bigend with “ready tasks” (you only generate *epics* right now).

---

## 0) Architectural red flags to address first

1. **Keybinding collisions are guaranteed**

   * You currently have **Interview: `1-3` adopt tradeoff** and **Global: `1-4` switch tabs**. That’s a direct conflict.
   * You also use `Tab` as “toggle research overlay” while `Tab` is commonly “focus next field” in TUIs (and is often expected in text-input contexts).

2. **“Land users in Bigend with ready tasks” is not implemented**

   * Your Phase 4 generates epics, but there’s no explicit **story/task generation**, no dependency resolution into a “ready queue,” and no initial task state population.

3. **Async Pollard streaming + question flow can create stale suggestions**

   * If Pollard results arrive after the user answers, you risk:

     * UI flicker / layout shift
     * Suggestions appearing for the wrong question
     * “Default” selections silently changing (UX antipattern)

4. **Bigend’s role conflict (if Bigend is intended read-only)**

   * Your dev guide frames Bigend as “observes, doesn’t control.” This onboarding flow writes specs/tasks. If this onboarding runs *inside Bigend*, you’re changing Bigend’s contract.
   * That’s not necessarily wrong, but it should be explicit: either **Bigend becomes a write-capable orchestrator**, or onboarding lives in a separate “autarch tui” orchestrator.

---

## 1) Architecture: parallel Pollard + Gurgeh (soundness, races, state)

### The parallel approach is sound, but you need explicit “run identity” + cancellation semantics

**Main risk:** Pollard is asynchronous and Gurgeh is stateful. Without a *RunID* and *QuestionID* mapping, you’ll get out-of-order updates and stale UI.

**Actionable architecture fixes:**

**A. Introduce a stable research run identity**

* When kickoff starts Pollard, generate a `ResearchRunID` and attach it to every Pollard update message.
* In Bubble Tea, *every* async message should include `{RunID, ProjectID}` and your model should ignore messages not matching the current active run.

**B. Add cancellation + replacement**

* If the user:

  * switches projects from recents,
  * edits the kickoff prompt materially (if you allow it),
  * or quits onboarding,
    you should `cancel()` the Pollard context.
* Otherwise you’ll keep updating UI for a project the user is no longer in.

**C. Never mutate UI state from goroutines**

* Pollard should publish updates via `tea.Msg` only.
* Your TUI model should remain single-threaded (Bubble Tea’s Update loop) and treat Pollard as a message stream.

**D. Explicit question-scoping to prevent “wrong question” suggestions**
You need a mapping layer such as:

* `QuestionID` (stable per interview step)
* `TopicKey` (e.g., `"platform"`, `"auth"`, `"storage"`, `"deploy"`)
* Pollard insights tagged with topic keys

Then your “relevant findings for current question” becomes deterministic:

* `FindingsFor(QuestionID)` or `FindingsFor(TopicKey)`

**E. Guard against “late-arriving defaults”**
If a question supports “research-informed defaults,” you must define:

* defaults applied **only once** (on question entry) **if user hasn’t interacted**
* once user interacts, defaults are locked unless user explicitly chooses “Apply suggestion”

A simple pattern:

* `QuestionState{Touched bool, DefaultApplied bool, Answer ...}`
* Pollard updates should not auto-modify answers when `Touched == true`

### Race conditions to watch for

* **Overlay toggle while Pollard updates**: ensure overlay rendering is robust if the findings list changes mid-scroll (avoid index out of range).
* **Question navigation while updates arrive**: if you store “current findings” by index, it will drift; store by stable IDs.
* **Program exit**: Pollard goroutines must stop cleanly or you’ll get “send on closed channel” style issues.

---

## 2) UX Flow: progressive disclosure (teaser → Tab → expand)

### Conceptually great for power users, but avoid layout-shift + mode confusion

Power users like dense, skimmable info—your approach aligns. The friction is in **timing** and **screen stability**.

**Friction points and fixes:**

**A. Avoid layout shift**
If the teaser box appears/disappears depending on “relevant findings,” your interview UI will jump.

* Fix: reserve a fixed-height teaser area from the start:

  * When empty: show “Research: running… [Tab]”
  * When available: show the summary
  * When disabled: show “Research: off (toggle in settings)” (optional)

**B. Make the overlay “modal” and clearly labeled**
A Tab-based overlay can feel like a mode switch. That’s fine if it’s obvious.

* Add a strong header: `RESEARCH (Tab/Esc to return)`
* Don’t reuse Enter behavior across modes unless it’s consistent.

**C. Provide a “pin” option for power users**
Many power users would prefer:

* a **side panel** that stays open (if terminal width allows), vs. a full-screen overlay.
  Actionable approach:
* If `width >= threshold`, render research as a right-side panel.
* Else fall back to overlay modal.

**D. Teasers should be “answer-adjacent,” not general**
If teasers are too generic, they become noise. Ensure teaser content is always tied to the current question:

* “For deployment: 3 common paths…”
* “For storage: tradeoffs between sqlite/postgres…”
  Not “market landscape” fluff unless you’re in a high-level section.

**E. Handle “research still running” explicitly at key transitions**
At spec completion:

* If Pollard still running, give a choice:

  * `Enter` generate epics now
  * `r` wait/refresh (or show partial)
  * `Tab` view research
    This avoids the feeling of “I left value on the table.”

**UX antipattern to avoid**

* **Auto-selecting options because research arrived** (silent state change)
* **Overusing popularity stats (“73% of projects”)** without context (sample bias); keep it but label it clearly as observational and sourced.

---

## 3) Implementation phases: missing pieces and overlooked edge cases

### Missing pieces (big ones)

**A. Story/task generation is missing**
You promise “ready tasks” in Bigend, but only generate epics.
Add a phase explicitly:

* **Phase 4b: Epic → Stories/Tasks**

  * generate initial tasks per epic
  * assign dependencies
  * mark tasks with readiness (no blockers)
  * persist to `.coldwine/specs/` and/or `.coldwine/plan/`

**B. Persistence + resume logic**
Onboarding will be interrupted (users quit, terminal closes, etc.).
You need:

* auto-save spec progress (draft)
* resume prompt on next start:

  * “Resume onboarding?” / “Go to dashboard?” / “Edit spec?”

**C. Error/empty states**
Pollard is networked and API-key sensitive. You need UX for:

* missing tokens / rate limits
* hunter failures (partial results)
* offline mode
* “0 findings” but still allow progress

**D. Project creation semantics**
Kickoff “create project, trigger parallel kickoff” leaves questions:

* What’s a “project” in filesystem terms? Where is it created?
* If user selects a recent project, do you:

  * go to Bigend dashboard immediately?
  * resume Gurgeh if spec incomplete?
  * rerun Pollard or reuse cached insights?

**E. Source attribution UX**
You mention attribution by hunter. Consider also:

* stable source identifiers (URL/repo)
* ability to copy/open (even if “open” means print URL)
* de-dup across hunters

### Edge cases to explicitly design for

* **Terminal resize**: overlay/panels must reflow without breaking scroll positions.
* **Long text inputs**: teaser/suggestions shouldn’t steal vertical space excessively.
* **Huge research result sets**: add filtering/search (`/` to search within research).
* **User answers faster than Pollard**: suggestions may arrive after the fact; show them as “FYI” without trying to re-open answered questions.
* **Regenerate epics after edits**: if user edited epics, regeneration must:

  * warn about overwriting
  * or support “regenerate selected epic only”
  * or preserve manual edits with a merge strategy

---

## 4) Data flow: Kickoff → Pollard → Gurgeh → Coldwine → Bigend

### The pipeline is reasonable, but you need clearer contracts and idempotency

**What’s good:**

* Pollard feeds Gurgeh during interview (contextual intelligence)
* Gurgeh outputs a spec
* Coldwine turns spec into execution units
* Bigend is the landing/dashboard

**Where it needs tightening:**

**A. Define canonical artifacts + where they live**
Right now you implicitly span:

* `.gurgeh/specs/*.yaml`
* `.pollard/*`
* `.coldwine/specs/*.yaml`
* Bigend reads everything

Make this explicit in the plan:

* “Spec is canonical in `.gurgeh/specs/`”
* “Epics/tasks canonical in `.coldwine/specs/`”
* “Research snapshots referenced from spec by Insight IDs (or file paths)”

**B. Make research references stable**
If Gurgeh spec says “decision came from research,” store:

* `InsightID` (stable) or a snapshot reference
* not just a hunter name string
  Otherwise, you can’t reliably show “which research led to this” later in Task Detail.

**C. Idempotency**
If a user re-runs onboarding for an existing project:

* do you create another spec?
* do you overwrite epics?
* do you create duplicates?

Add deterministic rules:

* Same ProjectID + SpecID overwrites draft until “finalized”
* Epic generation is “proposal” until accepted; accepted outputs replace prior outputs only after confirmation

**D. Consider using your event spine (`pkg/events`) as glue**
You already have an “event spine” SQLite DB. This is an ideal fit for:

* `ProjectCreated`
* `ResearchRunStarted/Updated/Completed`
* `SpecDraftUpdated/Finalized`
* `EpicsProposed/Accepted`
* `TasksCreated`

This reduces tight coupling and also makes resume/replay easier.

**E. Bigend write vs read-only decision**
If you want Bigend to remain read-only:

* onboarding should be its own “autarch tui” orchestrator that writes artifacts
* Bigend just detects them and displays

If you’re okay with Bigend writing:

* be explicit and update the “observes, doesn’t control” principle
* otherwise future contributors will fight the architecture

---

## 5) Keyboard navigation: intuitiveness + TUI convention alignment

### Current map is close, but needs consistency and conflict resolution

**A. Resolve conflicts**

* **Remove “Global 1-4 switch tabs”** or change it to `Alt+1..4` / `Ctrl+1..4`.
* Keep `1-3` for tradeoff adoption in Interview (that’s efficient).

**B. Rethink `Tab`**
Using `Tab` as a global overlay toggle is workable, but:

* In Kickoff view, Tab is typically focus-cycle (input ↔ recents)
* In text inputs, Tab is sometimes expected for completion

Two good options:

**Option 1 (safer, more conventional):**

* `Ctrl+R` or `Ctrl+I` = toggle research overlay (“R = research”)
* `Tab/Shift+Tab` = focus-cycle
* `Esc` = close overlay/back
* `Ctrl+C` = quit

**Option 2 (keep Tab, but constrain it):**

* `Tab` toggles research only **in Interview view**
* `Tab` focus-cycles in Kickoff/spec/epic screens
* `Esc` always closes modals
* `Ctrl+C` always quits

**C. Fix `q` behavior**
`q` as “Quit/back” will annoy people if a text input is focused.

* Make `Esc` = back/close modal
* Make `q` = quit only when **no text input is focused** (or require confirmation)
* Keep `Ctrl+C` = always quit (standard)

**D. Agent selection keys (`c/x/a`)**

* `c = Claude` is mnemonic
* `a = Aider` is mnemonic
* `x = Codex` is not mnemonic

Better patterns:

* `c` cycle agent, `Enter` start (common pattern)
* Or `d` for co**d**ex, `l` for c**l**aude, `a` for aider
* Or show a small selector: `←→` choose agent, `Enter` start (fast + discoverable)

**E. Add vim keys as optional sugar**
For power users:

* `j/k` scroll lists (in addition to arrows)
* `/` search within research overlay

**F. Always show contextual help**
Your `?` help is good; make it per-view and dynamic:

* show active keymap
* show conflicts resolved by mode (e.g., “1-3 suggestions (Interview only)”)

---

## 6) Testing strategy: sufficiency of verification plan

### Manual plan is fine; automation needs async + race coverage

**What you have is a good start**, but it doesn’t stress the real failure modes: async streaming, cancellation, out-of-order messages, and “resume after restart.”

**Add these tests and checks:**

**A. Run Go race detector in CI for TUIs**

* `go test ./... -race`
  This catches accidental concurrent state mutation (common with streaming updates).

**B. Deterministic fake Pollard in tests**
Create a Pollard stub that emits:

* hunter started
* partial findings
* delayed completion
* error for one hunter
* out-of-order messages (simulate network timing)

Then assert:

* no panics
* UI state remains consistent
* suggestions only attach to the intended question

**C. View-model property tests**
For models like ResearchOverlay:

* scrolling never goes negative
* selected index never exceeds list length after updates
* expanding/collapsing doesn’t corrupt state

**D. Resume/rehydration tests**

* Start onboarding → answer 2 questions → quit
* Relaunch → verify it resumes correctly (or prompts correctly)
  This requires you to implement persistence, which is currently missing from the plan.

**E. Golden tests (snapshot-ish) for rendering**
Not full “pixel perfect,” but validate key strings appear:

* “◉ hunters running…”
* “RESEARCH” header
* ✓/✗ formatting
  These tests catch accidental regressions in key UX affordances.

**F. Integration test: end-to-end artifacts**
Your `internal/tui/integration_test.go` should assert filesystem outputs:

* `.gurgeh/specs/*.yaml` exists and contains recorded decisions + insight refs
* `.coldwine/specs/*.yaml` contains generated epics **and tasks** (after you add task gen)
* Bigend aggregator sees them as “ready”

---

## Recommended plan edits (tight, actionable)

1. **Add a “ResearchRunID + QuestionID mapping” section** (architecture contract)
2. **Add Phase 4b: Stories/Tasks generation** (so Bigend has “ready tasks”)
3. **Add a persistence/resume story** (draft spec + in-progress research)
4. **Fix keymap conflicts** (remove global `1-4`, add `Esc` semantics, reconsider `Tab`)
5. **Decide whether onboarding is a new orchestrator or part of Bigend** (write-capability contract)
6. **Expand tests to cover async + cancellation + resume + -race**

If you want a single “most important” change: **make Pollard integration run-ID scoped and non-destructive** (never auto-change answers after user interaction). That’s the difference between “power-user magic” and “mysterious UI that changes under me.”
