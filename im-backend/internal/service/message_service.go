package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"d-im/internal/models"
	"d-im/internal/repository"
	"d-im/pkg/logger"
)

var ErrCannotReplyToSystemUser = errors.New("cannot reply to system user")

// MessageService 消息服务
type MessageService struct {
	repo                *repository.MessageRepository
	conversationRepo    *repository.ConversationRepository
	memberRepo          *repository.ConversationMemberRepository
	conversationService *ConversationService
	groupRepo           *repository.GroupRepository
	groupMemberRepo     *repository.GroupMemberRepository
	sessionService      *SessionService
	wsManager           *WSManager
	redisClient         *redis.Client
	userRepo            *repository.UserRepository
}

// NewMessageService 创建消息服务
func NewMessageService(
	repo *repository.MessageRepository,
	conversationRepo *repository.ConversationRepository,
	memberRepo *repository.ConversationMemberRepository,
	conversationService *ConversationService,
	groupRepo *repository.GroupRepository,
	groupMemberRepo *repository.GroupMemberRepository,
	sessionService *SessionService,
	wsManager *WSManager,
	redisClient *redis.Client,
	userRepo *repository.UserRepository,
) *MessageService {
	return &MessageService{
		repo:                repo,
		conversationRepo:    conversationRepo,
		memberRepo:          memberRepo,
		conversationService: conversationService,
		groupRepo:           groupRepo,
		groupMemberRepo:     groupMemberRepo,
		sessionService:      sessionService,
		wsManager:           wsManager,
		redisClient:         redisClient,
		userRepo:            userRepo,
	}
}

// SendMessageToConversationHTTP 通过 HTTP 发送会话消息，并异步推送给会话成员。
func (s *MessageService) SendMessageToConversationHTTP(ctx context.Context, senderID string, conversationID primitive.ObjectID, clientMessageID string, content string, msgType *models.MessageType, payload *models.Payload) (*models.Message, error) {
	message, created, err := s.SendMessageToConversation(ctx, senderID, conversationID, clientMessageID, content, msgType, payload)
	if err != nil {
		return nil, err
	}
	if !created {
		return message, nil
	}

	go func(msg *models.Message) {
		// 异步推送不可使用 HTTP 请求 context（响应结束后会被 cancel）
		if err := s.FanoutMessage(context.Background(), msg); err != nil {
			logger.Error("failed to fanout message via ws", zap.String("message_id", msg.ID.Hex()), zap.Error(err))
		}
	}(message)

	return message, nil
}

// SendMessageToConversation 发送会话消息。单聊和群聊统一按 conversation_members 做权限与未读控制。
func (s *MessageService) SendMessageToConversation(
	ctx context.Context,
	senderID string,
	conversationID primitive.ObjectID,
	clientMessageID string,
	content string,
	msgType *models.MessageType,
	payload *models.Payload,
) (*models.Message, bool, error) {
	conversation, err := s.conversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get conversation: %w", err)
	}
	if _, err := s.memberRepo.GetActive(ctx, conversation.ID, senderID); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, false, ErrConversationAccessDenied
		}
		return nil, false, fmt.Errorf("failed to get conversation member: %w", err)
	}

	if conversation.Type == models.ConversationTypePrivate {
		peerID, err := resolvePrivatePeerID(conversation, senderID)
		if err != nil {
			return nil, false, err
		}

		// 检查对方是否为系统用户（只读），普通用户不能回复系统消息
		if s.userRepo != nil {
			if peerUser, err := s.userRepo.GetByID(ctx, peerID); err == nil && peerUser.Type == models.UserTypeSystem {
				return nil, false, ErrCannotReplyToSystemUser
			}
		}
	} else if conversation.Type == models.ConversationTypeGroup {
		if err := s.ensureCanSendGroupMessage(ctx, conversation, senderID); err != nil {
			return nil, false, err
		}
	} else {
		return nil, false, fmt.Errorf("unsupported conversation type")
	}

	if clientMessageID != "" {
		existing, err := s.repo.FindByClientMessageID(ctx, conversation.ID, senderID, clientMessageID)
		if err == nil {
			existing.DecodePayload()
			return existing, false, nil
		}
		if err != mongo.ErrNoDocuments {
			return nil, false, fmt.Errorf("failed to find message by client id: %w", err)
		}
	}

	// 如果msgType为nil，则默认使用文本类型
	msgTypeValue := models.MessageTypeText
	if msgType != nil {
		msgTypeValue = *msgType
	}

	seq, err := s.conversationRepo.NextMessageSeq(ctx, conversation.ID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to allocate message seq: %w", err)
	}

	now := time.Now()
	message := &models.Message{
		ID:              primitive.NewObjectID(),
		ClientMessageID: clientMessageID,
		ConversationID:  conversation.ID,
		Seq:             seq,
		SenderID:        senderID,
		Content:         content,
		Type:            msgTypeValue,
		Payload:         payload,
		Status:          models.MessageStatusSent,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 非文本消息由后端统一生成摘要
	message.GenerateDigest()

	logger.Debug(zap.NewDevelopmentEncoderConfig().MessageKey, zap.Any("message", message))

	// 保存消息到数据库
	result, err := s.repo.Save(ctx, message)
	if err != nil {
		if clientMessageID != "" && mongo.IsDuplicateKeyError(err) {
			existing, findErr := s.repo.FindByClientMessageID(ctx, conversation.ID, senderID, clientMessageID)
			if findErr == nil {
				existing.DecodePayload()
				return existing, false, nil
			}
		}
		logger.Error("SendMessage Save", zap.Any("error", err))
		return nil, false, fmt.Errorf("failed to save message: %w", err)
	}

	if err := s.conversationService.UpdateLastMessage(ctx, conversation.ID, result); err != nil {
		return nil, false, fmt.Errorf("failed to update conversation last message: %w", err)
	}
	if err := s.memberRepo.IncrementUnreadForOthers(ctx, conversation.ID, senderID, result.Seq, result.ID, result.CreatedAt); err != nil {
		return nil, false, fmt.Errorf("failed to update conversation unread state: %w", err)
	}

	return result, true, nil
}

func resolvePrivatePeerID(conversation *models.Conversation, senderID string) (string, error) {
	if conversation.Type != models.ConversationTypePrivate || len(conversation.Participants) != 2 {
		return "", fmt.Errorf("unsupported conversation type")
	}

	for _, participantID := range conversation.Participants {
		if participantID != senderID {
			return participantID, nil
		}
	}

	return "", ErrConversationAccessDenied
}

func (s *MessageService) ensureCanSendGroupMessage(ctx context.Context, conversation *models.Conversation, senderID string) error {
	if conversation.GroupID == nil || s.groupRepo == nil || s.groupMemberRepo == nil {
		return fmt.Errorf("invalid group conversation")
	}
	group, err := s.groupRepo.Get(ctx, *conversation.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}
	if !group.IsActive() {
		return ErrGroupDissolved
	}
	return nil
}

func (s *MessageService) FanoutMessage(ctx context.Context, message *models.Message) error {
	if message == nil {
		return nil
	}

	conversation, err := s.conversationRepo.GetConversation(ctx, message.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation for fanout: %w", err)
	}

	recipientIDs, err := s.resolveFanoutRecipients(ctx, conversation.ID)
	if err != nil {
		return err
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, userID := range recipientIDs {
		if err := s.pushToUser(ctx, userID, messageBytes); err != nil {
			logger.Error("failed to push message via ws", zap.String("user_id", userID), zap.Error(err))
		}
	}
	return nil
}

func (s *MessageService) resolveFanoutRecipients(ctx context.Context, conversationID primitive.ObjectID) ([]string, error) {
	members, err := s.memberRepo.ListActiveByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	results := make([]string, 0, len(members))
	for _, member := range members {
		if member.UserID == "" {
			continue
		}
		if _, ok := seen[member.UserID]; ok {
			continue
		}
		seen[member.UserID] = struct{}{}
		results = append(results, member.UserID)
	}
	return results, nil
}

func (s *MessageService) CreateSystemEvent(
	ctx context.Context,
	conversationID primitive.ObjectID,
	operatorID string,
	content string,
	payload *models.Payload,
) error {
	seq, err := s.conversationRepo.NextMessageSeq(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to allocate system event seq: %w", err)
	}

	now := time.Now()
	message := &models.Message{
		ID:             primitive.NewObjectID(),
		ConversationID: conversationID,
		Seq:            seq,
		SenderID:       operatorID,
		Type:           models.MessageTypeSystemEvent,
		Content:        content,
		Payload:        payload,
		Status:         models.MessageStatusSent,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result, err := s.repo.Save(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to save system event: %w", err)
	}
	if err := s.conversationService.UpdateLastMessage(ctx, conversationID, result); err != nil {
		return fmt.Errorf("failed to update conversation last message: %w", err)
	}
	if err := s.memberRepo.IncrementUnreadForOthers(ctx, conversationID, operatorID, result.Seq, result.ID, result.CreatedAt); err != nil {
		return fmt.Errorf("failed to update system event unread state: %w", err)
	}
	return s.FanoutMessage(ctx, result)
}

func (s *MessageService) pushToUser(ctx context.Context, userID string, messageBytes []byte) error {
	if userID == "" {
		return nil
	}

	// 同进程直连（cmd/server 单进程模式）
	if s.wsManager != nil && s.wsManager.TryDeliver(userID, messageBytes) {
		return nil
	}

	// 跨进程：API server → Redis → WS server
	if s.redisClient == nil {
		return nil
	}

	event := WSPushEvent{
		UserID:  userID,
		Message: json.RawMessage(messageBytes),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ws push event: %w", err)
	}

	return s.redisClient.Publish(context.Background(), wsPushChannel, payload).Err()
}

// FindMessagesByConversationID 根据会话ID获取消息列表
func (s *MessageService) FindMessagesByConversationID(ctx context.Context, conversationID primitive.ObjectID, currentUserID string, beforeID *primitive.ObjectID, afterID *primitive.ObjectID, limit int64) ([]models.Message, error) {
	if _, err := s.memberRepo.GetActive(ctx, conversationID, currentUserID); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConversationAccessDenied
		}
		return nil, err
	}

	messages, err := s.repo.FindMessagesByConversationID(ctx, conversationID, beforeID, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	for i := range messages {
		messages[i].DecodePayload()
	}
	return messages, nil
}

// ProcessWebSocketMessage 处理 WebSocket 消息
func (s *MessageService) ProcessWebSocketMessage(
	ctx context.Context,
	senderID string,
	message []byte,
) (*models.Message, error) {
	var raw struct {
		ConversationID  string             `json:"conversation_id"`
		ClientMessageID string             `json:"client_message_id"`
		Content         string             `json:"content"`
		Type            models.MessageType `json:"type"`
		Payload         *models.Payload    `json:"payload"`
	}
	if err := json.Unmarshal(message, &raw); err != nil {
		return nil, err
	}

	conversationID, err := primitive.ObjectIDFromHex(raw.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("invalid conversation_id")
	}

	msgType := raw.Type
	if msgType == "" {
		msgType = models.MessageTypeText
	}

	result, created, err := s.SendMessageToConversation(ctx, senderID, conversationID, raw.ClientMessageID, raw.Content, &msgType, raw.Payload)
	if err != nil {
		return nil, err
	}
	if created {
		if err := s.FanoutMessage(ctx, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
