package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"

	"d-im/internal/models"
	"d-im/internal/repository"
	jwtpkg "d-im/pkg/jwt"
)

var (
	ErrInvalidAuthToken    = errors.New("invalid auth token")
	ErrAuthTokenMissing    = errors.New("auth token missing")
	ErrAuthSessionNotFound = errors.New("auth session not found")
	ErrAuthSessionExpired  = errors.New("auth session expired")
)

type AuthCookieConfig struct {
	Name     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

type TokenResult struct {
	AccessToken      string
	AccessExpiresIn  int
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type AuthService struct {
	jwtService   *jwtpkg.Service
	sessionRepo  *repository.AuthSessionRepository
	cookieConfig AuthCookieConfig
}

func NewAuthService(
	jwtService *jwtpkg.Service,
	sessionRepo *repository.AuthSessionRepository,
	cookieConfig AuthCookieConfig,
) *AuthService {
	return &AuthService{
		jwtService:   jwtService,
		sessionRepo:  sessionRepo,
		cookieConfig: cookieConfig,
	}
}

func (s *AuthService) CreateSession(ctx context.Context, userID string) (*TokenResult, error) {
	now := time.Now()
	sessionID := uuid.NewString()
	sessionStart := now

	refreshToken, refreshExpiresAt, err := s.jwtService.SignRefreshToken(userID, sessionID, sessionStart)
	if err != nil {
		return nil, err
	}
	accessToken, accessExpiresIn, err := s.jwtService.SignAccessToken(userID, sessionID, sessionStart)
	if err != nil {
		return nil, err
	}

	err = s.sessionRepo.Create(ctx, &models.AuthSession{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hashToken(refreshToken),
		ExpiresAt:        refreshExpiresAt,
		CreatedAt:        sessionStart,
		UpdatedAt:        now,
	})
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:      accessToken,
		AccessExpiresIn:  accessExpiresIn,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *AuthService) Exchange(ctx context.Context, accessToken string) (*TokenResult, error) {
	claims := &jwtpkg.AuthTokenClaims{}
	if err := s.jwtService.ParseAccessToken(accessToken, claims); err != nil {
		return nil, normalizeJWTError(err)
	}

	session, err := s.getSession(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != claims.Subject {
		return nil, ErrInvalidAuthToken
	}

	return s.rotateSession(ctx, session, "")
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenResult, error) {
	claims := &jwtpkg.AuthTokenClaims{}
	if err := s.jwtService.ParseRefreshToken(refreshToken, claims); err != nil {
		return nil, normalizeJWTError(err)
	}

	session, err := s.getSession(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != claims.Subject {
		return nil, ErrInvalidAuthToken
	}

	return s.rotateSession(ctx, session, refreshToken)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims := &jwtpkg.AuthTokenClaims{}
	if err := s.jwtService.ParseRefreshToken(refreshToken, claims); err != nil {
		return normalizeJWTError(err)
	}
	return s.sessionRepo.Revoke(ctx, claims.SessionID, time.Now())
}

func (s *AuthService) ReadRefreshToken(c *gin.Context) (string, error) {
	token, err := c.Cookie(s.cookieConfig.Name)
	if err != nil {
		return "", ErrAuthTokenMissing
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrAuthTokenMissing
	}
	return token, nil
}

func (s *AuthService) SetRefreshCookie(c *gin.Context, refreshToken string, expiresAt time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     s.cookieConfig.Name,
		Value:    refreshToken,
		Path:     "/",
		Domain:   s.cookieConfig.Domain,
		HttpOnly: true,
		Secure:   s.cookieConfig.Secure,
		SameSite: s.cookieConfig.SameSite,
		Expires:  expiresAt,
		MaxAge:   max(int(time.Until(expiresAt).Seconds()), 0),
	})
}

func (s *AuthService) ClearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     s.cookieConfig.Name,
		Value:    "",
		Path:     "/",
		Domain:   s.cookieConfig.Domain,
		HttpOnly: true,
		Secure:   s.cookieConfig.Secure,
		SameSite: s.cookieConfig.SameSite,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func (s *AuthService) rotateSession(
	ctx context.Context,
	session *models.AuthSession,
	currentRefreshToken string,
) (*TokenResult, error) {
	now := time.Now()
	if session.RevokedAt != nil {
		return nil, ErrInvalidAuthToken
	}
	if !session.ExpiresAt.After(now) {
		return nil, ErrAuthSessionExpired
	}

	sessionStart := session.CreatedAt
	if sessionStart.IsZero() {
		sessionStart = now
	}

	refreshToken, refreshExpiresAt, err := s.jwtService.SignRefreshToken(session.UserID, session.ID, sessionStart)
	if err != nil {
		return nil, err
	}
	accessToken, accessExpiresIn, err := s.jwtService.SignAccessToken(session.UserID, session.ID, sessionStart)
	if err != nil {
		return nil, err
	}

	currentHash := session.RefreshTokenHash
	if currentRefreshToken != "" {
		currentHash = hashToken(currentRefreshToken)
	}
	if err := s.sessionRepo.RotateRefreshToken(
		ctx,
		session.ID,
		currentHash,
		hashToken(refreshToken),
		now,
		refreshExpiresAt,
	); err != nil {
		if errors.Is(err, repository.ErrAuthSessionConflict) {
			return nil, ErrInvalidAuthToken
		}
		return nil, err
	}

	return &TokenResult{
		AccessToken:      accessToken,
		AccessExpiresIn:  accessExpiresIn,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *AuthService) getSession(ctx context.Context, sessionID string) (*models.AuthSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAuthSessionNotFound
		}
		return nil, err
	}
	if session == nil {
		return nil, ErrAuthSessionNotFound
	}
	return session, nil
}

func normalizeJWTError(err error) error {
	switch {
	case errors.Is(err, jwtpkg.ErrSessionExpired):
		return ErrAuthSessionExpired
	case errors.Is(err, jwtpkg.ErrInvalidToken):
		return ErrInvalidAuthToken
	default:
		return ErrInvalidAuthToken
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ParseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
