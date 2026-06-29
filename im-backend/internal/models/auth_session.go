package models

import "time"

type AuthSession struct {
	ID               string     `bson:"id" json:"id"`
	UserID           string     `bson:"user_id" json:"user_id"`
	Platform         string     `bson:"platform" json:"platform"`
	DeviceID         string     `bson:"device_id" json:"device_id"`
	DeviceName       string     `bson:"device_name,omitempty" json:"device_name,omitempty"`
	AppVersion       string     `bson:"app_version,omitempty" json:"app_version,omitempty"`
	PushToken        string     `bson:"push_token,omitempty" json:"-"`
	RefreshTokenHash string     `bson:"refresh_token_hash" json:"-"`
	ExpiresAt        time.Time  `bson:"expires_at" json:"expires_at"`
	LastActiveAt     time.Time  `bson:"last_active_at" json:"last_active_at"`
	CreatedAt        time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `bson:"updated_at" json:"updated_at"`
	RevokedAt        *time.Time `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}
