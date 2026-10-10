package typesense

import (
	"context"
	"net/http"
	"strings"

	"typesense/api"
)

/* CurationSetItems manages the items within a named curation set. */
type CurationSetItems struct {
	apiClient api.ClientWithResponsesInterface
	setName   string
}

/* Retrieve lists all items in the set. */
func (c *CurationSetItems) Retrieve(ctx context.Context) ([]api.CurationItemSchema, error) {
	response, err := c.apiClient.RetrieveCurationSetItemsWithResponse(ctx, c.setName)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Upsert creates or replaces an item selected by ID. */
func (c *CurationSetItems) Upsert(ctx context.Context, id string, schema *api.CurationItemCreateSchema) (*api.CurationItemSchema, error) {
	return (&CurationSetItem{apiClient: c.apiClient, setName: c.setName, itemID: id}).Upsert(ctx, schema)
}
