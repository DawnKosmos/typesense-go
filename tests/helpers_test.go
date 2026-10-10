package tests

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"slices"
	"testing"

	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func requireHTTPError(t *testing.T, err error, status int) {
	t.Helper()
	var httpErr *typesense.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError %d, got %v", status, err)
	}
	if httpErr.Status != status || len(httpErr.Body) == 0 {
		t.Fatalf("unexpected HTTP error: %+v", httpErr)
	}
}

func createCollection(t *testing.T, client *typesense.Client, name string) *typesense.Collection[data.Product] {
	t.Helper()
	return createCollectionOf[data.Product](t, client, name)
}

func createCollectionOf[T any](t *testing.T, client *typesense.Client, name string, configure ...func(*api.CollectionSchema)) *typesense.Collection[T] {
	t.Helper()
	schema, err := typesense.ToSchema(reflect.TypeFor[T]())
	requireNoError(t, err)
	schema.Name = name
	for _, apply := range configure {
		apply(schema)
	}
	_, err = client.Collections().Create(t.Context(), schema)
	requireNoError(t, err)
	selected := client.Collection[T](name)
	cleanupDelete(t, "collection "+name, selected.Delete)
	return selected
}

/* Cleanup runs with its own deadline, even when the test context has been canceled. */
func cleanupDelete[T any](t *testing.T, name string, remove func(context.Context) (T, error), missingStatuses ...int) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := cleanupContext()
		defer cancel()
		_, err := remove(ctx)
		var httpErr *typesense.HTTPError
		if errors.As(err, &httpErr) && (httpErr.Status == 404 || slices.Contains(missingStatuses, httpErr.Status)) {
			return
		}
		if err != nil {
			t.Errorf("clean up %s: %v", name, err)
		}
	})
}

func fixtureCollection[T any](t *testing.T, client *typesense.Client, name, fixture string) *typesense.Collection[T] {
	t.Helper()
	collection := createCollectionOf[T](t, client, name)
	seedDocuments(t, collection.Documents(), loadFixture[T](t, fixture))
	return collection
}

func loadFixture[T any](t *testing.T, name string) []T {
	t.Helper()
	file, err := data.Fixtures.Open(name)
	requireNoError(t, err)
	defer file.Close()
	decoder := json.NewDecoder(file)
	var documents []T
	for {
		var document T
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("fixture %s, document %d: %v", name, len(documents), err)
		}
		documents = append(documents, document)
	}
	if len(documents) == 0 {
		t.Fatalf("fixture %s is empty", name)
	}
	return documents
}

func seedDocuments(t *testing.T, documents *typesense.Documents, fixtures any) {
	t.Helper()
	results, err := documents.Import(t.Context(), fixtures, &api.ImportDocumentsParams_ImportDocumentsParameters{ReturnID: optional.True})
	requireNoError(t, err)
	if len(results) != reflect.ValueOf(fixtures).Len() {
		t.Fatalf("import returned %d results", len(results))
	}
	for i, result := range results {
		if result == nil || !result.Success || result.ID == "" {
			t.Fatalf("fixture import failed at %d: %+v", i, result)
		}
	}
}

func authenticatedClient(t *testing.T, server, key string) *typesense.Client {
	t.Helper()
	client, err := typesense.NewClient(typesense.WithServer(server), typesense.WithAPIKey(key))
	requireNoError(t, err)
	return client
}

func search(t *testing.T, documents *typesense.Documents, params api.SearchParameters) *api.SearchResult {
	t.Helper()
	if params.Q == "" {
		params.Q = "*"
	}
	if params.QueryBy == "" {
		params.QueryBy = "name"
	}
	result, err := documents.Search(t.Context(), &api.SearchCollectionParams{SearchParameters: params})
	requireNoError(t, err)
	return result
}

func hitIDs(t *testing.T, hits []api.SearchResultHit) []string {
	t.Helper()
	ids := make([]string, len(hits))
	for i, hit := range hits {
		id, ok := hit.Document["id"].(string)
		if !ok {
			t.Fatalf("search hit has no string ID: %+v", hit.Document)
		}
		ids[i] = id
	}
	return ids
}

func requireSearchIDs(t *testing.T, result *api.SearchResult, expected ...string) {
	t.Helper()
	if result.Found != len(expected) {
		t.Fatalf("found %d, want %d", result.Found, len(expected))
	}
	requireHitIDs(t, result.Hits, expected...)
}

func requireHitIDs(t *testing.T, hits []api.SearchResultHit, expected ...string) {
	t.Helper()
	actual, expected := hitIDs(t, hits), slices.Clone(expected)
	slices.Sort(actual)
	slices.Sort(expected)
	if !slices.Equal(actual, expected) {
		t.Fatalf("IDs %v, want %v", actual, expected)
	}
}

func requireFields(t *testing.T, object map[string]any, expected ...string) {
	t.Helper()
	actual := make([]string, 0, len(object))
	for key := range object {
		actual = append(actual, key)
	}
	expected = slices.Clone(expected)
	slices.Sort(actual)
	slices.Sort(expected)
	if !slices.Equal(actual, expected) {
		t.Fatalf("fields %v, want %v: %+v", actual, expected, object)
	}
}

func joinedObject(t *testing.T, document map[string]any, name string) map[string]any {
	t.Helper()
	object, ok := document[name].(map[string]any)
	if !ok {
		t.Fatalf("joined %s is not an object: %+v", name, document)
	}
	return object
}

func searchParams(query string) *api.SearchCollectionParams {
	return &api.SearchCollectionParams{SearchParameters: api.SearchParameters{Q: query, QueryBy: "name"}}
}

func cleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), data.RequestTimeout)
}
