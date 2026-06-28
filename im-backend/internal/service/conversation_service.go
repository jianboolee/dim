package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	maxConversationSearchUsers  = int64(100)
)

var ErrInvalidConversationCursor = errors.New("invalid conversation cursor")
var ErrInvalidConversationID = errors.New("invalid conversation id")
var ErrConversationAccessDenied = errors.New("conversation access denied")

type conversationCursor struct {
	SortAt time.Time `json:"sort_at"`
	ID     string    `json:"id"`
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

	conversation.GetLastActivity(currentUserID)
	return s.toConversationDTO(ctx, conversation, currentUserID), nil
}

// GetUserConversations 获取用户的所有会话
func (s *ConversationService) GetUserConversations(ctx context.Context, senderID string, limit int64, cursorValue string, keyword string, activeConversationID string) (*dto.ConversationListResponse, error) {
	limit = normalizeConversationLimit(limit)
	keyword = strings.TrimSpace(keyword)
	activeConversationID = strings.TrimSpace(activeConversationID)

	filter := bson.M{"participants": senderID}
	if keyword != "" {
		if s.userRepo == nil {
			return emptyConversationList(), nil
		}

		users, err := s.userRepo.Search(ctx, keyword, maxConversationSearchUsers)
		if err != nil {
			return nil, fmt.Errorf("failed to search conversation users: %w", err)
		}

		peerIDs := make([]string, 0, len(users))
		seenPeerIDs := map[string]struct{}{}
		for _, user := range users {
			if user.ID == "" || user.ID == senderID {
				continue
			}
			if _, ok := seenPeerIDs[user.ID]; ok {
				continue
			}
			seenPeerIDs[user.ID] = struct{}{}
			peerIDs = append(peerIDs, user.ID)
		}
		if len(peerIDs) == 0 {
			return emptyConversationList(), nil
		}

		filter["participants"] = bson.M{"$all": []string{senderID}, "$in": peerIDs}
	}

	var cursorSortAt time.Time
	var cursorID primitive.ObjectID
	if cursorValue != "" {
		cursor, err := decodeConversationCursor(cursorValue)
		if err != nil {
			return nil, err
		}
		cursorObjectID, err := primitive.ObjectIDFromHex(cursor.ID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid cursor id", ErrInvalidConversationCursor)
		}
		cursorSortAt = cursor.SortAt
		cursorID = cursorObjectID
	}

	if activeConversationID != "" && cursorValue == "" && keyword == "" {
		activeID, err := primitive.ObjectIDFromHex(activeConversationID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid active conversation id", ErrInvalidConversationID)
		}
		if _, err := s.ActivateConversation(ctx, activeID, senderID); err != nil {
			return nil, err
		}
	}

	conversations, err := s.repo.ListUserConversations(ctx, senderID, filter, limit+1, cursorSortAt, cursorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	hasMore := int64(len(conversations)) > limit
	if hasMore {
		conversations = conversations[:limit]
	}

	for i := range conversations {
		conversations[i].GetLastActivity(senderID)
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

func emptyConversationList() *dto.ConversationListResponse {
	return &dto.ConversationListResponse{
		Items:   []*dto.ConversationDTO{},
		HasMore: false,
	}
}

// UpdateLastMessage 更新会话的最后一条消息
func (s *ConversationService) UpdateLastMessage(ctx context.Context, conversationID primitive.ObjectID, message *models.Message) error {
	update := bson.M{
		"$set": bson.M{
			"last_message": message,
			"updated_at":   time.Now(),
		},
	}

	// 如果消息带有图片，更新会话预览图
	if message.Payload != nil && message.Payload.ImageURL != "" {
		update["$set"].(bson.M)["image_url"] = message.Payload.ImageURL
	}

	return s.repo.UpdateConversation(ctx, conversationID, update)
}

// UpdateUnreadCount 更新未读消息数（扣减时不会低于 0）
func (s *ConversationService) UpdateUnreadCount(ctx context.Context, conversationID primitive.ObjectID, userID string, increment int) error {
	return s.repo.UpdateUnreadCount(ctx, conversationID, userID, increment)
}

// MarkConversationRead 标记当前用户已读该会话，清空会话级未读数。
func (s *ConversationService) MarkConversationRead(ctx context.Context, conversationID primitive.ObjectID, userID string) error {
	return s.repo.MarkConversationRead(ctx, conversationID, userID, time.Now())
}

// GetConversations 获取会话列表
func (s *ConversationService) GetConversations(ctx context.Context, userID string) ([]*dto.ConversationDTO, error) {
	conversations, err := s.repo.ListUserConversations(ctx, userID, bson.M{"participants": userID}, 100, time.Time{}, primitive.NilObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// 设置最后活动时间
	for i := range conversations {
		conversations[i].GetLastActivity(userID)
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
		SortAt: conversation.LastActivity,
		ID:     conversation.ID.Hex(),
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
	if cursor.SortAt.IsZero() || cursor.ID == "" {
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
			UserStates:   conversation.UserStates,
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

func (s *ConversationService) ActivateConversation(ctx context.Context, id primitive.ObjectID, currentUserID string) (*dto.ConversationDTO, error) {
	conversation, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if !conversation.HasParticipant(currentUserID) {
		return nil, ErrConversationAccessDenied
	}

	if err := s.repo.ActivateConversation(ctx, id, currentUserID, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to activate conversation: %w", err)
	}

	conversation, err = s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activated conversation: %w", err)
	}
	conversation.GetLastActivity(currentUserID)

	return s.toConversationDTO(ctx, conversation, currentUserID), nil
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
