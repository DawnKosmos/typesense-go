package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

/* SynonymSetItem manages an item selected by set name and ID. */
type SynonymSetItem struct {
	apiClient api.ClientWithResponsesInterface
	setName   string
	itemID    string
}

/* Retrieve returns the item. */
func (s *SynonymSetItem) Retrieve(ctx context.Context) (*api.SynonymItemSchema, error) {
	response, err := s.apiClient.RetrieveSynonymSetItemWithResponse(ctx, s.setName, s.itemID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Upsert creates or replaces the item. */
func (s *SynonymSetItem) Upsert(ctx context.Context, schema *api.SynonymItemUpsertSchema) (*api.SynonymItemSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: synonym item schema must not be nil")
	}
	response, err := s.apiClient.UpsertSynonymSetItemWithResponse(ctx, s.setName, s.itemID, *schema)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Delete removes the item. */
func (s *SynonymSetItem) Delete(ctx context.Context) (*api.SynonymItemDeleteSchema, error) {
	response, err := s.apiClient.DeleteSynonymSetItemWithResponse(ctx, s.setName, s.itemID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
