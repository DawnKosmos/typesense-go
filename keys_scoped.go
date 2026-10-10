package typesense

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

/*
GenerateScopedSearchKey embeds search restrictions in a key derived from a
parent key with only documents:search permissions. No server request is made.
Use NewScopedSearchKeyBuilder for a chainable alternative to constructing params.

params is a flat map of search parameter names to JSON values. Use the same
names and syntax as a normal Typesense search request, without wrapping them
in searchParameters. Strings contain filter expressions or comma-separated
lists; numbers and booleans are passed as Go numbers and booleans.

Any supported search parameter can be embedded. Common examples include:
  - filter_by: "tenant_id:=42 && price:<50" restricts matching documents.
  - include_fields: "id,name,price" selects returned document fields.
  - exclude_fields: "embedding,private_notes" removes returned fields.
  - q, query_by, query_by_weights, prefix, num_typos: configure text matching.
  - sort_by: "price:asc"; facet_by: "category"; group_by: "brand".
  - page, per_page, limit_hits: configure pagination and result limits.
  - highlight_fields, highlight_full_fields: configure highlighting.
  - vector_query: "embedding:([0.1,0.2,0.3], k:100)" for a matching vector schema.
  - curation_sets, synonym_sets: select named search rules.
  - use_cache, cache_ttl: configure caching (cache_ttl is scoped-key-only).
  - nl_query, nl_model_id, conversation and related parameters: enable
    configured natural language or conversational search.

JOIN expressions also work when the collections have schema references:
  - filter_by: "$users(id:=user-1)" filters through a related collection.
  - include_fields: "id,name,$users(name)" includes selected joined fields.
  - exclude_fields: "$users(email)" excludes a joined field; "$users(*)"
    excludes the entire joined document.
  - Nested JOINs use "$orders($users(id:=user-1))" and corresponding nested
    include_fields expressions. JOIN sort_by syntax is "$users(rank:asc)".

Example (users must be related to the searched collection):

	key, err := typesense.GenerateScopedSearchKey(parentSearchKey, map[string]any{
		"filter_by":      "tenant_id:=42 && $users(id:=user-1)",
		"include_fields": "id,name,price,$users(name)",
		"exclude_fields": "embedding,private_notes,$users(email)",
		"limit_hits":     200,
		"expires_at":     time.Now().Add(time.Hour).Unix(),
	})

expires_at is an additional key option: an absolute Unix timestamp in seconds,
earlier than the parent key's expiry. Typesense enforces the embedded search
parameters; clients cannot override them to bypass the restrictions. Actions
and permitted collections come from the parent key. Joined data can be exposed
through searches on a permitted collection. The configured Prefix is applied to
JOIN collection names in filter_by, include_fields, exclude_fields, sort_by and
facet_by, including nested JOINs. Already prefixed names and quoted values stay
unchanged. The caller's parameter map is not modified.

Generate the key on your backend and use it as X-TYPESENSE-API-KEY for searches.
Revoking the parent invalidates its scoped keys. The format is:

	base64(base64(HMAC-SHA256(parentKey, paramsJSON)) + parentKey[:4] + paramsJSON)

The parameters are signed, not encrypted, and remain readable in the key.
This function encodes parameters; Typesense validates their meaning at search time.
Complete parameter syntax and scoped-key behavior:
https://typesense.org/docs/30.2/api/search.html
https://typesense.org/docs/30.2/api/joins.html
https://typesense.org/docs/30.2/api/api-keys.html#generate-scoped-search-key
*/
func GenerateScopedSearchKey(searchKey string, params map[string]any) (string, error) {
	if len(searchKey) < 4 {
		return "", fmt.Errorf("typesense: parent search key must contain at least 4 bytes")
	}
	var normalized map[string]any
	if params != nil {
		normalized = make(map[string]any, len(params))
		for name, value := range params {
			normalized[name] = prefixSearchParam(name, value)
		}
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(searchKey))
	_, _ = mac.Write(encoded)
	digest := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(digest + searchKey[:4] + string(encoded))), nil
}

/*
ScopedSearchKeyBuilder collects search parameters for generating a scoped search key.

	key, err := typesense.NewScopedSearchKeyBuilder(parentSearchKey).
		FilterBy("organization_id:=org-7 && user_id:=user-42").
		IncludeFields("id", "title", "organization_id", "user_id").
		ExcludeFields("embedding", "internal_notes").
		LimitHits(200).
		ExpiresAt(time.Now().Add(time.Hour)).
		Build()
	key, err := typesense.NewScopedSearchKeyBuilder(parentSearchKey).
		FilterBy("organization_id:=org-7 && (user_id:=user-42 || accessible_to_user_ids:=user-42)").
		IncludeFields("id", "title", "$users(id,name)").
		ExcludeFields("embedding", "internal_notes", "$users(email)").
		SortBy("_text_match:desc", "created_at:desc").
		PerPage(20).
		LimitHits(200).
		ExpiresAt(time.Now().Add(time.Hour)).
		Build()
*/
type ScopedSearchKeyBuilder struct {
	searchKey string
	params    map[string]any
}

/*
NewScopedSearchKeyBuilder creates a builder using a parent search key.

	builder := typesense.NewScopedSearchKeyBuilder(parentSearchKey)
*/
func NewScopedSearchKeyBuilder(searchKey string) *ScopedSearchKeyBuilder {
	return &ScopedSearchKeyBuilder{searchKey: searchKey, params: make(map[string]any)}
}

/*
Param sets or replaces a scoped-key parameter by its JSON name.

	builder.Param("filter_by", "organization_id:=org-7 && user_id:=user-42")
	builder.Param("include_fields", "id,title,$users(name)")
	builder.Param("per_page", 20)
	builder.Param("use_cache", false)
	builder.Param("expires_at", time.Now().Add(time.Hour).Unix())
*/
func (b *ScopedSearchKeyBuilder) Param(name string, value any) *ScopedSearchKeyBuilder {
	if b.params == nil {
		b.params = make(map[string]any)
	}
	b.params[name] = value
	return b
}

/*
FilterBy sets the filter expression that restricts matching documents.

	builder.FilterBy("organization_id:=org-7 && user_id:=user-42")
	builder.FilterBy("organization_id:=org-7 && (user_id:=user-42 || accessible_to_user_ids:=user-42)")
	builder.FilterBy("organization_id:=org-7 && $users(id:=user-42)")
*/
func (b *ScopedSearchKeyBuilder) FilterBy(expression string) *ScopedSearchKeyBuilder {
	return b.Param("filter_by", expression)
}

/*
IncludeFields selects the document and joined fields returned in search results.

	builder.IncludeFields("id", "title", "$users(id,name)")
	builder.IncludeFields("id", "title", "$users(*)")
	builder.IncludeFields("id", "title", "$users(id,name, strategy: nest_array)")
*/
func (b *ScopedSearchKeyBuilder) IncludeFields(fields ...string) *ScopedSearchKeyBuilder {
	return b.Param("include_fields", strings.Join(fields, ","))
}

/*
ExcludeFields removes selected document and joined fields from search results.

	builder.ExcludeFields("embedding", "internal_notes", "$users(email)")
	builder.ExcludeFields("$users(*)")
*/
func (b *ScopedSearchKeyBuilder) ExcludeFields(fields ...string) *ScopedSearchKeyBuilder {
	return b.Param("exclude_fields", strings.Join(fields, ","))
}

/*
Query sets the search text, or "*" to match all documents subject to filters.

	builder.Query("hiking backpack")
	builder.Query("*")
*/
func (b *ScopedSearchKeyBuilder) Query(query string) *ScopedSearchKeyBuilder {
	return b.Param("q", query)
}

/*
QueryBy selects the fields used for text or embedding search.

	builder.QueryBy("title", "description")
*/
func (b *ScopedSearchKeyBuilder) QueryBy(fields ...string) *ScopedSearchKeyBuilder {
	return b.Param("query_by", strings.Join(fields, ","))
}

/*
SortBy sets the result sorting expressions in priority order.

	builder.SortBy("price:asc", "created_at:desc")
	builder.SortBy("_text_match:desc", "$users(rank:asc)")
*/
func (b *ScopedSearchKeyBuilder) SortBy(expressions ...string) *ScopedSearchKeyBuilder {
	return b.Param("sort_by", strings.Join(expressions, ","))
}

/*
FacetBy selects fields for facet counts and statistics.

	builder.FacetBy("category", "brand")
*/
func (b *ScopedSearchKeyBuilder) FacetBy(fields ...string) *ScopedSearchKeyBuilder {
	return b.Param("facet_by", strings.Join(fields, ","))
}

/*
GroupBy groups search results by the selected fields.

	builder.GroupBy("brand", "category")
*/
func (b *ScopedSearchKeyBuilder) GroupBy(fields ...string) *ScopedSearchKeyBuilder {
	return b.Param("group_by", strings.Join(fields, ","))
}

/*
Page sets the result page, starting at 1.

	builder.Page(2)
*/
func (b *ScopedSearchKeyBuilder) Page(page int) *ScopedSearchKeyBuilder {
	return b.Param("page", page)
}

/*
PerPage sets the number of hits returned per page.

	builder.PerPage(20)
*/
func (b *ScopedSearchKeyBuilder) PerPage(count int) *ScopedSearchKeyBuilder {
	return b.Param("per_page", count)
}

/*
LimitHits caps the number of hits accessible through pagination.

	builder.LimitHits(200)
*/
func (b *ScopedSearchKeyBuilder) LimitHits(count int) *ScopedSearchKeyBuilder {
	return b.Param("limit_hits", count)
}

/*
ExpiresAt sets the scoped key expiry as an absolute time.

	builder.ExpiresAt(time.Now().Add(time.Hour))
	builder.ExpiresAt(time.Unix(1906054106, 0))
*/
func (b *ScopedSearchKeyBuilder) ExpiresAt(expiry time.Time) *ScopedSearchKeyBuilder {
	return b.Param("expires_at", expiry.Unix())
}

/*
UseCache enables or disables search result caching.
*/
func (b *ScopedSearchKeyBuilder) UseCache(enabled bool) *ScopedSearchKeyBuilder {
	return b.Param("use_cache", enabled)
}

/*
CacheTTL sets the search cache lifetime in seconds.

	builder.UseCache(true).CacheTTL(60)
*/
func (b *ScopedSearchKeyBuilder) CacheTTL(seconds int) *ScopedSearchKeyBuilder {
	return b.Param("cache_ttl", seconds)
}

/*
Build signs the current parameters and returns a scoped search key.

	key, err := builder.Build()
*/
func (b *ScopedSearchKeyBuilder) Build() (string, error) {
	return GenerateScopedSearchKey(b.searchKey, b.params)
}
