package permissions

import (
	"strings"

	"clove/apps/api/internal/repos"
	"clove/apps/api/internal/users"
)

type Role string

const (
	RoleOwner      Role = "owner"
	RoleAdmin      Role = "admin"
	RoleMaintainer Role = "maintainer"
	RoleWrite      Role = "write"
	RoleTriage     Role = "triage"
	RoleRead       Role = "read"
)

const (
	OwnerTypeUser         = "user"
	OwnerTypeOrganization = "organization"

	VisibilityPublic   = "public"
	VisibilityPrivate  = "private"
	VisibilityInternal = "internal"
)

type Subject struct {
	User              users.User
	OrganizationRoles map[string]Role
	RepositoryRoles   map[string]Role
}

func NewSubject(user users.User) Subject {
	return Subject{
		User:              user,
		OrganizationRoles: map[string]Role{},
		RepositoryRoles:   map[string]Role{},
	}
}

func CanViewRepo(user Subject, repo repos.Repo) bool {
	if normalize(repo.Visibility) == VisibilityPublic {
		return true
	}
	if normalize(repo.Visibility) == VisibilityInternal && user.Authenticated() {
		return true
	}
	return user.effectiveRole(repo).AtLeast(RoleRead)
}

func CanPushRepo(user Subject, repo repos.Repo) bool {
	return user.effectiveRole(repo).AtLeast(RoleWrite)
}

func CanCreatePullRequest(user Subject, repo repos.Repo) bool {
	return CanViewRepo(user, repo)
}

func CanReviewPullRequest(user Subject, repo repos.Repo) bool {
	return user.effectiveRole(repo).AtLeast(RoleTriage)
}

func CanMergePullRequest(user Subject, repo repos.Repo) bool {
	return user.effectiveRole(repo).AtLeast(RoleMaintainer)
}

func CanAdminRepo(user Subject, repo repos.Repo) bool {
	return user.effectiveRole(repo).AtLeast(RoleAdmin)
}

func (s Subject) Authenticated() bool {
	return strings.TrimSpace(s.User.ID) != ""
}

func (s Subject) effectiveRole(repo repos.Repo) Role {
	role := Role("")

	if normalize(repo.OwnerType) == OwnerTypeUser && repo.OwnerID != "" && repo.OwnerID == s.User.ID {
		role = maxRole(role, RoleOwner)
	}

	if normalize(repo.OwnerType) == OwnerTypeOrganization && repo.OwnerID != "" {
		role = maxRole(role, normalizeRole(s.OrganizationRoles[repo.OwnerID]))
	}

	if repo.ID != "" {
		role = maxRole(role, normalizeRole(s.RepositoryRoles[repo.ID]))
	}

	return role
}

func (r Role) AtLeast(required Role) bool {
	return roleRank(normalizeRole(r)) >= roleRank(required)
}

func maxRole(left, right Role) Role {
	if right.AtLeast(left) {
		return right
	}
	return left
}

func normalizeRole(role Role) Role {
	switch Role(normalize(string(role))) {
	case RoleOwner:
		return RoleOwner
	case RoleAdmin:
		return RoleAdmin
	case RoleMaintainer:
		return RoleMaintainer
	case RoleWrite:
		return RoleWrite
	case RoleTriage:
		return RoleTriage
	case RoleRead:
		return RoleRead
	default:
		return ""
	}
}

func roleRank(role Role) int {
	switch normalizeRole(role) {
	case RoleOwner:
		return 6
	case RoleAdmin:
		return 5
	case RoleMaintainer:
		return 4
	case RoleWrite:
		return 3
	case RoleTriage:
		return 2
	case RoleRead:
		return 1
	default:
		return 0
	}
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
