package api

import (
	"encoding/json"
	"testing"
)

func TestExplicitNumericZeroInSearchRequests(t *testing.T) {
	parameters := SearchParameters{
		Q: "*", QueryBy: "name", DropTokensThreshold: new(0), TypoTokensThreshold: new(0),
		FacetQueryNumTypos: new(0), PerPage: new(0), Limit: new(0),
		HighlightAffixNumTokens: new(0), SnippetThreshold: new(0),
		MaxExtraPrefix: new(0), MaxExtraSuffix: new(0),
	}
	names := []string{"drop_tokens_threshold", "typo_tokens_threshold", "facet_query_num_typos", "per_page", "limit", "highlight_affix_num_tokens", "snippet_threshold", "max_extra_prefix", "max_extra_suffix"}
	request, err := NewSearchCollectionRequest("http://example.invalid", "products", &SearchCollectionParams{SearchParameters: parameters})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if value := request.URL.Query().Get(name); value != "0" {
			t.Errorf("query parameter %s = %q, want explicit zero", name, value)
		}
	}
	unset, err := NewSearchCollectionRequest("http://example.invalid", "products", &SearchCollectionParams{SearchParameters: SearchParameters{Q: "*", QueryBy: "name"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if unset.URL.Query().Has(name) {
			t.Errorf("unset query parameter %s was sent", name)
		}
	}
	common := MultiSearchParameters{
		DropTokensThreshold: new(0), TypoTokensThreshold: new(0), FacetQueryNumTypos: new(0),
		PerPage: new(0), Limit: new(0), HighlightAffixNumTokens: new(0),
		SnippetThreshold: new(0), MaxExtraPrefix: new(0), MaxExtraSuffix: new(0),
	}
	search := MultiSearchCollectionParameters{
		Collection: "products", Q: "*", QueryBy: "name", DropTokensThreshold: new(0),
		TypoTokensThreshold: new(0), FacetQueryNumTypos: new(0), PerPage: new(0), Limit: new(0),
		HighlightAffixNumTokens: new(0), SnippetThreshold: new(0), MaxExtraPrefix: new(0), MaxExtraSuffix: new(0),
	}
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"SearchParameters", parameters},
		{"MultiSearchParameters", common},
		{"MultiSearchCollectionParameters", search},
	} {
		t.Run(tc.name, func(t *testing.T) { requireNumericFields(t, tc.value, names...) })
	}
	request, err = NewMultiSearchRequest("http://example.invalid", &MultiSearchParams{MultiSearchParameters: common}, MultiSearchJSONRequestBody{Searches: []MultiSearchCollectionParameters{search}})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if value := request.URL.Query().Get(name); value != "0" {
			t.Errorf("multi-search query parameter %s = %q, want explicit zero", name, value)
		}
	}
	var preset PresetUpsertSchema_Value
	if err := preset.FromSearchParameters(parameters); err != nil {
		t.Fatal(err)
	}
	requireNumericFields(t, preset, names...)
}

func TestExplicitZeroSamplingParameters(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"Base", NLSearchModelBase{Temperature: new(float32(0)), TopP: new(float32(0))}},
		{"Create", NLSearchModelCreateSchema{Temperature: new(float32(0)), TopP: new(float32(0))}},
		{"Update", NLSearchModelUpdateSchema{Temperature: new(float32(0)), TopP: new(float32(0))}},
		{"Response", NLSearchModelSchema{Temperature: new(float32(0)), TopP: new(float32(0))}},
	} {
		t.Run(tc.name, func(t *testing.T) { requireNumericFields(t, tc.value, "temperature", "top_p") })
	}
	var decoded NLSearchModelCreateSchema
	if err := json.Unmarshal([]byte(`{"temperature":0,"top_p":0}`), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Temperature == nil || decoded.TopP == nil || *decoded.Temperature != 0 || *decoded.TopP != 0 {
		t.Fatalf("zero presence lost during decoding: %+v", decoded)
	}
	var unset NLSearchModelCreateSchema
	if err := json.Unmarshal([]byte(`{}`), &unset); err != nil {
		t.Fatal(err)
	}
	if unset.Temperature != nil || unset.TopP != nil {
		t.Fatal("omitted sampling parameters have values")
	}
}

func requireNumericFields(t *testing.T, value any, names ...string) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if value, exists := fields[name]; !exists || value != float64(0) {
			t.Errorf("JSON field %s missing or not zero: %s", name, encoded)
		}
	}
}
