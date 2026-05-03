package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"sync"
	"time"

	"clove/apps/api/internal/users"
	"clove/apps/api/pkg/apierror"
	"clove/apps/api/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/workos/workos-go/v7"
)

type contextKey string

const principalKey contextKey = "auth_principal"

type Principal struct {
	User           users.User `json:"user"`
	SessionID      string     `json:"session_id"`
	OrganizationID string     `json:"organization_id,omitempty"`
	Role           string     `json:"role,omitempty"`
	Permissions    []string   `json:"permissions,omitempty"`
	Entitlements   []string   `json:"entitlements,omitempty"`
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RequestMetadata struct {
	IPAddress string
	UserAgent string
}

type Session struct {
	AccessToken    string
	RefreshToken   string
	User           users.User
	SessionID      string
	OrganizationID string
	Role           string
	Permissions    []string
	Entitlements   []string
}

type Authenticator interface {
	Register(ctx context.Context, req RegisterRequest, meta RequestMetadata) (Session, error)
	Login(ctx context.Context, req LoginRequest, meta RequestMetadata) (Session, error)
	AuthenticateCode(ctx context.Context, code string, meta RequestMetadata) (Session, error)
	Refresh(ctx context.Context, refreshToken string, meta RequestMetadata) (Session, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (Principal, error)
	RevokeSession(ctx context.Context, sessionID string) error
	AuthorizationURL(screenHint, state, loginHint string) (string, error)
	Enabled() bool
}

type Service struct {
	cfg        config.Config
	client     *workos.Client
	jwksMu     sync.Mutex
	jwksKeys   map[string]any
	jwksExpiry time.Time
	now        func() time.Time
}

func NewService(cfg config.Config) *Service {
	var client *workos.Client
	if cfg.WorkOSAPIKey != "" && cfg.WorkOSClientID != "" {
		opts := []workos.ClientOption{
			workos.WithClientID(cfg.WorkOSClientID),
			workos.WithAppInfo(cfg.AppName, "0.1.0", ""),
		}
		if cfg.WorkOSBaseURL != "" {
			opts = append(opts, workos.WithBaseURL(cfg.WorkOSBaseURL))
		}
		client = workos.NewClient(cfg.WorkOSAPIKey, opts...)
	}

	return &Service{
		cfg:      cfg,
		client:   client,
		jwksKeys: map[string]any{},
		now:      time.Now,
	}
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey).(Principal)
	return principal, ok
}

func (s *Service) Enabled() bool {
	return s.client != nil && s.cfg.WorkOSClientID != ""
}

func (s *Service) Register(ctx context.Context, req RegisterRequest, meta RequestMetadata) (Session, error) {
	if err := s.requireEnabled(); err != nil {
		return Session{}, err
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	if req.Email == "" || req.Password == "" || req.Username == "" {
		return Session{}, apierror.BadRequest("username, email, and password are required")
	}

	firstName, lastName := namesFromRegisterRequest(req)
	user, err := s.client.UserManagement().Create(ctx, &workos.UserManagementCreateParams{
		Email:     req.Email,
		FirstName: optionalString(firstName),
		LastName:  optionalString(lastName),
		Metadata: map[string]string{
			"username": req.Username,
		},
		Password: workos.UserManagementPasswordPlaintext{Password: req.Password},
	})
	if err != nil {
		return Session{}, fmt.Errorf("create workos user: %w", err)
	}

	session, err := s.Login(ctx, LoginRequest{Email: req.Email, Password: req.Password}, meta)
	if err != nil {
		return Session{}, err
	}
	if session.User.ID == "" {
		session.User = userFromWorkOS(user)
	}

	return session, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest, meta RequestMetadata) (Session, error) {
	if err := s.requireEnabled(); err != nil {
		return Session{}, err
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return Session{}, apierror.BadRequest("email and password are required")
	}

	resp, err := s.client.UserManagement().AuthenticateWithPassword(ctx, &workos.UserManagementAuthenticateWithPasswordParams{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: optionalString(meta.IPAddress),
		UserAgent: optionalString(meta.UserAgent),
	})
	if err != nil {
		return Session{}, fmt.Errorf("authenticate with workos password: %w", err)
	}

	return s.sessionFromAuthResponse(ctx, resp)
}

func (s *Service) AuthenticateCode(ctx context.Context, code string, meta RequestMetadata) (Session, error) {
	if err := s.requireEnabled(); err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(code) == "" {
		return Session{}, apierror.BadRequest("code is required")
	}

	resp, err := s.client.UserManagement().AuthenticateWithCode(ctx, &workos.UserManagementAuthenticateWithCodeParams{
		Code:      code,
		IPAddress: optionalString(meta.IPAddress),
		UserAgent: optionalString(meta.UserAgent),
	})
	if err != nil {
		return Session{}, fmt.Errorf("authenticate workos code: %w", err)
	}

	return s.sessionFromAuthResponse(ctx, resp)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, meta RequestMetadata) (Session, error) {
	if err := s.requireEnabled(); err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(refreshToken) == "" {
		return Session{}, apierror.Unauthorized("refresh token is required")
	}

	resp, err := s.client.UserManagement().AuthenticateWithRefreshToken(ctx, &workos.UserManagementAuthenticateWithRefreshTokenParams{
		RefreshToken: refreshToken,
		IPAddress:    optionalString(meta.IPAddress),
		UserAgent:    optionalString(meta.UserAgent),
	})
	if err != nil {
		return Session{}, fmt.Errorf("refresh workos session: %w", err)
	}

	return s.sessionFromAuthResponse(ctx, resp)
}

func (s *Service) ValidateAccessToken(ctx context.Context, accessToken string) (Principal, error) {
	if err := s.requireEnabled(); err != nil {
		return Principal{}, err
	}

	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Principal{}, apierror.Unauthorized("missing access token")
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		accessToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected jwt signing method %q", token.Header["alg"])
			}
			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, errors.New("missing jwt key id")
			}
			return s.keyForID(ctx, kid)
		},
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return Principal{}, apierror.Unauthorized("invalid or expired session")
	}

	issuer := stringClaim(claims, "iss")
	if !issuerMatches(issuer, s.cfg.WorkOSIssuer) {
		return Principal{}, apierror.Unauthorized("invalid or expired session")
	}

	userID, _ := claims["sub"].(string)
	sessionID, _ := claims["sid"].(string)
	if userID == "" || sessionID == "" {
		return Principal{}, apierror.Unauthorized("invalid session claims")
	}

	user := users.User{ID: userID}
	if email, _ := claims["email"].(string); email != "" {
		user.Email = email
		user.Username = usernameFromEmail(email)
	}

	return Principal{
		User:           user,
		SessionID:      sessionID,
		OrganizationID: stringClaim(claims, "org_id"),
		Role:           stringClaim(claims, "role"),
		Permissions:    stringSliceClaim(claims, "permissions"),
		Entitlements:   stringSliceClaim(claims, "entitlements"),
	}, nil
}

func (s *Service) RevokeSession(ctx context.Context, sessionID string) error {
	if err := s.requireEnabled(); err != nil {
		return err
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}

	return s.client.UserManagement().RevokeSession(ctx, &workos.UserManagementRevokeSessionParams{SessionID: sessionID})
}

func (s *Service) AuthorizationURL(screenHint, state, loginHint string) (string, error) {
	if err := s.requireEnabled(); err != nil {
		return "", err
	}

	provider := workos.UserManagementAuthenticationProviderAuthkit
	params := &workos.UserManagementGetAuthorizationURLParams{
		Provider:    &provider,
		RedirectURI: s.cfg.WorkOSRedirectURI,
	}

	if state != "" {
		params.State = &state
	}
	if loginHint != "" {
		params.LoginHint = &loginHint
	}
	if screenHint != "" {
		hint := workos.UserManagementAuthenticationScreenHint(screenHint)
		params.ScreenHint = &hint
	}

	return s.client.UserManagement().GetAuthorizationURL(params), nil
}

func (s *Service) sessionFromAuthResponse(ctx context.Context, resp *workos.AuthenticateResponse) (Session, error) {
	if resp == nil || resp.AccessToken == "" || resp.RefreshToken == "" {
		return Session{}, apierror.Unauthorized("authentication did not return a complete session")
	}

	principal, err := s.ValidateAccessToken(ctx, resp.AccessToken)
	if err != nil {
		return Session{}, err
	}
	if resp.User != nil {
		principal.User = userFromWorkOS(resp.User)
	}

	return Session{
		AccessToken:    resp.AccessToken,
		RefreshToken:   resp.RefreshToken,
		User:           principal.User,
		SessionID:      principal.SessionID,
		OrganizationID: principal.OrganizationID,
		Role:           principal.Role,
		Permissions:    principal.Permissions,
		Entitlements:   principal.Entitlements,
	}, nil
}

func (s *Service) requireEnabled() error {
	if s.Enabled() {
		return nil
	}
	return apierror.ServiceUnavailable("WorkOS authentication is not configured")
}

func (s *Service) keyForID(ctx context.Context, kid string) (any, error) {
	s.jwksMu.Lock()
	if key, ok := s.jwksKeys[kid]; ok && s.now().Before(s.jwksExpiry) {
		s.jwksMu.Unlock()
		return key, nil
	}
	s.jwksMu.Unlock()

	jwks, err := s.client.UserManagement().GetJWKS(ctx, s.cfg.WorkOSClientID)
	if err != nil {
		return nil, fmt.Errorf("fetch workos jwks: %w", err)
	}

	keys := map[string]any{}
	for _, key := range jwks.Keys {
		publicKey, err := rsaPublicKey(key)
		if err != nil {
			continue
		}
		keys[key.Kid] = publicKey
	}

	s.jwksMu.Lock()
	s.jwksKeys = keys
	s.jwksExpiry = s.now().Add(10 * time.Minute)
	key, ok := s.jwksKeys[kid]
	s.jwksMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("workos jwks key %q not found", kid)
	}
	return key, nil
}

func rsaPublicKey(key *workos.JWKSResponseKeys) (*rsa.PublicKey, error) {
	if key == nil {
		return nil, errors.New("nil jwks key")
	}
	if len(key.X5C) > 0 {
		certBytes, err := base64.StdEncoding.DecodeString(key.X5C[0])
		if err != nil {
			return nil, err
		}
		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			return nil, err
		}
		publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("jwks certificate is not an rsa key")
		}
		return publicKey, nil
	}

	nBytes, err := decodeBase64URL(key.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := decodeBase64URL(key.E)
	if err != nil {
		return nil, err
	}

	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func decodeBase64URL(value string) ([]byte, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	return base64.URLEncoding.DecodeString(value)
}

func userFromWorkOS(user *workos.User) users.User {
	if user == nil {
		return users.User{}
	}

	displayName := strings.TrimSpace(strings.Join(nonEmptyStrings(deref(user.FirstName), deref(user.LastName)), " "))
	username := user.Metadata["username"]
	if username == "" {
		username = usernameFromEmail(user.Email)
	}

	return users.User{
		ID:          user.ID,
		Username:    username,
		Email:       user.Email,
		DisplayName: displayName,
		AvatarURL:   deref(user.ProfilePictureURL),
	}
}

func namesFromRegisterRequest(req RegisterRequest) (string, string) {
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	if firstName != "" || lastName != "" {
		return firstName, lastName
	}

	parts := strings.Fields(req.DisplayName)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func usernameFromEmail(email string) string {
	if before, _, ok := strings.Cut(email, "@"); ok && before != "" {
		return before
	}
	return email
}

func nonEmptyStrings(values ...string) []string {
	var result []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

func stringClaim(claims jwt.MapClaims, key string) string {
	value, _ := claims[key].(string)
	return value
}

func stringSliceClaim(claims jwt.MapClaims, key string) []string {
	raw, ok := claims[key].([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		if value, ok := item.(string); ok {
			values = append(values, value)
		}
	}
	return values
}

func issuerMatches(issuer, expected string) bool {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	expected = strings.TrimRight(strings.TrimSpace(expected), "/")
	if issuer == "" || expected == "" {
		return false
	}
	if issuer == expected {
		return true
	}
	issuerURL, err := url.Parse(issuer)
	if err != nil || issuerURL.Host == "" {
		return false
	}
	expectedURL, err := url.Parse(expected)
	if err != nil || expectedURL.Host == "" {
		return false
	}
	return strings.EqualFold(issuerURL.Host, expectedURL.Host)
}
