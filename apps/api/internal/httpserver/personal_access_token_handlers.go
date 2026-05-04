package httpserver

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"clove/apps/api/internal/auth"
	"clove/apps/api/pkg/apierror"

	"github.com/google/uuid"
)

type personalAccessTokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type createPersonalAccessTokenResponse struct {
	PersonalAccessToken personalAccessTokenResponse `json:"personal_access_token"`
	Token               string                      `json:"token"`
	RequestID           string                      `json:"request_id"`
}

func (s *Server) handlePersonalAccessTokens(w http.ResponseWriter, r *http.Request) {
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
		SELECT id, name, last_used_at, created_at
		FROM personal_access_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, principal.User.ID)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load personal access tokens."))
		return
	}
	defer rows.Close()

	tokens := []personalAccessTokenResponse{}
	for rows.Next() {
		var token personalAccessTokenResponse
		if err := rows.Scan(&token.ID, &token.Name, &token.LastUsedAt, &token.CreatedAt); err != nil {
			apierror.Write(w, r, apierror.Internal("Could not load personal access tokens."))
			return
		}
		tokens = append(tokens, token)
	}
	if err := rows.Err(); err != nil {
		apierror.Write(w, r, apierror.Internal("Could not load personal access tokens."))
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"personal_access_tokens": tokens,
		"request_id":             RequestID(r.Context()),
	})
}

func (s *Server) handleCreatePersonalAccessToken(w http.ResponseWriter, r *http.Request) {
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
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		apierror.Write(w, r, apierror.BadRequest("name is required"))
		return
	}

	rawToken, err := newPersonalAccessToken()
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not create personal access token."))
		return
	}

	var token personalAccessTokenResponse
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO personal_access_tokens (id, user_id, name, token_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, last_used_at, created_at
	`, uuid.NewString(), principal.User.ID, req.Name, hashPersonalAccessToken(rawToken)).Scan(
		&token.ID,
		&token.Name,
		&token.LastUsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not create personal access token."))
		return
	}

	apierror.Respond(w, r, http.StatusCreated, createPersonalAccessTokenResponse{
		PersonalAccessToken: token,
		Token:               rawToken,
		RequestID:           RequestID(r.Context()),
	})
}

func (s *Server) handleDeletePersonalAccessToken(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if !s.requireDatabase(w, r) {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		apierror.Write(w, r, apierror.NotFound("personal access token not found"))
		return
	}

	result, err := s.db.ExecContext(r.Context(), `
		DELETE FROM personal_access_tokens
		WHERE id = $1 AND user_id = $2
	`, id, principal.User.ID)
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not delete personal access token."))
		return
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not delete personal access token."))
		return
	}
	if deleted == 0 {
		apierror.Write(w, r, apierror.NotFound("personal access token not found"))
		return
	}

	apierror.Respond(w, r, http.StatusNoContent, nil)
}

func newPersonalAccessToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "clove_pat_" + base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func hashPersonalAccessToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
