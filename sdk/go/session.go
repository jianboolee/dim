package dim

import (
	"context"
	"net/url"
)

type ensureUsersRequest struct {
	Users []UserInput `json:"users"`
}

type Session struct {
	client       *client
	userID       string
	token        string
	refreshToken string
	sessionID    string
	expiresIn    int
	redirectURL  string
}

type RedirectOption func(*redirectOptions)

type redirectOptions struct {
	conversationID string
}

func WithConversationID(conversationID string) RedirectOption {
	return func(options *redirectOptions) {
		options.conversationID = conversationID
	}
}

func (c *Client) EnsureUser(ctx context.Context, user UserInput) error {
	return c.EnsureUsers(ctx, user)
}

func (c *Client) EnsureUsers(ctx context.Context, users ...UserInput) error {
	if len(users) == 0 {
		return nil
	}
	return c.client.do(ctx, "POST", "/im/api/integration/users", ensureUsersRequest{Users: users}, nil)
}

func (c *Client) Login(ctx context.Context, userID string) (*Session, error) {
	var out LoginResponse
	if err := c.client.do(ctx, "POST", "/im/api/integration/login", LoginRequest{UserID: userID}, &out); err != nil {
		return nil, err
	}
	return &Session{
		client:       c.client.withToken(out.Token),
		userID:       userID,
		token:        out.Token,
		refreshToken: out.RefreshToken,
		sessionID:    out.SessionID,
		expiresIn:    out.ExpiresIn,
		redirectURL:  out.RedirectURL,
	}, nil
}

func (s *Session) Token() string {
	return s.token
}

func (s *Session) UserID() string {
	return s.userID
}

func (s *Session) RefreshToken() string {
	return s.refreshToken
}

func (s *Session) SessionID() string {
	return s.sessionID
}

func (s *Session) ExpiresIn() int {
	return s.expiresIn
}

func (s *Session) RedirectURL(options ...RedirectOption) string {
	redirectURL := s.redirectURL
	var opts redirectOptions
	for _, option := range options {
		option(&opts)
	}
	if opts.conversationID == "" || redirectURL == "" {
		return redirectURL
	}
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		return redirectURL
	}
	values := parsed.Query()
	values.Set("conversation_id", opts.conversationID)
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func (s *Session) Conversations() *ConversationService {
	return &ConversationService{client: s.client}
}

func (s *Session) Messages() *MessageService {
	return &MessageService{client: s.client}
}

func (s *Session) Groups() *GroupService {
	return &GroupService{client: s.client}
}

func (s *Session) Users() *UserService {
	return &UserService{client: s.client}
}
