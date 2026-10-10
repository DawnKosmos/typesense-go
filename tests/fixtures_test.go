package tests

import (
	"testing"
	"typesense"
	data "typesense/tests_data"
)

type accessCollections struct {
	organizations *typesense.Collection[data.Organization]
	users         *typesense.Collection[data.User]
	projects      *typesense.Collection[data.Project]
	grants        *typesense.Collection[data.Grant]
}

/* Create referenced collections first; cleanup runs in reverse dependency order. */
func accessFixture(t *testing.T, client *typesense.Client) accessCollections {
	t.Helper()
	return accessCollections{
		organizations: fixtureCollection[data.Organization](t, client, data.OrganizationsCollection, "organizations.jsonl"),
		users:         fixtureCollection[data.User](t, client, data.UsersCollection, "users.jsonl"),
		projects:      fixtureCollection[data.Project](t, client, data.ProjectsCollection, "projects.jsonl"),
		grants:        fixtureCollection[data.Grant](t, client, data.GrantsCollection, "grants.jsonl"),
	}
}
