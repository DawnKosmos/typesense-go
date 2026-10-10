package tests

import (
	"fmt"
	"slices"
	"testing"
	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func joins(t *testing.T, client *typesense.Client) {
	fixtures := accessFixture(t, client)
	users, organizations, projects, grants := typesense.Prefix(data.UsersCollection), typesense.Prefix(data.OrganizationsCollection), typesense.Prefix(data.ProjectsCollection), typesense.Prefix(data.GrantsCollection)
	t.Run("AutomaticSearchPrefixes", func(t *testing.T) {
		params := api.SearchParameters{
			FilterBy:      "archived:=false && $users($organizations(id:=" + data.AlphaOrganizationID + "))",
			IncludeFields: "id,name,$users(id,name,$organizations(id,name))",
			SortBy:        "$users(rank:asc)",
		}
		result := search(t, fixtures.projects.Documents(), params)
		requireSearchIDs(t, result, data.SharedProjectID, data.PrivateProjectID)
		if result.Hits[0].Document["id"] != data.PrivateProjectID {
			t.Fatalf("joined sorting not applied: %+v", result.Hits)
		}
		owner := joinedObject(t, result.Hits[0].Document, data.UsersCollection)
		organization := joinedObject(t, owner, data.OrganizationsCollection)
		if organization["id"] != data.AlphaOrganizationID {
			t.Fatalf("incorrect nested join: %+v", organization)
		}
	})
	t.Run("PrefixedReferences", func(t *testing.T) {
		schema, err := fixtures.projects.Retrieve(t.Context())
		requireNoError(t, err)
		references := map[string]string{}
		for _, field := range schema.Fields {
			if field.Reference != "" {
				references[field.Name] = field.Reference
			}
		}
		if references["owner_id"] != users+".id" || references["organization_id"] != organizations+".id" {
			t.Fatalf("incorrect references: %+v", references)
		}
	})
	t.Run("ForwardJoinProjection", func(t *testing.T) {
		result := search(t, fixtures.projects.Documents(), api.SearchParameters{
			FilterBy:      "id:=" + data.SharedProjectID,
			IncludeFields: fmt.Sprintf("id,name,$%s(id,name),$%s(id,name)", users, organizations),
		})
		requireSearchIDs(t, result, data.SharedProjectID)
		document := result.Hits[0].Document
		requireFields(t, document, "id", "name", users, organizations)
		owner := joinedObject(t, document, data.UsersCollection)
		requireFields(t, owner, "id", "name")
		if owner["id"] != data.AliceID {
			t.Fatalf("wrong owner: %+v", owner)
		}
		organization := joinedObject(t, document, data.OrganizationsCollection)
		requireFields(t, organization, "id", "name")
		if organization["id"] != data.AlphaOrganizationID {
			t.Fatalf("wrong organization: %+v", organization)
		}
	})
	t.Run("ReverseJoinArray", func(t *testing.T) {
		result := search(t, fixtures.users.Documents(), api.SearchParameters{
			FilterBy:      fmt.Sprintf("$%s(organization_id:=%s)", projects, data.AlphaOrganizationID),
			IncludeFields: fmt.Sprintf("id,name,$%s(id,name,strategy:nest_array)", projects),
		})
		requireSearchIDs(t, result, data.AliceID, data.BobID)
		for _, hit := range result.Hits {
			joined, ok := hit.Document[projects].([]any)
			if !ok {
				t.Fatalf("reverse join is not an array: %+v", hit.Document)
			}
			ids := make([]string, len(joined))
			for i, value := range joined {
				project, ok := value.(map[string]any)
				if !ok {
					t.Fatalf("invalid joined project: %+v", value)
				}
				requireFields(t, project, "id", "name")
				ids[i], ok = project["id"].(string)
				if !ok {
					t.Fatal("joined project has no ID")
				}
			}
			want := []string{data.PrivateProjectID}
			if hit.Document["id"] == data.AliceID {
				want = []string{data.ArchivedProjectID, data.SharedProjectID}
			}
			slices.Sort(ids)
			slices.Sort(want)
			if !slices.Equal(ids, want) {
				t.Fatalf("wrong reverse join for %s: %v", hit.Document["id"], ids)
			}
		}
	})
	t.Run("NestedJoinFilter", func(t *testing.T) {
		result := search(t, fixtures.projects.Documents(), api.SearchParameters{
			FilterBy:      fmt.Sprintf("archived:=false && $%s($%s(plan:=pro))", users, organizations),
			IncludeFields: fmt.Sprintf("id,name,$%s(id,name,$%s(id,name))", users, organizations),
		})
		requireSearchIDs(t, result, data.SharedProjectID, data.PrivateProjectID)
		for _, hit := range result.Hits {
			owner := joinedObject(t, hit.Document, data.UsersCollection)
			organization := joinedObject(t, owner, data.OrganizationsCollection)
			if organization["id"] != data.AlphaOrganizationID {
				t.Fatalf("nested join matched another organization: %+v", organization)
			}
		}
	})
	t.Run("ManyToManyAccess", func(t *testing.T) {
		result := search(t, fixtures.projects.Documents(), api.SearchParameters{
			FilterBy:      fmt.Sprintf("organization_id:=%s && archived:=false && $%s(user_id:=%s)", data.AlphaOrganizationID, grants, data.AliceID),
			IncludeFields: fmt.Sprintf("id,name,$%s(user_id,project_id,strategy:nest_array)", grants),
		})
		requireSearchIDs(t, result, data.SharedProjectID)
		joined, ok := result.Hits[0].Document[grants].([]any)
		if !ok || len(joined) != 1 {
			t.Fatalf("unexpected access grants: %+v", result.Hits[0].Document)
		}
		grant, ok := joined[0].(map[string]any)
		if !ok || grant["user_id"] != data.AliceID || grant["project_id"] != data.SharedProjectID {
			t.Fatalf("wrong access grant: %+v", joined)
		}
	})
	t.Run("SortByJoinedField", func(t *testing.T) {
		result := search(t, fixtures.projects.Documents(), api.SearchParameters{FilterBy: "organization_id:=" + data.AlphaOrganizationID + " && archived:=false", SortBy: "$" + users + "(rank:asc)"})
		if !slices.Equal(hitIDs(t, result.Hits), []string{data.PrivateProjectID, data.SharedProjectID}) {
			t.Fatalf("joined sort order is wrong: %+v", result.Hits)
		}
	})
	t.Run("RejectMissingReference", func(t *testing.T) {
		_, err := fixtures.projects.Documents().Create(t.Context(), data.Project{ID: data.InvalidID, Name: "Invalid reference", OrganizationID: data.AlphaOrganizationID, OwnerID: data.MissingName}, nil)
		requireHTTPError(t, err, 400)
		_, err = fixtures.projects.Document(data.InvalidID).Retrieve(t.Context())
		requireHTTPError(t, err, 404)
	})
}
