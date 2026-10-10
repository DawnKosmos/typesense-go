package tests

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func scopedKeys(t *testing.T, client *typesense.Client, server string) {
	fixtures := accessFixture(t, client)
	parent := createSearchKey(t, client, data.ProjectsCollection)
	users, organizations := typesense.Prefix(data.UsersCollection), typesense.Prefix(data.OrganizationsCollection)
	filter := fmt.Sprintf("organization_id:=%s && archived:=false && $%s(user_id:=%s)", data.AlphaOrganizationID, data.GrantsCollection, data.AliceID)
	projection := fmt.Sprintf("id,name,organization_id,$%s(id,name),$%s(id,name)", data.UsersCollection, data.OrganizationsCollection)
	scoped, err := typesense.GenerateScopedSearchKey(parent.Value, map[string]any{
		"filter_by": filter, "include_fields": projection,
		"exclude_fields": fmt.Sprintf("confidential,$%s(email),$%s(secret),$%s(*)", data.UsersCollection, data.OrganizationsCollection, data.GrantsCollection),
	})
	requireNoError(t, err)
	restricted := authenticatedClient(t, server, scoped)
	projects := restricted.Collection[data.Project](data.ProjectsCollection).Documents()
	t.Run("FixturesContainForbiddenDocuments", func(t *testing.T) {
		result := search(t, fixtures.projects.Documents(), api.SearchParameters{})
		requireSearchIDs(t, result, data.SharedProjectID, data.PrivateProjectID, data.BetaProjectID, data.ArchivedProjectID)
		if result.Hits[0].Document["confidential"] == nil {
			t.Fatal("fixture does not contain stored confidential fields")
		}
	})
	for _, tc := range []struct {
		name   string
		params api.SearchParameters
		ids    []string
	}{
		{"TenantAndUserFilter", api.SearchParameters{}, []string{data.SharedProjectID}},
		{"CannotWidenProjection", api.SearchParameters{IncludeFields: fmt.Sprintf("*,$%s(*),$%s(*)", users, organizations)}, []string{data.SharedProjectID}},
		{"CannotClearExclusions", api.SearchParameters{ExcludeFields: "owner_id"}, []string{data.SharedProjectID}},
		{"CannotWidenFilter", api.SearchParameters{FilterBy: "organization_id:=" + data.BetaOrganizationID + " || id:=" + data.PrivateProjectID}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := search(t, projects, tc.params)
			requireSearchIDs(t, result, tc.ids...)
			if len(result.Hits) == 0 {
				return
			}
			document := result.Hits[0].Document
			requireFields(t, document, "id", "name", "organization_id", users, organizations)
			owner := joinedObject(t, document, data.UsersCollection)
			requireFields(t, owner, "id", "name")
			if owner["id"] != data.AliceID {
				t.Fatalf("wrong joined owner: %+v", owner)
			}
			organization := joinedObject(t, document, data.OrganizationsCollection)
			requireFields(t, organization, "id", "name")
			if organization["id"] != data.AlphaOrganizationID {
				t.Fatalf("wrong joined organization: %+v", organization)
			}
		})
	}
	t.Run("SignedExclusionsWithoutAllowlist", func(t *testing.T) {
		key, err := typesense.GenerateScopedSearchKey(parent.Value, map[string]any{"filter_by": filter, "exclude_fields": "confidential"})
		requireNoError(t, err)
		result := search(t, authenticatedClient(t, server, key).Collection[data.Project](data.ProjectsCollection).Documents(), api.SearchParameters{ExcludeFields: "name"})
		requireSearchIDs(t, result, data.SharedProjectID)
		if _, exists := result.Hits[0].Document["confidential"]; exists {
			t.Fatal("caller overrode signed exclusions")
		}
	})
	t.Run("BuilderAndLimitHits", func(t *testing.T) {
		key, err := typesense.NewScopedSearchKeyBuilder(parent.Value).FilterBy("organization_id:="+data.AlphaOrganizationID).IncludeFields("id", "name").LimitHits(1).ExpiresAt(time.Now().Add(time.Hour)).Build()
		requireNoError(t, err)
		documents := authenticatedClient(t, server, key).Collection[data.Project](data.ProjectsCollection).Documents()
		result := search(t, documents, api.SearchParameters{PerPage: new(1), Page: 1, SortBy: "name:asc"})
		if len(result.Hits) != 1 {
			t.Fatalf("signed hit limit not applied: %+v", result)
		}
		page := search(t, documents, api.SearchParameters{Page: 2, PerPage: new(1), LimitHits: 100})
		if len(page.Hits) != 0 {
			t.Fatalf("caller bypassed signed pagination limit: %+v", page)
		}
		override := search(t, documents, api.SearchParameters{Page: 1, PerPage: new(100), LimitHits: 100})
		if len(override.Hits) != 1 {
			t.Fatalf("caller bypassed signed hit limit: %+v", override)
		}
	})
	t.Run("CannotSearchJoinedCollectionDirectly", func(t *testing.T) {
		_, err := restricted.Collection[data.User](data.UsersCollection).Documents().Search(t.Context(), searchParams("*"))
		requireHTTPError(t, err, 401)
	})
	t.Run("CannotWriteOrListCollections", func(t *testing.T) {
		_, err := projects.Create(t.Context(), loadFixture[data.Project](t, "projects.jsonl")[0], nil)
		requireHTTPError(t, err, 401)
		_, err = restricted.Collections().Retrieve(t.Context(), nil)
		requireHTTPError(t, err, 401)
	})
	t.Run("ExpiredKey", func(t *testing.T) {
		key, err := typesense.NewScopedSearchKeyBuilder(parent.Value).FilterBy(filter).ExpiresAt(time.Now().Add(-time.Minute)).Build()
		requireNoError(t, err)
		_, err = authenticatedClient(t, server, key).Collection[data.Project](data.ProjectsCollection).Documents().Search(t.Context(), searchParams("*"))
		requireHTTPError(t, err, 401)
	})
	t.Run("ModifiedSignedParameters", func(t *testing.T) {
		wire, err := base64.StdEncoding.DecodeString(scoped)
		requireNoError(t, err)
		modified := strings.Replace(string(wire), data.AlphaOrganizationID, data.BetaOrganizationID, 1)
		if modified == string(wire) {
			t.Fatal("signature tampering fixture did not change the parameters")
		}
		key := base64.StdEncoding.EncodeToString([]byte(modified))
		_, err = authenticatedClient(t, server, key).Collection[data.Project](data.ProjectsCollection).Documents().Search(t.Context(), searchParams("*"))
		requireHTTPError(t, err, 401)
	})
	t.Run("MultiSearchEnforcesScopes", func(t *testing.T) {
		result, err := restricted.MultiSearch().Perform(t.Context(), nil, api.MultiSearchSearchesParameter{Searches: []api.MultiSearchCollectionParameters{
			{Collection: typesense.Prefix(data.ProjectsCollection), Q: "*", QueryBy: "name", IncludeFields: "*"},
		}})
		requireNoError(t, err)
		if len(result.Results) != 1 || result.Results[0].Found != 1 || len(result.Results[0].Hits) != 1 || result.Results[0].Hits[0].Document["id"] != data.SharedProjectID {
			t.Fatalf("multi-search bypassed restrictions: %+v", result)
		}
		requireFields(t, result.Results[0].Hits[0].Document, "id", "name", "organization_id", users, organizations)
	})
	t.Run("RevokedParentInvalidatesChild", func(t *testing.T) {
		_, err := client.Key(parent.ID).Delete(t.Context())
		requireNoError(t, err)
		_, err = projects.Search(t.Context(), searchParams("*"))
		requireHTTPError(t, err, 401)
	})
}
