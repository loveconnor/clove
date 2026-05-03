package permissions

import (
	"testing"

	"clove/apps/api/internal/repos"
	"clove/apps/api/internal/users"
)

func TestCanViewRepoByVisibility(t *testing.T) {
	t.Parallel()

	anonymous := Subject{}
	alice := subject("user_alice")

	if !CanViewRepo(anonymous, repo("repo_public", "user", "user_owner", "public")) {
		t.Fatal("expected anonymous users to view public repos")
	}
	if CanViewRepo(anonymous, repo("repo_internal", "user", "user_owner", "internal")) {
		t.Fatal("expected anonymous users not to view internal repos")
	}
	if !CanViewRepo(alice, repo("repo_internal", "user", "user_owner", "internal")) {
		t.Fatal("expected authenticated users to view internal repos")
	}
	if CanViewRepo(alice, repo("repo_private", "user", "user_owner", "private")) {
		t.Fatal("expected unrelated users not to view private repos")
	}
}

func TestPersonalRepoOwnerGetsAdminPermissions(t *testing.T) {
	t.Parallel()

	alice := subject("user_alice")
	repository := repo("repo_private", "user", "user_alice", "private")

	if !CanViewRepo(alice, repository) {
		t.Fatal("expected owner to view personal repo")
	}
	if !CanPushRepo(alice, repository) {
		t.Fatal("expected owner to push personal repo")
	}
	if !CanAdminRepo(alice, repository) {
		t.Fatal("expected owner to admin personal repo")
	}
}

func TestRepoRoleThresholds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		role      Role
		view      bool
		push      bool
		createPR  bool
		reviewPR  bool
		mergePR   bool
		adminRepo bool
	}{
		{name: "read", role: RoleRead, view: true, createPR: true},
		{name: "triage", role: RoleTriage, view: true, createPR: true, reviewPR: true},
		{name: "write", role: RoleWrite, view: true, push: true, createPR: true, reviewPR: true},
		{name: "maintainer", role: RoleMaintainer, view: true, push: true, createPR: true, reviewPR: true, mergePR: true},
		{name: "admin", role: RoleAdmin, view: true, push: true, createPR: true, reviewPR: true, mergePR: true, adminRepo: true},
		{name: "owner", role: RoleOwner, view: true, push: true, createPR: true, reviewPR: true, mergePR: true, adminRepo: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			alice := subject("user_alice")
			alice.RepositoryRoles["repo_private"] = tc.role
			repository := repo("repo_private", "organization", "org_clove", "private")

			assertBool(t, "view", CanViewRepo(alice, repository), tc.view)
			assertBool(t, "push", CanPushRepo(alice, repository), tc.push)
			assertBool(t, "create pull request", CanCreatePullRequest(alice, repository), tc.createPR)
			assertBool(t, "review pull request", CanReviewPullRequest(alice, repository), tc.reviewPR)
			assertBool(t, "merge pull request", CanMergePullRequest(alice, repository), tc.mergePR)
			assertBool(t, "admin repo", CanAdminRepo(alice, repository), tc.adminRepo)
		})
	}
}

func TestOrganizationRoleCanGrantRepoAccess(t *testing.T) {
	t.Parallel()

	alice := subject("user_alice")
	alice.OrganizationRoles["org_clove"] = RoleAdmin
	repository := repo("repo_private", "organization", "org_clove", "private")

	if !CanAdminRepo(alice, repository) {
		t.Fatal("expected org admin to admin organization repo")
	}
}

func TestRepositoryRoleWinsOverLowerOrganizationRole(t *testing.T) {
	t.Parallel()

	alice := subject("user_alice")
	alice.OrganizationRoles["org_clove"] = RoleRead
	alice.RepositoryRoles["repo_private"] = RoleMaintainer
	repository := repo("repo_private", "organization", "org_clove", "private")

	if !CanMergePullRequest(alice, repository) {
		t.Fatal("expected repository maintainer role to allow merges")
	}
	if CanAdminRepo(alice, repository) {
		t.Fatal("expected maintainer role not to allow repo administration")
	}
}

func subject(id string) Subject {
	return NewSubject(users.User{ID: id, Username: id})
}

func repo(id, ownerType, ownerID, visibility string) repos.Repo {
	return repos.Repo{
		ID:         id,
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		Name:       id,
		Visibility: visibility,
	}
}

func assertBool(t *testing.T, name string, got, want bool) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %s to be %t, got %t", name, want, got)
	}
}
