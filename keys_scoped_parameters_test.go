package typesense

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
)

func TestScopedKeyPreservesJoinNames(t *testing.T) {
	params := map[string]any{
		"filter_by":      "$grants(user_id:=alice && $projects(organization_id:=alpha))",
		"include_fields": "id,$users(name)", "exclude_fields": "$users(email)",
		"sort_by": "$users(rank:asc)", "facet_by": "$users(rank)",
		"q": "$users(name)", "per_page": 0, "use_cache": false,
	}
	before, _ := json.Marshal(params)
	key, err := GenerateScopedSearchKey("parent-search-key", params)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(wire[48:], &decoded); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"filter_by":      "$grants(user_id:=alice && $projects(organization_id:=alpha))",
		"include_fields": "id,$users(name)", "exclude_fields": "$users(email)",
		"sort_by": "$users(rank:asc)", "facet_by": "$users(rank)",
		"q": "$users(name)", "per_page": float64(0), "use_cache": false,
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("signed parameters: %+v, want %+v", decoded, want)
	}
	after, _ := json.Marshal(params)
	if string(before) != string(after) {
		t.Fatal("scoped-key parameters were mutated")
	}
}
