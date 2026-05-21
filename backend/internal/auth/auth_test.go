package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenServiceIssueAndParseAccessToken(t *testing.T) {
	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	service := newTestTokenService(t, "secret", time.Hour, fixedNow)

	token, err := service.IssueAccessToken("user-123", "org-456", "admin")
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	claims, err := service.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}

	if claims.Subject != "user-123" {
		t.Fatalf("expected subject %q, got %q", "user-123", claims.Subject)
	}

	if claims.OrgID != "org-456" {
		t.Fatalf("expected org ID %q, got %q", "org-456", claims.OrgID)
	}

	if claims.Role != "admin" {
		t.Fatalf("expected role %q, got %q", "admin", claims.Role)
	}

	if got := claims.IssuedAt.Time; !got.Equal(fixedNow) {
		t.Fatalf("expected issued at %s, got %s", fixedNow, got)
	}

	if got := claims.ExpiresAt.Time; !got.Equal(fixedNow.Add(time.Hour)) {
		t.Fatalf("expected expires at %s, got %s", fixedNow.Add(time.Hour), got)
	}
}

func TestTokenServiceIssueAccessTokenUsesSnakeCaseClaims(t *testing.T) {
	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	service := newTestTokenService(t, "secret", time.Hour, fixedNow)

	token, err := service.IssueAccessToken("user-123", "org-456", "admin")
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	payload := decodeTokenPayload(t, token)

	if _, ok := payload["org_id"]; !ok {
		t.Fatal("expected org_id claim")
	}

	if _, ok := payload["OrgID"]; ok {
		t.Fatal("did not expect OrgID claim")
	}
}

func TestTokenServiceParseAccessTokenRejectsExpiredToken(t *testing.T) {
	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	service := newTestTokenService(t, "secret", time.Hour, fixedNow)

	token, err := service.IssueAccessToken("user-123", "org-456", "admin")
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	service.now = func() time.Time {
		return fixedNow.Add(2 * time.Hour)
	}

	if _, err := service.ParseAccessToken(token); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestTokenServiceParseAccessTokenRejectsWrongSecret(t *testing.T) {
	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	issuer := newTestTokenService(t, "secret", time.Hour, fixedNow)
	parser := newTestTokenService(t, "different-secret", time.Hour, fixedNow)

	token, err := issuer.IssueAccessToken("user-123", "org-456", "admin")
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	if _, err := parser.ParseAccessToken(token); err == nil {
		t.Fatal("expected wrong secret error")
	}
}

func TestTokenServiceParseAccessTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	service := newTestTokenService(t, "secret", time.Hour, fixedNow)
	claims := AccessClaims{
		OrgID: "org-456",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			IssuedAt:  jwt.NewNumericDate(fixedNow),
			ExpiresAt: jwt.NewNumericDate(fixedNow.Add(time.Hour)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	if _, err := service.ParseAccessToken(token); err == nil {
		t.Fatal("expected unexpected signing method error")
	}
}

func TestTokenServiceParseAccessTokenRejectsMalformedToken(t *testing.T) {
	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	service := newTestTokenService(t, "secret", time.Hour, fixedNow)

	if _, err := service.ParseAccessToken("not-a-token"); err == nil {
		t.Fatal("expected malformed token error")
	}
}

func TestTokenServiceParseAccessTokenRejectsMissingClaims(t *testing.T) {
	tests := []struct {
		name      string
		claims    AccessClaims
		expiresAt bool
	}{
		{
			name: "subject",
			claims: AccessClaims{
				OrgID: "org-456",
				Role:  "admin",
			},
			expiresAt: true,
		},
		{
			name: "org ID",
			claims: AccessClaims{
				Role: "admin",
			},
			expiresAt: true,
		},
		{
			name: "role",
			claims: AccessClaims{
				OrgID: "org-456",
			},
			expiresAt: true,
		},
		{
			name: "expires at",
			claims: AccessClaims{
				OrgID: "org-456",
				Role:  "admin",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "user-123",
				},
			},
		},
	}

	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	service := newTestTokenService(t, "secret", time.Hour, fixedNow)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expiresAt {
				tt.claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(fixedNow.Add(time.Hour))
			}

			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, tt.claims).SignedString([]byte("secret"))
			if err != nil {
				t.Fatalf("sign access token: %v", err)
			}

			if _, err := service.ParseAccessToken(token); err == nil {
				t.Fatal("expected missing claim error")
			}
		})
	}
}

func TestNewTokenServiceValidatesInputs(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		ttl    time.Duration
	}{
		{
			name:   "secret",
			secret: "",
			ttl:    time.Hour,
		},
		{
			name:   "zero ttl",
			secret: "secret",
			ttl:    0,
		},
		{
			name:   "negative ttl",
			secret: "secret",
			ttl:    -time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewTokenService(tt.secret, tt.ttl); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestTokenServiceIssueAccessTokenValidatesInputs(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		orgID  string
		role   string
	}{
		{
			name:   "user ID",
			userID: "",
			orgID:  "org-456",
			role:   "admin",
		},
		{
			name:   "org ID",
			userID: "user-123",
			orgID:  "",
			role:   "admin",
		},
		{
			name:   "role",
			userID: "user-123",
			orgID:  "org-456",
			role:   "",
		},
	}

	fixedNow := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	service := newTestTokenService(t, "secret", time.Hour, fixedNow)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := service.IssueAccessToken(tt.userID, tt.orgID, tt.role); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func newTestTokenService(t *testing.T, secret string, ttl time.Duration, now time.Time) *TokenService {
	t.Helper()

	service, err := NewTokenService(secret, ttl)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}

	service.now = func() time.Time {
		return now
	}

	return service
}

func decodeTokenPayload(t *testing.T, token string) map[string]any {
	t.Helper()

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected token to have 3 parts, got %d", len(parts))
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode token payload: %v", err)
	}

	payload := make(map[string]any)
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("unmarshal token payload: %v", err)
	}

	return payload
}
