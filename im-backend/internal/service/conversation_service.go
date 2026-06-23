package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"d-im/internal/models"
	"d-im/internal/repository"
	"d-im/pkg/logger"
)

type ConversationService struct {
	repo *repository.ConversationRepository
}

func NewConversationService(repo *repository.ConversationRepository) *ConversationService {
	return &ConversationService{
		repo: repo,
	}
}

// CreatePrivateConversation 创建单聊会话
func (s *ConversationService) CreatePrivateConversation(ctx context.Context, senderID, receiverID string) (*models.Conversation, error) {
	// 创建新的会话
	participants := []string{senderID, receiverID}
	now := time.Now()
	conversationID, err := s.repo.UpsertConversationByParticipants(ctx, participants, bson.M{
		"$set": bson.M{
			"type":         models.ConversationTypePrivate,
			"participants": participants,
			"created_at":   now,
			"updated_at":   now,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	conversation, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conversation, nil
}

// GetConversation 获取会话详情
func (s *ConversationService) GetConversation(ctx context.Context, id primitive.ObjectID) (*models.Conversation, error) {
	conversation, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	conversation.GetLastActivity()
	return conversation, nil
}

// GetUserConversations 获取用户的所有会话
func (s *ConversationService) GetUserConversations(ctx context.Context, senderID string, limit int64, beforeID *primitive.ObjectID) ([]*models.Conversation, error) {

	filter := bson.M{"participants": senderID}

	if beforeID == nil {
		filter["updated_at"] = bson.M{"$lt": time.Now()}
	} else {
		filter["updated_at"] = bson.M{"$lt": beforeID}
	}

	conversations, err := s.repo.ListConversations(ctx, filter, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// TODO: 为每个会话添加对方的用户信息
	for i := range conversations {
		// 设置最后活动时间
		conversations[i].GetLastActivity()
	}

	return conversations, nil
}

// UpdateLastMessage 更新会话的最后一条消息
func (s *ConversationService) UpdateLastMessage(ctx context.Context, conversationID primitive.ObjectID, message *models.Message) error {
	update := bson.M{
		"$set": bson.M{
			"last_message": message,
			"updated_at":   time.Now(),
		},
	}

	// 如果是卡片消息，更新会话的图片
	if message.Type == models.MessageTypeCard && message.Payload != nil {
		update["$set"].(bson.M)["image_url"] = message.Payload.ImageURL
	}

	return s.repo.UpdateConversation(ctx, conversationID, update)
}

// UpdateUnreadCount 更新未读消息数
func (s *ConversationService) UpdateUnreadCount(ctx context.Context, conversationID primitive.ObjectID, userID string, increment int) error {
	update := bson.M{
		"$inc": bson.M{"unread_counts." + userID: increment},
		"$set": bson.M{"updated_at": time.Now()},
	}

	return s.repo.UpdateConversation(ctx, conversationID, update)
}

// GetConversations 获取会话列表
func (s *ConversationService) GetConversations(ctx context.Context, userID string) ([]*models.Conversation, error) {
	conversations, err := s.repo.ListConversations(ctx, bson.M{"participants": userID}, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// 设置最后活动时间
	for i := range conversations {
		conversations[i].GetLastActivity()
	}

	return conversations, nil
}

// GetUnreadCount 获取未读消息数
func (s *ConversationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// GetConversationByParticipants 根据参与者获取会话
func (s *ConversationService) GetConversationByParticipants(ctx context.Context, participants []string) (*models.Conversation, error) {
	return s.repo.GetConversationByParticipants(ctx, participants)
}

// UpdateConversation 更新会话
func (s *ConversationService) UpdateConversation(ctx context.Context, conversationID primitive.ObjectID, update bson.M) error {
	return s.repo.UpdateConversation(ctx, conversationID, update)
}

// CreateOrUpdateConversation 创建或更新会话（适用于私聊）
func (s *ConversationService) CreateOrUpdateConversation(ctx context.Context, conversation *models.Conversation) (*models.Conversation, error) {
	// 按参与者查找是否已有会话（目前适用于 1v1）
	existingConv, err := s.repo.GetConversationByParticipants(ctx, conversation.Participants)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	// ✅ 如果不存在，创建新会话
	if err == mongo.ErrNoDocuments {
		if conversation.ID.IsZero() {
			conversation.ID = primitive.NewObjectID()
		}

		_, err := s.repo.CreateConversation(ctx, conversation)
		if err != nil {
			return nil, err
		}
		return conversation, nil
	}

	// ✅ 已存在，更新会话内容

	update := bson.M{
		"last_message": conversation.LastMessage,
		"updated_at":   time.Now(),
	}

	err = s.repo.UpdateConversation(ctx, existingConv.ID, update)
	if err != nil {
		return nil, err
	}

	// ✅ 返回更新后的会话（可以选择重新查一遍）
	existingConv.LastMessage = conversation.LastMessage
	existingConv.UpdatedAt = time.Now()
	return existingConv, nil
}

// CreateOrUpdateConversationByMessage 根据消息创建或更新会话
func (s *ConversationService) CreateOrUpdateConversationByMessage(ctx context.Context, message *models.Message) (*models.Conversation, error) {
	now := time.Now()

	conversation := &models.Conversation{
		Type:         models.ConversationTypePrivate,
		Participants: []string{message.SenderID, message.ReceiverID},
		LastMessage:  message,
		UnreadCounts: map[string]int64{message.ReceiverID: 0},
		CreatedAt:    message.CreatedAt,
		UpdatedAt:    now,
	}

	set := bson.M{
		"updated_at": now,
	}

	// 如果是卡片消息，更新会话的图片
	if message.Payload != nil && message.Type == models.MessageTypeCard {
		set["image_url"] = message.Payload.ImageURL
	}

	// Upsert 执行：根据 participants，插入或更新会话
	upsertedID, err := s.repo.UpsertConversationByParticipants(ctx, conversation.Participants, bson.M{
		"$set": set,
		"$setOnInsert": bson.M{
			"type":         conversation.Type,
			"participants": conversation.Participants,
			"created_at":   conversation.CreatedAt,
		},
		"$inc": bson.M{
			"unread_counts." + message.ReceiverID: 1,
		},
	})
	if err != nil {
		return nil, err
	}

	message.ConversationID = upsertedID

	conversation.ID = upsertedID

	// 更新会话的最后一条消息， 异步执行
	go func() {
		err = s.repo.UpdateConversation(context.Background(), upsertedID, bson.M{
			"$set": bson.M{
				"last_message": message,
			},
		})
		if err != nil {
			logger.Error("UpdateConversation", zap.Any("error", err))
		}
	}()

	return conversation, nil
}
