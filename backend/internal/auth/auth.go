package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	OrgID string
	Role  string
	jwt.RegisteredClaims
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenService(secret string, ttl time.Duration) (*TokenService, error) {
	if secret == "" {
		return nil, fmt.Errorf("null string")
	}

	if ttl < 0 {
		return nil, fmt.Errorf("ttl < 0")
	}

	tokenSerivce := &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}

	return tokenSerivce, nil
}

func (s *TokenService) IssueAccessToken(userId string, orgID string, role string) (string, error) {
	if userId == "" {
		return "", errors.New("userID is required")
	}

	if orgID == "" {
		return "", errors.New("orgID is required")
	}

	if role == "" {
		return "", errors.New("role is required")
	}

	claims := AccessClaims{
		OrgID: orgID,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return ss, nil
}

func (s *TokenService) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(s.secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("parsing token error: %w", err)
	}

	claims, ok := token.Claims.(*AccessClaims)

	if !ok {
		return nil, fmt.Errorf("parsing claims error: %w", err)
	}

	return claims, nil
}
