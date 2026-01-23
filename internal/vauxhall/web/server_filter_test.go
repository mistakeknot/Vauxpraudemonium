package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/aggregator"
	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/config"
	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/tmux"
)

type fakeStatusClient struct {
	bySession map[string]tmux.Status
}

func (f fakeStatusClient) DetectStatus(name string) tmux.Status {
	if status, ok := f.bySession[name]; ok {
		return status
	}
	return tmux.StatusUnknown
}

func TestHandleSessionsFiltersByQuery(t *testing.T) {
	agg := &fakeAgg{state: aggregator.State{
		Sessions: []aggregator.TmuxSession{{Name: "codex"}, {Name: "claude"}},
	}}
	srv := NewServer(config.ServerConfig{}, agg)
	srv.statusClient = fakeStatusClient{bySession: map[string]tmux.Status{"codex": tmux.StatusWaiting}}

	req := httptest.NewRequest(http.MethodGet, "/sessions?q=!waiting", nil)
	w := httptest.NewRecorder()
	srv.handleSessions(w, req)
	body := w.Body.String()
	if !strings.Contains(body, "font-mono\">codex") || strings.Contains(body, "font-mono\">claude") {
		t.Fatalf("expected filtered sessions in response")
	}
}

func TestSessionsTemplateShowsQueryValue(t *testing.T) {
	agg := &fakeAgg{state: aggregator.State{Sessions: []aggregator.TmuxSession{{Name: "codex"}}}}
	srv := NewServer(config.ServerConfig{}, agg)
	req := httptest.NewRequest(http.MethodGet, "/sessions?q=codex", nil)
	w := httptest.NewRecorder()
	srv.handleSessions(w, req)
	if !strings.Contains(w.Body.String(), "value=\"codex\"") {
		t.Fatalf("expected query value in input")
	}
}
