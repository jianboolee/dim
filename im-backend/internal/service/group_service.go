package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"d-im/internal/dto"
	"d-im/internal/models"
	"d-im/internal/repository"
	"d-im/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const maxGroupMembers = 500

var (
	ErrGroupNotFound         = errors.New("group not found")
	ErrGroupDissolved        = errors.New("group dissolved")
	ErrGroupAccessDenied     = errors.New("group access denied")
	ErrGroupPermissionDenied = errors.New("group permission denied")
	ErrGroupMemberLimit      = errors.New("group member limit exceeded")
	ErrGroupOwnerRequired    = errors.New("group owner required")
)

type GroupService struct {
	groupRepo              *repository.GroupRepository
	memberRepo             *repository.GroupMemberRepository
	conversationMemberRepo *repository.ConversationMemberRepository
	conversationRepo       *repository.ConversationRepository
	userRepo               *repository.UserRepository
	messageService         *MessageService
}

func NewGroupService(
	groupRepo *repository.GroupRepository,
	memberRepo *repository.GroupMemberRepository,
	conversationMemberRepo *repository.ConversationMemberRepository,
	conversationRepo *repository.ConversationRepository,
	userRepo *repository.UserRepository,
	messageService *MessageService,
) *GroupService {
	return &GroupService{
		groupRepo:              groupRepo,
		memberRepo:             memberRepo,
		conversationMemberRepo: conversationMemberRepo,
		conversationRepo:       conversationRepo,
		userRepo:               userRepo,
		messageService:         messageService,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, creatorID string, req dto.GroupCreateRequest) (*dto.GroupDetailResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "群聊"
	}

	groupID := primitive.NewObjectID()
	memberIDs := normalizeGroupMemberIDs(append([]string{creatorID}, req.MemberIDs...))
	if len(memberIDs) > maxGroupMembers {
		return nil, ErrGroupMemberLimit
	}
	conversation, err := s.conversationRepo.CreateGroupConversation(ctx, groupID, name, strings.TrimSpace(req.AvatarURL), memberIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create group conversation: %w", err)
	}

	group, err := s.groupRepo.Create(ctx, &models.Group{
		ID:             groupID,
		ConversationID: conversation.ID,
		Name:           name,
		AvatarURL:      strings.TrimSpace(req.AvatarURL),
		OwnerID:        creatorID,
		MemberCount:    len(memberIDs),
		Status:         models.GroupStatusActive,
	})
	if err != nil {
		_ = s.conversationRepo.DeleteConversation(ctx, conversation.ID)
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	if err := s.upsertMembers(ctx, group.ID, creatorID, creatorID, memberIDs); err != nil {
		s.cleanupFailedGroupCreate(ctx, group, conversation.ID)
		return nil, err
	}
	if err := s.upsertConversationMembers(ctx, conversation.ID, memberIDs, group.OwnerID); err != nil {
		s.cleanupFailedGroupCreate(ctx, group, conversation.ID)
		return nil, err
	}
	if err := s.syncGroupMembers(ctx, group); err != nil {
		s.cleanupFailedGroupCreate(ctx, group, conversation.ID)
		return nil, err
	}
	_ = s.emitGroupEvent(ctx, group, creatorID, models.SystemEventGroupCreated, memberIDs, "", name)

	return s.GetGroup(ctx, group.ID, creatorID)
}

func (s *GroupService) cleanupFailedGroupCreate(ctx context.Context, group *models.Group, conversationID primitive.ObjectID) {
	if group != nil {
		_ = s.memberRepo.DeleteByGroup(ctx, group.ID)
		_ = s.conversationMemberRepo.DeleteByConversation(ctx, conversationID)
		_ = s.groupRepo.Delete(ctx, group.ID)
	}
	_ = s.conversationRepo.DeleteConversation(ctx, conversationID)
}

func (s *GroupService) GetGroup(ctx context.Context, groupID primitive.ObjectID, currentUserID string) (*dto.GroupDetailResponse, error) {
	group, err := s.getActiveAccessibleGroup(ctx, groupID, currentUserID)
	if err != nil {
		return nil, err
	}
	members, err := s.memberRepo.ListActiveByGroup(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	return &dto.GroupDetailResponse{
		Group:   dto.ConvertToGroupDTO(group),
		Members: s.toMemberDTOs(ctx, members),
	}, nil
}

func (s *GroupService) ListMembers(ctx context.Context, groupID primitive.ObjectID, currentUserID string) ([]dto.GroupMemberDTO, error) {
	group, err := s.getActiveAccessibleGroup(ctx, groupID, currentUserID)
	if err != nil {
		return nil, err
	}
	members, err := s.memberRepo.ListActiveByGroup(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	return s.toMemberDTOs(ctx, members), nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, groupID primitive.ObjectID, currentUserID string, req dto.GroupUpdateRequest) (*dto.GroupDetailResponse, error) {
	group, err := s.getActiveAccessibleGroup(ctx, groupID, currentUserID)
	if err != nil {
		return nil, err
	}

	set := bson.M{}
	var eventType, beforeValue, afterValue string
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			name = "群聊"
		}
		if name != group.Name {
			set["name"] = name
			eventType = models.SystemEventGroupNameUpdated
			beforeValue = group.Name
			afterValue = name
			group.Name = name
		}
	}
	if req.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*req.AvatarURL)
		if avatarURL != group.AvatarURL {
			set["avatar_url"] = avatarURL
			eventType = models.SystemEventGroupAvatarUpdated
			beforeValue = group.AvatarURL
			afterValue = avatarURL
			group.AvatarURL = avatarURL
		}
	}
	if len(set) > 0 {
		now := time.Now()
		set["updated_at"] = now
		if err := s.groupRepo.Update(ctx, group.ID, bson.M{"$set": set}); err != nil {
			return nil, err
		}
		if err := s.conversationRepo.UpdateConversation(ctx, group.ConversationID, bson.M{"$set": bson.M{
			"display_name": group.Name,
			"image_url":    group.AvatarURL,
			"updated_at":   now,
		}}); err != nil {
			return nil, err
		}
		_ = s.emitGroupEvent(ctx, group, currentUserID, eventType, nil, beforeValue, afterValue)
	}

	return s.GetGroup(ctx, group.ID, currentUserID)
}

func (s *GroupService) AddMembers(ctx context.Context, groupID primitive.ObjectID, currentUserID string, userIDs []string) (*dto.GroupDetailResponse, error) {
	group, err := s.getActiveAccessibleGroup(ctx, groupID, currentUserID)
	if err != nil {
		return nil, err
	}

	requestedIDs := normalizeGroupMemberIDs(userIDs)
	if len(requestedIDs) == 0 {
		return s.GetGroup(ctx, group.ID, currentUserID)
	}

	activeMembers, err := s.memberRepo.ListActiveByGroup(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	activeSet := map[string]struct{}{}
	for _, member := range activeMembers {
		activeSet[member.UserID] = struct{}{}
	}

	newIDs := make([]string, 0, len(requestedIDs))
	for _, userID := range requestedIDs {
		if _, exists := activeSet[userID]; exists {
			continue
		}
		newIDs = append(newIDs, userID)
	}
	if len(activeSet)+len(newIDs) > maxGroupMembers {
		return nil, ErrGroupMemberLimit
	}
	if len(newIDs) == 0 {
		return s.GetGroup(ctx, group.ID, currentUserID)
	}

	if err := s.upsertMembers(ctx, group.ID, currentUserID, group.OwnerID, newIDs); err != nil {
		return nil, err
	}
	if err := s.upsertConversationMembers(ctx, group.ConversationID, newIDs, group.OwnerID); err != nil {
		return nil, err
	}
	if err := s.syncGroupMembers(ctx, group); err != nil {
		return nil, err
	}
	_ = s.emitGroupEvent(ctx, group, currentUserID, models.SystemEventMemberJoined, newIDs, "", "")

	return s.GetGroup(ctx, group.ID, currentUserID)
}

func (s *GroupService) KickMember(ctx context.Context, groupID primitive.ObjectID, currentUserID string, targetUserID string) (*dto.GroupDetailResponse, error) {
	group, err := s.getActiveAccessibleGroup(ctx, groupID, currentUserID)
	if err != nil {
		return nil, err
	}
	actor, err := s.memberRepo.GetActive(ctx, group.ID, currentUserID)
	if err != nil {
		return nil, ErrGroupAccessDenied
	}
	target, err := s.memberRepo.GetActive(ctx, group.ID, targetUserID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	if !actor.CanKick(target) {
		return nil, ErrGroupPermissionDenied
	}

	if err := s.memberRepo.UpdateStatus(ctx, group.ID, targetUserID, models.GroupMemberStatusKicked); err != nil {
		return nil, err
	}
	if err := s.conversationMemberRepo.SetStatus(ctx, group.ConversationID, targetUserID, models.ConversationMemberStatusKicked); err != nil {
		return nil, err
	}
	if err := s.syncGroupMembers(ctx, group); err != nil {
		return nil, err
	}
	_ = s.emitGroupEvent(ctx, group, currentUserID, models.SystemEventMemberKicked, []string{targetUserID}, "", "")

	return s.GetGroup(ctx, group.ID, currentUserID)
}

func (s *GroupService) LeaveGroup(ctx context.Context, groupID primitive.ObjectID, currentUserID string) error {
	group, err := s.getActiveAccessibleGroup(ctx, groupID, currentUserID)
	if err != nil {
		return err
	}
	member, err := s.memberRepo.GetActive(ctx, group.ID, currentUserID)
	if err != nil {
		return ErrGroupAccessDenied
	}

	if member.Role == models.GroupMemberRoleOwner {
		_ = s.emitGroupEvent(ctx, group, currentUserID, models.SystemEventGroupDissolved, nil, "", "")
		if err := s.groupRepo.Dissolve(ctx, group.ID); err != nil {
			return err
		}
		if err := s.memberRepo.SetAllActiveStatus(ctx, group.ID, models.GroupMemberStatusLeft); err != nil {
			return err
		}
		if err := s.conversationMemberRepo.SetAllActiveStatus(ctx, group.ConversationID, models.ConversationMemberStatusLeft); err != nil {
			return err
		}
		if err := s.groupRepo.SetMemberCount(ctx, group.ID, 0); err != nil {
			return err
		}
		return nil
	}

	if err := s.memberRepo.UpdateStatus(ctx, group.ID, currentUserID, models.GroupMemberStatusLeft); err != nil {
		return err
	}
	if err := s.conversationMemberRepo.SetStatus(ctx, group.ConversationID, currentUserID, models.ConversationMemberStatusLeft); err != nil {
		return err
	}
	if err := s.syncGroupMembers(ctx, group); err != nil {
		return err
	}
	_ = s.emitGroupEvent(ctx, group, currentUserID, models.SystemEventMemberLeft, []string{currentUserID}, "", "")
	return nil
}

func (s *GroupService) AddAdmin(ctx context.Context, groupID primitive.ObjectID, currentUserID string, targetUserID string) (*dto.GroupDetailResponse, error) {
	return s.setAdmin(ctx, groupID, currentUserID, targetUserID, true)
}

func (s *GroupService) RemoveAdmin(ctx context.Context, groupID primitive.ObjectID, currentUserID string, targetUserID string) (*dto.GroupDetailResponse, error) {
	return s.setAdmin(ctx, groupID, currentUserID, targetUserID, false)
}

func (s *GroupService) setAdmin(ctx context.Context, groupID primitive.ObjectID, currentUserID string, targetUserID string, enabled bool) (*dto.GroupDetailResponse, error) {
	group, err := s.getActiveAccessibleGroup(ctx, groupID, currentUserID)
	if err != nil {
		return nil, err
	}
	actor, err := s.memberRepo.GetActive(ctx, group.ID, currentUserID)
	if err != nil || actor.Role != models.GroupMemberRoleOwner {
		return nil, ErrGroupOwnerRequired
	}
	target, err := s.memberRepo.GetActive(ctx, group.ID, targetUserID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	if enabled {
		if target.Role != models.GroupMemberRoleMember {
			return nil, ErrGroupPermissionDenied
		}
		if err := s.memberRepo.SetRole(ctx, group.ID, targetUserID, models.GroupMemberRoleAdmin); err != nil {
			return nil, err
		}
		_ = s.emitGroupEvent(ctx, group, currentUserID, models.SystemEventAdminAdded, []string{targetUserID}, "", "")
	} else {
		if target.Role != models.GroupMemberRoleAdmin {
			return nil, ErrGroupPermissionDenied
		}
		if err := s.memberRepo.SetRole(ctx, group.ID, targetUserID, models.GroupMemberRoleMember); err != nil {
			return nil, err
		}
		_ = s.emitGroupEvent(ctx, group, currentUserID, models.SystemEventAdminRemoved, []string{targetUserID}, "", "")
	}

	return s.GetGroup(ctx, group.ID, currentUserID)
}

func (s *GroupService) getActiveAccessibleGroup(ctx context.Context, groupID primitive.ObjectID, currentUserID string) (*models.Group, error) {
	group, err := s.groupRepo.Get(ctx, groupID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	if !group.IsActive() {
		return nil, ErrGroupDissolved
	}
	if _, err := s.memberRepo.GetActive(ctx, group.ID, currentUserID); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrGroupAccessDenied
		}
		return nil, err
	}
	return group, nil
}

func (s *GroupService) upsertMembers(ctx context.Context, groupID primitive.ObjectID, invitedBy string, ownerID string, userIDs []string) error {
	usersByID := map[string]*models.User{}
	if s.userRepo != nil && len(userIDs) > 0 {
		if users, err := s.userRepo.FindByIDs(ctx, userIDs); err == nil {
			usersByID = users
		}
	}

	for _, userID := range userIDs {
		role := models.GroupMemberRoleMember
		if userID == ownerID {
			role = models.GroupMemberRoleOwner
		} else if user := usersByID[userID]; user != nil && user.Type == models.UserTypeBot {
			role = models.GroupMemberRoleBot
		}
		if _, err := s.memberRepo.UpsertActive(ctx, &models.GroupMember{
			GroupID:   groupID,
			UserID:    userID,
			Role:      role,
			Status:    models.GroupMemberStatusActive,
			InvitedBy: invitedBy,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupService) upsertConversationMembers(ctx context.Context, conversationID primitive.ObjectID, userIDs []string, ownerID string) error {
	if s.conversationMemberRepo == nil {
		return nil
	}
	now := time.Now()
	for _, userID := range userIDs {
		role := string(models.GroupMemberRoleMember)
		if userID == ownerID {
			role = string(models.GroupMemberRoleOwner)
		}
		if _, err := s.conversationMemberRepo.UpsertActive(ctx, conversationID, userID, role, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupService) syncGroupMembers(ctx context.Context, group *models.Group) error {
	participantIDs, err := s.memberRepo.ListActiveUserIDs(ctx, group.ID)
	if err != nil {
		return err
	}
	participantIDs = utils.NormalizeParticipantIDs(participantIDs)
	if err := s.conversationRepo.SetParticipants(ctx, group.ConversationID, participantIDs); err != nil {
		return err
	}
	return s.groupRepo.SetMemberCount(ctx, group.ID, len(participantIDs))
}

func (s *GroupService) emitGroupEvent(
	ctx context.Context,
	group *models.Group,
	operatorID string,
	eventType string,
	targetUserIDs []string,
	beforeValue string,
	afterValue string,
) error {
	if s.messageService == nil || eventType == "" {
		return nil
	}
	content := buildGroupEventContent(eventType, afterValue)
	payload := &models.Payload{
		EventType:     eventType,
		OperatorID:    operatorID,
		TargetUserIDs: targetUserIDs,
		GroupID:       group.ID.Hex(),
		GroupName:     group.Name,
		BeforeValue:   beforeValue,
		AfterValue:    afterValue,
	}
	return s.messageService.CreateSystemEvent(ctx, group.ConversationID, operatorID, content, payload)
}

func (s *GroupService) toMemberDTOs(ctx context.Context, members []*models.GroupMember) []dto.GroupMemberDTO {
	userIDs := make([]string, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}
	usersByID := map[string]*models.User{}
	if s.userRepo != nil && len(userIDs) > 0 {
		if users, err := s.userRepo.FindByIDs(ctx, userIDs); err == nil {
			usersByID = users
		}
	}

	results := make([]dto.GroupMemberDTO, 0, len(members))
	for _, member := range members {
		item := dto.GroupMemberDTO{
			ID:            member.ID,
			GroupID:       member.GroupID,
			UserID:        member.UserID,
			Role:          member.Role,
			Status:        member.Status,
			GroupNickname: member.GroupNickname,
			JoinedAt:      member.JoinedAt,
			InvitedBy:     member.InvitedBy,
		}
		if user := usersByID[member.UserID]; user != nil {
			item.UserInfo = dto.ConvertToUserInfoDto(user)
		}
		results = append(results, item)
	}
	return results
}

func normalizeGroupMemberIDs(ids []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func buildGroupEventContent(eventType string, value string) string {
	switch eventType {
	case models.SystemEventGroupCreated:
		return "创建了群聊"
	case models.SystemEventMemberJoined:
		return "加入了群聊"
	case models.SystemEventMemberKicked:
		return "被移出群聊"
	case models.SystemEventMemberLeft:
		return "退出了群聊"
	case models.SystemEventGroupDissolved:
		return "群聊已解散"
	case models.SystemEventGroupNameUpdated:
		if value != "" {
			return "群名修改为 " + value
		}
		return "修改了群名"
	case models.SystemEventGroupAvatarUpdated:
		return "修改了群头像"
	case models.SystemEventAdminAdded:
		return "设置了管理员"
	case models.SystemEventAdminRemoved:
		return "取消了管理员"
	default:
		return "群聊状态已更新"
	}
}
