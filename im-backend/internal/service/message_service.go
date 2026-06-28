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
		conversationService: conversationService,
		groupRepo:           groupRepo,
		groupMemberRepo:     groupMemberRepo,
		sessionService:      sessionService,
		wsManager:           wsManager,
		redisClient:         redisClient,
		userRepo:            userRepo,
	}
}

// SendMessageToConversationHTTP 通过 HTTP 发送会话消息，并异步推送给接收方。
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

// SendMessageToConversation 发送会话消息。当前项目只支持单聊，因此接收方由会话参与者推导。
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
	if !conversation.HasParticipant(senderID) {
		return nil, false, ErrConversationAccessDenied
	}

	receiverID := ""
	if conversation.Type == models.ConversationTypePrivate {
		var err error
		receiverID, err = resolvePrivateReceiverID(conversation, senderID)
		if err != nil {
			return nil, false, err
		}

		// 检查对方是否为系统用户（只读），普通用户不能回复系统消息
		if s.userRepo != nil {
			if peerUser, err := s.userRepo.GetByID(ctx, receiverID); err == nil && peerUser.Type == models.UserTypeSystem {
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

	now := time.Now()
	message := &models.Message{
		ID:              primitive.NewObjectID(),
		ClientMessageID: clientMessageID,
		ConversationID:  conversation.ID,
		SenderID:        senderID,
		ReceiverID:      receiverID,
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

	return result, true, nil
}

func resolvePrivateReceiverID(conversation *models.Conversation, senderID string) (string, error) {
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
	if _, err := s.groupMemberRepo.GetActive(ctx, group.ID, senderID); err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrConversationAccessDenied
		}
		return err
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

	recipientIDs, err := s.resolveFanoutRecipients(ctx, conversation)
	if err != nil {
		return err
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, userID := range recipientIDs {
		if err := s.pushToUser(ctx, userID, messageBytes); err != nil {
			logger.Error("failed to push message via ws", zap.String("receiver_id", userID), zap.Error(err))
		}
	}
	return nil
}

func (s *MessageService) resolveFanoutRecipients(ctx context.Context, conversation *models.Conversation) ([]string, error) {
	var ids []string
	switch conversation.Type {
	case models.ConversationTypePrivate, models.ConversationTypeSystem, models.ConversationTypeChannel:
		ids = append(ids, conversation.Participants...)
	case models.ConversationTypeGroup:
		if conversation.GroupID == nil || s.groupMemberRepo == nil {
			return nil, fmt.Errorf("invalid group conversation")
		}
		activeIDs, err := s.groupMemberRepo.ListActiveUserIDs(ctx, *conversation.GroupID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, activeIDs...)
	default:
		return nil, fmt.Errorf("unsupported conversation type")
	}

	seen := map[string]struct{}{}
	results := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		results = append(results, id)
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
	now := time.Now()
	message := &models.Message{
		ID:             primitive.NewObjectID(),
		ConversationID: conversationID,
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
	return s.FanoutMessage(ctx, result)
}

// SendMessageWithWs 将消息推送给在线用户（本进程 WS 或经 Redis 转发到 WS 进程）
func (s *MessageService) SendMessageWithWs(ctx context.Context, receiverID string, message *models.Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return s.pushToUser(ctx, receiverID, messageBytes)
}

func (s *MessageService) pushToUser(ctx context.Context, receiverID string, messageBytes []byte) error {
	if receiverID == "" {
		return nil
	}

	// 同进程直连（cmd/server 单进程模式）
	if s.wsManager != nil && s.wsManager.TryDeliver(receiverID, messageBytes) {
		return nil
	}

	// 跨进程：API server → Redis → WS server
	if s.redisClient == nil {
		return nil
	}

	event := WSPushEvent{
		ReceiverID: receiverID,
		Message:    json.RawMessage(messageBytes),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ws push event: %w", err)
	}

	return s.redisClient.Publish(context.Background(), wsPushChannel, payload).Err()
}

// FindMessagesByConversationID 根据会话ID获取消息列表
func (s *MessageService) FindMessagesByConversationID(ctx context.Context, conversationID primitive.ObjectID, currentUserID string, beforeID *primitive.ObjectID, afterID *primitive.ObjectID, limit int64) ([]models.Message, error) {
	conversation, err := s.conversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if !conversation.HasParticipant(currentUserID) {
		return nil, ErrConversationAccessDenied
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

// 注意：GetConversations方法应该在ConversationService中实现，这里删除

// MarkMessageAsDelivered 标记消息为已送达
func (s *MessageService) MarkMessageAsDelivered(ctx context.Context, messageID primitive.ObjectID, currentUserID string) error {
	// 先获取消息信息
	message, err := s.repo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to find message: %w", err)
	}

	// 检查权限：只有消息接收者可以标记消息为已送达
	if message.ReceiverID != currentUserID {
		return fmt.Errorf("permission denied: only message recipient can mark message as delivered")
	}

	return s.repo.UpdateStatus(ctx, messageID, models.MessageStatusDelivered)
}

// MarkMessageAsRead 标记消息为已读
func (s *MessageService) MarkMessageAsRead(ctx context.Context, messageID primitive.ObjectID, currentUserID string) error {
	// 先获取消息信息
	message, err := s.repo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to find message: %w", err)
	}

	// 检查权限：只有消息接收者可以标记消息为已读
	if message.ReceiverID != currentUserID {
		return fmt.Errorf("permission denied: only message recipient can mark message as read")
	}

	// 已读消息不重复扣减未读数
	if message.Status == models.MessageStatusRead {
		return nil
	}

	// 更新消息状态
	if err := s.repo.UpdateStatus(ctx, messageID, models.MessageStatusRead); err != nil {
		return err
	}

	// 如果消息被标记为已读，更新会话的未读数
	if err := s.conversationService.UpdateUnreadCount(ctx, message.ConversationID, currentUserID, -1); err != nil {
		return fmt.Errorf("failed to update conversation unread count: %w", err)
	}

	return nil
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
