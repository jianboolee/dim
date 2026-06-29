package repository

import (
	"context"
	"errors"
	"time"

	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	metadata bson.M,
) error {
	set := bson.M{
		"refresh_token_hash": nextHash,
		"updated_at":         now,
		"last_active_at":     now,
		"expires_at":         expiresAt,
	}
	for key, value := range metadata {
		set[key] = value
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{
		"id":                 id,
		"refresh_token_hash": currentHash,
		"revoked_at":         bson.M{"$exists": false},
		"expires_at":         bson.M{"$gt": now},
	}, bson.M{
		"$set": set,
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrAuthSessionConflict
	}
	return nil
}

func (r *AuthSessionRepository) Touch(ctx context.Context, id string, now time.Time) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{
		"id":         id,
		"revoked_at": bson.M{"$exists": false},
		"expires_at": bson.M{"$gt": now},
	}, bson.M{
		"$set": bson.M{
			"last_active_at": now,
			"updated_at":     now,
		},
	})
	return err
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

func (r *AuthSessionRepository) RevokeOverflowActiveSessions(ctx context.Context, userID string, keep int64, now time.Time) error {
	if keep <= 0 {
		_, err := r.collection.UpdateMany(ctx, bson.M{
			"user_id":    userID,
			"revoked_at": bson.M{"$exists": false},
			"expires_at": bson.M{"$gt": now},
		}, bson.M{"$set": bson.M{"revoked_at": now, "updated_at": now}})
		return err
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":    userID,
		"revoked_at": bson.M{"$exists": false},
		"expires_at": bson.M{"$gt": now},
	}, options.Find().
		SetSort(bson.D{{Key: "last_active_at", Value: -1}, {Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(keep).
		SetProjection(bson.M{"id": 1}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	ids := []string{}
	for cursor.Next(ctx) {
		var item struct {
			ID string `bson:"id"`
		}
		if err := cursor.Decode(&item); err != nil {
			return err
		}
		if item.ID != "" {
			ids = append(ids, item.ID)
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	_, err = r.collection.UpdateMany(ctx, bson.M{
		"id":         bson.M{"$in": ids},
		"revoked_at": bson.M{"$exists": false},
	}, bson.M{"$set": bson.M{"revoked_at": now, "updated_at": now}})
	return err
}
