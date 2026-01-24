package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/bigend/aggregator"
	"github.com/mistakeknot/vauxpraudemonium/internal/bigend/agentmail"
	"github.com/mistakeknot/vauxpraudemonium/internal/bigend/config"
	"github.com/mistakeknot/vauxpraudemonium/internal/bigend/discovery"
	"github.com/mistakeknot/vauxpraudemonium/internal/bigend/coldwine"
)

type fakeAgg struct {
	state          aggregator.State
	restartCalled  bool
	restartName    string
	restartProject string
	restartType    string
}

func (f *fakeAgg) GetState() aggregator.State { return f.state }
func (f *fakeAgg) Refresh(ctx context.Context) error { return nil }
func (f *fakeAgg) GetProject(path string) *discovery.Project { return nil }
func (f *fakeAgg) GetProjectTasks(projectPath string) (map[string][]coldwine.Task, error) {
	return nil, nil
}
func (f *fakeAgg) GetAgent(name string) *aggregator.Agent { return nil }
func (f *fakeAgg) GetAgentMailAgent(name string) (*agentmail.Agent, error) { return nil, nil }
func (f *fakeAgg) GetAgentMessages(agentID int, limit int) ([]agentmail.Message, error) { return nil, nil }
func (f *fakeAgg) GetAgentReservations(agentID int) ([]agentmail.FileReservation, error) { return nil, nil }
func (f *fakeAgg) GetActiveReservations() ([]agentmail.FileReservation, error) { return nil, nil }
func (f *fakeAgg) NewSession(name, projectPath, agentType string) error { return nil }
func (f *fakeAgg) RenameSession(oldName, newName string) error          { return nil }
func (f *fakeAgg) ForkSession(name, projectPath, agentType string) error { return nil }
func (f *fakeAgg) AttachSession(name string) error                       { return nil }
func (f *fakeAgg) StartMCP(ctx context.Context, projectPath, component string) error  { return nil }
func (f *fakeAgg) StopMCP(projectPath, component string) error           { return nil }

func (f *fakeAgg) RestartSession(name, projectPath, agentType string) error {
	f.restartCalled = true
	f.restartName = name
	f.restartProject = projectPath
	f.restartType = agentType
	return nil
}

func TestRestartSessionEndpoint(t *testing.T) {
	agg := &fakeAgg{
		state: aggregator.State{
			Sessions: []aggregator.TmuxSession{{
				Name:        "demo",
				ProjectPath: "/root/projects/demo",
				AgentType:   "claude",
			}},
		},
	}
	srv := NewServer(config.ServerConfig{}, agg)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/demo/restart", nil)
	w := httptest.NewRecorder()

	srv.handleSessionAction(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !agg.restartCalled {
		t.Fatalf("expected restart call")
	}
	if agg.restartProject != "/root/projects/demo" || agg.restartType != "claude" {
		t.Fatalf("unexpected restart args: %q %q", agg.restartProject, agg.restartType)
	}
}
