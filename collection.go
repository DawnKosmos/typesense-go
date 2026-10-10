package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

// Collection provides operations on a named collection with document type T.
type Collection[T any] struct {
	apiClient api.ClientWithResponsesInterface
	name      string
}

// Document selects a document by ID using the collection's document type T.
func (c *Collection[T]) Document(id string) *Document[T] {
	return &Document[T]{apiClient: c.apiClient, collectionName: c.name, documentID: id}
}

// Documents provides operations on documents in this collection.
func (c *Collection[T]) Documents() *Documents {
	return &Documents{apiClient: c.apiClient, collectionName: c.name}
}

// Retrieve returns the collection schema and statistics.
func (c *Collection[T]) Retrieve(ctx context.Context) (*api.CollectionResponse, error) {
	response, err := c.apiClient.GetCollectionWithResponse(ctx, c.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Delete removes the collection and its documents.
func (c *Collection[T]) Delete(ctx context.Context) (*api.CollectionResponse, error) {
	response, err := c.apiClient.DeleteCollectionWithResponse(ctx, c.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Update applies changes to the collection schema.
func (c *Collection[T]) Update(ctx context.Context, schema *api.CollectionUpdateSchema) (*api.CollectionUpdateSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: collection update schema must not be nil")
	}
	body := *schema
	response, err := c.apiClient.UpdateCollectionWithResponse(ctx, c.name, body)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
