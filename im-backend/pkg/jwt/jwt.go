package jwt

import "time"

func InitService(
	secret string,
	expiresIn, refreshExpire, maxSession time.Duration,
	issuer string,
) (*Service, error) {
	return NewService(secret, expiresIn, refreshExpire, maxSession, issuer)
}
