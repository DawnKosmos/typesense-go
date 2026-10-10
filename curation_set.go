package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

// CurationSet provides operations on a named curation set.
type CurationSet struct {
	apiClient api.ClientWithResponsesInterface
	name      string
}

/* Items provides listing and upserts for items in this set. */
func (c *CurationSet) Items() *CurationSetItems {
	return &CurationSetItems{apiClient: c.apiClient, setName: c.name}
}

/* Item selects an item in this set by ID. */
func (c *CurationSet) Item(id string) *CurationSetItem {
	return &CurationSetItem{apiClient: c.apiClient, setName: c.name, itemID: id}
}

// Retrieve returns the curation set and its items.
func (c *CurationSet) Retrieve(ctx context.Context) (*api.CurationSetSchema, error) {
	response, err := c.apiClient.RetrieveCurationSetWithResponse(ctx, c.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Upsert creates or updates the curation set.
func (c *CurationSet) Upsert(ctx context.Context, schema *api.CurationSetCreateSchema) (*api.CurationSetSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: curation set schema must not be nil")
	}
	response, err := c.apiClient.UpsertCurationSetWithResponse(ctx, c.name, api.UpsertCurationSetJSONRequestBody(*schema))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Delete removes the curation set.
func (c *CurationSet) Delete(ctx context.Context) (*api.CurationSetDeleteSchema, error) {
	response, err := c.apiClient.DeleteCurationSetWithResponse(ctx, c.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
