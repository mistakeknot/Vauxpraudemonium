package inbox

import (
	"context"
	"log"
	"time"
)

const defaultInterval = 2 * time.Second

// Processor handles inbox processing.
type Processor interface {
	ProcessInbox(ctx context.Context) error
}

// Handler polls an inbox processor on an interval.
type Handler struct {
	processor Processor
	interval  time.Duration
}

// NewHandler creates a new inbox handler.
func NewHandler(processor Processor, interval time.Duration) *Handler {
	if interval <= 0 {
		interval = defaultInterval
	}
	return &Handler{processor: processor, interval: interval}
}

// Run starts polling until the context is canceled.
func (h *Handler) Run(ctx context.Context) {
	if h == nil || h.processor == nil {
		return
	}
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		if err := h.processor.ProcessInbox(ctx); err != nil {
			log.Printf("[pollard] inbox: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
