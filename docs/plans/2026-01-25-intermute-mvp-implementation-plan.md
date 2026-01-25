# Intermute MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `Vauxpraudemonium-1wl` (Task reference)

**Goal:** Stand up a new Intermute repo with a durable REST+WS MVP (agent registry + heartbeats + messaging) and integrate Autarch modules via a shared Go client.

**Architecture:** Intermute is a standalone service with a transport-agnostic core, SQLite event log, REST as the source of truth, and WebSocket for real-time delivery. Bigend owns session I/O; Intermute handles coordination only.

**Tech Stack:** Go 1.24+, net/http, SQLite, WebSocket (nhooyr.io/websocket or equivalent), structured logging.

---

## Preconditions
- User requested **no worktrees**; plan assumes current repo for Autarch and a new sibling repo `../Intermute`.
- Confirm dependency policy for WebSocket library (preferred: `nhooyr.io/websocket`).

---

### Task 1: Create the Intermute repository skeleton

**Files:**
- Create: `../Intermute/README.md`
- Create: `../Intermute/go.mod`
- Create: `../Intermute/.gitignore`
- Create: `../Intermute/cmd/intermute/main.go`
- Create: `../Intermute/internal/server/server.go`

**Step 1: Write the failing test**

Create `../Intermute/internal/server/server_test.go`:

```go
package server

import "testing"

func TestServerStarts(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		// placeholder: should fail until config validation added
		t.Fatalf("expected error without config")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ../Intermute && go test ./internal/server -v`
Expected: FAIL (New not implemented)

**Step 3: Write minimal implementation**

Create `../Intermute/internal/server/server.go` with:

```go
package server

import "fmt"

type Config struct {
	Addr string
}

type Server struct {
	cfg Config
}

func New(cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("addr required")
	}
	return &Server{cfg: cfg}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd ../Intermute && go test ./internal/server -v`
Expected: PASS

**Step 5: Commit**

```bash
git -C ../Intermute add .
git -C ../Intermute commit -m "chore: bootstrap intermute server skeleton"
```

---

### Task 2: Define core domain models and storage interface

**Files:**
- Create: `../Intermute/internal/core/models.go`
- Create: `../Intermute/internal/storage/storage.go`
- Create: `../Intermute/internal/storage/storage_test.go`

**Step 1: Write the failing test**

```go
package storage

import "testing"

func TestAppendEventReturnsCursor(t *testing.T) {
	st := NewInMemory()
	cursor, err := st.AppendEvent(Event{Type: "message.created"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cursor == 0 {
		t.Fatalf("expected non-zero cursor")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ../Intermute && go test ./internal/storage -v`
Expected: FAIL (NewInMemory not implemented)

**Step 3: Write minimal implementation**

Add core types in `models.go` and a minimal in-memory storage in `storage.go` for test scaffolding.

**Step 4: Run test to verify it passes**

Run: `cd ../Intermute && go test ./internal/storage -v`
Expected: PASS

**Step 5: Commit**

```bash
git -C ../Intermute add .
git -C ../Intermute commit -m "feat(core): add core models and storage interface"
```

---

### Task 3: Implement SQLite event log + inbox indexes

**Files:**
- Create: `../Intermute/internal/storage/sqlite/sqlite.go`
- Create: `../Intermute/internal/storage/sqlite/schema.sql`
- Modify: `../Intermute/internal/storage/storage.go`
- Create: `../Intermute/internal/storage/sqlite/sqlite_test.go`

**Step 1: Write the failing test**

```go
func TestSQLiteInboxSinceCursor(t *testing.T) {
	st := NewSQLiteTest(t)
	_, _ = st.AppendEvent(Event{Type: "message.created", Agent: "a"})
	_, _ = st.AppendEvent(Event{Type: "message.created", Agent: "a"})
	msgs, err := st.InboxSince("a", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message after cursor=1, got %d", len(msgs))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ../Intermute && go test ./internal/storage/sqlite -v`
Expected: FAIL (sqlite storage not implemented)

**Step 3: Write minimal implementation**

Implement schema and SQLite storage with:
- `events` append-only table
- `messages` materialized table
- `inbox_index` for `since_cursor`

**Step 4: Run test to verify it passes**

Run: `cd ../Intermute && go test ./internal/storage/sqlite -v`
Expected: PASS

**Step 5: Commit**

```bash
git -C ../Intermute add .
git -C ../Intermute commit -m "feat(storage): sqlite event log and inbox indexes"
```

---

### Task 4: REST API for agent registry + heartbeats

**Files:**
- Create: `../Intermute/internal/http/handlers_agents.go`
- Create: `../Intermute/internal/http/router.go`
- Create: `../Intermute/internal/http/handlers_agents_test.go`

**Step 1: Write the failing test**

```go
func TestRegisterAgent(t *testing.T) {
	srv := newTestServer(t)
	resp := doJSON(t, srv.URL+"/api/agents", map[string]string{"name":"agent-a"})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ../Intermute && go test ./internal/http -v`
Expected: FAIL (handlers missing)

**Step 3: Write minimal implementation**

Implement `POST /api/agents` and `POST /api/agents/{id}/heartbeat` with basic validation and storage.

**Step 4: Run test to verify it passes**

Run: `cd ../Intermute && go test ./internal/http -v`
Expected: PASS

**Step 5: Commit**

```bash
git -C ../Intermute add .
git -C ../Intermute commit -m "feat(api): agent registry and heartbeats"
```

---

### Task 5: REST API for messaging + inbox

**Files:**
- Create: `../Intermute/internal/http/handlers_messages.go`
- Create: `../Intermute/internal/http/handlers_messages_test.go`

**Step 1: Write the failing test**

```go
func TestSendMessageAndFetchInbox(t *testing.T) {
	srv := newTestServer(t)
	send := doJSON(t, srv.URL+"/api/messages", map[string]any{
		"from":"a", "to":[]string{"b"}, "body":"hi",
	})
	if send.StatusCode != 200 {
		t.Fatalf("send failed: %d", send.StatusCode)
	}
	inbox := doGET(t, srv.URL+"/api/inbox/b")
	if inbox.StatusCode != 200 {
		t.Fatalf("inbox failed: %d", inbox.StatusCode)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ../Intermute && go test ./internal/http -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Implement:
- `POST /api/messages`
- `GET /api/inbox/{agent}?since_cursor=...`
- `POST /api/messages/{id}/ack`
- `POST /api/messages/{id}/read`

**Step 4: Run test to verify it passes**

Run: `cd ../Intermute && go test ./internal/http -v`
Expected: PASS

**Step 5: Commit**

```bash
git -C ../Intermute add .
git -C ../Intermute commit -m "feat(api): messaging and inbox"
```

---

### Task 6: WebSocket real-time delivery

**Files:**
- Create: `../Intermute/internal/ws/gateway.go`
- Create: `../Intermute/internal/ws/gateway_test.go`
- Modify: `../Intermute/internal/http/router.go`

**Step 1: Write the failing test**

```go
func TestWSReceivesMessageEvents(t *testing.T) {
	srv := newTestServer(t)
	ws := connectWS(t, srv.URL, "agent-b", 0)
	doJSON(t, srv.URL+"/api/messages", map[string]any{
		"from":"a", "to":[]string{"agent-b"}, "body":"hi",
	})
	event := readEvent(t, ws)
	if event.Type != "message.created" {
		t.Fatalf("expected message.created")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ../Intermute && go test ./internal/ws -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Implement WebSocket hub that tails events and pushes to connected agent streams. Use REST for replay on reconnect.

**Step 4: Run test to verify it passes**

Run: `cd ../Intermute && go test ./internal/ws -v`
Expected: PASS

**Step 5: Commit**

```bash
git -C ../Intermute add .
git -C ../Intermute commit -m "feat(ws): realtime delivery"
```

---

### Task 7: Intermute Go client library

**Files:**
- Create: `../Intermute/client/client.go`
- Create: `../Intermute/client/client_test.go`

**Step 1: Write the failing test**

```go
func TestClientSendAndInbox(t *testing.T) {
	c := New("http://localhost:7338")
	_, err := c.SendMessage(ctx, Message{From:"a", To:[]string{"b"}, Body:"hi"})
	if err == nil {
		t.Fatalf("expected failure without server")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ../Intermute && go test ./client -v`
Expected: FAIL (client not implemented)

**Step 3: Write minimal implementation**

Implement REST client with:
- Register, Heartbeat
- SendMessage, InboxSince, Ack, Read
- WS Subscribe (optional in MVP client)

**Step 4: Run test to verify it passes**

Run: `cd ../Intermute && go test ./client -v`
Expected: PASS

**Step 5: Commit**

```bash
git -C ../Intermute add .
git -C ../Intermute commit -m "feat(client): intermute go client"
```

---

### Task 8: Autarch integration - shared client wiring

**Files:**
- Modify: `go.mod` (Autarch)
- Create: `internal/bigend/intermute/client.go`
- Create: `internal/gurgeh/intermute/client.go`
- Create: `internal/coldwine/intermute/client.go`
- Create: `internal/pollard/intermute/client.go`
- Modify: module startup points to register/heartbeat

**Step 1: Write the failing test**

Add a small unit test in `internal/bigend/intermute/client_test.go` verifying client config is loaded and registration is attempted via interface.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bigend/intermute -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Wire Intermute client into each module with feature-flagged env var `INTERMUTE_URL`. Use `replace ../Intermute` in Autarch `go.mod` for local dev.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/bigend/intermute -v`
Expected: PASS

**Step 5: Commit**

```bash
git add go.mod internal/bigend/intermute
 git commit -m "feat(bigend): intermute client wiring"
```

---

### Task 9: Autarch integration - Pollard inbox migration

**Files:**
- Modify: `internal/pollard/api/scanner.go`
- Modify: `internal/pollard/api/*` as needed
- Create: `internal/pollard/intermute/` if needed

**Step 1: Write the failing test**

Add a test that expects Pollard to call Intermute client when `INTERMUTE_URL` is set instead of writing file inbox.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pollard/api -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Introduce a switch: file-based inbox remains default; Intermute message path is used when configured.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/pollard/api -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/pollard/api
 git commit -m "feat(pollard): intermute-based messaging"
```

---

### Task 10: End-to-end verification

**Files:**
- Modify: `../Intermute/README.md` (add run instructions)
- Modify: `README.md` (Autarch) with Intermute integration notes

**Step 1: Write a failing check**

Add a minimal integration script (or manual steps) documented in README.

**Step 2: Run verification**

Commands:
- `cd ../Intermute && go test ./...`
- `cd /root/projects/Autarch && go test ./...`

Expected: PASS

**Step 3: Commit**

```bash
git -C ../Intermute add README.md
 git -C ../Intermute commit -m "docs: add run instructions"
 git add README.md
 git commit -m "docs: add intermute integration notes"
```

---

## Notes
- If WebSocket dependency policy is strict, implement WS using stdlib + minimal custom handshake (higher effort).
- Keep Idempotency-Key handling on all POSTs; tests should confirm retries do not duplicate events.

