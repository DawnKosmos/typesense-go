package typesense

import (
	"net/http"
	"time"
	"typesense/api"

	"github.com/sony/gobreaker"
)

type ClientConfig struct {
	ServerURL                   string
	NearestNode                 string // optional
	Nodes                       []string
	NumRetries                  int
	RetryInterval               time.Duration
	HealthcheckInterval         time.Duration
	APIKey                      string
	ConnectionTimeout           time.Duration
	CircuitBreakerName          string
	CircuitBreakerMaxRequests   uint32
	CircuitBreakerInterval      time.Duration
	CircuitBreakerTimeout       time.Duration
	CircuitBreakerReadyToTrip   func(gobreaker.Counts) bool
	CircuitBreakerOnStateChange func(string, gobreaker.State, gobreaker.State)
	CustomHTTPClient            *http.Client
}

type ClientOption func(*Client)

// WithAPIClient supplies a concrete API client and bypasses transport configuration.
func WithAPIClient(apiClient *api.ClientWithResponses) ClientOption {
	return func(c *Client) {
		c.c = apiClient
	}
}

// WithServer sets the API server URL
func WithServer(serverURL string) ClientOption {
	return func(c *Client) {
		c.apiConfig.ServerURL = serverURL
	}
}

// WithNearestNode sets the Load Balanced endpoint.
func WithNearestNode(nodeURL string) ClientOption {
	return func(c *Client) {
		c.apiConfig.NearestNode = nodeURL
	}
}

// WithNodes sets multiple hostnames to load balance reads & writes across all nodes.
func WithNodes(nodeURLs []string) ClientOption {
	return func(c *Client) {
		c.apiConfig.Nodes = append([]string(nil), nodeURLs...)
	}
}

// WithNumRetries sets the number of retries per request.
// Default value is the number of nodes (+1 if nearestNode is specified).
func WithNumRetries(num int) ClientOption {
	return func(c *Client) {
		c.apiConfig.NumRetries = num
	}
}

// WithRetryInterval sets the wait time between each retry.
// Default value is 100 milliseconds.
func WithRetryInterval(duration time.Duration) ClientOption {
	return func(c *Client) {
		c.apiConfig.RetryInterval = duration
	}
}

// WithHealthcheckInterval sets the wait time for an unhealthy node to become healthy again.
// A node is marked as unhealthy if its response status code is 5xx or has an error (e.g. timeout).
// Default value is 1 minute.
func WithHealthcheckInterval(duration time.Duration) ClientOption {
	return func(c *Client) {
		c.apiConfig.HealthcheckInterval = duration
	}
}

// WithAPIKey sets the API token.
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiConfig.APIKey = apiKey
	}
}

// WithConnectionTimeout sets the connection timeout of http client.
// Default value is 5 seconds.
func WithConnectionTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.apiConfig.ConnectionTimeout = timeout
	}
}

// WithCircuitBreakerName sets the name of the CircuitBreaker.
// Default value is "typesenseClient".
func WithCircuitBreakerName(name string) ClientOption {
	return func(c *Client) {
		c.apiConfig.CircuitBreakerName = name
	}
}

// WithCircuitBreakerMaxRequests sets the maximum number of requests allowed to pass
// through when the CircuitBreaker is half-open. If MaxRequests is 0,
// CircuitBreaker allows only 1 request.
// Default value is 50 requests.
func WithCircuitBreakerMaxRequests(maxRequests uint32) ClientOption {
	return func(c *Client) {
		c.apiConfig.CircuitBreakerMaxRequests = maxRequests
	}
}

// WithCircuitBreakerInterval sets the cyclic period of the closed state for CircuitBreaker
// to clear the internal Counts, described in gobreaker documentation. If Interval is 0,
// CircuitBreaker doesn't clear the internal Counts during the closed state.
// Default value is 2 minutes.
func WithCircuitBreakerInterval(interval time.Duration) ClientOption {
	return func(c *Client) {
		c.apiConfig.CircuitBreakerInterval = interval
	}
}

// WithCircuitBreakerTimeout sets the period of the open state, after which the state of
// CircuitBreaker becomes half-open. If Timeout is 0, the timeout value of CircuitBreaker is set
// to 60 seconds.
// Default value is 1 minute.
func WithCircuitBreakerTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.apiConfig.CircuitBreakerTimeout = timeout
	}
}

// WithCircuitBreakerReadyToTrip sets the function that is called with a copy of Counts
// whenever a request fails in the closed state.
// If ReadyToTrip returns true, CircuitBreaker will be placed into the open state.
// If ReadyToTrip is nil, default ReadyToTrip is used. Default ReadyToTrip returns true when
// number of requests more than 100 and the percent of failures is more than 50 percents.
func WithCircuitBreakerReadyToTrip(readyToTrip func(gobreaker.Counts) bool) ClientOption {
	return func(c *Client) {
		c.apiConfig.CircuitBreakerReadyToTrip = readyToTrip
	}
}

// WithCircuitBreakerOnStateChange sets the function that is called whenever
// the state of CircuitBreaker changes.
func WithCircuitBreakerOnStateChange(onStateChange func(string, gobreaker.State, gobreaker.State)) ClientOption {
	return func(c *Client) {
		c.apiConfig.CircuitBreakerOnStateChange = onStateChange
	}
}

// WithClientConfig allows to pass all configs at once
func WithClientConfig(config *ClientConfig) ClientOption {
	return func(c *Client) {
		if config != nil {
			*c.apiConfig = *config
			c.apiConfig.Nodes = append([]string(nil), config.Nodes...)
		}
	}
}

func WithCustomHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.apiConfig.CustomHTTPClient = client
	}
}
