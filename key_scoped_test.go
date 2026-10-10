package typesense_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"typesense"
)

func TestGenerateScopedSearchKey(t *testing.T) {
	const parent = "test-parent-search-key"
	params := map[string]any{
		"filter_by":      "user_id:=user-1 && organization_id:=org-1",
		"include_fields": "id,name,$users(name,email)",
		"exclude_fields": "embedding,$users(email)",
		"use_cache":      false, "per_page": float64(0), "q": "", "expires_at": float64(1900000000),
	}
	key, err := typesense.GenerateScopedSearchKey(parent, params)
	requireNoError(t, err)
	wire, err := base64.StdEncoding.DecodeString(key)
	requireNoError(t, err)
	const digestSize = 44
	if len(wire) < digestSize+4 || string(wire[digestSize:digestSize+4]) != parent[:4] {
		t.Fatal("invalid signature or parent prefix layout")
	}
	digest, err := base64.StdEncoding.DecodeString(string(wire[:digestSize]))
	requireNoError(t, err)
	payload := wire[digestSize+4:]
	mac := hmac.New(sha256.New, []byte(parent))
	_, _ = mac.Write(payload)
	if !hmac.Equal(digest, mac.Sum(nil)) {
		t.Fatal("signature does not authenticate the parameters")
	}
	var decoded map[string]any
	requireNoError(t, json.Unmarshal(payload, &decoded))
	if !reflect.DeepEqual(decoded, params) {
		t.Fatalf("parameters changed: %#v", decoded)
	}
	again, err := typesense.GenerateScopedSearchKey(parent, params)
	requireNoError(t, err)
	if key != again {
		t.Fatal("same key and parameters produced different signatures")
	}
	changed, err := typesense.GenerateScopedSearchKey(parent, map[string]any{"filter_by": "user_id:=user-2"})
	requireNoError(t, err)
	if key == changed {
		t.Fatal("changed parameters did not change the signature")
	}
}

func TestScopedSearchKeyBuilder(t *testing.T) {
	const parent = "test-parent-search-key"
	expiry := time.Unix(1900000000, 123456789)
	builder := typesense.NewScopedSearchKeyBuilder(parent).
		FilterBy("user_id:=user-1 && organization_id:=org-1").
		IncludeFields("id", "name", "$users(name,email)").ExcludeFields("embedding", "$users(email)").
		Query("shoes").QueryBy("name", "description").SortBy("price:asc", "$users(rank:asc)").
		FacetBy("category", "brand").GroupBy("brand").Page(1).PerPage(0).LimitHits(200).
		ExpiresAt(expiry).UseCache(false).CacheTTL(0).Param("highlight_fields", "")
	key, err := builder.Build()
	requireNoError(t, err)
	want, err := typesense.GenerateScopedSearchKey(parent, map[string]any{
		"filter_by":      "user_id:=user-1 && organization_id:=org-1",
		"include_fields": "id,name,$users(name,email)", "exclude_fields": "embedding,$users(email)",
		"q": "shoes", "query_by": "name,description", "sort_by": "price:asc,$users(rank:asc)",
		"facet_by": "category,brand", "group_by": "brand", "page": 1, "per_page": 0,
		"limit_hits": 200, "expires_at": int64(1900000000), "use_cache": false, "cache_ttl": 0, "highlight_fields": "",
	})
	requireNoError(t, err)
	if key != want {
		t.Fatal("builder did not preserve all parameters")
	}
	again, err := builder.Build()
	requireNoError(t, err)
	if again != key {
		t.Fatal("Build changed the parameters")
	}
	changed, err := builder.FilterBy("organization_id:=org-2").IncludeFields("id").Build()
	requireNoError(t, err)
	if changed == key {
		t.Fatal("setters did not replace parameters")
	}
	var zero typesense.ScopedSearchKeyBuilder
	if _, err := zero.Param("per_page", 1).Build(); err == nil {
		t.Fatal("zero builder accepted a missing parent key")
	}
	if _, err := typesense.NewScopedSearchKeyBuilder(parent).Param("invalid", make(chan int)).Build(); err == nil {
		t.Fatal("builder accepted a non-JSON value")
	}
}

func TestGenerateScopedSearchKeyErrors(t *testing.T) {
	for _, parent := range []string{"", "abc"} {
		key, err := typesense.GenerateScopedSearchKey(parent, nil)
		if err == nil || key != "" {
			t.Fatalf("short parent %q produced key %q, error %v", parent, key, err)
		}
	}
	if key, err := typesense.GenerateScopedSearchKey("abcd", map[string]any{"invalid": make(chan int)}); err == nil || key != "" {
		t.Fatal("non-JSON parameters produced a key")
	}
	if _, err := typesense.GenerateScopedSearchKey("abcd", map[string]any{}); err != nil {
		t.Fatal(err)
	}
}
