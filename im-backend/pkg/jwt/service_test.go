package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSignAccessTokenExpiresInMatchesJWTExp(t *testing.T) {
	svc, err := NewService("test-secret", 10*time.Minute, time.Hour, time.Hour, "d-im")
	if err != nil {
		t.Fatal(err)
	}

	sessionStart := time.Now()
	token, expiresIn, err := svc.SignAccessToken("user1", "sess1", sessionStart, "web", "device1")
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("invalid jwt format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("missing exp claim: %#v", claims)
	}

	jwtTTL := int(exp) - int(time.Now().Unix())
	if abs(jwtTTL-expiresIn) > 2 {
		t.Fatalf("expires_in=%d, jwt exp-now≈%d, diff too large", expiresIn, jwtTTL)
	}
	if expiresIn != 600 && expiresIn != 599 {
		t.Fatalf("expected expires_in around 600 for 10m config, got %d", expiresIn)
	}
}

func TestNewServiceRejectsSubSecondDurations(t *testing.T) {
	svc, err := NewService("test-secret", 3600, time.Hour, time.Hour, "d-im")
	if err != nil {
		t.Fatal(err)
	}

	if svc.expiresIn != DefaultExpiresIn {
		t.Fatalf("expiresIn=%v, want default %v for legacy bare-number nanoseconds", svc.expiresIn, DefaultExpiresIn)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
