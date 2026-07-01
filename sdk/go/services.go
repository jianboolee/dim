package dim

type Services struct {
	client *Client
}

func (c *Client) Services() *Services {
	return &Services{client: c}
}
