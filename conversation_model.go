package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

// ConversationModel provides operations on a conversation model configuration.
type ConversationModel struct {
	apiClient api.ClientWithResponsesInterface
	modelID   string
}

// Retrieve returns the conversation model configuration.
func (c *ConversationModel) Retrieve(ctx context.Context) (*api.ConversationModelSchema, error) {
	response, err := c.apiClient.RetrieveConversationModelWithResponse(ctx, c.modelID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Update changes the conversation model configuration.
func (c *ConversationModel) Update(ctx context.Context, schema *api.ConversationModelUpdateSchema) (*api.ConversationModelSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: conversation model update schema must not be nil")
	}
	response, err := c.apiClient.UpdateConversationModelWithResponse(ctx, c.modelID, api.UpdateConversationModelJSONRequestBody(*schema))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Delete removes the conversation model configuration.
func (c *ConversationModel) Delete(ctx context.Context) (*api.ConversationModelSchema, error) {
	response, err := c.apiClient.DeleteConversationModelWithResponse(ctx, c.modelID)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
