package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"d-im/internal/models"
	"d-im/internal/repository"
	"d-im/pkg/logger"
)

// MessageService 消息服务
type MessageService struct {
	repo                *repository.MessageRepository
	conversationRepo    *repository.ConversationRepository
	conversationService *ConversationService
	sessionService      *SessionService
	wsManager           *WSManager
	redisClient         *redis.Client
}

// NewMessageService 创建消息服务
func NewMessageService(
	repo *repository.MessageRepository,
	conversationRepo *repository.ConversationRepository,
	conversationService *ConversationService,
	sessionService *SessionService,
	wsManager *WSManager,
	redisClient *redis.Client,
) *MessageService {
	return &MessageService{
		repo:                repo,
		conversationRepo:    conversationRepo,
		conversationService: conversationService,
		sessionService:      sessionService,
		wsManager:           wsManager,
		redisClient:         redisClient,
	}
}

// SendMessageToConversationHTTP 通过 HTTP 发送会话消息，并异步推送给接收方。
func (s *MessageService) SendMessageToConversationHTTP(ctx context.Context, senderID string, conversationID primitive.ObjectID, content string, msgType *models.MessageType, payload *models.Payload) (*models.Message, error) {
	message, err := s.SendMessageToConversation(ctx, senderID, conversationID, content, msgType, payload)
	if err != nil {
		return nil, err
	}

	// 异步推送不可使用 HTTP 请求 context（响应结束后会被 cancel）
	go func(msg *models.Message) {
		if err := s.SendMessageWithWs(context.Background(), msg.ReceiverID, msg); err != nil {
			logger.Error("failed to push message via ws", zap.String("receiver_id", msg.ReceiverID), zap.Error(err))
		}
		if msg.SenderID != msg.ReceiverID {
			if err := s.SendMessageWithWs(context.Background(), msg.SenderID, msg); err != nil {
				logger.Error("failed to echo message via ws", zap.String("sender_id", msg.SenderID), zap.Error(err))
			}
		}
	}(message)

	return message, nil
}

// SendMessageToConversation 发送会话消息。当前项目只支持单聊，因此接收方由会话参与者推导。
func (s *MessageService) SendMessageToConversation(
	ctx context.Context,
	senderID string,
	conversationID primitive.ObjectID,
	content string,
	msgType *models.MessageType,
	payload *models.Payload,
) (*models.Message, error) {
	conversation, err := s.conversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if !conversation.HasParticipant(senderID) {
		return nil, ErrConversationAccessDenied
	}

	receiverID, err := resolvePrivateReceiverID(conversation, senderID)
	if err != nil {
		return nil, err
	}

	// 如果msgType为nil，则默认使用文本类型
	msgTypeValue := models.MessageTypeText
	if msgType != nil {
		msgTypeValue = *msgType
	}

	now := time.Now()
	message := &models.Message{
		ID:             primitive.NewObjectID(),
		ConversationID: conversation.ID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Content:        content,
		Type:           msgTypeValue,
		Payload:        payload, // 这里需要传递指针
		Status:         models.MessageStatusSent,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	logger.Debug(zap.NewDevelopmentEncoderConfig().MessageKey, zap.Any("message", message))

	// 保存消息到数据库
	result, err := s.repo.Save(ctx, message)
	if err != nil {
		logger.Error("SendMessage Save", zap.Any("error", err))
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	if err := s.conversationService.UpdateLastMessage(ctx, conversation.ID, result); err != nil {
		return nil, fmt.Errorf("failed to update conversation last message: %w", err)
	}
	if err := s.conversationService.UpdateUnreadCount(ctx, conversation.ID, receiverID, 1); err != nil {
		return nil, fmt.Errorf("failed to update conversation unread count: %w", err)
	}

	return result, nil
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
		ConversationID string             `json:"conversation_id"`
		Content        string             `json:"content"`
		Type           models.MessageType `json:"type"`
		Payload        *models.Payload    `json:"payload"`
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

	result, err := s.SendMessageToConversation(ctx, senderID, conversationID, raw.Content, &msgType, raw.Payload)
	if err != nil {
		return nil, err
	}

	return result, nil
}
