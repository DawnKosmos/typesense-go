package api

import (
	"encoding/json"
	"testing"
)

func TestCollectionUpdatePresence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema CollectionUpdateSchema
		want   string
	}{
		{"Omitted", CollectionUpdateSchema{}, `{}`},
		{"ClearSynonyms", CollectionUpdateSchema{SynonymSets: []string{}}, `{"synonym_sets":[]}`},
		{"ClearMetadata", CollectionUpdateSchema{Metadata: map[string]any{}}, `{"metadata":{}}`},
		{"ClearBoth", CollectionUpdateSchema{Metadata: map[string]any{}, SynonymSets: []string{}}, `{"metadata":{},"synonym_sets":[]}`},
		{"KeepEmptyFieldsOmitted", CollectionUpdateSchema{Fields: []Field{}, Metadata: map[string]any{}}, `{"metadata":{}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.schema)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != tc.want {
				t.Fatalf("payload %s, want %s", encoded, tc.want)
			}
		})
	}
}
