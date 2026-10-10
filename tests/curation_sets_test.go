package tests

import (
	"testing"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func curationSets(t *testing.T, client *typesense.Client) {
	created, err := client.CurationSets().Upsert(t.Context(), data.CurationSetName, &api.CurationSetCreateSchema{Items: []api.CurationItemCreateSchema{}, Description: "integration curations"})
	requireNoError(t, err)
	set := client.CurationSet(data.CurationSetName)
	cleanupDelete(t, "curation set", set.Delete)
	if created.Name != data.CurationSetName {
		t.Fatalf("wrong set: %+v", created)
	}
	_, err = set.Upsert(t.Context(), &api.CurationSetCreateSchema{Items: []api.CurationItemCreateSchema{}})
	requireNoError(t, err)
	schema := &api.CurationItemCreateSchema{Rule: api.CurationRule{Query: "backpack", Match: api.Exact}, Includes: []api.CurationInclude{{ID: data.FirstID, Position: 1}}}
	item, err := set.Items().Upsert(t.Context(), data.SetItemID, schema)
	requireNoError(t, err)
	if item.ID != data.SetItemID {
		t.Fatalf("wrong item: %+v", item)
	}
	item, err = set.Item(data.SetItemID).Retrieve(t.Context())
	requireNoError(t, err)
	if item.Rule.Query != "backpack" || len(item.Includes) != 1 {
		t.Fatalf("wrong item contents: %+v", item)
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
	list, err := client.CurationSets().Retrieve(t.Context())
	requireNoError(t, err)
	if len(list) != 1 || list[0].Name != data.CurationSetName {
		t.Fatalf("wrong sets: %+v", list)
	}
	schema.Rule.Query = "mug"
	updated, err := set.Item(data.SetItemID).Upsert(t.Context(), schema)
	requireNoError(t, err)
	if updated.Rule.Query != "mug" {
		t.Fatalf("update not applied: %+v", updated)
	}
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
