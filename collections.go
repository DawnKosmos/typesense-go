package typesense

import (
	"context"
	"fmt"
	"strings"

	"typesense/api"
)

// Collections provides collection creation and listing operations.
type Collections struct {
	apiClient api.ClientWithResponsesInterface
}

// CreateFromStruct creates a collection from a named struct's JSON names and tyse options.
func (c *Collections) CreateFromStruct(ctx context.Context, obj any) (*api.CollectionResponse, error) {
	schema, err := ToSchema(obj)
	if err != nil {
		return nil, err
	}
	if schema.Name == "" {
		return nil, fmt.Errorf("typesense: collection name cannot be inferred from an anonymous struct")
	}
	return c.Create(ctx, schema)
}

// Create creates a collection with the given schema.
func (c *Collections) Create(ctx context.Context, schema *api.CollectionSchema) (*api.CollectionResponse, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: collection schema must not be nil")
	}
	body := *schema
	response, err := c.apiClient.CreateCollectionWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if response.JSON201 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON201, nil
}

// Retrieve lists collections using the supplied parameters.
func (c *Collections) Retrieve(ctx context.Context, params *api.GetCollectionsParams) ([]*api.CollectionResponse, error) {
	response, err := c.apiClient.GetCollectionsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
