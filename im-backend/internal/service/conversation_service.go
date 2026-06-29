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
	repo            *repository.ConversationRepository
	memberRepo      *repository.ConversationMemberRepository
	userRepo        *repository.UserRepository
	groupRepo       *repository.GroupRepository
	groupMemberRepo *repository.GroupMemberRepository
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

func NewConversationService(
	repo *repository.ConversationRepository,
	memberRepo *repository.ConversationMemberRepository,
	userRepo *repository.UserRepository,
	groupRepo *repository.GroupRepository,
	groupMemberRepo *repository.GroupMemberRepository,
) *ConversationService {
	return &ConversationService{
		repo:            repo,
		memberRepo:      memberRepo,
		userRepo:        userRepo,
		groupRepo:       groupRepo,
		groupMemberRepo: groupMemberRepo,
	}
}

// CreatePrivateConversation 创建或获取单聊会话
func (s *ConversationService) CreatePrivateConversation(ctx context.Context, userID, peerID string) (*models.Conversation, error) {
	return s.getOrCreatePrivateConversation(ctx, userID, peerID)
}

// GetOrCreatePrivateConversation 获取或创建单聊会话（参与者顺序无关）
func (s *ConversationService) GetOrCreatePrivateConversation(ctx context.Context, userAID, userBID string) (*models.Conversation, error) {
	return s.getOrCreatePrivateConversation(ctx, userAID, userBID)
}

func (s *ConversationService) getOrCreatePrivateConversation(ctx context.Context, userAID, userBID string) (*models.Conversation, error) {
	conversation, err := s.repo.UpsertConversationByParticipants(ctx, []string{userAID, userBID})
	if err != nil {
		return nil, err
	}
	if s.memberRepo != nil {
		now := time.Now()
		if _, err := s.memberRepo.UpsertActive(ctx, conversation.ID, userAID, "", now); err != nil {
			return nil, err
		}
		if _, err := s.memberRepo.UpsertActive(ctx, conversation.ID, userBID, "", now); err != nil {
			return nil, err
		}
	}
	return conversation, nil
}

// GetConversation 获取会话详情
func (s *ConversationService) GetConversation(ctx context.Context, id primitive.ObjectID, currentUserID string) (*dto.ConversationDTO, error) {
	conversation, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	member, err := s.memberRepo.GetActive(ctx, id, currentUserID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConversationAccessDenied
		}
		return nil, fmt.Errorf("failed to get conversation member: %w", err)
	}
	if member == nil || !member.IsActive() {
		return nil, ErrConversationAccessDenied
	}

	conversation.LastActivity = conversation.GetLastActivityWithMember(member)
	return s.toConversationDTO(ctx, conversation, currentUserID, map[primitive.ObjectID]*models.ConversationMember{id: member}), nil
}

// GetUserConversations 获取用户的所有会话
func (s *ConversationService) GetUserConversations(ctx context.Context, senderID string, limit int64, cursorValue string, keyword string, activeConversationID string) (*dto.ConversationListResponse, error) {
	limit = normalizeConversationLimit(limit)
	keyword = strings.TrimSpace(keyword)
	activeConversationID = strings.TrimSpace(activeConversationID)

	// 搜索分支：先在 DB 层面搜到匹配的会话 ID，再取 member
	if keyword != "" {
		return s.searchUserConversations(ctx, senderID, limit, cursorValue, keyword)
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

	if activeConversationID != "" && cursorValue == "" {
		activeID, err := primitive.ObjectIDFromHex(activeConversationID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid active conversation id", ErrInvalidConversationID)
		}
		if _, err := s.ActivateConversation(ctx, activeID, senderID); err != nil {
			return nil, err
		}
	}

	memberPage, err := s.memberRepo.ListByUser(ctx, senderID, limit+1, cursorSortAt, cursorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation members: %w", err)
	}
	memberByConversationID := map[primitive.ObjectID]*models.ConversationMember{}
	conversationIDs := make([]primitive.ObjectID, 0, len(memberPage))
	for _, member := range memberPage {
		memberByConversationID[member.ConversationID] = member
		conversationIDs = append(conversationIDs, member.ConversationID)
	}

	conversationsByID, err := s.repo.GetConversationsByIDs(ctx, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	conversations := make([]*models.Conversation, 0, len(memberPage))
	for _, member := range memberPage {
		conversation := conversationsByID[member.ConversationID]
		if conversation == nil {
			continue
		}
		conversation.LastActivity = conversation.GetLastActivityWithMember(member)
		conversations = append(conversations, conversation)
	}

	hasMore := int64(len(conversations)) > limit
	if hasMore {
		conversations = conversations[:limit]
	}

	nextCursor := ""
	if hasMore && len(conversations) > 0 {
		last := conversations[len(conversations)-1]
		if member := memberByConversationID[last.ID]; member != nil {
			nextCursor = encodeConversationMemberCursor(member)
		}
	}

	return &dto.ConversationListResponse{
		Items:      s.toConversationDTOs(ctx, conversations, senderID, memberByConversationID),
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// searchUserConversations 按关键词搜索当前用户的会话。
//
// 流程: user keyword → peerIDs → SearchConversationIDs → ListByUserAndConversationIDs → 拼装结果。
func (s *ConversationService) searchUserConversations(
	ctx context.Context,
	senderID string,
	limit int64,
	cursorValue string,
	keyword string,
) (*dto.ConversationListResponse, error) {
	if s.userRepo == nil {
		return emptyConversationList(), nil
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

	// 1. 搜用户，收集 peerIDs
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

	// 2. 按参与者 ID + 群名搜会话 ID
	convIDs, err := s.repo.SearchConversationIDs(ctx, peerIDs, keyword, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search conversation ids: %w", err)
	}
	if len(convIDs) == 0 {
		return emptyConversationList(), nil
	}

	// 3. 取当前用户在这些会话中的 member 记录
	memberPage, err := s.memberRepo.ListByUserAndConversationIDs(ctx, senderID, convIDs, limit+1, cursorSortAt, cursorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation members: %w", err)
	}

	// 4. 分页
	hasMore := int64(len(memberPage)) > limit
	if hasMore {
		memberPage = memberPage[:limit]
	}

	// 5. 取会话详情
	conversationIDs := make([]primitive.ObjectID, 0, len(memberPage))
	memberByConversationID := map[primitive.ObjectID]*models.ConversationMember{}
	for _, m := range memberPage {
		conversationIDs = append(conversationIDs, m.ConversationID)
		memberByConversationID[m.ConversationID] = m
	}
	conversationsByID, err := s.repo.GetConversationsByIDs(ctx, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	conversations := make([]*models.Conversation, 0, len(memberPage))
	for _, m := range memberPage {
		conv := conversationsByID[m.ConversationID]
		if conv == nil {
			continue
		}
		conv.LastActivity = conv.GetLastActivityWithMember(m)
		conversations = append(conversations, conv)
	}

	nextCursor := ""
	if hasMore && len(conversations) > 0 {
		last := conversations[len(conversations)-1]
		if member := memberByConversationID[last.ID]; member != nil {
			nextCursor = encodeConversationMemberCursor(member)
		}
	}

	return &dto.ConversationListResponse{
		Items:      s.toConversationDTOs(ctx, conversations, senderID, memberByConversationID),
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

// UpdateLastMessage 更新会话的最后一条消息快照
func (s *ConversationService) UpdateLastMessage(ctx context.Context, conversationID primitive.ObjectID, message *models.Message) error {
	snapshot := &models.LastMessageSnapshot{
		Content:   message.Content,
		Type:      string(message.Type),
		CreatedAt: message.CreatedAt,
	}

	update := bson.M{
		"$set": bson.M{
			"last_message": snapshot,
			"updated_at":   time.Now(),
		},
	}

	// 仅卡片消息更新会话预览图
	if message.Type == models.MessageTypeCard && message.Payload != nil && message.Payload.ImageURL != "" {
		update["$set"].(bson.M)["image_url"] = message.Payload.ImageURL
	}

	return s.repo.UpdateConversation(ctx, conversationID, update)
}

// MarkConversationRead 标记当前用户已读该会话，清空会话级未读数。
func (s *ConversationService) MarkConversationRead(ctx context.Context, conversationID primitive.ObjectID, userID string) error {
	conversation, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}
	if _, err := s.memberRepo.GetActive(ctx, conversationID, userID); err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrConversationAccessDenied
		}
		return err
	}
	return s.memberRepo.MarkRead(ctx, conversationID, userID, conversation.MessageSeq, primitive.NilObjectID, time.Now())
}

// GetConversations 获取会话列表
func (s *ConversationService) GetConversations(ctx context.Context, userID string) ([]*dto.ConversationDTO, error) {
	response, err := s.GetUserConversations(ctx, userID, 100, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	return response.Items, nil
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

func encodeConversationMemberCursor(member *models.ConversationMember) string {
	payload, err := json.Marshal(conversationCursor{
		SortAt: member.SortAt,
		ID:     member.ID.Hex(),
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

func (s *ConversationService) toConversationDTO(ctx context.Context, conversation *models.Conversation, currentUserID string, membersByConversationID map[primitive.ObjectID]*models.ConversationMember) *dto.ConversationDTO {
	dtos := s.toConversationDTOs(ctx, []*models.Conversation{conversation}, currentUserID, membersByConversationID)
	if len(dtos) == 0 {
		return nil
	}
	return dtos[0]
}

func (s *ConversationService) toConversationDTOs(ctx context.Context, conversations []*models.Conversation, currentUserID string, membersByConversationID map[primitive.ObjectID]*models.ConversationMember) []*dto.ConversationDTO {
	peerIDs := make([]string, 0, len(conversations))
	groupIDs := make([]primitive.ObjectID, 0, len(conversations))
	seen := map[string]struct{}{}

	for _, conversation := range conversations {
		if conversation.Type == models.ConversationTypeGroup && conversation.GroupID != nil {
			groupIDs = append(groupIDs, *conversation.GroupID)
		}
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

	groupInfoByID := s.loadGroupSummaries(ctx, groupIDs)

	results := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		item := &dto.ConversationDTO{
			ID:            conversation.ID,
			Type:          conversation.Type,
			Participants:  conversation.Participants,
			LastMessage:   conversation.LastMessage,
			DisplayName:   conversation.DisplayName,
			DisplayAvatar: conversation.ImageURL,
			GroupID:       conversation.GroupID,
			ImageURL:      conversation.ImageURL,
			LastActivity:  conversation.LastActivity,
			CreatedAt:     conversation.CreatedAt,
			UpdatedAt:     conversation.UpdatedAt,
		}
		if member := membersByConversationID[conversation.ID]; member != nil {
			item.MemberState = &dto.ConversationMemberStateDTO{
				Status:          member.Status,
				LastReadSeq:     member.LastReadSeq,
				LastReadAt:      member.LastReadAt,
				LastActivatedAt: member.LastActivatedAt,
				UnreadCount:     member.UnreadCount,
				MentionCount:    member.MentionCount,
				Muted:           member.Muted,
				Pinned:          member.Pinned,
			}
		}

		if conversation.Type == models.ConversationTypeGroup && conversation.GroupID != nil {
			if groupInfo := groupInfoByID[conversation.GroupID.Hex()]; groupInfo != nil {
				item.GroupInfo = groupInfo
				item.DisplayName = groupInfo.Name
				item.DisplayAvatar = groupInfo.AvatarURL
			}
		} else {
			for _, participantID := range conversation.Participants {
				if participantID == currentUserID {
					continue
				}
				if user := usersByID[participantID]; user != nil {
					item.PeerUserInfo = dto.ConvertToUserInfoDto(user)
					item.DisplayName = item.PeerUserInfo.Nickname
					item.DisplayAvatar = item.PeerUserInfo.Avatar
				}
				break
			}
		}

		results = append(results, item)
	}

	return results
}

func (s *ConversationService) loadGroupSummaries(ctx context.Context, groupIDs []primitive.ObjectID) map[string]*dto.GroupSummaryDTO {
	results := map[string]*dto.GroupSummaryDTO{}
	if s.groupRepo == nil || s.groupMemberRepo == nil || len(groupIDs) == 0 {
		return results
	}

	for _, groupID := range groupIDs {
		if _, exists := results[groupID.Hex()]; exists {
			continue
		}

		group, err := s.groupRepo.Get(ctx, groupID)
		if err != nil {
			logger.Error("Get group summary", zap.String("group_id", groupID.Hex()), zap.Error(err))
			continue
		}

		members, err := s.groupMemberRepo.ListActiveByGroup(ctx, groupID)
		if err != nil {
			logger.Error("List group members for summary", zap.String("group_id", groupID.Hex()), zap.Error(err))
			continue
		}

		briefMembers := make([]dto.GroupMemberBriefDTO, 0, min(len(members), 4))
		userIDs := make([]string, 0, min(len(members), 4))
		for i, member := range members {
			if i >= 4 {
				break
			}
			userIDs = append(userIDs, member.UserID)
		}

		usersByID := map[string]*models.User{}
		if s.userRepo != nil && len(userIDs) > 0 {
			if users, err := s.userRepo.FindByIDs(ctx, userIDs); err == nil {
				usersByID = users
			}
		}

		for i, member := range members {
			if i >= 4 {
				break
			}
			brief := dto.GroupMemberBriefDTO{
				UserID:        member.UserID,
				Role:          member.Role,
				GroupNickname: member.GroupNickname,
			}
			if user := usersByID[member.UserID]; user != nil {
				brief.UserInfo = dto.ConvertToUserInfoDto(user)
			}
			briefMembers = append(briefMembers, brief)
		}

		results[groupID.Hex()] = &dto.GroupSummaryDTO{
			ID:          group.ID,
			Name:        group.Name,
			AvatarURL:   group.AvatarURL,
			MemberCount: group.MemberCount,
			Members:     briefMembers,
		}
	}

	return results
}

func (s *ConversationService) ActivateConversation(ctx context.Context, id primitive.ObjectID, currentUserID string) (*dto.ConversationDTO, error) {
	conversation, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if _, err := s.memberRepo.GetActive(ctx, id, currentUserID); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConversationAccessDenied
		}
		return nil, err
	}

	if err := s.memberRepo.Activate(ctx, id, currentUserID, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to activate conversation: %w", err)
	}

	conversation, err = s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activated conversation: %w", err)
	}
	member, err := s.memberRepo.GetActive(ctx, id, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activated conversation member: %w", err)
	}
	conversation.LastActivity = conversation.GetLastActivityWithMember(member)
	return s.toConversationDTO(ctx, conversation, currentUserID, map[primitive.ObjectID]*models.ConversationMember{id: member}), nil
}

// GetUnreadCount 获取未读消息数
func (s *ConversationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.memberRepo.SumUnreadByUser(ctx, userID)
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
