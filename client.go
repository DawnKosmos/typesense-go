package typesense

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"typesense/api"

	"github.com/sony/gobreaker"
)

func Version() string {
	return api.Version
}

// Client provides Typesense operations with configured authentication and transport.
type Client struct {
	c         api.ClientWithResponsesInterface
	apiConfig *ClientConfig
}

// Alias selects a collection alias by name.
func (c *Client) Alias(name string) *Alias {
	return &Alias{client: c.c, name: name}
}

// Aliases provides operations on all collection aliases.
func (c *Client) Aliases() *Aliases {
	return &Aliases{client: c.c}
}

// Analytics provides analytics events and rules operations.
func (c *Client) Analytics() *Analytics {
	return &Analytics{apiClient: c.c}
}

// Conversations provides conversation model configuration operations.
func (c *Client) Conversations() *Conversations {
	return &Conversations{apiClient: c.c}
}

// CurationSets provides curation set creation, updates, and listing.
func (c *Client) CurationSets() *CurationSets {
	return &CurationSets{apiClient: c.c}
}

// CurationSet selects a curation set by name.
func (c *Client) CurationSet(name string) *CurationSet {
	return &CurationSet{apiClient: c.c, name: name}
}

// Keys provides API key creation and listing.
func (c *Client) Keys() *Keys {
	return &Keys{apiClient: c.c}
}

// Key selects an API key by its numeric ID.
func (c *Client) Key(id int64) *Key {
	return &Key{apiClient: c.c, keyID: id}
}

// Metrics provides server resource usage metrics.
func (c *Client) Metrics() *Metrics {
	return &Metrics{apiClient: c.c}
}

/* MultiSearch provides batched and union searches across collections. */
func (c *Client) MultiSearch() *MultiSearch {
	return &MultiSearch{apiClient: c.c}
}

/* NLSearchModels provides natural-language search model creation and listing. */
func (c *Client) NLSearchModels() *NLSearchModels {
	return &NLSearchModels{apiClient: c.c}
}

/* NLSearchModel selects a natural-language search model by ID. */
func (c *Client) NLSearchModel(id string) *NLSearchModel {
	return &NLSearchModel{apiClient: c.c, modelID: id}
}

/* Operations provides snapshot creation and cluster leader election. */
func (c *Client) Operations() *Operations {
	return &Operations{apiClient: c.c}
}

/* Presets provides saved search parameter creation, updates, and listing. */
func (c *Client) Presets() *Presets {
	return &Presets{apiClient: c.c}
}

/* Preset selects a saved search configuration by name. */
func (c *Client) Preset(name string) *Preset {
	return &Preset{apiClient: c.c, name: name}
}

/* Stats provides API request rate and latency statistics. */
func (c *Client) Stats() *Stats {
	return &Stats{apiClient: c.c}
}

/* Stemming provides custom word-to-root dictionary operations. */
func (c *Client) Stemming() *Stemming {
	return &Stemming{apiClient: c.c}
}

/* Stopwords provides stopword set creation, updates, and listing. */
func (c *Client) Stopwords() *Stopwords {
	return &Stopwords{apiClient: c.c}
}

/* Stopword selects a stopword set by ID. */
func (c *Client) Stopword(id string) *Stopword {
	return &Stopword{apiClient: c.c, setID: id}
}

/* SynonymSets provides synonym set creation, updates, and listing. */
func (c *Client) SynonymSets() *SynonymSets {
	return &SynonymSets{apiClient: c.c}
}

/* SynonymSet selects a synonym set by name. */
func (c *Client) SynonymSet(name string) *SynonymSet {
	return &SynonymSet{apiClient: c.c, name: name}
}

// Collections provides collection creation and listing operations.
func (c *Client) Collections() *Collections {
	return &Collections{apiClient: c.c}
}

// Collection selects a collection with response type T.
func (c *Client) Collection[T any](name string) *Collection[T] {
	return &Collection[T]{apiClient: c.c, name: name}
}

// GenericCollection selects a collection with the response type T.
// Prefer client.Collection[T](name) on Go 1.27 and later.
func GenericCollection[T any](client *Client, name string) *Collection[T] {
	return client.Collection[T](name)
}

const (
	defaultRetryInterval       = 100 * time.Millisecond
	defaultHealthcheckInterval = time.Minute
	defaultConnectionTimeout   = 5 * time.Second
	defaultCircuitBreakerName  = "typesenseClient"
)

func defaultCircuitBreakerReadyToTrip(counts gobreaker.Counts) bool {
	return counts.Requests > 100 && float64(counts.TotalFailures)/float64(counts.Requests) > 0.5
}

// NewClient initializes the API client. An injected API client keeps its own configuration.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{apiConfig: &ClientConfig{
		RetryInterval:             defaultRetryInterval,
		HealthcheckInterval:       defaultHealthcheckInterval,
		ConnectionTimeout:         defaultConnectionTimeout,
		CircuitBreakerName:        defaultCircuitBreakerName,
		CircuitBreakerMaxRequests: 50,
		CircuitBreakerInterval:    2 * time.Minute,
		CircuitBreakerTimeout:     time.Minute,
		CircuitBreakerReadyToTrip: defaultCircuitBreakerReadyToTrip,
	}}
	for _, opt := range opts {
		opt(c)
	}
	if c.c != nil {
		return c, nil
	}
	config := c.apiConfig
	if config.NumRetries < 0 || config.RetryInterval < 0 || config.HealthcheckInterval < 0 || config.ConnectionTimeout < 0 || config.CircuitBreakerInterval < 0 || config.CircuitBreakerTimeout < 0 {
		return nil, fmt.Errorf("typesense: retry counts and durations must be nonnegative")
	}
	server := config.ServerURL
	if server == "" {
		server = config.NearestNode
	}
	if server == "" && len(config.Nodes) > 0 {
		server = config.Nodes[0]
	}
	if server == "" {
		return nil, fmt.Errorf("typesense: provide a server URL, nearest node, or nodes")
	}
	addresses := append([]string{server}, config.Nodes...)
	if config.NearestNode != "" {
		addresses = append(addresses, config.NearestNode)
	}
	for _, address := range addresses {
		parsed, err := url.Parse(address)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("typesense: invalid server or node URL %q", address)
		}
	}
	httpClient := config.CustomHTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.ConnectionTimeout}
	}
	transport := newClientTransport(httpClient, config)
	generated, err := api.NewClientWithResponses(server,
		api.WithHTTPClient(transport),
		api.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("X-TYPESENSE-API-KEY", config.APIKey)
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("typesense: initialize API client: %w", err)
	}
	c.c = generated
	return c, nil
}
