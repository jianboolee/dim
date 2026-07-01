package dim

import "context"

func (s *Services) GetOrCreatePrivateConversation(
	ctx context.Context,
	user UserInput,
	peerUser UserInput,
	options ...PrivateConversationServiceOption,
) (*ConversationSession, error) {
	opts := defaultPrivateConversationServiceOptions()
	for _, option := range options {
		option.applyPrivateConversationServiceOption(&opts)
	}

	if opts.ensureUsers {
		if err := s.client.EnsureUsers(ctx, user, peerUser); err != nil {
			return nil, err
		}
	}

	session, err := s.client.Login(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	conversation, err := session.Conversations().GetOrCreatePrivate(ctx, peerUser.ID, buildPrivateConversationOptions(user, peerUser, opts)...)
	if err != nil {
		return nil, err
	}

	return &ConversationSession{
		session:      session,
		Conversation: conversation,
	}, nil
}

func buildPrivateConversationOptions(
	user UserInput,
	peerUser UserInput,
	options privateConversationServiceOptions,
) []CreatePrivateConversationOption {
	conversationOptions := make([]CreatePrivateConversationOption, 0, len(options.initialMemberSettings)+2)
	for userID, settings := range options.initialMemberSettings {
		if settings.Muted != nil {
			conversationOptions = append(conversationOptions, WithInitialMemberMuted(userID, *settings.Muted))
		}
	}
	if options.initialSenderMuted != nil {
		conversationOptions = append(conversationOptions, WithInitialMemberMuted(user.ID, *options.initialSenderMuted))
	}
	if options.initialPeerMuted != nil {
		conversationOptions = append(conversationOptions, WithInitialMemberMuted(peerUser.ID, *options.initialPeerMuted))
	}
	return conversationOptions
}
