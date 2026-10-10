package typesense_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func TestClientConfiguration(t *testing.T) {
	t.Run("NodesOnlyConfiguration", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}`))
		}))
		defer server.Close()
		client, err := typesense.NewClient(typesense.WithNodes([]string{server.URL}))
		requireNoError(t, err)
		ok, err := client.Health(t.Context(), data.RequestTimeout)
		requireNoError(t, err)
		if !ok {
			t.Fatal("node-only configuration did not select the first node")
		}
	})
	t.Run("InvalidConfiguration", func(t *testing.T) {
		for _, options := range [][]typesense.ClientOption{
			nil,
			{typesense.WithServer("localhost:8108")},
			{typesense.WithServer("ftp://localhost")},
			{typesense.WithServer("http://user:password@localhost")},
			{typesense.WithServer("http://localhost?query=1")},
			{typesense.WithServer("http://localhost"), typesense.WithNumRetries(-1)},
			{typesense.WithServer("http://localhost"), typesense.WithConnectionTimeout(-time.Second)},
		} {
			if _, err := typesense.NewClient(options...); err == nil {
				t.Fatal("invalid client configuration accepted")
			}
		}
	})
	t.Run("AuthenticationRetriesAndBodyReplay", func(t *testing.T) {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-TYPESENSE-API-KEY") != data.APIKey {
				t.Error("API key missing")
			}
			if r.Method != http.MethodPost || r.URL.Path != "/collections" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			var schema api.CollectionSchema
			if err := json.NewDecoder(r.Body).Decode(&schema); err != nil || schema.Name != data.CollectionsName {
				t.Errorf("request body not replayed: %+v %v", schema, err)
			}
			w.Header().Set("Content-Type", "application/json")
			if requests.Add(1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"message":"retry"}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"name":"collections","fields":[]}`))
		}))
		defer server.Close()
		client, err := typesense.NewClient(typesense.WithServer(server.URL), typesense.WithAPIKey(data.APIKey), typesense.WithCustomHTTPClient(server.Client()), typesense.WithNumRetries(1), typesense.WithRetryInterval(time.Millisecond))
		requireNoError(t, err)
		_, err = client.Collections().Create(t.Context(), &api.CollectionSchema{Name: data.CollectionsName, Fields: []api.Field{}})
		requireNoError(t, err)
		if requests.Load() != 2 {
			t.Fatalf("expected 2 attempts, got %d", requests.Load())
		}
	})
	t.Run("NodeFailover", func(t *testing.T) {
		var failed, healthy atomic.Int32
		bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { failed.Add(1); w.WriteHeader(503) }))
		defer bad.Close()
		good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			healthy.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}`))
		}))
		defer good.Close()
		client, err := typesense.NewClient(typesense.WithNearestNode(bad.URL), typesense.WithNodes([]string{good.URL}), typesense.WithRetryInterval(time.Millisecond))
		requireNoError(t, err)
		ok, err := client.Health(t.Context(), data.RequestTimeout)
		requireNoError(t, err)
		if !ok || failed.Load() != 1 || healthy.Load() != 1 {
			t.Fatalf("wrong failover: failed=%d healthy=%d", failed.Load(), healthy.Load())
		}
	})
	t.Run("TimeoutAndCircuitBreaker", func(t *testing.T) {
		slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
		defer slow.Close()
		client, err := typesense.NewClient(typesense.WithServer(slow.URL))
		requireNoError(t, err)
		_, err = client.Health(t.Context(), 20*time.Millisecond)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected deadline error, got %v", err)
		}
		var requests atomic.Int32
		failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(503)
			w.Write([]byte(`{"message":"unavailable"}`))
		}))
		defer failing.Close()
		client, err = typesense.NewClient(typesense.WithServer(failing.URL), typesense.WithCircuitBreakerReadyToTrip(func(counts gobreaker.Counts) bool { return counts.TotalFailures >= 1 }))
		requireNoError(t, err)
		_, err = client.Health(t.Context(), data.RequestTimeout)
		requireHTTPError(t, err, 503)
		_, err = client.Health(t.Context(), data.RequestTimeout)
		if !errors.Is(err, gobreaker.ErrOpenState) || requests.Load() != 1 {
			t.Fatalf("circuit did not open: %v, requests=%d", err, requests.Load())
		}
	})
	t.Run("ConcreteAPIClientInjection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}`))
		}))
		defer server.Close()
		generated, err := api.NewClientWithResponses(server.URL)
		requireNoError(t, err)
		client, err := typesense.NewClient(typesense.WithAPIClient(generated))
		requireNoError(t, err)
		ok, err := client.Health(t.Context(), data.RequestTimeout)
		requireNoError(t, err)
		if !ok {
			t.Fatal("injected client health failed")
		}
	})
}
