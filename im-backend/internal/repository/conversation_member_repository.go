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

type ConversationMemberRepository struct {
	collection *mongo.Collection
}

func NewConversationMemberRepository(db *mongo.Database) *ConversationMemberRepository {
	return &ConversationMemberRepository{
		collection: db.Collection(models.CollectionConversationMember),
	}
}

func (r *ConversationMemberRepository) UpsertActive(
	ctx context.Context,
	conversationID primitive.ObjectID,
	userID string,
	roleSnapshot string,
	joinedAt time.Time,
) (*models.ConversationMember, error) {
	now := time.Now()
	if joinedAt.IsZero() {
		joinedAt = now
	}

	filter := bson.M{"conversation_id": conversationID, "user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"status":        models.ConversationMemberStatusActive,
			"role_snapshot": roleSnapshot,
			"updated_at":    now,
		},
		"$setOnInsert": bson.M{
			"_id":             primitive.NewObjectID(),
			"conversation_id": conversationID,
			"user_id":         userID,
			"joined_at":       joinedAt,
			"sort_at":         joinedAt,
			"last_read_seq":   int64(0),
			"unread_count":    int64(0),
			"mention_count":   int64(0),
			"muted":           false,
			"pinned":          false,
			"created_at":      now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, conversationID, userID)
}

func (r *ConversationMemberRepository) Get(ctx context.Context, conversationID primitive.ObjectID, userID string) (*models.ConversationMember, error) {
	var member models.ConversationMember
	err := r.collection.FindOne(ctx, bson.M{"conversation_id": conversationID, "user_id": userID}).Decode(&member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *ConversationMemberRepository) GetActive(ctx context.Context, conversationID primitive.ObjectID, userID string) (*models.ConversationMember, error) {
	var member models.ConversationMember
	err := r.collection.FindOne(ctx, bson.M{
		"conversation_id": conversationID,
		"user_id":         userID,
		"status":          models.ConversationMemberStatusActive,
	}).Decode(&member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *ConversationMemberRepository) ListActiveByConversation(ctx context.Context, conversationID primitive.ObjectID) ([]*models.ConversationMember, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"conversation_id": conversationID,
		"status":          models.ConversationMemberStatusActive,
	}, options.Find().SetSort(bson.D{{Key: "joined_at", Value: 1}, {Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*models.ConversationMember
	for cursor.Next(ctx) {
		var member models.ConversationMember
		if err := cursor.Decode(&member); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	return members, cursor.Err()
}

func (r *ConversationMemberRepository) ListActiveUserIDs(ctx context.Context, conversationID primitive.ObjectID) ([]string, error) {
	members, err := r.ListActiveByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.UserID)
	}
	return ids, nil
}

func (r *ConversationMemberRepository) ListByUser(
	ctx context.Context,
	userID string,
	limit int64,
	cursorPinned bool,
	cursorSortAt time.Time,
	cursorID primitive.ObjectID,
) ([]*models.ConversationMember, error) {
	filter := bson.M{
		"user_id": userID,
		"status":  models.ConversationMemberStatusActive,
	}
	if !cursorSortAt.IsZero() && !cursorID.IsZero() {
		filter["$or"] = []bson.M{
			{"pinned": bson.M{"$lt": cursorPinned}},
			{
				"pinned":  cursorPinned,
				"sort_at": bson.M{"$lt": cursorSortAt},
			},
			{
				"pinned":  cursorPinned,
				"sort_at": cursorSortAt,
				"_id":     bson.M{"$lt": cursorID},
			},
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "pinned", Value: -1}, {Key: "sort_at", Value: -1}, {Key: "_id", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*models.ConversationMember
	for cursor.Next(ctx) {
		var member models.ConversationMember
		if err := cursor.Decode(&member); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	return members, cursor.Err()
}

// ListByUserAndConversationIDs 获取用户在指定会话列表中的成员记录。
func (r *ConversationMemberRepository) ListByUserAndConversationIDs(
	ctx context.Context,
	userID string,
	conversationIDs []primitive.ObjectID,
	limit int64,
	cursorPinned bool,
	cursorSortAt time.Time,
	cursorID primitive.ObjectID,
) ([]*models.ConversationMember, error) {
	if len(conversationIDs) == 0 {
		return nil, nil
	}

	filter := bson.M{
		"user_id":         userID,
		"status":          models.ConversationMemberStatusActive,
		"conversation_id": bson.M{"$in": conversationIDs},
	}
	if !cursorSortAt.IsZero() && !cursorID.IsZero() {
		filter["$or"] = []bson.M{
			{"pinned": bson.M{"$lt": cursorPinned}},
			{
				"pinned":  cursorPinned,
				"sort_at": bson.M{"$lt": cursorSortAt},
			},
			{
				"pinned":  cursorPinned,
				"sort_at": cursorSortAt,
				"_id":     bson.M{"$lt": cursorID},
			},
		}
	}
	opts := options.Find().SetSort(bson.D{{Key: "pinned", Value: -1}, {Key: "sort_at", Value: -1}, {Key: "_id", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*models.ConversationMember
	for cursor.Next(ctx) {
		var member models.ConversationMember
		if err := cursor.Decode(&member); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	return members, cursor.Err()
}

func (r *ConversationMemberRepository) UpdateSettings(ctx context.Context, conversationID primitive.ObjectID, userID string, pinned *bool, muted *bool) (*models.ConversationMember, error) {
	now := time.Now()
	set := bson.M{"updated_at": now}
	if pinned != nil {
		set["pinned"] = *pinned
		if *pinned {
			set["pinned_at"] = now
		} else {
			set["pinned_at"] = time.Time{}
		}
	}
	if muted != nil {
		set["muted"] = *muted
		if *muted {
			set["muted_at"] = now
		} else {
			set["muted_at"] = time.Time{}
		}
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{
		"conversation_id": conversationID,
		"user_id":         userID,
		"status":          models.ConversationMemberStatusActive,
	}, bson.M{"$set": set})
	if err != nil {
		return nil, err
	}
	return r.GetActive(ctx, conversationID, userID)
}

func (r *ConversationMemberRepository) SetStatus(ctx context.Context, conversationID primitive.ObjectID, userID string, status models.ConversationMemberStatus) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"conversation_id": conversationID, "user_id": userID}, bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *ConversationMemberRepository) SetAllActiveStatus(ctx context.Context, conversationID primitive.ObjectID, status models.ConversationMemberStatus) error {
	_, err := r.collection.UpdateMany(ctx, bson.M{
		"conversation_id": conversationID,
		"status":          models.ConversationMemberStatusActive,
	}, bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *ConversationMemberRepository) TouchMembers(ctx context.Context, conversationID primitive.ObjectID, userIDs []string, at time.Time) error {
	if len(userIDs) == 0 {
		return nil
	}
	_, err := r.collection.UpdateMany(ctx, bson.M{
		"conversation_id": conversationID,
		"user_id":         bson.M{"$in": userIDs},
		"status":          models.ConversationMemberStatusActive,
	}, bson.M{
		"$set": bson.M{"sort_at": at, "updated_at": at},
	})
	return err
}

func (r *ConversationMemberRepository) MarkRead(ctx context.Context, conversationID primitive.ObjectID, userID string, seq int64, messageID primitive.ObjectID, readAt time.Time) error {
	set := bson.M{
		"unread_count":  0,
		"last_read_at":  readAt,
		"last_read_seq": seq,
		"updated_at":    readAt,
	}
	if !messageID.IsZero() {
		set["last_read_message_id"] = messageID
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{
		"conversation_id": conversationID,
		"user_id":         userID,
		"status":          models.ConversationMemberStatusActive,
	}, bson.M{"$set": set})
	return err
}

func (r *ConversationMemberRepository) Activate(ctx context.Context, conversationID primitive.ObjectID, userID string, activatedAt time.Time) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{
		"conversation_id": conversationID,
		"user_id":         userID,
		"status":          models.ConversationMemberStatusActive,
	}, bson.M{
		"$set": bson.M{
			"last_activated_at": activatedAt,
			"sort_at":           activatedAt,
			"updated_at":        activatedAt,
		},
	})
	return err
}

func (r *ConversationMemberRepository) IncrementUnreadForOthers(ctx context.Context, conversationID primitive.ObjectID, senderID string, seq int64, messageID primitive.ObjectID, at time.Time) error {
	if _, err := r.collection.UpdateMany(ctx, bson.M{
		"conversation_id": conversationID,
		"user_id":         bson.M{"$ne": senderID},
		"status":          models.ConversationMemberStatusActive,
	}, bson.M{
		"$inc": bson.M{"unread_count": int64(1)},
		"$set": bson.M{"sort_at": at, "updated_at": at},
	}); err != nil {
		return err
	}
	return r.MarkRead(ctx, conversationID, senderID, seq, messageID, at)
}

func (r *ConversationMemberRepository) SumUnreadByUser(ctx context.Context, userID string) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"status":  models.ConversationMemberStatusActive,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"total_unread": bson.M{"$sum": bson.M{
				"$max": bson.A{"$unread_count", int64(0)},
			}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalUnread int64 `bson:"total_unread"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
	}
	return result.TotalUnread, cursor.Err()
}

func (r *ConversationMemberRepository) DeleteByConversation(ctx context.Context, conversationID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"conversation_id": conversationID})
	return err
}
