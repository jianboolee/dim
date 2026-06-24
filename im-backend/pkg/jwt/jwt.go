package jwt

import "time"

func InitService(secret string, expireSeconds int, issuer string) (*Service, error) {
	expiresIn := time.Duration(expireSeconds) * time.Second
	return NewService(secret, expiresIn, issuer)
}
