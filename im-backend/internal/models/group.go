package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupStatus string

const (
	GroupStatusActive    GroupStatus = "active"
	GroupStatusDissolved GroupStatus = "dissolved"
)

type Group struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID `bson:"conversation_id" json:"conversation_id"`
	Name           string             `bson:"name" json:"name"`
	AvatarURL      string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	OwnerID        string             `bson:"owner_id" json:"owner_id"`
	MemberCount    int                `bson:"member_count" json:"member_count"`
	Status         GroupStatus        `bson:"status" json:"status"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

func (g *Group) IsActive() bool {
	return g != nil && g.Status == GroupStatusActive
}
