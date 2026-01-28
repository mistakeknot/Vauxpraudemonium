package signals

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mistakeknot/autarch/pkg/httpapi"
)

func TestServerPublishHappyPath(t *testing.T) {
	broker := NewBroker()
	srv := NewServer(broker)
	srv.routes()

	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	sub := broker.Subscribe(nil)
	defer sub.Close()

	sig := Signal{
		Type:   SignalAssumptionDecayed,
		Source: "gurgeh",
		Title:  "Assumption decayed",
		Detail: "Confidence dropped",
	}
	body, err := json.Marshal(sig)
	if err != nil {
		t.Fatalf("marshal signal: %v", err)
	}

	resp, err := http.Post(ts.URL+"/api/signals", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", resp.StatusCode)
	}

	var env httpapi.Envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !env.OK {
		t.Fatalf("expected ok response, got error: %+v", env.Error)
	}

	select {
	case got := <-sub.Chan():
		if got.Type != sig.Type || got.Source != sig.Source || got.Title != sig.Title {
			t.Fatalf("unexpected signal: %+v", got)
		}
		if got.CreatedAt.IsZero() {
			t.Fatalf("expected created_at to be set")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected signal to be published")
	}
}

func TestServerPublishRejectsInvalidJSON(t *testing.T) {
	srv := NewServer(NewBroker())
	srv.routes()
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/signals", "application/json", bytes.NewBufferString("{bad json"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestServerPublishRejectsMissingFields(t *testing.T) {
	srv := NewServer(NewBroker())
	srv.routes()
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/signals", "application/json", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}
