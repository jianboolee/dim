package dim

import "context"

type ConversationSession struct {
	session      *Session
	Conversation *Conversation
}

func (c *ConversationSession) ID() string {
	if c == nil || c.Conversation == nil {
		return ""
	}
	return c.Conversation.ID
}

func (c *ConversationSession) SendMessage(ctx context.Context, message MessageInput) (*Message, error) {
	return c.session.Messages().Send(ctx, c.ID(), message)
}

func (c *ConversationSession) SendTextMessage(ctx context.Context, content string) (*Message, error) {
	return c.SendMessage(ctx, NewMessage(TextMessage(content)))
}

func (c *ConversationSession) SendCardMessage(ctx context.Context, card CardInput) (*Message, error) {
	return c.SendMessage(ctx, NewMessage(CardMessage(card)))
}

func (s *Services) GetConversation(ctx context.Context, user UserInput, conversationID string) (*ConversationSession, error) {
	session, err := s.client.Login(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	conversation, err := session.Conversations().Get(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return &ConversationSession{
		session:      session,
		Conversation: conversation,
	}, nil
}
