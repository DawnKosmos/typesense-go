package typesense

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/sony/gobreaker"
)

func TestClientOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}
	ready := func(gobreaker.Counts) bool { return true }
	changed := false
	onChange := func(string, gobreaker.State, gobreaker.State) { changed = true }
	client := &Client{apiConfig: &ClientConfig{}}
	for _, option := range []ClientOption{
		WithServer("http://localhost:8108"), WithNearestNode("http://localhost:8109"),
		WithNodes([]string{"http://localhost:8110"}), WithNumRetries(2), WithRetryInterval(time.Second),
		WithHealthcheckInterval(2 * time.Second), WithAPIKey("test"), WithConnectionTimeout(3 * time.Second),
		WithCircuitBreakerName("test-breaker"), WithCircuitBreakerMaxRequests(4),
		WithCircuitBreakerInterval(5 * time.Second), WithCircuitBreakerTimeout(6 * time.Second),
		WithCircuitBreakerReadyToTrip(ready), WithCircuitBreakerOnStateChange(onChange), WithCustomHTTPClient(httpClient),
	} {
		option(client)
	}
	config := client.apiConfig
	if config.ServerURL != "http://localhost:8108" || config.NearestNode != "http://localhost:8109" || config.NumRetries != 2 || config.RetryInterval != time.Second || config.HealthcheckInterval != 2*time.Second || config.APIKey != "test" || config.ConnectionTimeout != 3*time.Second || config.CircuitBreakerName != "test-breaker" || config.CircuitBreakerMaxRequests != 4 || config.CircuitBreakerInterval != 5*time.Second || config.CircuitBreakerTimeout != 6*time.Second || config.CustomHTTPClient != httpClient || !config.CircuitBreakerReadyToTrip(gobreaker.Counts{}) {
		t.Fatalf("options not applied: %+v", config)
	}
	config.CircuitBreakerOnStateChange("test", gobreaker.StateClosed, gobreaker.StateOpen)
	if !changed {
		t.Fatal("state callback not applied")
	}
	original := *config
	WithClientConfig(nil)(client)
	current := *config
	original.CircuitBreakerReadyToTrip, current.CircuitBreakerReadyToTrip = nil, nil
	original.CircuitBreakerOnStateChange, current.CircuitBreakerOnStateChange = nil, nil
	if !reflect.DeepEqual(current, original) || !config.CircuitBreakerReadyToTrip(gobreaker.Counts{}) {
		t.Fatal("nil configuration changed defaults")
	}
	replacement := &ClientConfig{ServerURL: "http://replacement", Nodes: []string{"http://node"}}
	WithClientConfig(replacement)(client)
	replacement.Nodes[0] = "changed"
	if config.ServerURL != "http://replacement" || config.Nodes[0] != "http://node" {
		t.Fatal("WithClientConfig did not copy the nodes")
	}
	nodes := []string{"http://another"}
	WithNodes(nodes)(client)
	nodes[0] = "changed"
	if config.Nodes[0] != "http://another" {
		t.Fatal("WithNodes retained caller-owned storage")
	}
	if defaultCircuitBreakerReadyToTrip(gobreaker.Counts{Requests: 100, TotalFailures: 100}) || defaultCircuitBreakerReadyToTrip(gobreaker.Counts{Requests: 101, TotalFailures: 50}) || !defaultCircuitBreakerReadyToTrip(gobreaker.Counts{Requests: 101, TotalFailures: 51}) {
		t.Fatal("incorrect default circuit threshold")
	}
}
