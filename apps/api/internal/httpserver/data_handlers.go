package httpserver

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"clove/apps/api/internal/auth"
	"clove/apps/api/pkg/apierror"

	"github.com/google/uuid"
)

type organizationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	OwnerID     string    `json:"owner_id"`
	Role        string    `json:"role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
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
		SELECT o.id, o.name, COALESCE(o.display_name, ''), o.owner_id,
			COALESCE(om.role, CASE WHEN o.owner_id = $1 THEN 'owner' ELSE '' END) AS role,
			o.created_at
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
			&organization.OwnerID,
			&organization.Role,
			&organization.CreatedAt,
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
		SELECT o.id, o.name, COALESCE(o.display_name, ''), o.owner_id,
			COALESCE(om.role, CASE WHEN o.owner_id = $1 THEN 'owner' ELSE '' END) AS role,
			o.created_at
		FROM organizations o
		LEFT JOIN organization_members om
			ON om.organization_id = o.id AND om.user_id = $1
		WHERE o.name = $2 AND (o.owner_id = $1 OR om.user_id = $1)
	`, principal.User.ID, owner).Scan(
		&organization.ID,
		&organization.Name,
		&organization.DisplayName,
		&organization.OwnerID,
		&organization.Role,
		&organization.CreatedAt,
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
