package tests

import (
	"testing"
	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func collections(t *testing.T, client *typesense.Client) {
	selected := createCollection(t, client, data.CollectionsName)
	schema, err := selected.Retrieve(t.Context())
	requireNoError(t, err)
	if schema.Name != data.CollectionsName {
		t.Fatalf("wrong collection name: %s", schema.Name)
	}
	list, err := client.Collections().Retrieve(t.Context(), &api.GetCollectionsParams{GetCollectionsParameters: &api.GetCollectionsParams_GetCollectionsParameters{Limit: 1}})
	requireNoError(t, err)
	if len(list) != 1 || list[0].Name != schema.Name {
		t.Fatalf("unexpected collections: %+v", list)
	}
	_, err = client.Collections().Create(t.Context(), &api.CollectionSchema{Name: data.CollectionsName, Fields: schema.Fields})
	requireHTTPError(t, err, 409)
	_, err = client.Collections().Create(t.Context(), nil)
	if err == nil {
		t.Fatal("nil schema accepted")
	}
	created, err := client.Collections().CreateFromStruct(t.Context(), data.Product{})
	requireNoError(t, err)
	cleanupDelete(t, "inferred collection", client.Collection[data.Product](created.Name).Delete)
	if created.Name != data.StructCollectionName {
		t.Fatalf("unexpected inferred name: %s", created.Name)
	}
	// Collection names and join references are preserved without mutating input.
	reference := &api.CollectionSchema{Name: data.ReferenceCollectionName, Fields: []api.Field{{Name: "product_id", Type: string(api.STRING), Index: optional.True, Reference: data.CollectionsName + ".id"}}}
	referenced, err := client.Collections().Create(t.Context(), reference)
	requireNoError(t, err)
	cleanupDelete(t, "reference collection", client.Collection[data.Product](data.ReferenceCollectionName).Delete)
	if reference.Name != data.ReferenceCollectionName || reference.Fields[0].Reference != data.CollectionsName+".id" {
		t.Fatal("Create mutated the caller schema")
	}
	if referenced.Name != data.ReferenceCollectionName || referenced.Fields[0].Reference != data.CollectionsName+".id" {
		t.Fatalf("incorrect reference: %+v", referenced)
	}
	_, err = client.Collections().CreateFromStruct(t.Context(), struct {
		Name string `json:"name" tyse:"index"`
	}{})
	if err == nil {
		t.Fatal("anonymous collection accepted")
	}
}
