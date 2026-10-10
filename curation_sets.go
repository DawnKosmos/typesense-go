package typesense

import (
	"context"
	"net/http"
	"strings"

	"typesense/api"
)

// CurationSets provides curation set creation, updates, and listing.
type CurationSets struct {
	apiClient api.ClientWithResponsesInterface
}

// Upsert creates or updates the named curation set.
func (c *CurationSets) Upsert(ctx context.Context, name string, schema *api.CurationSetCreateSchema) (*api.CurationSetSchema, error) {
	set := &CurationSet{apiClient: c.apiClient, name: name}
	return set.Upsert(ctx, schema)
}

// Retrieve lists all curation sets.
func (c *CurationSets) Retrieve(ctx context.Context) ([]api.CurationSetSchema, error) {
	response, err := c.apiClient.RetrieveCurationSetsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
