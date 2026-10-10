package typesense

import (
	"context"

	"typesense/api"
)

// Alias provides operations on a named collection alias.
type Alias struct {
	client api.ClientWithResponsesInterface
	name   string
}

// Retrieve returns the alias and the collection it points to.
func (a *Alias) Retrieve(ctx context.Context) (*api.CollectionAlias, error) {
	response, err := a.client.GetAliasWithResponse(ctx, a.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

// Delete removes the alias without deleting its collection.
func (a *Alias) Delete(ctx context.Context) (*api.CollectionAlias, error) {
	response, err := a.client.DeleteAliasWithResponse(ctx, a.name)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
