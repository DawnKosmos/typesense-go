package tests

import (
	"testing"
	"typesense"
	data "typesense/tests_data"
)

func document(t *testing.T, client *typesense.Client) {
	selected := createCollection(t, client, data.DocumentCollectionName)
	fixture := data.Products()[0]
	_, err := selected.Documents().Create(t.Context(), fixture, nil)
	requireNoError(t, err)
	doc := selected.Document(fixture.ID)
	retrieved, err := doc.Retrieve(t.Context())
	requireNoError(t, err)
	if retrieved != fixture {
		t.Fatalf("typed document mismatch (including int64 precision): got %+v, want %+v", retrieved, fixture)
	}
	updated, err := doc.Update(t.Context(), map[string]any{"price": 99.5}, nil)
	requireNoError(t, err)
	if updated.Price != 99.5 || updated.Sequence != data.LargeInteger {
		t.Fatalf("unexpected update: %+v", updated)
	}
	deleted, err := doc.Delete(t.Context())
	requireNoError(t, err)
	if deleted != updated {
		t.Fatalf("unexpected deleted document: %+v", deleted)
	}
	_, err = doc.Retrieve(t.Context())
	requireHTTPError(t, err, 404)
}
