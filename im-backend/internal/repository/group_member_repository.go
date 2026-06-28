package repository

import (
	"context"
	"time"

	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GroupMemberRepository struct {
	collection *mongo.Collection
}

func NewGroupMemberRepository(db *mongo.Database) *GroupMemberRepository {
	return &GroupMemberRepository{
		collection: db.Collection(models.CollectionGroupMember),
	}
}

func (r *GroupMemberRepository) UpsertActive(ctx context.Context, member *models.GroupMember) (*models.GroupMember, error) {
	now := time.Now()
	if member.ID.IsZero() {
		member.ID = primitive.NewObjectID()
	}
	if member.Status == "" {
		member.Status = models.GroupMemberStatusActive
	}
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}

	filter := bson.M{"group_id": member.GroupID, "user_id": member.UserID}
	update := bson.M{
		"$set": bson.M{
			"role":           member.Role,
			"status":         member.Status,
			"group_nickname": member.GroupNickname,
			"joined_at":      member.JoinedAt,
			"invited_by":     member.InvitedBy,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"_id":        member.ID,
			"group_id":   member.GroupID,
			"user_id":    member.UserID,
			"created_at": now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, member.GroupID, member.UserID)
}

func (r *GroupMemberRepository) Get(ctx context.Context, groupID primitive.ObjectID, userID string) (*models.GroupMember, error) {
	var member models.GroupMember
	err := r.collection.FindOne(ctx, bson.M{"group_id": groupID, "user_id": userID}).Decode(&member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *GroupMemberRepository) GetActive(ctx context.Context, groupID primitive.ObjectID, userID string) (*models.GroupMember, error) {
	var member models.GroupMember
	err := r.collection.FindOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  userID,
		"status":   models.GroupMemberStatusActive,
	}).Decode(&member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *GroupMemberRepository) ListActiveByGroup(ctx context.Context, groupID primitive.ObjectID) ([]*models.GroupMember, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"group_id": groupID,
		"status":   models.GroupMemberStatusActive,
	}, options.Find().SetSort(bson.D{{Key: "joined_at", Value: 1}, {Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*models.GroupMember
	for cursor.Next(ctx) {
		var member models.GroupMember
		if err := cursor.Decode(&member); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	return members, cursor.Err()
}

func (r *GroupMemberRepository) ListActiveUserIDs(ctx context.Context, groupID primitive.ObjectID) ([]string, error) {
	members, err := r.ListActiveByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.UserID)
	}
	return ids, nil
}

func (r *GroupMemberRepository) CountActive(ctx context.Context, groupID primitive.ObjectID) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"group_id": groupID,
		"status":   models.GroupMemberStatusActive,
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *GroupMemberRepository) UpdateStatus(ctx context.Context, groupID primitive.ObjectID, userID string, status models.GroupMemberStatus) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"group_id": groupID, "user_id": userID}, bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *GroupMemberRepository) SetRole(ctx context.Context, groupID primitive.ObjectID, userID string, role models.GroupMemberRole) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  userID,
		"status":   models.GroupMemberStatusActive,
	}, bson.M{
		"$set": bson.M{
			"role":       role,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *GroupMemberRepository) SetAllActiveStatus(ctx context.Context, groupID primitive.ObjectID, status models.GroupMemberStatus) error {
	_, err := r.collection.UpdateMany(ctx, bson.M{
		"group_id": groupID,
		"status":   models.GroupMemberStatusActive,
	}, bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *GroupMemberRepository) DeleteByGroup(ctx context.Context, groupID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"group_id": groupID})
	return err
}
