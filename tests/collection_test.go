package tests

import (
	"testing"
	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func collection(t *testing.T, client *typesense.Client) {
	createCollection(t, client, data.ReferenceCollectionName)
	selected := createCollection(t, client, data.CollectionName)
	update := &api.CollectionUpdateSchema{Fields: []api.Field{
		{Name: "rating", Type: string(api.FLOAT), Optional: optional.True},
		{Name: "owner_id", Type: string(api.STRING), Reference: data.ReferenceCollectionName + ".id", Optional: optional.True},
	}}
	response, err := selected.Update(t.Context(), update)
	requireNoError(t, err)
	if len(response.Fields) != 2 || response.Fields[0].Name != "rating" {
		t.Fatalf("unexpected update: %+v", response)
	}
	if update.Fields[1].Reference != data.ReferenceCollectionName+".id" {
		t.Fatal("updating changed the caller's reference")
	}
	schema, err := typesense.GenericCollection[data.Product](client, data.CollectionName).Retrieve(t.Context())
	requireNoError(t, err)
	found, referenceFound := false, false
	for _, field := range schema.Fields {
		if field.Name == "rating" {
			found = true
		}
		if field.Name == "owner_id" && field.Reference == data.ReferenceCollectionName+".id" {
			referenceFound = true
		}
	}
	if !found {
		t.Fatal("updated field missing")
	}
	if !referenceFound {
		t.Fatal("updated reference missing or incorrect")
	}
	t.Run("ClearSettingsWithoutDeletingCollection", func(t *testing.T) {
		_, err := client.SynonymSets().Upsert(t.Context(), data.SynonymSetName, &api.SynonymSetCreateSchema{Items: []api.SynonymItemSchema{}})
		requireNoError(t, err)
		cleanupDelete(t, "synonym set", client.SynonymSet(data.SynonymSetName).Delete)
		_, err = selected.Update(t.Context(), &api.CollectionUpdateSchema{
			Metadata: map[string]any{"owner": "integration"}, SynonymSets: []string{data.SynonymSetName},
		})
		requireNoError(t, err)
		_, err = selected.Update(t.Context(), &api.CollectionUpdateSchema{Metadata: map[string]any{"owner": "updated"}})
		requireNoError(t, err)
		current, err := selected.Retrieve(t.Context())
		requireNoError(t, err)
		if len(current.SynonymSets) != 1 || current.SynonymSets[0] != data.SynonymSetName || current.Metadata["owner"] != "updated" {
			t.Fatalf("omitting synonyms did not preserve the association: %+v", current)
		}
		_, err = selected.Update(t.Context(), &api.CollectionUpdateSchema{SynonymSets: []string{}})
		requireNoError(t, err)
		current, err = selected.Retrieve(t.Context())
		requireNoError(t, err)
		if len(current.SynonymSets) != 0 || current.Metadata["owner"] != "updated" {
			t.Fatalf("clearing synonyms changed metadata or kept the association: %+v", current)
		}
		_, err = selected.Update(t.Context(), &api.CollectionUpdateSchema{Metadata: map[string]any{}})
		requireNoError(t, err)
		current, err = selected.Retrieve(t.Context())
		requireNoError(t, err)
		if len(current.Metadata) != 0 || len(current.Fields) != len(schema.Fields) {
			t.Fatalf("clearing metadata kept values or changed collection fields: %+v", current)
		}
	})
	deleted, err := selected.Delete(t.Context())
	requireNoError(t, err)
	if deleted.Name != data.CollectionName {
		t.Fatalf("wrong deleted collection: %s", deleted.Name)
	}
	_, err = selected.Retrieve(t.Context())
	requireHTTPError(t, err, 404)
	_, err = selected.Update(t.Context(), nil)
	if err == nil {
		t.Fatal("nil update accepted")
	}
}
