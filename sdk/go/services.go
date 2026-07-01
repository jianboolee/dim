package dim

type Services struct {
	client *Client
}

type SendMessageResult struct {
	Conversation *Conversation
	Message      *Message
}

func (c *Client) Services() *Services {
	return &Services{client: c}
}
