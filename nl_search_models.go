package typesense

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"typesense/api"
)

/* NLSearchModels manages LLM configurations that translate queries into search parameters. */
type NLSearchModels struct {
	apiClient api.ClientWithResponsesInterface
}

/* Retrieve lists all natural-language search model configurations. */
func (n *NLSearchModels) Retrieve(ctx context.Context) ([]api.NLSearchModelSchema, error) {
	response, err := n.apiClient.RetrieveAllNLSearchModelsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* Create registers a natural-language search model configuration. */
func (n *NLSearchModels) Create(ctx context.Context, model *api.NLSearchModelCreateSchema) (*api.NLSearchModelSchema, error) {
	if model == nil {
		return nil, fmt.Errorf("typesense: natural-language search model must not be nil")
	}
	response, err := n.apiClient.CreateNLSearchModelWithResponse(ctx, *model)
	if err != nil {
		return nil, err
	}
	if response.JSON201 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON201, nil
}
