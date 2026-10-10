package tests

import (
	"testing"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func createSearchKey(t *testing.T, client *typesense.Client, collections ...string) *api.APIKey {
	t.Helper()
	names := make([]string, len(collections))
	for i, name := range collections {
		names[i] = typesense.Prefix(name)
	}
	created, err := client.Keys().Create(t.Context(), &api.APIKeySchema{Description: data.SearchKeyDescription, Actions: []string{"documents:search"}, Collections: names})
	requireNoError(t, err)
	cleanupDelete(t, "search key", client.Key(created.ID).Delete)
	if created.Value == "" {
		t.Fatal("created key has no value")
	}
	return created
}

func keys(t *testing.T, client *typesense.Client) {
	collection := createCollection(t, client, data.ScopedCollectionName)
	seedDocuments(t, collection.Documents(), data.Products())
	created := createSearchKey(t, client, data.ScopedCollectionName)
	key, err := client.Key(created.ID).Retrieve(t.Context())
	requireNoError(t, err)
	if key.ID != created.ID || key.Description != data.SearchKeyDescription {
		t.Fatalf("wrong key metadata: %+v", key)
	}
	list, err := client.Keys().Retrieve(t.Context())
	requireNoError(t, err)
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("wrong keys: %+v", list)
	}
	deleted, err := client.Key(created.ID).Delete(t.Context())
	requireNoError(t, err)
	if deleted.ID != created.ID {
		t.Fatalf("wrong deleted key: %+v", deleted)
	}
	_, err = client.Key(created.ID).Retrieve(t.Context())
	requireHTTPError(t, err, 404)
}
