// Package embedded provides wrapper for embedded Intermute server.
package embedded

import (
	"github.com/mistakeknot/intermute/pkg/embedded"
)

// Config is a re-export of the embedded server config
type Config = embedded.Config

// Server wraps the embedded Intermute server
type Server struct {
	*embedded.Server
}

// New creates a new embedded Intermute server
func New(cfg Config) (*Server, error) {
	srv, err := embedded.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Server{Server: srv}, nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		// Defaults are applied in embedded.New()
	}
}
