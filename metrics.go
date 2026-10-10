package typesense

import (
	"context"
	"net/http"
	"strings"

	"typesense/api"
)

// Metrics provides server resource usage metrics.
type Metrics struct {
	apiClient api.ClientWithResponsesInterface
}

// Retrieve returns RAM, CPU, disk, and network usage metrics.
func (m *Metrics) Retrieve(ctx context.Context) (map[string]any, error) {
	response, err := m.apiClient.RetrieveMetricsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
