package typesense

import (
	"context"
	"net/http"
	"strings"

	"typesense/api"
)

/* SynonymSets provides synonym set creation, updates, and listing. */
type SynonymSets struct {
	apiClient api.ClientWithResponsesInterface
}

/* Upsert creates or replaces the named synonym set. */
func (s *SynonymSets) Upsert(ctx context.Context, name string, schema *api.SynonymSetCreateSchema) (*api.SynonymSetSchema, error) {
	set := &SynonymSet{apiClient: s.apiClient, name: name}
	return set.Upsert(ctx, schema)
}

/* Retrieve lists all synonym sets. */
func (s *SynonymSets) Retrieve(ctx context.Context) ([]api.SynonymSetSchema, error) {
	response, err := s.apiClient.RetrieveSynonymSetsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
