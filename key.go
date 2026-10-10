package typesense

import (
	"context"

	"typesense/api"
)

// Key provides operations on an API key selected by ID.
type Key struct {
	apiClient api.ClientWithResponsesInterface
	keyID     int64
}

// Retrieve returns metadata about the API key.
func (k *Key) Retrieve(ctx context.Context) (*api.APIKey, error) {
	response, err := k.apiClient.GetKeyWithResponse(ctx, k.keyID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Delete revokes the API key.
func (k *Key) Delete(ctx context.Context) (*api.APIKeyDeleteResponse, error) {
	response, err := k.apiClient.DeleteKeyWithResponse(ctx, k.keyID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
