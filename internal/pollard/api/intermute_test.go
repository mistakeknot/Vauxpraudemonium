package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	ic "github.com/mistakeknot/intermute/client"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestSendResearchRequestUsesIntermuteWhenConfigured(t *testing.T) {
	called := make(chan struct{}, 1)

	originalClient := intermuteClient
	t.Cleanup(func() { intermuteClient = originalClient })
	intermuteClient = func() (*ic.Client, error) {
		transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method == http.MethodPost && r.URL.Path == "/api/messages" {
				called <- struct{}{}
				payload, _ := json.Marshal(map[string]any{
					"message_id": "msg-1",
					"cursor":     1,
				})
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewReader(payload)),
					Request:    r,
				}
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewReader(nil)),
				Request:    r,
			}, nil
		})
		httpClient := &http.Client{Transport: transport}
		return ic.New("http://intermute.test", ic.WithHTTPClient(httpClient)), nil
	}

	t.Setenv("INTERMUTE_URL", "http://intermute.test")

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
