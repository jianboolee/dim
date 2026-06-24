package jwt

import "time"

func InitService(secret string, expireSeconds, maxSessionSeconds int, issuer string) (*Service, error) {
	expiresIn := time.Duration(expireSeconds) * time.Second
	maxSession := time.Duration(maxSessionSeconds) * time.Second
	return NewService(secret, expiresIn, maxSession, issuer)
}
