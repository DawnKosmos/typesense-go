package typesense

import (
	"context"

	"typesense/api"
)

/* Stopword provides operations on a stopword set selected by ID. */
type Stopword struct {
	apiClient api.ClientWithResponsesInterface
	setID     string
}

/* Retrieve returns the stopword set. */
func (s *Stopword) Retrieve(ctx context.Context) (*api.StopwordsSetSchema, error) {
	response, err := s.apiClient.RetrieveStopwordsSetWithResponse(ctx, s.setID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return &response.JSON200.Stopwords, nil
}

/* Delete removes the stopword set. */
func (s *Stopword) Delete(ctx context.Context) (*api.DeleteStopwordsSet200JSONResponseBody, error) {
	response, err := s.apiClient.DeleteStopwordsSetWithResponse(ctx, s.setID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
