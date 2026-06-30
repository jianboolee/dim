package repository

import (
	"context"
	"errors"
	"time"

	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GroupRepository struct {
	collection *mongo.Collection
}

func NewGroupRepository(db *mongo.Database) *GroupRepository {
	return &GroupRepository{
		collection: db.Collection(models.CollectionGroup),
	}
}

func (r *GroupRepository) Create(ctx context.Context, group *models.Group) (*models.Group, error) {
	now := time.Now()
	if group.ID.IsZero() {
		group.ID = primitive.NewObjectID()
	}
	if group.Status == "" {
		group.Status = models.GroupStatusActive
	}
	group.CreatedAt = now
	group.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *GroupRepository) Get(ctx context.Context, id primitive.ObjectID) (*models.Group, error) {
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) GetByConversationID(ctx context.Context, conversationID primitive.ObjectID) (*models.Group, error) {
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{"conversation_id": conversationID}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) GetActiveByUniqueKey(ctx context.Context, scopeUserID string, uniqueKey string) (*models.Group, error) {
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{
		"scope_user_id": scopeUserID,
		"unique_key":    uniqueKey,
		"status":        models.GroupStatusActive,
	}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if update == nil {
		return errors.New("update data cannot be nil")
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *GroupRepository) SetMemberCount(ctx context.Context, id primitive.ObjectID, count int) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"member_count": count,
			"updated_at":   time.Now(),
		},
	})
	return err
}

func (r *GroupRepository) SetAvatar(ctx context.Context, id primitive.ObjectID, avatarURL string, filePath string, version int64, updatedAt time.Time) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"avatar_url":        avatarURL,
			"avatar_file_path":  filePath,
			"avatar_version":    version,
			"avatar_updated_at": updatedAt,
			"updated_at":        updatedAt,
		},
	})
	return err
}

func (r *GroupRepository) Dissolve(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"status":     models.GroupStatusDissolved,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *GroupRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
