package typesense

import (
	"context"

	"typesense/api"
)

/* StemmingDictionary provides access to a named word-to-root dictionary. */
type StemmingDictionary struct {
	apiClient    api.ClientWithResponsesInterface
	dictionaryID string
}

/* Retrieve returns the dictionary's word-to-root mappings. */
func (s *StemmingDictionary) Retrieve(ctx context.Context) (*api.StemmingDictionary, error) {
	response, err := s.apiClient.GetStemmingDictionaryWithResponse(ctx, s.dictionaryID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
