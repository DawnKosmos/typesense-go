package typesense

import (
	"context"
	"net/http"
	"strings"

	"typesense/api"
)

/* SynonymSetItems manages the items within a named synonym set. */
type SynonymSetItems struct {
	apiClient api.ClientWithResponsesInterface
	setName   string
}

/* Retrieve lists all items in the set. */
func (s *SynonymSetItems) Retrieve(ctx context.Context) ([]api.SynonymItemSchema, error) {
	response, err := s.apiClient.RetrieveSynonymSetItemsWithResponse(ctx, s.setName)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Upsert creates or replaces an item selected by ID. */
func (s *SynonymSetItems) Upsert(ctx context.Context, id string, schema *api.SynonymItemUpsertSchema) (*api.SynonymItemSchema, error) {
	return (&SynonymSetItem{apiClient: s.apiClient, setName: s.setName, itemID: id}).Upsert(ctx, schema)
}
