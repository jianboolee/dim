package dim

import "time"

type User struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Type      string `json:"type,omitempty"` // normal / system / bot
}

type Conversation struct {
	ID            string                   `json:"id"`
	Type          string                   `json:"type"`
	Participants  []string                 `json:"participants"`
	LastMessage   *LastMessageSnapshot     `json:"last_message,omitempty"`
	DisplayName   string                   `json:"display_name,omitempty"`
	DisplayAvatar string                   `json:"display_avatar,omitempty"`
	GroupID       string                   `json:"group_id,omitempty"`
	GroupInfo     *GroupSummary            `json:"group_info,omitempty"`
	PeerUserInfo  *User                    `json:"peer_user_info,omitempty"`
	ImageURL      string                   `json:"image_url,omitempty"`
	MemberState   *ConversationMemberState `json:"member_state,omitempty"`
	LastActivity  time.Time                `json:"last_activity"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

type ConversationMemberState struct {
	Status          string    `json:"status"`
	LastReadSeq     int64     `json:"last_read_seq"`
	LastReadAt      time.Time `json:"last_read_at,omitempty"`
	LastActivatedAt time.Time `json:"last_activated_at,omitempty"`
	UnreadCount     int64     `json:"unread_count"`
	MentionCount    int64     `json:"mention_count,omitempty"`
	Muted           bool      `json:"muted,omitempty"`
	Pinned          bool      `json:"pinned,omitempty"`
}

type Message struct {
	ID              string    `json:"id,omitempty"`
	ClientMessageID string    `json:"client_message_id,omitempty"`
	ConversationID  string    `json:"conversation_id,omitempty"`
	Seq             int64     `json:"seq,omitempty"`
	SenderID        string    `json:"sender_id,omitempty"`
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
	Title         string            `json:"title,omitempty"`
	Description   string            `json:"description,omitempty"`
	URL           string            `json:"url,omitempty"`
	ImageURL      string            `json:"image_url,omitempty"`
	PriceText     string            `json:"price_text,omitempty"`
	Meta          map[string]string `json:"meta,omitempty"`
	EventType     string            `json:"event_type,omitempty"`
	OperatorID    string            `json:"operator_id,omitempty"`
	TargetUserIDs []string          `json:"target_user_ids,omitempty"`
	GroupID       string            `json:"group_id,omitempty"`
	GroupName     string            `json:"group_name,omitempty"`
	BeforeValue   string            `json:"before_value,omitempty"`
	AfterValue    string            `json:"after_value,omitempty"`
}

type ConversationPage struct {
	Items      []Conversation `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type CreatePrivateConversationRequest struct {
	PeerID string `json:"peer_id"`
}

type IntegrationPrivateConversationRequest struct {
	User     User   `json:"user"`
	PeerUser User   `json:"peer_user"`
	Device   Device `json:"device,omitempty"`
}

type CreateConversationResponse struct {
	Token          string `json:"token"`
	ExpiresIn      int    `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
	SessionID      string `json:"session_id"`
	ConversationID string `json:"conversation_id"`
	RedirectURL    string `json:"redirect_url"`
}

type LoginRequest struct {
	User   User   `json:"user"`
	Device Device `json:"device,omitempty"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
	RedirectURL  string `json:"redirect_url"`
}

type Device struct {
	Platform   string `json:"platform,omitempty"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
	PushToken  string `json:"push_token,omitempty"`
}

type SendMessageRequest struct {
	ClientMessageID string   `json:"client_message_id,omitempty"`
	Type            string   `json:"type"`
	Content         string   `json:"content,omitempty"`
	Payload         *Payload `json:"payload,omitempty"`
}

type CreateGroupRequest struct {
	Name      string   `json:"name"`
	AvatarURL string   `json:"avatar_url,omitempty"`
	MemberIDs []string `json:"member_ids,omitempty"`
}

type GroupDetailResponse struct {
	Group   *Group        `json:"group"`
	Members []GroupMember `json:"members,omitempty"`
}

type Group struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Name           string    `json:"name"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	OwnerID        string    `json:"owner_id"`
	MemberCount    int       `json:"member_count"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type GroupMember struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	UserID        string    `json:"user_id"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	GroupNickname string    `json:"group_nickname,omitempty"`
	JoinedAt      time.Time `json:"joined_at"`
	InvitedBy     string    `json:"invited_by,omitempty"`
	UserInfo      *User     `json:"user_info,omitempty"`
}

type GroupMemberBrief struct {
	UserID        string `json:"user_id"`
	Role          string `json:"role"`
	GroupNickname string `json:"group_nickname,omitempty"`
	UserInfo      *User  `json:"user_info,omitempty"`
}

type GroupSummary struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	AvatarURL   string             `json:"avatar_url,omitempty"`
	MemberCount int                `json:"member_count"`
	Members     []GroupMemberBrief `json:"members,omitempty"`
}
