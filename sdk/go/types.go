package dim

import "time"

type User struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Type      string `json:"type,omitempty"` // normal / system / bot
}

type ConversationUserState struct {
	LastActivatedAt time.Time `json:"last_activated_at,omitempty"`
	LastReadAt      time.Time `json:"last_read_at,omitempty"`
	UnreadCount     int64     `json:"unread_count"`
}

type Conversation struct {
	ID           string                           `json:"id"`
	Type         string                           `json:"type"`
	Participants []string                         `json:"participants"`
	LastMessage  *LastMessageSnapshot             `json:"last_message,omitempty"`
	ToUserInfo   *User                            `json:"to_user_info,omitempty"`
	ImageURL     string                           `json:"image_url,omitempty"`
	UserStates   map[string]ConversationUserState `json:"user_states,omitempty"`
	LastActivity time.Time                        `json:"last_activity"`
	CreatedAt    time.Time                        `json:"created_at"`
	UpdatedAt    time.Time                        `json:"updated_at"`
}

type Message struct {
	ID              string    `json:"id,omitempty"`
	ClientMessageID string    `json:"client_message_id,omitempty"`
	ConversationID  string    `json:"conversation_id,omitempty"`
	SenderID        string    `json:"sender_id,omitempty"`
	ReceiverID      string    `json:"receiver_id,omitempty"`
	Type            string    `json:"type"`
	Content         string    `json:"content"`
	Status          string    `json:"status,omitempty"`
	Payload         *Payload  `json:"payload,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type LastMessageSnapshot struct {
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type Payload struct {
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	URL         string            `json:"url,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	Price       string            `json:"price,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type ConversationPage struct {
	Items      []Conversation `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type CreateConversationRequest struct {
	FromUser User `json:"from_user"`
	ToUser   User `json:"to_user"`
}

type CreateConversationResponse struct {
	Token          string `json:"token"`
	ExpiresIn      int    `json:"expires_in"`
	ConversationID string `json:"conversation_id"`
	RedirectURL    string `json:"redirect_url"`
}

type LoginRequest struct {
	User User `json:"user"`
}

type LoginResponse struct {
	Token       string `json:"token"`
	ExpiresIn   int    `json:"expires_in"`
	RedirectURL string `json:"redirect_url"`
}

type SendMessageRequest struct {
	ClientMessageID string   `json:"client_message_id,omitempty"`
	Type            string   `json:"type"`
	Content         string   `json:"content,omitempty"`
	Payload         *Payload `json:"payload,omitempty"`
}
