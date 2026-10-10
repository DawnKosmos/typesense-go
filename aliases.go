package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

// Aliases provides collection alias listing and creation operations.
type Aliases struct {
	client api.ClientWithResponsesInterface
}

// Retrieve lists all collection aliases.
func (a *Aliases) Retrieve(ctx context.Context) ([]*api.CollectionAlias, error) {
	response, err := a.client.GetAliasesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Aliases, nil
}

// Upsert creates or updates an alias to point to the given collection.
func (a *Aliases) Upsert(ctx context.Context, name string, schema *api.CollectionAliasSchema) (*api.CollectionAlias, error) {
	if schema == nil {
		return nil, fmt.Errorf("typesense: alias schema must not be nil")
	}
	body := *schema
	body.CollectionName = Prefix(body.CollectionName)
	response, err := a.client.UpsertAliasWithResponse(ctx, Prefix(name), body)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
