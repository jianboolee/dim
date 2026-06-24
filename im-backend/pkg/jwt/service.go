package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultExpiresIn = time.Hour
	DefaultIssuer    = "d-im"
)

type AuthTokenClaims struct {
	jwt.RegisteredClaims
}

type Service struct {
	secret    []byte
	expiresIn time.Duration
	issuer    string
}

func NewService(secret string, expiresIn time.Duration, issuer string) (*Service, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	if expiresIn <= 0 {
		expiresIn = DefaultExpiresIn
	}
	if issuer == "" {
		issuer = DefaultIssuer
	}

	return &Service{
		secret:    []byte(secret),
		expiresIn: expiresIn,
		issuer:    issuer,
	}, nil
}

func (s *Service) SignUserToken(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("user id is required")
	}

	now := time.Now()
	claims := AuthTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) Parse(tokenStr string, claims *AuthTokenClaims) error {
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	return err
}
