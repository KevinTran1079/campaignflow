package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	OrgID string `json:"org_id"`
	Role  string `json:"role"`

	jwt.RegisteredClaims
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenService(secret string, ttl time.Duration) (*TokenService, error) {
	if secret == "" {
		return nil, errors.New("secret is required")
	}

	if ttl <= 0 {
		return nil, errors.New("ttl must be greater than 0")
	}

	tokenService := &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}

	return tokenService, nil
}

func (s *TokenService) IssueAccessToken(userID string, orgID string, role string) (string, error) {
	if userID == "" {
		return "", errors.New("userID is required")
	}

	if orgID == "" {
		return "", errors.New("orgID is required")
	}

	if role == "" {
		return "", errors.New("role is required")
	}

	now := s.currentTime()
	claims := AccessClaims{
		OrgID: orgID,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
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
	if tokenString == "" {
		return nil, errors.New("access token is required")
	}

	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return s.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(s.currentTime),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("access token is invalid")
	}

	if claims.Subject == "" {
		return nil, errors.New("access token subject is required")
	}

	if claims.OrgID == "" {
		return nil, errors.New("access token org_id is required")
	}

	if claims.Role == "" {
		return nil, errors.New("access token role is required")
	}

	return claims, nil
}

func (s *TokenService) currentTime() time.Time {
	if s.now == nil {
		return time.Now()
	}

	return s.now()
}
