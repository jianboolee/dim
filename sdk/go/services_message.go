package dim

import "context"

func (s *Services) SendMessage(
	ctx context.Context,
	user UserInput,
	peerUser UserInput,
	message MessageInput,
	options ...SendMessageOption,
) (*SendMessageResult, error) {
	opts := defaultSendMessageOptions()
	for _, option := range options {
		option.applySendMessageOption(&opts)
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

	conversation, err := session.Conversations().GetOrCreatePrivate(ctx, peerUser.ID, buildCreatePrivateConversationOptions(user, peerUser, opts)...)
	if err != nil {
		return nil, err
	}

	sentMessage, err := session.Messages().Send(ctx, conversation.ID, message)
	if err != nil {
		return nil, err
	}

	return &SendMessageResult{
		Conversation: conversation,
		Message:      sentMessage,
	}, nil
}

func (s *Services) SendTextMessage(
	ctx context.Context,
	user UserInput,
	peerUser UserInput,
	content string,
	options ...SendMessageOption,
) (*SendMessageResult, error) {
	return s.SendMessage(ctx, user, peerUser, NewMessage(TextMessage(content)), options...)
}

func (s *Services) SendCardMessage(
	ctx context.Context,
	user UserInput,
	peerUser UserInput,
	card CardInput,
	options ...SendMessageOption,
) (*SendMessageResult, error) {
	return s.SendMessage(ctx, user, peerUser, NewMessage(CardMessage(card)), options...)
}

func buildCreatePrivateConversationOptions(
	user UserInput,
	peerUser UserInput,
	options sendMessageOptions,
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
