package jwt

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultExpiresIn = 10 * time.Minute // 默认过期时间（分钟）
	DefaultIssuer    = "api"            // 默认签发者
)

type JWTGenerator struct {
	privateKey *rsa.PrivateKey
	expiresIn  time.Duration
	issuer     string
}

// NewJWTGenerator 创建JWT生成器
//
// 参数:
//   - privateKey: 私钥
//   - expiresIn: 默认过期时间（分钟）
//   - issuer: 默认签发者
func NewJWTGenerator(privateKey *rsa.PrivateKey, expiresIn time.Duration, issuer string) (*JWTGenerator, error) {

	if expiresIn == 0 {
		expiresIn = DefaultExpiresIn
	}

	if issuer == "" {
		issuer = DefaultIssuer
	}

	if privateKey == nil {
		return nil, errors.New("privateKey is required")
	}

	return &JWTGenerator{
		privateKey: privateKey,
		expiresIn:  expiresIn,
		issuer:     issuer,
	}, nil
}

// Sign 签名
func (j *JWTGenerator) Sign(claims jwt.Claims) (string, error) {
	// 如果是 StandardClaims，可以设置默认值
	if sc, ok := claims.(jwt.RegisteredClaims); ok {
		if sc.Issuer == "" {
			sc.Issuer = j.issuer
		}
		if sc.ExpiresAt == nil {
			sc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(j.expiresIn))
		}
		if sc.IssuedAt == nil {
			sc.IssuedAt = jwt.NewNumericDate(time.Now())
		}
		claims = sc
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}
