package dim

import (
	"context"
	"errors"
)

var ErrGroupConversationIDMissing = errors.New("group conversation id is missing")

type GroupTarget struct {
	Name        string
	UniqueKey   string
	ScopeUserID string
	MemberUsers []UserInput
	MemberIDs   []string
}

func (s *Services) GetOrCreateGroupConversation(
	ctx context.Context,
	owner UserInput,
	target GroupTarget,
	options ...GroupConversationServiceOption,
) (*ConversationSession, error) {
	opts := defaultGroupConversationServiceOptions()
	for _, option := range options {
		option.applyGroupConversationServiceOption(&opts)
	}

	if opts.ensureUsers {
		if err := s.client.EnsureUsers(ctx, ensureGroupUsers(owner, target.MemberUsers)...); err != nil {
			return nil, err
		}
	}

	session, err := s.client.Login(ctx, owner.ID)
	if err != nil {
		return nil, err
	}

	req := CreateGroupRequest{
		Name:        target.Name,
		MemberIDs:   groupMemberIDs(target.MemberUsers, target.MemberIDs),
		UniqueKey:   target.UniqueKey,
		ScopeUserID: target.ScopeUserID,
	}

	var groupConversation *GroupConversationResponse
	if target.UniqueKey != "" {
		groupConversation, err = session.Groups().GetOrCreateConversation(ctx, req)
	} else {
		groupConversation, err = session.Groups().CreateConversation(ctx, req)
	}
	if err != nil {
		return nil, err
	}
	if groupConversation == nil || groupConversation.Group == nil || groupConversation.Group.ConversationID == "" {
		return nil, ErrGroupConversationIDMissing
	}

	return &ConversationSession{
		session: session,
		Conversation: &Conversation{
			ID: groupConversation.Group.ConversationID,
		},
	}, nil
}

func (s *Services) EnsureGroupMembers(
	ctx context.Context,
	operator UserInput,
	groupID string,
	members []UserInput,
	memberIDs []string,
	options ...GroupMemberServiceOption,
) error {
	opts := defaultGroupMemberServiceOptions()
	for _, option := range options {
		option.applyGroupMemberServiceOption(&opts)
	}

	if opts.ensureUsers {
		if err := s.client.EnsureUsers(ctx, ensureGroupUsers(operator, members)...); err != nil {
			return err
		}
	}

	session, err := s.client.Login(ctx, operator.ID)
	if err != nil {
		return err
	}

	ids := groupMemberIDs(members, memberIDs)
	if len(ids) == 0 {
		return nil
	}
	_, err = session.Groups().Invite(ctx, groupID, ids)
	return err
}

func ensureGroupUsers(owner UserInput, members []UserInput) []UserInput {
	users := make([]UserInput, 0, len(members)+1)
	if owner.ID != "" {
		users = append(users, owner)
	}
	users = append(users, members...)
	return users
}

func groupMemberIDs(users []UserInput, ids []string) []string {
	seen := map[string]struct{}{}
	results := make([]string, 0, len(users)+len(ids))
	for _, user := range users {
		if user.ID == "" {
			continue
		}
		if _, ok := seen[user.ID]; ok {
			continue
		}
		seen[user.ID] = struct{}{}
		results = append(results, user.ID)
	}
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
	return results
}
