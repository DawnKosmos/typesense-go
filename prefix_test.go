package typesense

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"typesense/api"
)

func TestPrefixJoins(t *testing.T) {
	SetPrefix("dev_")
	t.Cleanup(func() { SetPrefix("") })
	for _, tc := range []struct{ input, want string }{
		{"$users(id:=alice)", "$dev_users(id:=alice)"},
		{"!$users($organizations(id:=alpha))", "!$dev_users($dev_organizations(id:=alpha))"},
		{"id,$users(name,$organizations(name),strategy:nest_array)", "id,$dev_users(name,$dev_organizations(name),strategy:nest_array)"},
		{"$dev_users(rank:asc)", "$dev_users(rank:asc)"},
		{"$user-groups (rank:asc)", "$dev_user-groups (rank:asc)"},
		{"label:=`$users(id:=alice)` && $users(id:=bob)", "label:=`$users(id:=alice)` && $dev_users(id:=bob)"},
		{"label:=\"$users(foo)\" && $users(id:=bob)", "label:=\"$users(foo)\" && $dev_users(id:=bob)"},
		{"name:=O'Connor && $users(id:=bob)", "name:=O'Connor && $dev_users(id:=bob)"},
		{"label:=`escaped\\`$users(foo)` && $users(id:=bob)", "label:=`escaped\\`$users(foo)` && $dev_users(id:=bob)"},
		{"price:=$100 && $users(id:=bob)", "price:=$100 && $dev_users(id:=bob)"},
		{"$ && $users", "$ && $users"},
		{"name:=alice", "name:=alice"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			if got := prefixJoins(tc.input); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if got := prefixJoins(tc.want); got != tc.want {
				t.Fatalf("prefixing twice changed the expression: %q", got)
			}
		})
	}
	SetPrefix("")
	if got := prefixJoins("$users(id:=alice)"); got != "$users(id:=alice)" {
		t.Fatalf("unset prefix changed expression: %q", got)
	}
}

func TestPrefixCopiesParameters(t *testing.T) {
	SetPrefix("dev_")
	t.Cleanup(func() { SetPrefix("") })
	fields := []api.Field{{Name: "owner_id", Reference: "users.id"}, {Name: "name"}}
	if got := prefixFields(fields); got[0].Reference != "dev_users.id" || fields[0].Reference != "users.id" {
		t.Fatalf("incorrect or mutating field prefix: %+v / %+v", got, fields)
	}
	common := &api.MultiSearchParams{MultiSearchParameters: api.MultiSearchParameters{FilterBy: "$users(id:=alice)", FacetBy: "$users(rank)"}}
	body := api.MultiSearchSearchesParameter{Searches: []api.MultiSearchCollectionParameters{{Collection: "products", IncludeFields: "$users(name)"}}}
	gotCommon, gotBody := prefixMultiSearch(common, body)
	if gotCommon.MultiSearchParameters.FilterBy != "$dev_users(id:=alice)" || gotCommon.MultiSearchParameters.FacetBy != "$dev_users(rank)" || gotBody.Searches[0].Collection != "dev_products" || gotBody.Searches[0].IncludeFields != "$dev_users(name)" {
		t.Fatalf("incorrect multi-search prefixes: %+v / %+v", gotCommon, gotBody)
	}
	if common.MultiSearchParameters.FilterBy != "$users(id:=alice)" || body.Searches[0].Collection != "products" || body.Searches[0].IncludeFields != "$users(name)" {
		t.Fatal("caller parameters were changed")
	}
	if c, _ := prefixMultiSearch(nil, api.MultiSearchSearchesParameter{}); c != nil {
		t.Fatal("nil common parameters changed")
	}
}

func TestScopedKeyPrefixesJoins(t *testing.T) {
	SetPrefix("dev_")
	t.Cleanup(func() { SetPrefix("") })
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
		"filter_by":      "$dev_grants(user_id:=alice && $dev_projects(organization_id:=alpha))",
		"include_fields": "id,$dev_users(name)", "exclude_fields": "$dev_users(email)",
		"sort_by": "$dev_users(rank:asc)", "facet_by": "$dev_users(rank)",
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
