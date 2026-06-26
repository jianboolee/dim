package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultExpiresIn        = time.Hour
	DefaultRefreshExpiresIn = 24 * time.Hour
	DefaultMaxSession       = 24 * time.Hour
	DefaultIssuer           = "d-im"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrSessionExpired = errors.New("session expired")
)

type AuthTokenClaims struct {
	SessionID string `json:"sid,omitempty"`
	TokenType string `json:"typ,omitempty"`
	jwt.RegisteredClaims
}

type Service struct {
	secret           []byte
	expiresIn        time.Duration
	refreshExpiresIn time.Duration
	maxSession       time.Duration
	issuer           string
}

func NewService(secret string, expiresIn, refreshExpiresIn, maxSession time.Duration, issuer string) (*Service, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	if expiresIn < time.Second {
		expiresIn = DefaultExpiresIn
	}
	if refreshExpiresIn < time.Second {
		refreshExpiresIn = DefaultRefreshExpiresIn
	}
	if maxSession < time.Second {
		maxSession = DefaultMaxSession
	}
	if refreshExpiresIn > maxSession {
		refreshExpiresIn = maxSession
	}
	if issuer == "" {
		issuer = DefaultIssuer
	}

	return &Service{
		secret:           []byte(secret),
		expiresIn:        expiresIn,
		refreshExpiresIn: refreshExpiresIn,
		maxSession:       maxSession,
		issuer:           issuer,
	}, nil
}

func (s *Service) ExpiresInSeconds() int {
	return int(s.expiresIn.Seconds())
}

func (s *Service) RefreshExpiresInSeconds() int {
	return int(s.refreshExpiresIn.Seconds())
}

func (s *Service) MaxSessionSeconds() int {
	return int(s.maxSession.Seconds())
}

func (s *Service) SessionExpiry(sessionStart time.Time) time.Time {
	return sessionStart.Add(s.maxSession)
}

func (s *Service) AccessTokenExpiry(now, sessionStart time.Time) time.Time {
	accessExpiry := now.Add(s.expiresIn)
	sessionExpiry := s.SessionExpiry(sessionStart)
	if accessExpiry.After(sessionExpiry) {
		return sessionExpiry
	}
	return accessExpiry
}

func (s *Service) RefreshTokenExpiry(now, sessionStart time.Time) time.Time {
	refreshExpiry := now.Add(s.refreshExpiresIn)
	sessionExpiry := s.SessionExpiry(sessionStart)
	if refreshExpiry.After(sessionExpiry) {
		return sessionExpiry
	}
	return refreshExpiry
}

func (s *Service) SignAccessToken(userID, sessionID string, sessionStart time.Time) (string, int, error) {
	now := time.Now()
	return s.signToken(userID, sessionID, "access", sessionStart, now, s.AccessTokenExpiry(now, sessionStart))
}

func (s *Service) SignRefreshToken(userID, sessionID string, sessionStart time.Time) (string, time.Time, error) {
	now := time.Now()
	expiresAt := s.RefreshTokenExpiry(now, sessionStart)
	token, _, err := s.signToken(userID, sessionID, "refresh", sessionStart, now, expiresAt)
	return token, expiresAt, err
}

func (s *Service) ParseAccessToken(tokenStr string, claims *AuthTokenClaims) error {
	return s.parseToken(tokenStr, claims, "access", false)
}

func (s *Service) ParseRefreshToken(tokenStr string, claims *AuthTokenClaims) error {
	return s.parseToken(tokenStr, claims, "refresh", false)
}

func (s *Service) parseToken(
	tokenStr string,
	claims *AuthTokenClaims,
	expectedType string,
	withoutClaimsValidation bool,
) error {
	var opts []jwt.ParserOption
	if withoutClaimsValidation {
		opts = append(opts, jwt.WithoutClaimsValidation())
	}
	_, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc, opts...)
	if err != nil {
		return err
	}
	return s.validateClaims(claims, expectedType)
}

func (s *Service) validateClaims(claims *AuthTokenClaims, expectedType string) error {
	if claims.Subject == "" || claims.SessionID == "" || claims.TokenType != expectedType {
		return ErrInvalidToken
	}
	return s.validateSessionAge(claims)
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

func (s *Service) signToken(
	userID string,
	sessionID string,
	tokenType string,
	sessionStart time.Time,
	now time.Time,
	expiresAt time.Time,
) (string, int, error) {
	if userID == "" || sessionID == "" {
		return "", 0, errors.New("user id and session id are required")
	}
	claims := AuthTokenClaims{
		SessionID: sessionID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(sessionStart),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, err
	}
	ttl := int(time.Until(expiresAt).Round(time.Second).Seconds())
	if ttl < 0 {
		ttl = 0
	}
	return signed, ttl, nil
}

func (s *Service) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return s.secret, nil
}
