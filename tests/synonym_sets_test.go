package tests

import (
	"slices"
	"testing"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func synonymSets(t *testing.T, client *typesense.Client) {
	created, err := client.SynonymSets().Upsert(t.Context(), data.SynonymSetName, &api.SynonymSetCreateSchema{Items: []api.SynonymItemSchema{}})
	requireNoError(t, err)
	set := client.SynonymSet(data.SynonymSetName)
	cleanupDelete(t, "synonym set", set.Delete)
	if created.Name != data.SynonymSetName {
		t.Fatalf("wrong set: %+v", created)
	}
	_, err = set.Upsert(t.Context(), &api.SynonymSetCreateSchema{Items: []api.SynonymItemSchema{}})
	requireNoError(t, err)
	schema := &api.SynonymItemUpsertSchema{Synonyms: []string{"backpack", "rucksack"}}
	item, err := set.Items().Upsert(t.Context(), data.SetItemID, schema)
	requireNoError(t, err)
	if item.ID != data.SetItemID || !slices.Equal(item.Synonyms, schema.Synonyms) {
		t.Fatalf("wrong item: %+v", item)
	}
	item, err = set.Item(data.SetItemID).Retrieve(t.Context())
	requireNoError(t, err)
	if item.ID != data.SetItemID {
		t.Fatalf("wrong item: %+v", item)
	}
	items, err := set.Items().Retrieve(t.Context())
	requireNoError(t, err)
	if len(items) != 1 || items[0].ID != data.SetItemID {
		t.Fatalf("wrong items: %+v", items)
	}
	retrieved, err := set.Retrieve(t.Context())
	requireNoError(t, err)
	if len(retrieved.Items) != 1 {
		t.Fatalf("missing item: %+v", retrieved)
	}
	list, err := client.SynonymSets().Retrieve(t.Context())
	requireNoError(t, err)
	if len(list) != 1 || list[0].Name != data.SynonymSetName {
		t.Fatalf("wrong sets: %+v", list)
	}
	documents := createCollection(t, client, data.SynonymCollectionName).Documents()
	seedDocuments(t, documents, data.Products())
	t.Run("SearchUsingSynonyms", func(t *testing.T) {
		params := api.SearchParameters{Q: "rucksack", NumTypos: "0", Prefix: "false"}
		requireSearchIDs(t, search(t, documents, params))
		params.SynonymSets = data.SynonymSetName
		requireSearchIDs(t, search(t, documents, params), data.FirstID)
	})
	schema.Synonyms = []string{"mug", "cup"}
	updated, err := set.Item(data.SetItemID).Upsert(t.Context(), schema)
	requireNoError(t, err)
	if !slices.Equal(updated.Synonyms, schema.Synonyms) {
		t.Fatalf("update not applied: %+v", updated)
	}
	t.Run("UpdatedSynonymChangesSearch", func(t *testing.T) {
		requireSearchIDs(t, search(t, documents, api.SearchParameters{Q: "cup", NumTypos: "0", Prefix: "false", SynonymSets: data.SynonymSetName}), data.SecondID)
		requireSearchIDs(t, search(t, documents, api.SearchParameters{Q: "rucksack", NumTypos: "0", Prefix: "false", SynonymSets: data.SynonymSetName}))
	})
	deleted, err := set.Item(data.SetItemID).Delete(t.Context())
	requireNoError(t, err)
	if deleted.ID != data.SetItemID {
		t.Fatalf("wrong deleted item: %+v", deleted)
	}
	_, err = set.Item(data.SetItemID).Retrieve(t.Context())
	requireHTTPError(t, err, 404)
	_, err = set.Delete(t.Context())
	requireNoError(t, err)
	_, err = set.Retrieve(t.Context())
	requireHTTPError(t, err, 404)
}
