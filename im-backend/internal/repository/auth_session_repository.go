package repository

import (
	"context"
	"errors"
	"time"

	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrAuthSessionConflict = errors.New("auth session conflict")

type AuthSessionRepository struct {
	collection *mongo.Collection
}

func NewAuthSessionRepository(db *mongo.Database) *AuthSessionRepository {
	return &AuthSessionRepository{
		collection: db.Collection(models.CollectionAuthSession),
	}
}

func (r *AuthSessionRepository) Create(ctx context.Context, session *models.AuthSession) error {
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

func (r *AuthSessionRepository) GetByID(ctx context.Context, id string) (*models.AuthSession, error) {
	var session models.AuthSession
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AuthSessionRepository) RotateRefreshToken(
	ctx context.Context,
	id string,
	currentHash string,
	nextHash string,
	now time.Time,
	expiresAt time.Time,
) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{
		"id":                 id,
		"refresh_token_hash": currentHash,
		"revoked_at":         bson.M{"$exists": false},
		"expires_at":         bson.M{"$gt": now},
	}, bson.M{
		"$set": bson.M{
			"refresh_token_hash": nextHash,
			"updated_at":         now,
			"expires_at":         expiresAt,
		},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrAuthSessionConflict
	}
	return nil
}

func (r *AuthSessionRepository) Revoke(ctx context.Context, id string, now time.Time) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{
		"id":         id,
		"revoked_at": bson.M{"$exists": false},
	}, bson.M{
		"$set": bson.M{
			"revoked_at": now,
			"updated_at": now,
		},
	})
	return err
}
