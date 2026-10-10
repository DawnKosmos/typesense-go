package typesense

import (
	"context"
	"encoding/json"
	"strings"

	"typesense/api"
)

// Document provides operations on an individual document decoded into T.
type Document[T any] struct {
	apiClient      api.ClientWithResponsesInterface
	collectionName string
	documentID     string
}

// Retrieve returns the document.
func (d *Document[T]) Retrieve(ctx context.Context) (T, error) {
	var result T
	response, err := d.apiClient.GetDocumentWithResponse(ctx, d.collectionName, d.documentID)
	if err != nil {
		return result, err
	}
	if response.StatusCode() != 200 || !strings.Contains(response.ContentType(), "json") {
		return result, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	err = json.Unmarshal(response.Body, &result)
	return result, err
}

// Update applies a partial document update and returns the updated document.
func (d *Document[T]) Update(ctx context.Context, document any, params *api.DocumentIndexParameters) (T, error) {
	var result T
	var options *api.UpdateDocumentParams
	if params != nil && params.DirtyValues != nil {
		options = &api.UpdateDocumentParams{DirtyValues: *params.DirtyValues}
	}
	response, err := d.apiClient.UpdateDocumentWithResponse(ctx, d.collectionName, d.documentID, options, document)
	if err != nil {
		return result, err
	}
	if response.StatusCode() != 200 || !strings.Contains(response.ContentType(), "json") {
		return result, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	err = json.Unmarshal(response.Body, &result)
	return result, err
}

// Delete removes the document and returns its previous contents.
func (d *Document[T]) Delete(ctx context.Context) (T, error) {
	var result T
	response, err := d.apiClient.DeleteDocumentWithResponse(ctx, d.collectionName, d.documentID)
	if err != nil {
		return result, err
	}
	if response.StatusCode() != 200 || !strings.Contains(response.ContentType(), "json") {
		return result, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	err = json.Unmarshal(response.Body, &result)
	return result, err
}
