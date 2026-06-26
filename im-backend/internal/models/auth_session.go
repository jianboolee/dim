package models

import "time"

type AuthSession struct {
	ID               string     `bson:"id" json:"id"`
	UserID           string     `bson:"user_id" json:"user_id"`
	RefreshTokenHash string     `bson:"refresh_token_hash" json:"-"`
	ExpiresAt        time.Time  `bson:"expires_at" json:"expires_at"`
	CreatedAt        time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `bson:"updated_at" json:"updated_at"`
	RevokedAt        *time.Time `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}
