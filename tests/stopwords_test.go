package tests

import (
	"slices"
	"testing"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func stopwords(t *testing.T, client *typesense.Client) {
	schema := &api.StopwordsSetUpsertSchema{Stopwords: []string{"the", "a"}, Locale: "en"}
	created, err := client.Stopwords().Upsert(t.Context(), data.StopwordsID, schema)
	requireNoError(t, err)
	cleanupDelete(t, "stopword set", client.Stopword(data.StopwordsID).Delete)
	if created.ID != data.StopwordsID || len(created.Stopwords) != 2 {
		t.Fatalf("wrong stopwords: %+v", created)
	}
	retrieved, err := client.Stopword(data.StopwordsID).Retrieve(t.Context())
	requireNoError(t, err)
	slices.Sort(retrieved.Stopwords)
	slices.Sort(created.Stopwords)
	if retrieved.ID != created.ID || !slices.Equal(retrieved.Stopwords, created.Stopwords) {
		t.Fatalf("wrong retrieved stopwords: %+v", retrieved)
	}
	list, err := client.Stopwords().Retrieve(t.Context())
	requireNoError(t, err)
	if len(list) != 1 || list[0].ID != data.StopwordsID {
		t.Fatalf("wrong stopword sets: %+v", list)
	}
	t.Run("SearchUsingStopwords", func(t *testing.T) {
		documents := createCollection(t, client, data.StopwordCollectionName).Documents()
		seedDocuments(t, documents, data.Products())
		requireSearchIDs(t, search(t, documents, api.SearchParameters{Q: "the backpack", Stopwords: data.StopwordsID}), data.FirstID)
	})
	_, err = client.Stopword(data.StopwordsID).Delete(t.Context())
	requireNoError(t, err)
	_, err = client.Stopword(data.StopwordsID).Retrieve(t.Context())
	requireHTTPError(t, err, 404)
}
