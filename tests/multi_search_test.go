package tests

import (
	"encoding/json"
	"testing"
	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func multiSearch(t *testing.T, client *typesense.Client) {
	selected := createCollection(t, client, data.MultiSearchCollectionName)
	seedDocuments(t, selected.Documents(), data.Products())
	searches := api.MultiSearchSearchesParameter{Searches: []api.MultiSearchCollectionParameters{
		{Collection: data.MultiSearchCollectionName, Q: "backpack", QueryBy: "name"},
		{Collection: data.MultiSearchCollectionName, Q: "mug", QueryBy: "name"},
	}}
	result, err := client.MultiSearch().Perform(t.Context(), nil, searches)
	requireNoError(t, err)
	if len(result.Results) != 2 || result.Results[0].Found != 1 || result.Results[1].Found != 1 {
		t.Fatalf("wrong multi-search: %+v", result)
	}
	custom, err := client.MultiSearch().PerformWithContentType(t.Context(), nil, searches, "application/json")
	requireNoError(t, err)
	var decoded api.MultiSearchResult
	requireNoError(t, json.Unmarshal(custom.Body, &decoded))
	if len(decoded.Results) != 2 {
		t.Fatalf("wrong custom response: %s", custom.Body)
	}
	union, err := client.MultiSearch().PerformUnion(t.Context(), nil, searches)
	requireNoError(t, err)
	if union.Found != 2 || len(union.Hits) != 2 {
		t.Fatalf("wrong union: %+v", union)
	}
	if searches.Union != optional.None {
		t.Fatal("PerformUnion mutated caller parameters")
	}
	searches.Searches[1].Collection = data.MissingName
	partial, err := client.MultiSearch().Perform(t.Context(), nil, searches)
	requireNoError(t, err)
	if partial.Results[0].Found != 1 || partial.Results[1].Code != 404 || partial.Results[1].Error == "" {
		t.Fatalf("wrong partial errors: %+v", partial)
	}
	searches.Union = optional.False
	if _, err := client.MultiSearch().PerformUnion(t.Context(), nil, searches); err == nil {
		t.Fatal("explicit false union accepted")
	}
	t.Run("DifferentCollectionsAndFilters", func(t *testing.T) {
		fixtureCollection[data.CatalogItem](t, client, data.MultiSearchCatalogCollection, "catalog.jsonl")
		requests := api.MultiSearchSearchesParameter{Searches: []api.MultiSearchCollectionParameters{
			{Collection: data.MultiSearchCollectionName, Q: "*", QueryBy: "name", FilterBy: "category:=outdoors"},
			{Collection: data.MultiSearchCatalogCollection, Q: "*", QueryBy: "name", FilterBy: "in_stock:=true"},
		}}
		result, err := client.MultiSearch().Perform(t.Context(), nil, requests)
		requireNoError(t, err)
		if len(result.Results) != 2 || result.Results[0].Found != 1 || result.Results[1].Found != 3 {
			t.Fatalf("wrong federated results: %+v", result)
		}
		requireHitIDs(t, result.Results[0].Hits, data.FirstID)
		requireHitIDs(t, result.Results[1].Hits, data.CatalogPackID, data.CatalogMugID, data.CatalogBottleID)
		for i := range requests.Searches {
			requests.Searches[i].FilterBy = ""
			requests.Searches[i].SortBy = "price:desc"
		}
		union, err := client.MultiSearch().PerformUnion(t.Context(), nil, requests)
		requireNoError(t, err)
		requireSearchIDs(t, union, data.FirstID, data.SecondID, data.CatalogPackID, data.CatalogTentID, data.CatalogMugID, data.CatalogBottleID)
		if union.Hits[0].Document["id"] != data.CatalogTentID {
			t.Fatalf("cross-collection sort order is wrong: %+v", union.Hits)
		}
	})
}
