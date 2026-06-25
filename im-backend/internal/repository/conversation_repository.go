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
	hashID := GenerateConversationHashID(utils.NormalizeParticipantIDs(participants)).Hex()

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
	opts.SetSort(bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: -1}}) // 默认按更新时间倒序

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

func (r *ConversationRepository) ListUserConversations(
	ctx context.Context,
	userID string,
	filter bson.M,
	limit int64,
	cursorSortAt time.Time,
	cursorID primitive.ObjectID,
) ([]*models.Conversation, error) {
	if filter == nil {
		filter = bson.M{}
	}

	sortAtExpression := bson.M{
		"$max": bson.A{
			bson.M{"$ifNull": bson.A{"$last_message.created_at", "$updated_at"}},
			bson.M{"$ifNull": bson.A{fmt.Sprintf("$user_states.%s.last_activated_at", userID), "$updated_at"}},
			"$updated_at",
		},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$addFields", Value: bson.M{"sort_at": sortAtExpression}}},
	}

	if !cursorSortAt.IsZero() && !cursorID.IsZero() {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.M{
			"$or": []bson.M{
				{"sort_at": bson.M{"$lt": cursorSortAt}},
				{
					"sort_at": cursorSortAt,
					"_id":     bson.M{"$lt": cursorID},
				},
			},
		}}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{
		{Key: "sort_at", Value: -1},
		{Key: "_id", Value: -1},
	}}})

	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
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
	if err := cursor.Err(); err != nil {
		return nil, err
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

// UpdateUnreadCount 增减会话未读数；扣减时仅在当前未读数大于 0 时执行
func (r *ConversationRepository) UpdateUnreadCount(
	ctx context.Context,
	id primitive.ObjectID,
	userID string,
	delta int,
) error {
	field := "user_states." + userID + ".unread_count"
	filter := bson.M{"_id": id}
	if delta < 0 {
		filter[field] = bson.M{"$gt": 0}
	}

	update := bson.M{
		"$inc": bson.M{field: delta},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *ConversationRepository) MarkConversationRead(ctx context.Context, id primitive.ObjectID, userID string, readAt time.Time) error {
	fieldPrefix := "user_states." + userID
	update := bson.M{
		"$set": bson.M{
			fieldPrefix + ".unread_count": int64(0),
		},
		"$max": bson.M{
			fieldPrefix + ".last_read_at": readAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "participants": userID}, update)
	return err
}

func (r *ConversationRepository) ActivateConversation(ctx context.Context, id primitive.ObjectID, userID string, activatedAt time.Time) error {
	fieldPrefix := "user_states." + userID
	update := bson.M{
		"$set": bson.M{
			fieldPrefix + ".last_activated_at": activatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "participants": userID}, update)
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
				"unread": bson.M{
					"$max": []interface{}{
						0,
						bson.M{"$ifNull": []interface{}{fmt.Sprintf("$user_states.%s.unread_count", userID), 0}},
					},
				},
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

// UpsertConversationByParticipants 创建或更新私聊会话（参与者顺序无关）
func (r *ConversationRepository) UpsertConversationByParticipants(
	ctx context.Context,
	participants []string,
) (*models.Conversation, error) {
	normalized := utils.NormalizeParticipantIDs(participants)
	hashID := GenerateConversationHashID(normalized).Hex()
	now := time.Now()

	filter := bson.M{"hash_id": hashID}
	update := bson.M{
		"$set": bson.M{
			"type":         models.ConversationTypePrivate,
			"participants": normalized,
			"updated_at":   now,
		},
		"$setOnInsert": bson.M{
			"hash_id":     hashID,
			"user_states": bson.M{},
			"created_at":  now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	var conversation models.Conversation
	if err := r.collection.FindOne(ctx, filter).Decode(&conversation); err != nil {
		return nil, err
	}

	logger.Debug("UpsertConversationByParticipants", zap.Any("conversation", conversation))

	return &conversation, nil
}
