package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"clove/apps/git-http/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Principal struct {
	UserID   string
	Username string
	Email    string
}

type Service struct {
	client  *http.Client
	cfg     config.Config
	keysMu  sync.Mutex
	keys    map[string]any
	expiry  time.Time
	now     func() time.Time
	jwksURL string
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5C []string `json:"x5c"`
}

func NewService(cfg config.Config) *Service {
	return &Service{
		client:  &http.Client{Timeout: 10 * time.Second},
		cfg:     cfg,
		keys:    map[string]any{},
		now:     time.Now,
		jwksURL: jwksURL(cfg.WorkOSBaseURL, cfg.WorkOSClientID),
	}
}

func (s *Service) Configured() bool {
	return strings.TrimSpace(s.cfg.WorkOSClientID) != ""
}

func (s *Service) Authenticate(ctx context.Context, tokenValue string) (Principal, error) {
	if !s.Configured() {
		return Principal{}, errors.New("WorkOS authentication is not configured")
	}

	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return Principal{}, errors.New("missing access token")
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenValue,
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
		return Principal{}, errors.New("invalid or expired session")
	}

	issuer, _ := claims["iss"].(string)
	if !issuerMatches(issuer, s.cfg.WorkOSIssuer) {
		return Principal{}, errors.New("invalid issuer")
	}

	userID, _ := claims["sub"].(string)
	if userID == "" {
		return Principal{}, errors.New("missing subject")
	}

	email, _ := claims["email"].(string)
	return Principal{
		UserID:   userID,
		Username: usernameFromEmail(email),
		Email:    email,
	}, nil
}

func (s *Service) keyForID(ctx context.Context, kid string) (any, error) {
	s.keysMu.Lock()
	if key, ok := s.keys[kid]; ok && s.now().Before(s.expiry) {
		s.keysMu.Unlock()
		return key, nil
	}
	s.keysMu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.jwksURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch WorkOS JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("fetch WorkOS JWKS: status %d", resp.StatusCode)
	}

	var body jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	keys := map[string]any{}
	for _, key := range body.Keys {
		publicKey, err := rsaPublicKey(key)
		if err != nil {
			continue
		}
		keys[key.Kid] = publicKey
	}

	s.keysMu.Lock()
	s.keys = keys
	s.expiry = s.now().Add(10 * time.Minute)
	key, ok := s.keys[kid]
	s.keysMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("WorkOS JWKS key %q not found", kid)
	}
	return key, nil
}

func rsaPublicKey(key jwk) (*rsa.PublicKey, error) {
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

	if key.Kty != "" && key.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported jwk key type %q", key.Kty)
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
	if e == 0 {
		return nil, errors.New("invalid jwk exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func jwksURL(baseURL, clientID string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.workos.com"
	}
	return fmt.Sprintf("%s/sso/jwks/%s", baseURL, url.PathEscape(clientID))
}

func decodeBase64URL(value string) ([]byte, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	return base64.URLEncoding.DecodeString(value)
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
	return issuerURL.Scheme == expectedURL.Scheme && issuerURL.Host == expectedURL.Host
}

func usernameFromEmail(email string) string {
	if before, _, ok := strings.Cut(email, "@"); ok && before != "" {
		return before
	}
	return email
}
