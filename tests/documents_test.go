package tests

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func documents(t *testing.T, client *typesense.Client) {
	t.Run("CreateUpsertAndSearch", func(t *testing.T) {
		selected := createCollection(t, client, data.DocumentsCollectionName)
		docs := selected.Documents()
		fixture := data.Products()[0]
		fixture.Secret = "hidden"
		created, err := docs.Create(t.Context(), fixture, nil)
		requireNoError(t, err)
		if created["id"] != fixture.ID {
			t.Fatalf("wrong document: %+v", created)
		}
		_, err = docs.Create(t.Context(), fixture, nil)
		requireHTTPError(t, err, 409)
		fixture.Price = 19.5
		upserted, err := docs.Upsert(t.Context(), fixture, nil)
		requireNoError(t, err)
		if upserted["price"] != fixture.Price {
			t.Fatalf("upsert failed: %+v", upserted)
		}
		result, err := docs.Search(t.Context(), searchParams("backpack"))
		requireNoError(t, err)
		if result.Found != 1 || len(result.Hits) != 1 || result.Hits[0].Document["id"] != fixture.ID {
			t.Fatalf("unexpected search: %+v", result)
		}
		if _, exists := result.Hits[0].Document["secret"]; exists {
			t.Fatal("nostore field was stored")
		}
		retrieved, err := selected.Document(fixture.ID).Retrieve(t.Context())
		requireNoError(t, err)
		if retrieved.Secret != "" {
			t.Fatal("nostore field returned by retrieve")
		}
		_, err = docs.Search(t.Context(), nil)
		if err == nil {
			t.Fatal("nil search accepted")
		}
	})
	t.Run("ImportUpdateDeleteAndExport", func(t *testing.T) {
		selected := createCollection(t, client, data.DocumentsCollectionName)
		docs := selected.Documents()
		fixtures := data.Products()
		options := &api.ImportDocumentsParams_ImportDocumentsParameters{ReturnID: optional.True, ReturnDoc: optional.True}
		original := *options
		results, err := docs.Import(t.Context(), fixtures, options)
		requireNoError(t, err)
		if *options != original {
			t.Fatal("import mutated caller options")
		}
		if len(results) != len(fixtures) {
			t.Fatalf("expected two results, got %+v", results)
		}
		for i, result := range results {
			if !result.Success || result.ID != fixtures[i].ID {
				t.Fatalf("unexpected import result %d: %+v", i, result)
			}
		}
		updated, err := docs.Update(t.Context(), map[string]any{"price": 88.5}, &api.UpdateDocumentsParams{UpdateDocumentsParameters: &api.UpdateDocumentsParams_UpdateDocumentsParameters{FilterBy: "category:=outdoors"}})
		requireNoError(t, err)
		if updated != 1 {
			t.Fatalf("updated %d documents, expected 1", updated)
		}
		retrieved, err := selected.Document(data.FirstID).Retrieve(t.Context())
		requireNoError(t, err)
		if retrieved.Price != 88.5 {
			t.Fatalf("update not persisted: %+v", retrieved)
		}
		stream, err := docs.Export(t.Context(), nil)
		requireNoError(t, err)
		defer stream.Close()
		exported := map[string]data.Product{}
		decoder := json.NewDecoder(stream)
		for {
			var item data.Product
			err := decoder.Decode(&item)
			if err == io.EOF {
				break
			}
			requireNoError(t, err)
			exported[item.ID] = item
		}
		if len(exported) != 2 || exported[data.FirstID].Sequence != data.LargeInteger || exported[data.SecondID] != fixtures[1] {
			t.Fatalf("wrong export: %+v", exported)
		}
		filtered, err := docs.Export(t.Context(), &api.ExportDocumentsParams{ExportDocumentsParameters: &api.ExportDocumentsParams_ExportDocumentsParameters{FilterBy: "category:=kitchen", IncludeFields: "id,name"}})
		requireNoError(t, err)
		payload, err := io.ReadAll(filtered)
		filtered.Close()
		requireNoError(t, err)
		var item map[string]any
		requireNoError(t, json.Unmarshal(payload, &item))
		if len(item) != 2 || item["id"] != data.SecondID {
			t.Fatalf("wrong filtered export: %s", payload)
		}
		deleted, err := docs.Delete(t.Context(), &api.DeleteDocumentsParams{DeleteDocumentsParameters: &api.DeleteDocumentsParams_DeleteDocumentsParameters{FilterBy: "category:=kitchen", IgnoreNotFound: optional.False}})
		requireNoError(t, err)
		if deleted != 1 {
			t.Fatalf("deleted %d documents, expected 1", deleted)
		}
		_, err = selected.Document(data.SecondID).Retrieve(t.Context())
		requireHTTPError(t, err, 404)
	})
	t.Run("ImportArrayAndPerDocumentErrors", func(t *testing.T) {
		docs := createCollection(t, client, data.DocumentsCollectionName).Documents()
		fixtures := data.Products()
		results, err := docs.Import(t.Context(), [2]data.Product{fixtures[0], fixtures[1]}, nil)
		requireNoError(t, err)
		if len(results) != 2 || !results[0].Success || !results[1].Success {
			t.Fatalf("array import failed: %+v", results)
		}
		results, err = docs.Import(t.Context(), []any{fixtures[0], map[string]any{"id": data.InvalidID, "price": "wrong-type"}}, nil)
		requireNoError(t, err)
		if len(results) != 2 || results[0].Success || results[1].Success || results[0].Error == "" || results[1].Error == "" {
			t.Fatalf("expected duplicate and invalid-document errors: %+v", results)
		}
		for _, invalid := range []any{nil, data.Product{}, "not an array", []data.Product{}, &fixtures} {
			_, err := docs.Import(t.Context(), invalid, nil)
			if err == nil {
				t.Fatalf("invalid import input accepted: %T", invalid)
			}
		}
	})
	t.Run("JSONLFixtures", func(t *testing.T) {
		selected := createCollection(t, client, data.DocumentsCollectionName)
		fixture, err := data.Fixtures.Open("products.jsonl")
		requireNoError(t, err)
		defer fixture.Close()
		response, err := selected.Documents().ImportJsonl(t.Context(), fixture, &api.ImportDocumentsParams_ImportDocumentsParameters{ReturnID: optional.True})
		requireNoError(t, err)
		defer response.Close()
		decoder := json.NewDecoder(response)
		for _, expected := range data.Products() {
			var result api.ImportDocumentResponse
			requireNoError(t, decoder.Decode(&result))
			if !result.Success || result.ID != expected.ID {
				t.Fatalf("JSONL import failed: %+v", result)
			}
			actual, err := selected.Document(expected.ID).Retrieve(t.Context())
			requireNoError(t, err)
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("fixture mismatch: %+v", actual)
			}
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			t.Fatalf("unexpected extra import result: %v", err)
		}
	})
	t.Run("StreamHTTPErrors", func(t *testing.T) {
		docs := client.Collection[data.Product](data.MissingName).Documents()
		_, err := docs.Export(t.Context(), nil)
		requireHTTPError(t, err, 404)
		_, err = docs.ImportJsonl(t.Context(), strings.NewReader("{}\n"), nil)
		requireHTTPError(t, err, 404)
	})
}
