package typesense

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"typesense/api"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type observedBody struct {
	io.Reader
	closed bool
}

func (b *observedBody) Close() error { b.closed = true; return nil }

func testTransport(t *testing.T, transport http.RoundTripper, retries int) *clientTransport {
	t.Helper()
	return newClientTransport(&http.Client{Transport: transport}, &ClientConfig{NumRetries: retries, RetryInterval: time.Millisecond})
}

func TestTransportRetryBoundaries(t *testing.T) {
	t.Run("NonReplayableBody", func(t *testing.T) {
		attempts := 0
		transport := testTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
			attempts++
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("busy")), Header: make(http.Header)}, nil
		}), 3)
		req, _ := http.NewRequest(http.MethodPost, "http://example", io.NopCloser(strings.NewReader("body")))
		response, err := transport.Do(req)
		if err != nil || attempts != 1 {
			t.Fatalf("stream retried: attempts=%d error=%v", attempts, err)
		}
		response.Body.Close()
	})
	t.Run("BodyReplayFailure", func(t *testing.T) {
		sentinel := errors.New("replay failed")
		body := &observedBody{Reader: strings.NewReader("busy")}
		transport := testTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 503, Body: body, Header: make(http.Header)}, nil
		}), 1)
		req, _ := http.NewRequest(http.MethodPost, "http://example", strings.NewReader("body"))
		req.GetBody = func() (io.ReadCloser, error) { return nil, sentinel }
		_, err := transport.Do(req)
		if !errors.Is(err, sentinel) || !body.closed {
			t.Fatalf("replay error or cleanup lost: %v, closed=%v", err, body.closed)
		}
	})
	t.Run("CancellationBeforeRequest", func(t *testing.T) {
		transport := testTransport(t, roundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("canceled request sent"); return nil, nil }), 1)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example", nil)
		if _, err := transport.Do(req); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation lost: %v", err)
		}
	})
	t.Run("CancellationDuringRequest", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		transport := testTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
			cancel()
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("busy")), Header: make(http.Header)}, nil
		}), 1)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example", nil)
		if _, err := transport.Do(req); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation lost: %v", err)
		}
	})
	t.Run("CancellationDuringRetryDelay", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		transport := testTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
			time.AfterFunc(10*time.Millisecond, cancel)
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("busy")), Header: make(http.Header)}, nil
		}), 1)
		transport.retryInterval = time.Hour
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example", nil)
		if _, err := transport.Do(req); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation lost: %v", err)
		}
	})
	t.Run("RedirectFailure", func(t *testing.T) {
		body := &observedBody{Reader: strings.NewReader("redirect")}
		transport := testTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"http://redirected"}}, Body: body}, nil
		}), 0)
		sentinel := errors.New("redirect refused")
		transport.client.CheckRedirect = func(*http.Request, []*http.Request) error { return sentinel }
		req, _ := http.NewRequest(http.MethodGet, "http://example", nil)
		if _, err := transport.Do(req); !errors.Is(err, sentinel) || !body.closed {
			t.Fatalf("redirect error or body cleanup lost: %v", err)
		}
	})
}

func TestTransportNodeRecovery(t *testing.T) {
	parse := func(address string) *url.URL {
		result, err := url.Parse(address)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	first, second := &clientNode{url: parse("http://first")}, &clientNode{url: parse("http://second")}
	transport := &clientTransport{nodes: []*clientNode{first, second}, healthcheckInterval: time.Hour}
	transport.setHealth(first, true)
	transport.setHealth(second, true)
	if transport.nextNode() != first {
		t.Fatal("all-unhealthy fallback did not choose a node")
	}
	transport.setHealth(second, false)
	if transport.nextNode() != second {
		t.Fatal("healthy node not selected")
	}
	first.unhealthyAt = time.Now().Add(-2 * time.Hour)
	if transport.nextNode() != first {
		t.Fatal("node not eligible after healthcheck interval")
	}
}

type responseOnlyClient struct {
	api.ClientWithResponsesInterface
}

func TestStreamingRequiresRawClient(t *testing.T) {
	documents := &Documents{apiClient: &responseOnlyClient{}}
	if _, err := documents.Export(t.Context(), nil); err == nil {
		t.Fatal("export accepted a response-only client")
	}
	if _, err := documents.ImportJsonl(t.Context(), strings.NewReader("{}"), nil); err == nil {
		t.Fatal("import accepted a response-only client")
	}
	stemming := &StemmingDictionaries{apiClient: &responseOnlyClient{}}
	if _, err := stemming.UpsertJsonl(t.Context(), "dictionary", strings.NewReader("{}")); err == nil {
		t.Fatal("stemming accepted a response-only client")
	}
}
