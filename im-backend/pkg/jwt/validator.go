package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type AuthTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

type JWTValidator struct {
	publicKey *rsa.PublicKey
}

func NewJWTValidator(publicKey *rsa.PublicKey) (*JWTValidator, error) {
	if publicKey == nil {
		return nil, errors.New("publicKey is required")
	}

	return &JWTValidator{
		publicKey: publicKey,
	}, nil
}

// Parse 解析
func (j *JWTValidator) Parse(tokenStr string, claims jwt.Claims) error {
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return j.publicKey, nil
	})
	return err
}

// GetPublicKey 获取公钥
func (j *JWTValidator) GetPublicKeyPEM() (string, error) {
	publicKeyPEM, err := x509.MarshalPKIXPublicKey(j.publicKey)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyPEM,
	})), nil
}
