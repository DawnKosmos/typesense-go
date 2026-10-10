package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

/* SynonymSet provides operations on a named synonym set. */
type SynonymSet struct {
	apiClient api.ClientWithResponsesInterface
	name      string
}

/* Items provides listing and upserts for items in this set. */
func (s *SynonymSet) Items() *SynonymSetItems {
	return &SynonymSetItems{apiClient: s.apiClient, setName: s.name}
}

/* Item selects an item in this set by ID. */
func (s *SynonymSet) Item(id string) *SynonymSetItem {
	return &SynonymSetItem{apiClient: s.apiClient, setName: s.name, itemID: id}
}

/* Retrieve returns the synonym set and its items. */
func (s *SynonymSet) Retrieve(ctx context.Context) (*api.SynonymSetSchema, error) {
	response, err := s.apiClient.RetrieveSynonymSetWithResponse(ctx, s.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Upsert creates or replaces the synonym set. */
func (s *SynonymSet) Upsert(ctx context.Context, schema *api.SynonymSetCreateSchema) (*api.SynonymSetSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: synonym set schema must not be nil")
	}
	response, err := s.apiClient.UpsertSynonymSetWithResponse(ctx, s.name, api.UpsertSynonymSetJSONRequestBody(*schema))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Delete removes the synonym set. */
func (s *SynonymSet) Delete(ctx context.Context) (*api.SynonymSetDeleteSchema, error) {
	response, err := s.apiClient.DeleteSynonymSetWithResponse(ctx, s.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
