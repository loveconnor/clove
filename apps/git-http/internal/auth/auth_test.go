package auth

import "testing"

func TestJWKSURL(t *testing.T) {
	t.Parallel()

	got := jwksURL("https://api.workos.com/", "client_123")
	want := "https://api.workos.com/sso/jwks/client_123"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestIssuerMatchesSameHost(t *testing.T) {
	t.Parallel()

	if !issuerMatches("https://api.workos.com/", "https://api.workos.com") {
		t.Fatal("expected issuer to match")
	}
	if issuerMatches("https://example.com", "https://api.workos.com") {
		t.Fatal("expected issuer mismatch")
	}
}
