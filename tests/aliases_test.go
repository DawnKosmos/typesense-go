package tests

import (
	"testing"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func aliases(t *testing.T, client *typesense.Client) {
	createCollection(t, client, data.AliasCollectionName)
	createCollection(t, client, data.AliasTargetName)
	cleanupDelete(t, "alias", client.Alias(data.AliasName).Delete)
	for _, target := range []string{data.AliasCollectionName, data.AliasTargetName} {
		created, err := client.Aliases().Upsert(t.Context(), data.AliasName, &api.CollectionAliasSchema{CollectionName: target})
		requireNoError(t, err)
		if created.Name != data.AliasName || created.CollectionName != target {
			t.Fatalf("unexpected alias: %+v", created)
		}
		retrieved, err := client.Alias(data.AliasName).Retrieve(t.Context())
		requireNoError(t, err)
		if *retrieved != *created {
			t.Fatalf("alias mismatch: %+v", retrieved)
		}
		throughAlias, err := client.Collection[data.Product](data.AliasName).Retrieve(t.Context())
		requireNoError(t, err)
		if throughAlias.Name != target {
			t.Fatalf("alias selected wrong collection: %+v", throughAlias)
		}
	}
	list, err := client.Aliases().Retrieve(t.Context())
	requireNoError(t, err)
	if len(list) != 1 || list[0].Name != data.AliasName {
		t.Fatalf("unexpected aliases: %+v", list)
	}
	_, err = client.Alias(data.AliasName).Delete(t.Context())
	requireNoError(t, err)
	_, err = client.Alias(data.AliasName).Retrieve(t.Context())
	requireHTTPError(t, err, 404)
	_, err = client.Collection[data.Product](data.AliasTargetName).Retrieve(t.Context())
	requireNoError(t, err)
	_, err = client.Aliases().Upsert(t.Context(), data.AliasName, nil)
	if err == nil {
		t.Fatal("nil alias accepted")
	}
}
