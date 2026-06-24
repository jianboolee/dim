package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"d-im/internal/models"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection(models.CollectionUser),
	}
}

func (r *UserRepository) Upsert(ctx context.Context, user *models.User) error {
	now := time.Now()
	filter := bson.M{"id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"id":         user.ID,
			"bio":        user.Bio,
			"created_at": now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *UserRepository) UpsertMany(ctx context.Context, users []*models.User) error {
	for _, user := range users {
		if err := r.Upsert(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
