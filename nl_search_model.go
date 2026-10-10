package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

/* NLSearchModel manages a natural-language search model selected by ID. */
type NLSearchModel struct {
	apiClient api.ClientWithResponsesInterface
	modelID   string
}

/* Retrieve returns the model configuration. */
func (n *NLSearchModel) Retrieve(ctx context.Context) (*api.NLSearchModelSchema, error) {
	response, err := n.apiClient.RetrieveNLSearchModelWithResponse(ctx, n.modelID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Update changes the model configuration. */
func (n *NLSearchModel) Update(ctx context.Context, model *api.NLSearchModelUpdateSchema) (*api.NLSearchModelSchema, error) {
	if model == nil {
		return nil, fmt.Errorf("typesense: natural-language search model update must not be nil")
	}
	response, err := n.apiClient.UpdateNLSearchModelWithResponse(ctx, n.modelID, *model)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Delete removes the model configuration. */
func (n *NLSearchModel) Delete(ctx context.Context) (*api.NLSearchModelDeleteSchema, error) {
	response, err := n.apiClient.DeleteNLSearchModelWithResponse(ctx, n.modelID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
