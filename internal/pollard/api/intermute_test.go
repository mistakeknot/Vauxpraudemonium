package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendResearchRequestUsesIntermuteWhenConfigured(t *testing.T) {
	called := make(chan struct{}, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/messages" {
			called <- struct{}{}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message_id": "msg-1",
				"cursor":     1,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv("INTERMUTE_URL", srv.URL)

	payload := ResearchPayload{RequestType: "prd", Vision: "v", Problem: "p"}
	msg, err := SendResearchRequest(t.TempDir(), payload, "praude")
	if err != nil {
		t.Fatalf("send request failed: %v", err)
	}
	if msg == nil || msg.ID == "" {
		t.Fatalf("expected message id")
	}

	select {
	case <-called:
	default:
		t.Fatalf("expected intermute call")
	}
}
