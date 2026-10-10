package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

// Keys provides API key creation and listing.
type Keys struct {
	apiClient api.ClientWithResponsesInterface
}

// Create creates an API key with the specified permissions.
func (k *Keys) Create(ctx context.Context, schema *api.APIKeySchema) (*api.APIKey, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: API key schema must not be nil")
	}
	response, err := k.apiClient.CreateKeyWithResponse(ctx, api.CreateKeyJSONRequestBody(*schema))
	if err != nil {
		return nil, err
	}
	if response.JSON201 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON201, nil
}

// Retrieve lists metadata about all API keys.
func (k *Keys) Retrieve(ctx context.Context) ([]*api.APIKey, error) {
	response, err := k.apiClient.GetKeysWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Keys, nil
}
