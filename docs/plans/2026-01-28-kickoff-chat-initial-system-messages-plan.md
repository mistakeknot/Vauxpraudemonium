# Kickoff Chat Initial System Messages Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `[bead-id] (Task reference)` — mandatory line tying the plan to the active bead/Task Master item.

**Goal:** Show Kickoff guidance (“What do you want to build?”, “Tips”, “Shortcuts”) as initial chat history above the composer, every time the Kickoff view is entered.

**Architecture:** Add a helper in `KickoffView` that seeds the `ChatPanel` with system messages when the view is created or focused. This keeps content in the chat history (scrollable) and only affects the Kickoff view. Re-seeding on each Kickoff entry ensures guidance is always visible when starting a project.

**Tech Stack:** Go, Bubble Tea, shared `pkg/tui.ChatPanel`.

---

### Task 1: Kickoff chat seeding helper + test

**Files:**
- Modify: `internal/tui/views/kickoff.go`
- Create: `internal/tui/views/kickoff_chat_test.go`

**Step 1: Write the failing test**

```go
func TestKickoffSeedsChatHistory(t *testing.T) {
	v := NewKickoffView()
	msgs := v.ChatMessagesForTest()
	if len(msgs) == 0 {
		t.Fatal("expected seeded chat messages")
	}
	if msgs[0].Role != "system" {
		t.Fatalf("expected system role, got %q", msgs[0].Role)
	}
	if !strings.Contains(msgs[0].Content, "What do you want to build") {
		t.Fatalf("expected prompt message, got %q", msgs[0].Content)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/views -run TestKickoffSeedsChatHistory -v`
Expected: FAIL (helper missing)

**Step 3: Implement minimal seeding**

- Add `seedChat()` method in `KickoffView`:
  - `chatPanel.ClearMessages()`
  - `AddMessage("system", "What do you want to build?")`
  - `AddMessage("system", "Tips: ...")`
  - `AddMessage("system", "Shortcuts: ...")`
- Call `seedChat()` inside `NewKickoffView()` and `Focus()` to re-seed every entry.
- Add a small test-only accessor (e.g., `ChatMessagesForTest() []pkgtui.ChatMessage`) to inspect messages.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/views -run TestKickoffSeedsChatHistory -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/views/kickoff.go internal/tui/views/kickoff_chat_test.go
git commit -m "feat(tui): seed kickoff chat guidance"
```

---

## Final verification

Run:
- `go test ./internal/tui/views -run TestKickoffSeedsChatHistory -v`

---

Plan complete and saved to `docs/plans/2026-01-28-kickoff-chat-initial-system-messages-plan.md`.

Two execution options:
1) Subagent-Driven (this session)
2) Parallel Session (separate)

Which approach?
