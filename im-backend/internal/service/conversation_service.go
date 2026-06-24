package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"d-im/internal/dto"
	"d-im/internal/models"
	"d-im/internal/repository"
	"d-im/pkg/logger"
)

type ConversationService struct {
	repo     *repository.ConversationRepository
	userRepo *repository.UserRepository
}

const (
	defaultConversationPageSize = int64(20)
	maxConversationPageSize     = int64(50)
)

var ErrInvalidConversationCursor = errors.New("invalid conversation cursor")
var ErrConversationAccessDenied = errors.New("conversation access denied")

type conversationCursor struct {
	UpdatedAt time.Time `json:"updated_at"`
	ID        string    `json:"id"`
}

func NewConversationService(repo *repository.ConversationRepository, userRepo *repository.UserRepository) *ConversationService {
	return &ConversationService{
		repo:     repo,
		userRepo: userRepo,
	}
}

// CreatePrivateConversation 创建或获取单聊会话
func (s *ConversationService) CreatePrivateConversation(ctx context.Context, senderID, receiverID string) (*models.Conversation, error) {
	return s.repo.UpsertConversationByParticipants(ctx, []string{senderID, receiverID})
}

// GetOrCreatePrivateConversation 获取或创建单聊会话（参与者顺序无关）
func (s *ConversationService) GetOrCreatePrivateConversation(ctx context.Context, userAID, userBID string) (*models.Conversation, error) {
	return s.repo.UpsertConversationByParticipants(ctx, []string{userAID, userBID})
}

// GetConversation 获取会话详情
func (s *ConversationService) GetConversation(ctx context.Context, id primitive.ObjectID, currentUserID string) (*dto.ConversationDTO, error) {
	conversation, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if !conversation.HasParticipant(currentUserID) {
		return nil, ErrConversationAccessDenied
	}

	conversation.GetLastActivity()
	return s.toConversationDTO(ctx, conversation, currentUserID), nil
}

// GetUserConversations 获取用户的所有会话
func (s *ConversationService) GetUserConversations(ctx context.Context, senderID string, limit int64, cursorValue string) (*dto.ConversationListResponse, error) {
	limit = normalizeConversationLimit(limit)

	filter := bson.M{"participants": senderID}

	if cursorValue != "" {
		cursor, err := decodeConversationCursor(cursorValue)
		if err != nil {
			return nil, err
		}
		cursorID, err := primitive.ObjectIDFromHex(cursor.ID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid cursor id", ErrInvalidConversationCursor)
		}
		filter["$or"] = []bson.M{
			{"updated_at": bson.M{"$lt": cursor.UpdatedAt}},
			{
				"updated_at": cursor.UpdatedAt,
				"_id":        bson.M{"$lt": cursorID},
			},
		}
	}

	conversations, err := s.repo.ListConversations(ctx, filter, limit+1, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	hasMore := int64(len(conversations)) > limit
	if hasMore {
		conversations = conversations[:limit]
	}

	for i := range conversations {
		// 设置最后活动时间
		conversations[i].GetLastActivity()
	}

	nextCursor := ""
	if hasMore && len(conversations) > 0 {
		nextCursor = encodeConversationCursor(conversations[len(conversations)-1])
	}

	return &dto.ConversationListResponse{
		Items:      s.toConversationDTOs(ctx, conversations, senderID),
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
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

// UpdateUnreadCount 更新未读消息数（扣减时不会低于 0）
func (s *ConversationService) UpdateUnreadCount(ctx context.Context, conversationID primitive.ObjectID, userID string, increment int) error {
	return s.repo.UpdateUnreadCount(ctx, conversationID, userID, increment)
}

// GetConversations 获取会话列表
func (s *ConversationService) GetConversations(ctx context.Context, userID string) ([]*dto.ConversationDTO, error) {
	conversations, err := s.repo.ListConversations(ctx, bson.M{"participants": userID}, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// 设置最后活动时间
	for i := range conversations {
		conversations[i].GetLastActivity()
	}

	return s.toConversationDTOs(ctx, conversations, userID), nil
}

func normalizeConversationLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultConversationPageSize
	}
	if limit > maxConversationPageSize {
		return maxConversationPageSize
	}
	return limit
}

func encodeConversationCursor(conversation *models.Conversation) string {
	payload, err := json.Marshal(conversationCursor{
		UpdatedAt: conversation.UpdatedAt,
		ID:        conversation.ID.Hex(),
	})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}

func decodeConversationCursor(value string) (*conversationCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConversationCursor, err)
	}

	var cursor conversationCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConversationCursor, err)
	}
	if cursor.UpdatedAt.IsZero() || cursor.ID == "" {
		return nil, ErrInvalidConversationCursor
	}

	return &cursor, nil
}

func (s *ConversationService) toConversationDTO(ctx context.Context, conversation *models.Conversation, currentUserID string) *dto.ConversationDTO {
	dtos := s.toConversationDTOs(ctx, []*models.Conversation{conversation}, currentUserID)
	if len(dtos) == 0 {
		return nil
	}
	return dtos[0]
}

func (s *ConversationService) toConversationDTOs(ctx context.Context, conversations []*models.Conversation, currentUserID string) []*dto.ConversationDTO {
	peerIDs := make([]string, 0, len(conversations))
	seen := map[string]struct{}{}

	for _, conversation := range conversations {
		for _, participantID := range conversation.Participants {
			if participantID == currentUserID {
				continue
			}
			if _, ok := seen[participantID]; ok {
				continue
			}
			seen[participantID] = struct{}{}
			peerIDs = append(peerIDs, participantID)
		}
	}

	usersByID := map[string]*models.User{}
	if s.userRepo != nil && len(peerIDs) > 0 {
		users, err := s.userRepo.FindByIDs(ctx, peerIDs)
		if err != nil {
			logger.Error("Find conversation users", zap.Error(err))
		} else {
			usersByID = users
		}
	}

	results := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		item := &dto.ConversationDTO{
			ID:           conversation.ID,
			Type:         conversation.Type,
			Participants: conversation.Participants,
			LastMessage:  conversation.LastMessage,
			ImageURL:     conversation.ImageURL,
			UnreadCounts: conversation.UnreadCounts,
			LastActivity: conversation.LastActivity,
			CreatedAt:    conversation.CreatedAt,
			UpdatedAt:    conversation.UpdatedAt,
		}

		for _, participantID := range conversation.Participants {
			if participantID == currentUserID {
				continue
			}
			if user := usersByID[participantID]; user != nil {
				item.ToUserInfo = dto.ConvertToUserInfoDto(user)
			}
			break
		}

		results = append(results, item)
	}

	return results
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

	// Upsert 执行：根据 participants，插入或更新会话
	upsertedConversation, err := s.repo.UpsertConversationByParticipants(ctx, conversation.Participants)
	if err != nil {
		return nil, err
	}

	message.ConversationID = upsertedConversation.ID
	conversation.ID = upsertedConversation.ID

	// 更新会话的最后一条消息与未读数
	go func() {
		err = s.repo.UpdateConversation(context.Background(), upsertedConversation.ID, bson.M{
			"$set": bson.M{
				"last_message": message,
				"updated_at":   time.Now(),
			},
			"$inc": bson.M{
				"unread_counts." + message.ReceiverID: 1,
			},
		})
		if err != nil {
			logger.Error("UpdateConversation", zap.Any("error", err))
		}
	}()

	return conversation, nil
}
