package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

/* CurationSetItem manages an item selected by set name and ID. */
type CurationSetItem struct {
	apiClient api.ClientWithResponsesInterface
	setName   string
	itemID    string
}

/* Retrieve returns the item. */
func (c *CurationSetItem) Retrieve(ctx context.Context) (*api.CurationItemSchema, error) {
	response, err := c.apiClient.RetrieveCurationSetItemWithResponse(ctx, c.setName, c.itemID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Upsert creates or replaces the item. */
func (c *CurationSetItem) Upsert(ctx context.Context, schema *api.CurationItemCreateSchema) (*api.CurationItemSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: curation item schema must not be nil")
	}
	response, err := c.apiClient.UpsertCurationSetItemWithResponse(ctx, c.setName, c.itemID, *schema)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Delete removes the item. */
func (c *CurationSetItem) Delete(ctx context.Context) (*api.CurationItemDeleteSchema, error) {
	response, err := c.apiClient.DeleteCurationSetItemWithResponse(ctx, c.setName, c.itemID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
