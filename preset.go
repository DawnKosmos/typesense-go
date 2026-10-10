package typesense

import (
	"context"

	"typesense/api"
)

/* Preset provides operations on a named saved search configuration. */
type Preset struct {
	apiClient api.ClientWithResponsesInterface
	name      string
}

/* Retrieve returns the saved search configuration. */
func (p *Preset) Retrieve(ctx context.Context) (*api.PresetSchema, error) {
	response, err := p.apiClient.RetrievePresetWithResponse(ctx, p.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Delete removes the saved search configuration. */
func (p *Preset) Delete(ctx context.Context) (*api.PresetDeleteSchema, error) {
	response, err := p.apiClient.DeletePresetWithResponse(ctx, p.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
