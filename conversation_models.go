package typesense

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"typesense/api"
)

// ConversationModels provides conversation model creation and listing.
type ConversationModels struct {
	apiClient api.ClientWithResponsesInterface
}

// Create registers an LLM configuration for conversational search.
func (c *ConversationModels) Create(ctx context.Context, schema *api.ConversationModelCreateSchema) (*api.ConversationModelSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: conversation model schema must not be nil")
	}
	response, err := c.apiClient.CreateConversationModelWithResponse(ctx, api.CreateConversationModelJSONRequestBody(*schema))
	if err != nil {
		return nil, err
	}
	if response.JSON201 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON201, nil
}

// Retrieve lists all conversation model configurations.
func (c *ConversationModels) Retrieve(ctx context.Context) ([]*api.ConversationModelSchema, error) {
	response, err := c.apiClient.RetrieveAllConversationModelsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
