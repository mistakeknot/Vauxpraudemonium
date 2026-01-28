package signals

import (
	"net/http"
	"strings"
	"time"

	"github.com/mistakeknot/autarch/pkg/httpapi"
	"github.com/mistakeknot/autarch/pkg/netguard"
)

// Server wraps a Broker with HTTP + WS endpoints.
type Server struct {
	broker *Broker
	mux    *http.ServeMux
	srv    *http.Server
}

// NewServer creates a new signals server.
func NewServer(broker *Broker) *Server {
	if broker == nil {
		broker = NewBroker()
	}
	return &Server{broker: broker, mux: http.NewServeMux()}
}

// Broker returns the broker instance.
func (s *Server) Broker() *Broker {
	return s.broker
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	if err := netguard.EnsureLocalOnly(addr); err != nil {
		return err
	}
	s.routes()
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	return s.srv.ListenAndServe()
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ws", s.handleWS)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, httpapi.ErrInvalidRequest, "method not allowed", nil, false)
		return
	}
	httpapi.WriteOK(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, httpapi.ErrInvalidRequest, "method not allowed", nil, false)
		return
	}
	// Optional filter: ?types=competitor_shipped,assumption_decayed
	var types []SignalType
	if v := strings.TrimSpace(r.URL.Query().Get("types")); v != "" {
		for _, t := range strings.Split(v, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				types = append(types, SignalType(trimmed))
			}
		}
	}
	s.broker.ServeWS(w, r, types)
}
