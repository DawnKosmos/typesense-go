package tests

import (
	"io"
	"slices"
	"strings"
	"testing"
	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func stemming(t *testing.T, client *typesense.Client) {
	words := loadFixture[api.StemmingDictionaryWord](t, "stemming.jsonl")
	results, err := client.Stemming().Dictionaries().Upsert(t.Context(), data.DictionaryID, words)
	requireNoError(t, err)
	if len(results) != 2 || *results[0] != words[0] || *results[1] != words[1] {
		t.Fatalf("wrong import: %+v", results)
	}
	dictionary, err := client.Stemming().Dictionary(data.DictionaryID).Retrieve(t.Context())
	requireNoError(t, err)
	if dictionary.ID != data.DictionaryID || !slices.Equal(dictionary.Words, words) {
		t.Fatalf("wrong dictionary: %+v", dictionary)
	}
	list, err := client.Stemming().Dictionaries().Retrieve(t.Context())
	requireNoError(t, err)
	if !slices.Contains(list, data.DictionaryID) {
		t.Fatalf("dictionary missing: %+v", list)
	}
	stream, err := client.Stemming().Dictionaries().UpsertJsonl(t.Context(), data.JsonlDictionaryID, strings.NewReader("{\"word\":\"children\",\"root\":\"child\"}\n"))
	requireNoError(t, err)
	payload, err := io.ReadAll(stream)
	requireNoError(t, stream.Close())
	requireNoError(t, err)
	if !strings.Contains(string(payload), "child") {
		t.Fatalf("wrong JSONL response: %s", payload)
	}
	t.Run("SearchUsingDictionary", func(t *testing.T) {
		collection := createCollectionOf[data.Product](t, client, data.StemmedCollectionName, func(schema *api.CollectionSchema) {
			for i := range schema.Fields {
				if schema.Fields[i].Name == "name" {
					schema.Fields[i].Stem = optional.True
					schema.Fields[i].StemDictionary = data.DictionaryID
				}
			}
		})
		seedDocuments(t, collection.Documents(), loadFixture[data.Product](t, "stemmed_products.jsonl"))
		for _, tc := range []struct{ query, id string }{{"person", data.FirstID}, {"company", data.SecondID}} {
			requireSearchIDs(t, search(t, collection.Documents(), api.SearchParameters{Q: tc.query, NumTypos: "0", Prefix: "false"}), tc.id)
		}
	})
	_, err = client.Stemming().Dictionary(data.MissingName).Retrieve(t.Context())
	requireHTTPError(t, err, 404)
}
