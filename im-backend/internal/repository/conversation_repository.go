package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"d-im/internal/models"
	"d-im/pkg/logger"
	"d-im/pkg/utils"
)

type ConversationRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{
		db:         db,
		collection: db.Collection(models.CollectionConversation),
	}
}

// GenerateConversationHashID 根据参与者生成唯一、稳定的 ObjectID（顺序无关 + 去重）
func GenerateConversationHashID(participants []string) primitive.ObjectID {
	return utils.GenerateConversationHashID(participants)
}

// CreateConversation 新建会话
func (r *ConversationRepository) CreateConversation(ctx context.Context, conv *models.Conversation) (primitive.ObjectID, error) {
	if conv.ID.IsZero() {
		conv.ID = primitive.NewObjectID()
	}
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, conv)
	return conv.ID, err
}

// GetConversation 获取会话详情
func (r *ConversationRepository) GetConversation(ctx context.Context, id primitive.ObjectID) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&conversation)
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// GetConversationByParticipants 根据参与者获取会话
func (r *ConversationRepository) GetConversationByParticipants(ctx context.Context, participants []string) (*models.Conversation, error) {
	hashID := GenerateConversationHashID(participants)

	var conversation models.Conversation
	err := r.collection.FindOne(ctx, bson.M{"hash_id": hashID}).Decode(&conversation)
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// ListConversations 获取会话列表（可加分页/过滤）
func (r *ConversationRepository) ListConversations(ctx context.Context, filter bson.M, limit, skip int64) ([]*models.Conversation, error) {
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	opts.SetSort(bson.M{"updated_at": -1}) // 默认按更新时间倒序

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conversations []*models.Conversation
	for cursor.Next(ctx) {
		var conv models.Conversation
		if err := cursor.Decode(&conv); err != nil {
			return nil, err
		}
		conversations = append(conversations, &conv)
	}

	return conversations, nil
}

// UpdateConversation 更新会话信息（部分字段）
func (r *ConversationRepository) UpdateConversation(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if update == nil {
		return errors.New("update data cannot be nil")
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// DeleteConversation 删除会话
func (r *ConversationRepository) DeleteConversation(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// GetUnreadCount 获取未读消息数
func (r *ConversationRepository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	log.Printf("Getting unread count for user: %s", userID)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"participants": userID,
			},
		},
		{
			"$project": bson.M{
				"unread": bson.M{"$ifNull": []interface{}{fmt.Sprintf("$unread_counts.%s", userID), 0}},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"total_unread": bson.M{
					"$sum": "$unread",
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error executing aggregation: %v", err)
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalUnread int64 `bson:"total_unread"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Error decoding result: %v", err)
			return 0, err
		}
	}

	log.Printf("Found total unread messages: %d", result.TotalUnread)
	return result.TotalUnread, nil
}

// UpsertConversationByParticipants 创建或更新会话
func (r *ConversationRepository) UpsertConversationByParticipants(
	ctx context.Context,
	participants []string,
	update bson.M,
) (primitive.ObjectID, error) {
	hashID := GenerateConversationHashID(participants)

	filter := bson.M{
		"hash_id": hashID,
	}

	opts := options.Update().SetUpsert(true)

	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if result.UpsertedID != nil {
		if oid, ok := result.UpsertedID.(primitive.ObjectID); ok {
			return oid, nil
		}
	}

	// 如果是更新现有文档，直接返回 conversationID, 确保返回的ID是唯一的
	var conversation models.Conversation
	err = r.collection.FindOne(ctx, filter).Decode(&conversation)
	if err != nil {
		return primitive.NilObjectID, err
	}

	logger.Debug("UpsertConversationByParticipants", zap.Any("conversation", conversation))

	return conversation.ID, nil
}
