package httpserver

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"clove/apps/api/internal/auth"
	"clove/apps/api/pkg/apierror"

	"github.com/google/uuid"
)

var organizationNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,38}[a-z0-9])?$`)

var reservedOrganizationNames = map[string]bool{
	"new":           true,
	"login":         true,
	"logout":        true,
	"register":      true,
	"signup":        true,
	"signin":        true,
	"settings":      true,
	"dashboard":     true,
	"explore":       true,
	"api":           true,
	"admin":         true,
	"organizations": true,
	"organization":  true,
	"orgs":          true,
	"org":           true,
	"users":         true,
	"user":          true,
}

type organizationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	Role        string    `json:"role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type organizationMemberResponse struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type organizationInvitationResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type repositoryResponse struct {
	ID            string    `json:"id"`
	OwnerType     string    `json:"owner_type"`
	OwnerID       string    `json:"owner_id"`
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Visibility    string    `json:"visibility"`
	DefaultBranch string    `json:"default_branch"`
	GitPath       string    `json:"git_path"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (s *Server) handleOrganizations(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT o.id, o.name, COALESCE(o.display_name, ''), COALESCE(o.description, ''), o.owner_id,
			COALESCE(om.role, CASE WHEN o.owner_id = $1 THEN 'owner' ELSE '' END) AS role,
			o.created_at, o.updated_at
		FROM organizations o
		LEFT JOIN organization_members om
			ON om.organization_id = o.id AND om.user_id = $1
		WHERE o.owner_id = $1 OR om.user_id = $1
		ORDER BY COALESCE(o.display_name, o.name), o.name
	`, principal.User.ID)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load organizations."))
		return
	}
	defer rows.Close()

	organizations := []organizationResponse{}
	for rows.Next() {
		var organization organizationResponse
		if err := rows.Scan(
			&organization.ID,
			&organization.Name,
			&organization.DisplayName,
			&organization.Description,
			&organization.OwnerID,
			&organization.Role,
			&organization.CreatedAt,
			&organization.UpdatedAt,
		); err != nil {
			apierror.Write(w, r, apierror.Internal("Could not load organizations."))
			return
		}
		organizations = append(organizations, organization)
	}
	if err := rows.Err(); err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load organizations."))
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"organizations": organizations,
		"request_id":    RequestID(r.Context()),
	})
}

func (s *Server) handleOrganization(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	owner := strings.TrimSpace(r.PathValue("owner"))
	var organization organizationResponse
	err := s.db.QueryRowContext(r.Context(), `
		SELECT o.id, o.name, COALESCE(o.display_name, ''), COALESCE(o.description, ''), o.owner_id,
			COALESCE(om.role, CASE WHEN o.owner_id = $1 THEN 'owner' ELSE '' END) AS role,
			o.created_at, o.updated_at
		FROM organizations o
		LEFT JOIN organization_members om
			ON om.organization_id = o.id AND om.user_id = $1
		WHERE o.name = $2 AND (o.owner_id = $1 OR om.user_id = $1)
	`, principal.User.ID, owner).Scan(
		&organization.ID,
		&organization.Name,
		&organization.DisplayName,
		&organization.Description,
		&organization.OwnerID,
		&organization.Role,
		&organization.CreatedAt,
		&organization.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		apierror.Write(w, r, apierror.NotFound("organization not found"))
		return
	}
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load organization."))
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"organization": organization,
		"request_id":   RequestID(r.Context()),
	})
}

func (s *Server) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}
	req.Name = strings.ToLower(strings.TrimSpace(req.Name))
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Name == "" {
		apierror.Write(w, r, apierror.BadRequest("name is required"))
		return
	}
	if !organizationNamePattern.MatchString(req.Name) {
		apierror.Write(w, r, apierror.BadRequest("name must be 1-40 characters and contain only lowercase letters, numbers, and hyphens"))
		return
	}
	if reservedOrganizationNames[req.Name] {
		apierror.Write(w, r, apierror.BadRequest("name is reserved"))
		return
	}

	if err := s.ensureNameAvailable(r, req.Name); err != nil {
		apierror.Write(w, r, err)
		return
	}

	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not create organization."))
		return
	}
	defer func() { _ = tx.Rollback() }()

	organizationID := uuid.NewString()
	description := strings.TrimSpace(req.Description)
	var organization organizationResponse
	err = tx.QueryRowContext(r.Context(), `
		INSERT INTO organizations (id, name, display_name, description, owner_id)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5)
		RETURNING id, name, COALESCE(display_name, ''), COALESCE(description, ''),
			owner_id, 'owner', created_at, updated_at
	`, organizationID, req.Name, req.DisplayName, description, principal.User.ID).Scan(
		&organization.ID,
		&organization.Name,
		&organization.DisplayName,
		&organization.Description,
		&organization.OwnerID,
		&organization.Role,
		&organization.CreatedAt,
		&organization.UpdatedAt,
	)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not create organization."))
		return
	}

	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, organization.ID, principal.User.ID); err != nil {
		apierror.Write(w, r, apierror.Internal("Could not create organization membership."))
		return
	}

	if err := tx.Commit(); err != nil {
		apierror.Write(w, r, apierror.Internal("Could not create organization."))
		return
	}

	apierror.Respond(w, r, http.StatusCreated, map[string]any{
		"organization": organization,
		"request_id":   RequestID(r.Context()),
	})
}

func (s *Server) handleUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	owner := strings.TrimSpace(r.PathValue("owner"))
	role, organizationID, err := s.organizationRole(r, principal.User.ID, owner)
	if err != nil {
		apierror.Write(w, r, err)
		return
	}
	if role != "owner" && role != "admin" {
		apierror.Write(w, r, apierror.Forbidden("only organization owners and admins can edit settings"))
		return
	}

	var req struct {
		DisplayName *string `json:"display_name"`
		Description *string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}

	if req.DisplayName == nil && req.Description == nil {
		apierror.Write(w, r, apierror.BadRequest("at least one editable field is required"))
		return
	}

	if _, err := s.db.ExecContext(r.Context(), `
		UPDATE organizations
		SET display_name = CASE WHEN $1::boolean THEN NULLIF($2, '') ELSE display_name END,
			description = CASE WHEN $3::boolean THEN NULLIF($4, '') ELSE description END,
			updated_at = now()
		WHERE id = $5
	`,
		req.DisplayName != nil,
		stringValue(req.DisplayName),
		req.Description != nil,
		stringValue(req.Description),
		organizationID,
	); err != nil {
		apierror.Write(w, r, apierror.Internal("Could not update organization."))
		return
	}

	var organization organizationResponse
	err = s.db.QueryRowContext(r.Context(), `
		SELECT o.id, o.name, COALESCE(o.display_name, ''), COALESCE(o.description, ''), o.owner_id,
			COALESCE(om.role, CASE WHEN o.owner_id = $1 THEN 'owner' ELSE '' END) AS role,
			o.created_at, o.updated_at
		FROM organizations o
		LEFT JOIN organization_members om
			ON om.organization_id = o.id AND om.user_id = $1
		WHERE o.id = $2
	`, principal.User.ID, organizationID).Scan(
		&organization.ID,
		&organization.Name,
		&organization.DisplayName,
		&organization.Description,
		&organization.OwnerID,
		&organization.Role,
		&organization.CreatedAt,
		&organization.UpdatedAt,
	)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load organization."))
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"organization": organization,
		"request_id":   RequestID(r.Context()),
	})
}

func (s *Server) handleOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	owner := strings.TrimSpace(r.PathValue("owner"))
	_, organizationID, err := s.organizationRole(r, principal.User.ID, owner)
	if err != nil {
		apierror.Write(w, r, err)
		return
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT u.id, u.username, u.email,
			COALESCE(u.display_name, ''),
			COALESCE(u.avatar_url, ''),
			om.role, om.created_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1
		ORDER BY
			CASE om.role
				WHEN 'owner' THEN 0
				WHEN 'admin' THEN 1
				WHEN 'member' THEN 2
				ELSE 3
			END,
			u.username
	`, organizationID)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load organization members."))
		return
	}
	defer rows.Close()

	members := []organizationMemberResponse{}
	for rows.Next() {
		var member organizationMemberResponse
		if err := rows.Scan(
			&member.UserID,
			&member.Username,
			&member.Email,
			&member.DisplayName,
			&member.AvatarURL,
			&member.Role,
			&member.JoinedAt,
		); err != nil {
			apierror.Write(w, r, apierror.Internal("Could not load organization members."))
			return
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load organization members."))
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"members":    members,
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) handleCreateOrganizationInvitation(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	owner := strings.TrimSpace(r.PathValue("owner"))
	role, _, err := s.organizationRole(r, principal.User.ID, owner)
	if err != nil {
		apierror.Write(w, r, err)
		return
	}
	if role != "owner" && role != "admin" {
		apierror.Write(w, r, apierror.Forbidden("only organization owners and admins can invite members"))
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if req.Email == "" {
		apierror.Write(w, r, apierror.BadRequest("email is required"))
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "owner" && req.Role != "admin" && req.Role != "member" {
		apierror.Write(w, r, apierror.BadRequest("role must be owner, admin, or member"))
		return
	}

	apierror.Respond(w, r, http.StatusAccepted, map[string]any{
		"invitation": organizationInvitationResponse{
			ID:        uuid.NewString(),
			Email:     req.Email,
			Role:      req.Role,
			Status:    "pending_delivery",
			CreatedAt: time.Now().UTC(),
		},
		"note":       "Invitation delivery is not yet wired up. The request was accepted but no email has been sent.",
		"request_id": RequestID(r.Context()),
	})
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func (s *Server) ensureNameAvailable(r *http.Request, name string) error {
	var organizationExists, userExists bool
	if err := s.db.QueryRowContext(r.Context(), `
		SELECT EXISTS (SELECT 1 FROM organizations WHERE lower(name) = $1),
			EXISTS (SELECT 1 FROM users WHERE lower(username) = $1)
	`, name).Scan(&organizationExists, &userExists); err != nil {
		return apierror.Internal("Could not validate organization name.")
	}
	if organizationExists || userExists {
		return apierror.New(http.StatusConflict, "name_taken", "name is already taken")
	}
	return nil
}

func (s *Server) organizationRole(r *http.Request, userID, name string) (string, string, error) {
	var organizationID, role string
	err := s.db.QueryRowContext(r.Context(), `
		SELECT o.id,
			COALESCE(om.role, CASE WHEN o.owner_id = $1 THEN 'owner' ELSE '' END) AS role
		FROM organizations o
		LEFT JOIN organization_members om
			ON om.organization_id = o.id AND om.user_id = $1
		WHERE o.name = $2 AND (o.owner_id = $1 OR om.user_id = $1)
	`, userID, name).Scan(&organizationID, &role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", apierror.NotFound("organization not found")
	}
	if err != nil {
		return "", "", apierror.Internal("Could not load organization.")
	}
	return role, organizationID, nil
}

func (s *Server) handleRepositories(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	owner := strings.TrimSpace(r.URL.Query().Get("owner"))
	query := `
		SELECT r.id, r.owner_type, r.owner_id,
			CASE WHEN r.owner_type = 'organization' THEN o.name ELSE u.username END AS owner,
			r.name, COALESCE(r.description, ''), r.visibility, r.default_branch,
			r.git_path, r.created_at, r.updated_at
		FROM repositories r
		LEFT JOIN organizations o ON r.owner_type = 'organization' AND r.owner_id = o.id
		LEFT JOIN users u ON r.owner_type = 'user' AND r.owner_id = u.id
		LEFT JOIN organization_members om
			ON r.owner_type = 'organization'
			AND om.organization_id = r.owner_id
			AND om.user_id = $1
		WHERE (
			r.visibility = 'public'
			OR r.owner_id = $1
			OR om.user_id = $1
			OR (r.visibility = 'internal' AND $1 <> '')
		)
	`
	args := []any{principal.User.ID}
	if owner != "" {
		query += " AND (o.name = $2 OR u.username = $2)"
		args = append(args, owner)
	}
	query += " ORDER BY r.updated_at DESC, owner, r.name"

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load repositories."))
		return
	}
	defer rows.Close()

	repositories, err := scanRepositories(rows)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load repositories."))
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"repositories": repositories,
		"request_id":   RequestID(r.Context()),
	})
}

func (s *Server) handleRepository(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	owner := strings.TrimSpace(r.PathValue("owner"))
	repo := strings.TrimSpace(r.PathValue("repo"))
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT r.id, r.owner_type, r.owner_id,
			CASE WHEN r.owner_type = 'organization' THEN o.name ELSE u.username END AS owner,
			r.name, COALESCE(r.description, ''), r.visibility, r.default_branch,
			r.git_path, r.created_at, r.updated_at
		FROM repositories r
		LEFT JOIN organizations o ON r.owner_type = 'organization' AND r.owner_id = o.id
		LEFT JOIN users u ON r.owner_type = 'user' AND r.owner_id = u.id
		LEFT JOIN organization_members om
			ON r.owner_type = 'organization'
			AND om.organization_id = r.owner_id
			AND om.user_id = $1
		WHERE (o.name = $2 OR u.username = $2)
			AND r.name = $3
			AND (
				r.visibility = 'public'
				OR r.owner_id = $1
				OR om.user_id = $1
				OR (r.visibility = 'internal' AND $1 <> '')
			)
	`, principal.User.ID, owner, repo)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load repository."))
		return
	}
	defer rows.Close()

	repositories, err := scanRepositories(rows)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load repository."))
		return
	}
	if len(repositories) == 0 {
		apierror.Write(w, r, apierror.NotFound("repository not found"))
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"repository": repositories[0],
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) handleCreateRepository(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	var req struct {
		Owner       string `json:"owner"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}
	req.Owner = strings.TrimSpace(req.Owner)
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Visibility = strings.ToLower(strings.TrimSpace(req.Visibility))
	if req.Visibility == "" {
		req.Visibility = "private"
	}
	if req.Owner == "" || req.Name == "" {
		apierror.Write(w, r, apierror.BadRequest("owner and name are required"))
		return
	}
	if req.Visibility != "public" && req.Visibility != "private" && req.Visibility != "internal" {
		apierror.Write(w, r, apierror.BadRequest("visibility must be public, private, or internal"))
		return
	}

	ownerType, ownerID, err := s.resolveRepositoryOwner(r, principal, req.Owner)
	if err != nil {
		apierror.Write(w, r, err)
		return
	}

	repositoryID := uuid.NewString()
	gitPath := ownerType + "/" + ownerID + "/" + req.Name + ".git"
	var repository repositoryResponse
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO repositories (id, owner_type, owner_id, name, description, visibility, git_path)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
		RETURNING id, owner_type, owner_id, $8 AS owner, name, COALESCE(description, ''),
			visibility, default_branch, git_path, created_at, updated_at
	`, repositoryID, ownerType, ownerID, req.Name, req.Description, req.Visibility, gitPath, req.Owner).Scan(
		&repository.ID,
		&repository.OwnerType,
		&repository.OwnerID,
		&repository.Owner,
		&repository.Name,
		&repository.Description,
		&repository.Visibility,
		&repository.DefaultBranch,
		&repository.GitPath,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not create repository."))
		return
	}

	apierror.Respond(w, r, http.StatusCreated, map[string]any{
		"repository": repository,
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) resolveRepositoryOwner(r *http.Request, principal auth.Principal, owner string) (string, string, error) {
	if owner == principal.User.Username {
		return "user", principal.User.ID, nil
	}

	var organizationID string
	err := s.db.QueryRowContext(r.Context(), `
		SELECT o.id
		FROM organizations o
		LEFT JOIN organization_members om
			ON om.organization_id = o.id AND om.user_id = $1
		WHERE o.name = $2 AND (o.owner_id = $1 OR om.user_id = $1)
	`, principal.User.ID, owner).Scan(&organizationID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", apierror.NotFound("repository owner not found")
	}
	if err != nil {
		return "", "", apierror.Internal("Could not resolve repository owner.")
	}
	return "organization", organizationID, nil
}

func (s *Server) upsertUser(r *http.Request, principal auth.Principal) error {
	if s.db == nil || principal.User.ID == "" {
		return nil
	}
	username := strings.TrimSpace(principal.User.Username)
	if username == "" {
		username = principal.User.ID
	}

	_, err := s.db.ExecContext(r.Context(), `
		INSERT INTO users (id, username, email, display_name, avatar_url)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url
	`, principal.User.ID, username, principal.User.Email, principal.User.DisplayName, principal.User.AvatarURL)
	if err != nil {
		return apierror.Internal("Could not store current user.")
	}
	return nil
}

func (s *Server) requireDatabase(w http.ResponseWriter, r *http.Request) bool {
	if s.db != nil {
		return true
	}
	apierror.Write(w, r, apierror.ServiceUnavailable("database is not configured"))
	return false
}

func scanRepositories(rows *sql.Rows) ([]repositoryResponse, error) {
	repositories := []repositoryResponse{}
	for rows.Next() {
		var repository repositoryResponse
		if err := rows.Scan(
			&repository.ID,
			&repository.OwnerType,
			&repository.OwnerID,
			&repository.Owner,
			&repository.Name,
			&repository.Description,
			&repository.Visibility,
			&repository.DefaultBranch,
			&repository.GitPath,
			&repository.CreatedAt,
			&repository.UpdatedAt,
		); err != nil {
			return nil, err
		}
		repositories = append(repositories, repository)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return repositories, nil
}
