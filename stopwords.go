package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

/* Stopwords provides stopword set creation, updates, and listing. */
type Stopwords struct {
	apiClient api.ClientWithResponsesInterface
}

/* Retrieve lists all stopword sets. */
func (s *Stopwords) Retrieve(ctx context.Context) ([]api.StopwordsSetSchema, error) {
	response, err := s.apiClient.RetrieveStopwordsSetsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Stopwords, nil
}

/* Upsert creates or replaces the stopword set with the given ID. */
func (s *Stopwords) Upsert(ctx context.Context, id string, schema *api.StopwordsSetUpsertSchema) (*api.StopwordsSetSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: stopword set schema must not be nil")
	}
	response, err := s.apiClient.UpsertStopwordsSetWithResponse(ctx, id, api.UpsertStopwordsSetJSONRequestBody(*schema))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
