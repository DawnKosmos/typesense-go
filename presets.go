package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

/* Presets provides saved search parameter creation, updates, and listing. */
type Presets struct {
	apiClient api.ClientWithResponsesInterface
}

/* Retrieve lists all saved search configurations. */
func (p *Presets) Retrieve(ctx context.Context) ([]*api.PresetSchema, error) {
	response, err := p.apiClient.RetrieveAllPresetsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Presets, nil
}

/* Upsert creates or replaces the named search configuration. */
func (p *Presets) Upsert(ctx context.Context, name string, schema *api.PresetUpsertSchema) (*api.PresetSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: preset schema must not be nil")
	}
	response, err := p.apiClient.UpsertPresetWithResponse(ctx, name, api.UpsertPresetJSONRequestBody(*schema))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
