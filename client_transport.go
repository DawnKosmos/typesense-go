package typesense

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

type clientNode struct {
	url         *url.URL
	unhealthyAt time.Time
}

type clientTransport struct {
	client              *http.Client
	breaker             *gobreaker.CircuitBreaker
	mu                  sync.Mutex
	nodes               []*clientNode
	nearest             *clientNode
	next                int
	retries             int
	retryInterval       time.Duration
	healthcheckInterval time.Duration
}

func newClientTransport(client *http.Client, config *ClientConfig) *clientTransport {
	t := &clientTransport{client: client, retries: config.NumRetries, retryInterval: config.RetryInterval, healthcheckInterval: config.HealthcheckInterval}
	for _, address := range config.Nodes {
		parsed, _ := url.Parse(address)
		t.nodes = append(t.nodes, &clientNode{url: parsed})
	}
	if config.NearestNode != "" {
		parsed, _ := url.Parse(config.NearestNode)
		t.nearest = &clientNode{url: parsed}
	}
	if t.retries == 0 {
		t.retries = len(t.nodes)
		if t.nearest != nil {
			t.retries++
		}
	}
	readyToTrip := config.CircuitBreakerReadyToTrip
	if readyToTrip == nil {
		readyToTrip = defaultCircuitBreakerReadyToTrip
	}
	t.breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: config.CircuitBreakerName, MaxRequests: config.CircuitBreakerMaxRequests,
		Interval: config.CircuitBreakerInterval, Timeout: config.CircuitBreakerTimeout,
		ReadyToTrip: readyToTrip, OnStateChange: config.CircuitBreakerOnStateChange,
	})
	return t
}

// Do preserves streaming bodies; retrying a body requires a GetBody replay function.
func (t *clientTransport) Do(req *http.Request) (*http.Response, error) {
	result, err := t.breaker.Execute(func() (any, error) {
		rsp, err := t.do(req)
		if err == nil && rsp.StatusCode >= 500 {
			return rsp, fmt.Errorf("typesense: HTTP %d", rsp.StatusCode)
		}
		return rsp, err
	})
	rsp, _ := result.(*http.Response)
	// HTTP errors remain responses so the generated parser can expose their bodies.
	if rsp != nil && err != nil {
		return rsp, nil
	}
	return rsp, err
}

func (t *clientTransport) do(req *http.Request) (*http.Response, error) {
	attempts := t.retries + 1
	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		cloned := req.Clone(req.Context())
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			cloned.Body = body
		}
		node := t.nextNode()
		if node != nil {
			cloned.URL.Scheme, cloned.URL.Host = node.url.Scheme, node.url.Host
			cloned.Host = node.url.Host
		}
		rsp, err := t.client.Do(cloned)
		if rsp != nil && err != nil {
			rsp.Body.Close()
			rsp = nil
		}
		failed := err != nil || rsp.StatusCode >= 500
		t.setHealth(node, failed)
		if !failed || attempt == attempts-1 {
			return rsp, err
		}
		if rsp != nil {
			rsp.Body.Close()
		}
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		timer := time.NewTimer(t.retryInterval)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}
	return nil, fmt.Errorf("typesense: no request attempts")
}

func (t *clientTransport) nextNode() *clientNode {
	t.mu.Lock()
	defer t.mu.Unlock()
	available := func(n *clientNode) bool {
		return n.unhealthyAt.IsZero() || time.Since(n.unhealthyAt) >= t.healthcheckInterval
	}
	if t.nearest != nil && available(t.nearest) {
		return t.nearest
	}
	for range t.nodes {
		node := t.nodes[t.next%len(t.nodes)]
		t.next = (t.next + 1) % len(t.nodes)
		if available(node) {
			return node
		}
	}
	if len(t.nodes) > 0 {
		return t.nodes[t.next%len(t.nodes)]
	}
	return t.nearest
}

func (t *clientTransport) setHealth(node *clientNode, failed bool) {
	if node == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if failed {
		node.unhealthyAt = time.Now()
	} else {
		node.unhealthyAt = time.Time{}
	}
}
