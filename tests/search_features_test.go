package tests

import (
	"slices"
	"testing"
	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func searchFeatures(t *testing.T, client *typesense.Client) {
	collection := fixtureCollection[data.CatalogItem](t, client, data.CatalogCollection, "catalog.jsonl")
	documents := collection.Documents()
	t.Run("ExplicitZeroPerPage", func(t *testing.T) {
		result := search(t, documents, api.SearchParameters{FacetBy: "category", PerPage: new(0)})
		if len(result.Hits) != 0 || result.Found != 4 || len(result.FacetCounts) == 0 {
			t.Fatalf("zero per_page did not produce a facet-only result: %+v", result)
		}
	})
	t.Run("ExplicitZeroDisablesTokenDropping", func(t *testing.T) {
		params := api.SearchParameters{Q: "hiking unavailable", NumTypos: "0", Prefix: "false"}
		withDefaults := search(t, documents, params)
		requireSearchIDs(t, withDefaults, data.CatalogPackID)
		params.DropTokensThreshold = new(0)
		withoutDropping := search(t, documents, params)
		requireSearchIDs(t, withoutDropping)
	})
	t.Run("SchemaAndTypedNestedDocument", func(t *testing.T) {
		schema, err := collection.Retrieve(t.Context())
		requireNoError(t, err)
		if schema.EnableNestedFields != optional.True || schema.NumDocuments != 4 {
			t.Fatalf("wrong catalog schema: %+v", schema)
		}
		fields := map[string]api.Field{}
		for _, field := range schema.Fields {
			fields[field.Name] = field
		}
		if fields["embedding"].NumDim != 3 || fields["location"].Type != string(api.GEOPOINT) {
			t.Fatalf("vector or geo type missing: %+v", fields)
		}
		item, err := collection.Document(data.CatalogPackID).Retrieve(t.Context())
		requireNoError(t, err)
		if item.Details.Color != "blue" || item.Details.Weight != 800 || len(item.Embedding) != 3 || item.Location != (api.GeoPoint{52.52, 13.405}) {
			t.Fatalf("typed nested values did not round-trip: %+v", item)
		}
	})
	for _, tc := range []struct {
		name   string
		filter string
		ids    []string
	}{
		{"NumericRangeAndBoolean", "price:>=25 && in_stock:=true", []string{data.CatalogPackID, data.CatalogBottleID}},
		{"ArrayFilter", "tags:=travel", []string{data.CatalogPackID, data.CatalogTentID, data.CatalogBottleID}},
		{"NestedObjectFilter", "details.color:=blue && details.weight:<500", []string{data.CatalogBottleID}},
		{"NoMatches", "price:>1000", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requireSearchIDs(t, search(t, documents, api.SearchParameters{FilterBy: tc.filter}), tc.ids...)
		})
	}
	t.Run("SortAndPagination", func(t *testing.T) {
		result := search(t, documents, api.SearchParameters{SortBy: "price:asc"})
		want := []string{data.CatalogMugID, data.CatalogBottleID, data.CatalogPackID, data.CatalogTentID}
		if !slices.Equal(hitIDs(t, result.Hits), want) {
			t.Fatalf("incorrect numeric sorting: %+v", result.Hits)
		}
		page := search(t, documents, api.SearchParameters{SortBy: "price:asc", Page: 2, PerPage: new(2)})
		if page.Found != 4 || !slices.Equal(hitIDs(t, page.Hits), want[2:]) {
			t.Fatalf("incorrect second page: %+v", page)
		}
	})
	t.Run("Facets", func(t *testing.T) {
		result := search(t, documents, api.SearchParameters{FacetBy: "category,tags"})
		facets := map[string]map[string]int{}
		for _, facet := range result.FacetCounts {
			values := map[string]int{}
			for _, value := range facet.Counts {
				values[value.Value] = value.Count
			}
			facets[facet.FieldName] = values
		}
		if facets["category"]["outdoors"] != 2 || facets["category"]["kitchen"] != 2 || facets["tags"]["travel"] != 3 {
			t.Fatalf("wrong facet counts: %+v", facets)
		}
	})
	t.Run("GroupedResults", func(t *testing.T) {
		result := search(t, documents, api.SearchParameters{GroupBy: "category", GroupLimit: 1, SortBy: "price:asc"})
		if len(result.GroupedHits) != 2 {
			t.Fatalf("wrong groups: %+v", result.GroupedHits)
		}
		for _, group := range result.GroupedHits {
			if group.Found != 2 || len(group.GroupKey) != 1 || len(group.Hits) != 1 {
				t.Fatalf("wrong group contents: %+v", group)
			}
			want := data.CatalogMugID
			if group.GroupKey[0] == "outdoors" {
				want = data.CatalogPackID
			} else if group.GroupKey[0] != "kitchen" {
				t.Fatalf("unexpected group: %+v", group)
			}
			if group.Hits[0].Document["id"] != want {
				t.Fatalf("wrong top result for group: %+v", group)
			}
		}
	})
	t.Run("GeoRadius", func(t *testing.T) {
		result := search(t, documents, api.SearchParameters{FilterBy: "location:(52.5200,13.4050,2 km)"})
		requireSearchIDs(t, result, data.CatalogPackID, data.CatalogTentID)
	})
	t.Run("VectorSearchWithoutLLM", func(t *testing.T) {
		result := search(t, documents, api.SearchParameters{VectorQuery: "embedding:([1,0,0],k:2)", PerPage: new(2)})
		if !slices.Equal(hitIDs(t, result.Hits), []string{data.CatalogPackID, data.CatalogTentID}) {
			t.Fatalf("wrong nearest vectors: %+v", result.Hits)
		}
	})
	t.Run("FieldProjection", func(t *testing.T) {
		result := search(t, documents, api.SearchParameters{IncludeFields: "id,name,details", ExcludeFields: "embedding"})
		for _, hit := range result.Hits {
			requireFields(t, hit.Document, "id", "name", "details")
			if _, ok := hit.Document["details"].(map[string]any); !ok {
				t.Fatalf("nested object not preserved: %+v", hit.Document)
			}
		}
	})
	t.Run("InvalidVectorDimensions", func(t *testing.T) {
		item := loadFixture[data.CatalogItem](t, "catalog.jsonl")[0]
		item.ID = data.InvalidID
		item.Embedding = []float64{1, 0}
		_, err := documents.Create(t.Context(), item, nil)
		requireHTTPError(t, err, 400)
	})
}
