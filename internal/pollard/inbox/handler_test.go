package inbox

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type fakeProcessor struct {
	count int32
}

func (f *fakeProcessor) ProcessInbox(_ context.Context) error {
	atomic.AddInt32(&f.count, 1)
	return nil
}

func TestHandlerRunInvokesProcessor(t *testing.T) {
	fp := &fakeProcessor{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := NewHandler(fp, 10*time.Millisecond)
	go h.Run(ctx)

	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&fp.count) > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected processor to be invoked")
}
