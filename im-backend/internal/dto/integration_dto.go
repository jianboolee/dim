package dto

type IntegrationUserInput struct {
	ID       string `json:"id" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type IntegrationCreateConversationRequest struct {
	FromUser IntegrationUserInput `json:"from_user" binding:"required"`
	ToUser   IntegrationUserInput `json:"to_user" binding:"required"`
}

type IntegrationCreateConversationResponse struct {
	Token          string `json:"token"`
	ConversationID string `json:"conversation_id"`
	RedirectURL    string `json:"redirect_url"`
}
