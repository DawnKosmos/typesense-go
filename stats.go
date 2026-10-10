package typesense

import (
	"context"

	"typesense/api"
)

/* Stats provides API request rate and latency statistics. */
type Stats struct {
	apiClient api.ClientWithResponsesInterface
}

/* Retrieve returns API statistics for the server's recent reporting window. */
func (s *Stats) Retrieve(ctx context.Context) (*api.APIStatsResponse, error) {
	response, err := s.apiClient.RetrieveAPIStatsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
