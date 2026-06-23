// pkg/jwt/jwt.go
package jwt

import (
	"fmt"
	"path/filepath"
	"time"
)

type JWTManagerConfig struct {
	KeyDir         string        // 密钥目录
	PrivateKeyFile string        // 私钥文件
	PublicKeyFile  string        // 公钥文件
	Issuer         string        // 签发者
	Expiration     time.Duration // 默认过期时间（分钟）
}

// Init 初始化JWT管理器
func Init(cfg *JWTManagerConfig) (*JWTManager, error) {

	privateKeyPath := filepath.Join(cfg.KeyDir, cfg.PrivateKeyFile)
	publicKeyPath := filepath.Join(cfg.KeyDir, cfg.PublicKeyFile)

	privateKey, err := LoadRSAPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := LoadRSAPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	jwtManager, err := NewJWTManager(privateKey, publicKey, cfg.Expiration, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}

	return jwtManager, nil
}

// InitJwtGenerator 初始化JWT生成器
func InitJwtGenerator(privateKeyPath string, expiration time.Duration, issuer string) (*JWTGenerator, error) {

	privateKey, err := LoadRSAPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return NewJWTGenerator(privateKey, expiration, issuer)
}

// InitJwtVerifitor 初始化JWT验证器
func InitJwtValidator(publicKeyPath string) (*JWTValidator, error) {

	publicKey, err := LoadRSAPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return NewJWTValidator(publicKey)
}
