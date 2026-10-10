package testsdata

const (
	OrganizationsCollection = "organizations"
	UsersCollection         = "users"
	ProjectsCollection      = "projects"
	GrantsCollection        = "grants"
	AlphaOrganizationID     = "org-alpha"
	BetaOrganizationID      = "org-beta"
	AliceID                 = "alice"
	BobID                   = "bob"
	MalloryID               = "mallory"
	SharedProjectID         = "project-shared"
	PrivateProjectID        = "project-private"
	BetaProjectID           = "project-beta"
	ArchivedProjectID       = "project-archived"
)

type Organization struct {
	ID     string `json:"id"`
	Name   string `json:"name" tyse:"index"`
	Plan   string `json:"plan" tyse:"facet"`
	Secret string `json:"secret" tyse:"index"`
}

type User struct {
	ID             string `json:"id"`
	Name           string `json:"name" tyse:"index"`
	Email          string `json:"email" tyse:"index"`
	OrganizationID string `json:"organization_id" tyse:"index,reference:organizations.id"`
	Rank           int32  `json:"rank" tyse:"index,sort"`
}

type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name" tyse:"index,sort"`
	OrganizationID string `json:"organization_id" tyse:"index,reference:organizations.id"`
	OwnerID        string `json:"owner_id" tyse:"index,reference:users.id"`
	Archived       bool   `json:"archived" tyse:"facet"`
	Confidential   string `json:"confidential" tyse:"index"`
}

type Grant struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id" tyse:"index,reference:users.id"`
	ProjectID string `json:"project_id" tyse:"index,reference:projects.id"`
}
