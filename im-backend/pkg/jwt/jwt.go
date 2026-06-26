package jwt

import "time"

func InitService(secret string, expireSeconds, refreshExpireSeconds, maxSessionSeconds int, issuer string) (*Service, error) {
	expiresIn := time.Duration(expireSeconds) * time.Second
	refreshExpiresIn := time.Duration(refreshExpireSeconds) * time.Second
	maxSession := time.Duration(maxSessionSeconds) * time.Second
	return NewService(secret, expiresIn, refreshExpiresIn, maxSession, issuer)
}
