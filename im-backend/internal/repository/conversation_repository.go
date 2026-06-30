package repository

import (
	"context"
	"errors"
	"regexp"
	"strings"
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

func (r *ConversationRepository) CreateGroupConversation(
	ctx context.Context,
	groupID primitive.ObjectID,
	name string,
	participants []string,
) (*models.Conversation, error) {
	now := time.Now()
	conv := &models.Conversation{
		ID:           primitive.NewObjectID(),
		Type:         models.ConversationTypeGroup,
		GroupID:      &groupID,
		Participants: participants,
		DisplayName:  name,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	conv.NormalizeParticipants()

	_, err := r.collection.InsertOne(ctx, conv)
	if err != nil {
		return nil, err
	}
	return conv, nil
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

func (r *ConversationRepository) GetConversationsByIDs(ctx context.Context, ids []primitive.ObjectID) (map[primitive.ObjectID]*models.Conversation, error) {
	if len(ids) == 0 {
		return map[primitive.ObjectID]*models.Conversation{}, nil
	}
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	conversations := map[primitive.ObjectID]*models.Conversation{}
	for cursor.Next(ctx) {
		var conversation models.Conversation
		if err := cursor.Decode(&conversation); err != nil {
			return nil, err
		}
		conversations[conversation.ID] = &conversation
	}
	return conversations, cursor.Err()
}

func (r *ConversationRepository) NextMessageSeq(ctx context.Context, id primitive.ObjectID) (int64, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var conversation models.Conversation
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"message_seq": int64(1)},
		"$set": bson.M{"updated_at": time.Now()},
	}, opts).Decode(&conversation)
	if err != nil {
		return 0, err
	}
	return conversation.MessageSeq, nil
}

func (r *ConversationRepository) GetConversationByGroupID(ctx context.Context, groupID primitive.ObjectID) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.collection.FindOne(ctx, bson.M{"group_id": groupID}).Decode(&conversation)
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

// UpdateConversation 更新会话信息（部分字段）
func (r *ConversationRepository) UpdateConversation(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if update == nil {
		return errors.New("update data cannot be nil")
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *ConversationRepository) SetParticipants(ctx context.Context, id primitive.ObjectID, participants []string) error {
	normalized := utils.NormalizeParticipantIDs(participants)
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"participants": normalized,
			"updated_at":   time.Now(),
		},
	})
	return err
}

// SearchConversationIDs 根据参与者 ID 和/或群名关键词搜索匹配的会话 ID。
//
//   - participantIDs: 匹配其中任意一个参与者的私聊会话
//   - groupKeyword:   按 display_name 模糊匹配群聊（为空则忽略）
func (r *ConversationRepository) SearchConversationIDs(
	ctx context.Context,
	participantIDs []string,
	groupKeyword string,
	limit int64,
) ([]primitive.ObjectID, error) {
	if len(participantIDs) == 0 && groupKeyword == "" {
		return nil, nil
	}

	conditions := []bson.M{}
	if len(participantIDs) > 0 {
		conditions = append(conditions, bson.M{
			"type":         models.ConversationTypePrivate,
			"participants": bson.M{"$in": participantIDs},
		})
	}
	if kw := strings.TrimSpace(groupKeyword); kw != "" {
		conditions = append(conditions, bson.M{
			"type":         models.ConversationTypeGroup,
			"display_name": bson.M{"$regex": regexp.QuoteMeta(kw), "$options": "i"},
		})
	}

	if len(conditions) == 0 {
		return nil, nil
	}

	filter := bson.M{"$or": conditions}
	opts := options.Find().SetProjection(bson.M{"_id": 1})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	ids := make([]primitive.ObjectID, 0)
	for cursor.Next(ctx) {
		var result struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		ids = append(ids, result.ID)
	}
	return ids, cursor.Err()
}

// DeleteConversation 删除会话
func (r *ConversationRepository) DeleteConversation(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
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
			"hash_id":    hashID,
			"created_at": now,
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
