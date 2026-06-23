// pkg/jwt/jwt.go
package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	generator *JWTGenerator
	verifitor *JWTValidator
}

// NewJWTManager 创建一个新的 JWT 管理器
func NewJWTManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, expiresIn time.Duration, issuer string) (*JWTManager, error) {
	if expiresIn == 0 {
		expiresIn = 10 * time.Minute
	}

	generator, err := NewJWTGenerator(privateKey, expiresIn, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT generator: %w", err)
	}

	verifitor, err := NewJWTValidator(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT verifitor: %w", err)
	}

	return &JWTManager{
		generator: generator,
		verifitor: verifitor,
	}, nil
}

// Sign 签名
func (j *JWTManager) Sign(claims jwt.Claims) (string, error) {
	return j.generator.Sign(claims)
}

// Parse 解析
func (j *JWTManager) Parse(tokenStr string, claims jwt.Claims) error {
	return j.verifitor.Parse(tokenStr, claims)
}

// GetPublicKey 获取公钥
func (j *JWTManager) GetPublicKeyPEM() (string, error) {
	return j.verifitor.GetPublicKeyPEM()
}
