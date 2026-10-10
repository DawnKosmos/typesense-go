package tests

import (
	"testing"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func presets(t *testing.T, client *typesense.Client) {
	schema := &api.PresetUpsertSchema{}
	requireNoError(t, schema.Value.FromSearchParameters(api.SearchParameters{Q: "*", QueryBy: "name", PerPage: new(5)}))
	created, err := client.Presets().Upsert(t.Context(), data.PresetName, schema)
	requireNoError(t, err)
	cleanupDelete(t, "preset", client.Preset(data.PresetName).Delete)
	if created.Name != data.PresetName {
		t.Fatalf("wrong preset: %+v", created)
	}
	retrieved, err := client.Preset(data.PresetName).Retrieve(t.Context())
	requireNoError(t, err)
	params, err := retrieved.Value.AsSearchParameters()
	requireNoError(t, err)
	if params.Q != "*" || params.QueryBy != "name" || params.PerPage == nil || *params.PerPage != 5 {
		t.Fatalf("wrong saved parameters: %+v", params)
	}
	list, err := client.Presets().Retrieve(t.Context())
	requireNoError(t, err)
	if len(list) != 1 || list[0].Name != data.PresetName {
		t.Fatalf("wrong presets: %+v", list)
	}
	deleted, err := client.Preset(data.PresetName).Delete(t.Context())
	requireNoError(t, err)
	if deleted.Name != data.PresetName {
		t.Fatalf("wrong deleted preset: %+v", deleted)
	}
	_, err = client.Preset(data.PresetName).Retrieve(t.Context())
	requireHTTPError(t, err, 404)
}
