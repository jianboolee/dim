package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
}

// NewMessageService 创建消息服务
func NewMessageService(
	repo *repository.MessageRepository,
	conversationRepo *repository.ConversationRepository,
	conversationService *ConversationService,
	sessionService *SessionService,
	wsManager *WSManager,
) *MessageService {
	return &MessageService{
		repo:                repo,
		conversationRepo:    conversationRepo,
		conversationService: conversationService,
		sessionService:      sessionService,
		wsManager:           wsManager,
	}
}

// SendMessageHTTP 通过 HTTP 发送消息
func (s *MessageService) SendMessageHTTP(ctx context.Context, senderID string, receiverID string, content string, msgType *models.MessageType, payload *models.Payload) (*models.Message, error) {
	message, err := s.SendMessage(ctx, senderID, receiverID, content, msgType, payload)
	if err != nil {
		return nil, err
	}

	go s.SendMessageWithWs(ctx, receiverID, message)
	return message, nil
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(
	ctx context.Context,
	senderID string,
	receiverID string,
	content string,
	msgType *models.MessageType,
	payload *models.Payload,
) (*models.Message, error) {

	// 如果msgType为nil，则默认使用文本类型
	msgTypeValue := models.MessageTypeText
	if msgType != nil {
		msgTypeValue = *msgType
	}

	now := time.Now()
	message := &models.Message{
		ID:         primitive.NewObjectID(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Type:       msgTypeValue,
		Payload:    payload, // 这里需要传递指针
		Status:     models.MessageStatusSent,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 创建或更新会话
	conversation, err := s.conversationService.CreateOrUpdateConversationByMessage(ctx, message)
	if err != nil {
		logger.Error("SendMessage CreateOrUpdateConversationByMessage", zap.Any("error", err))
		return nil, fmt.Errorf("failed to create/update conversation: %w", err)
	}

	message.ConversationID = conversation.ID

	logger.Debug(zap.NewDevelopmentEncoderConfig().MessageKey, zap.Any("message", message))

	// 保存消息到数据库
	result, err := s.repo.Save(ctx, message)
	if err != nil {
		logger.Error("SendMessage Save", zap.Any("error", err))
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return result, nil
}

// SendMessageWithWs 通过 WebSocket 发送消息
func (s *MessageService) SendMessageWithWs(ctx context.Context, receiverID string, message *models.Message) error {
	if s.wsManager == nil {
		return fmt.Errorf("wsManager is not initialized")
	}

	// 检查在线状态
	online, err := s.sessionService.IsOnline(receiverID)
	if err != nil {
		return fmt.Errorf("failed to check online status: %w", err)
	}

	if !online {
		return nil
	}

	// 发送消息
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return s.wsManager.SendMessage(receiverID, messageBytes)
}

// GetMessages 获取消息列表
func (s *MessageService) GetMessages(ctx context.Context, senderID string, receiverID string, beforeID *primitive.ObjectID, limit int64) ([]models.Message, error) {

	conversation, err := s.conversationService.GetConversationByParticipants(ctx, []string{senderID, receiverID})
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	messages, err := s.repo.FindMessagesByConversationID(ctx, conversation.ID, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	for i := range messages {
		messages[i].DecodePayload()
	}
	return messages, nil
}

// FindMessagesByConversationID 根据会话ID获取消息列表
func (s *MessageService) FindMessagesByConversationID(ctx context.Context, conversationID primitive.ObjectID, beforeID *primitive.ObjectID, limit int64) ([]models.Message, error) {
	messages, err := s.repo.FindMessagesByConversationID(ctx, conversationID, beforeID, limit)
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

	var inMessage models.Message
	if err := json.Unmarshal(message, &inMessage); err != nil {
		return nil, err
	}

	inMessage.SenderID = senderID

	result, err := s.SendMessage(ctx, senderID, inMessage.ReceiverID, inMessage.Content, &inMessage.Type, inMessage.Payload)
	if err != nil {
		return nil, err
	}

	return result, nil
}
