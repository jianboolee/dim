package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultExpiresIn   = time.Hour
	DefaultMaxSession  = 24 * time.Hour
	DefaultIssuer      = "d-im"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrSessionExpired = errors.New("session expired")
)

type AuthTokenClaims struct {
	jwt.RegisteredClaims
}

type Service struct {
	secret      []byte
	expiresIn   time.Duration
	maxSession  time.Duration
	issuer      string
}

func NewService(secret string, expiresIn, maxSession time.Duration, issuer string) (*Service, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	if expiresIn <= 0 {
		expiresIn = DefaultExpiresIn
	}
	if maxSession <= 0 {
		maxSession = DefaultMaxSession
	}
	if issuer == "" {
		issuer = DefaultIssuer
	}

	return &Service{
		secret:     []byte(secret),
		expiresIn:  expiresIn,
		maxSession: maxSession,
		issuer:     issuer,
	}, nil
}

func (s *Service) ExpiresInSeconds() int {
	return int(s.expiresIn.Seconds())
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
	_, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc)
	return err
}

// ParseForRefresh 校验签名与绝对会话上限，允许 access token 已过期（用于续期接口）
func (s *Service) ParseForRefresh(tokenStr string, claims *AuthTokenClaims) error {
	_, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc, jwt.WithoutClaimsValidation())
	if err != nil {
		return err
	}

	if claims.Subject == "" {
		return ErrInvalidToken
	}

	if err := s.validateSessionAge(claims); err != nil {
		return err
	}

	return nil
}

// RefreshToken 滑动续期：保留原始 iat 作为会话锚点，重置 exp
func (s *Service) RefreshToken(claims *AuthTokenClaims) (string, int, error) {
	if claims == nil || claims.Subject == "" {
		return "", 0, ErrInvalidToken
	}

	if err := s.validateSessionAge(claims); err != nil {
		return "", 0, err
	}

	now := time.Now()
	sessionStart := now
	if claims.IssuedAt != nil {
		sessionStart = claims.IssuedAt.Time
	}

	newClaims := AuthTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Subject,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(sessionStart),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, err
	}

	return signed, s.ExpiresInSeconds(), nil
}

func (s *Service) validateSessionAge(claims *AuthTokenClaims) error {
	if s.maxSession <= 0 || claims.IssuedAt == nil {
		return nil
	}

	if time.Since(claims.IssuedAt.Time) > s.maxSession {
		return ErrSessionExpired
	}

	return nil
}

func (s *Service) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return s.secret, nil
}
