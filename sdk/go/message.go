package dim

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type MessageService struct {
	client *client
}

func TextMessage(content string) MessageBody {
	return MessageBody{Type: "text", Content: content}
}

func CardMessage(input CardInput) MessageBody {
	return MessageBody{
		Type:    "card",
		Content: input.Title,
		Payload: &Payload{
			Title:       input.Title,
			Description: input.Description,
			URL:         input.URL,
			ImageURL:    input.ImageURL,
			PriceText:   input.PriceText,
		},
	}
}

func ImageMessage(url string) MessageBody {
	return MessageBody{Type: "image", Content: "[图片]", Payload: &Payload{URL: url}}
}

func VideoMessage(url string) MessageBody {
	return MessageBody{Type: "video", Content: "[视频]", Payload: &Payload{URL: url}}
}

func AudioMessage(url string) MessageBody {
	return MessageBody{Type: "audio", Content: "[语音]", Payload: &Payload{URL: url}}
}

func LinkMessage(input CardInput) MessageBody {
	return MessageBody{
		Type:    "link",
		Content: input.Title,
		Payload: &Payload{
			Title:       input.Title,
			Description: input.Description,
			URL:         input.URL,
			ImageURL:    input.ImageURL,
		},
	}
}

func NewMessage(body MessageBody) MessageInput {
	return MessageInput{Body: body}
}

func (s *MessageService) Send(ctx context.Context, conversationID string, msg MessageInput) (*Message, error) {
	req := SendMessageRequest{
		ClientMessageID: msg.ClientMessageID,
		Type:            msg.Body.Type,
		Content:         msg.Body.Content,
		Payload:         msg.Body.Payload,
	}
	var out Message
	if err := s.client.do(ctx, http.MethodPost, "/im/api/conversations/"+url.PathEscape(conversationID)+"/messages", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MessageService) List(ctx context.Context, conversationID string, params ListMessagesParams) ([]Message, error) {
	values := url.Values{}
	if params.Limit > 0 {
		values.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.BeforeID != "" {
		values.Set("before_id", params.BeforeID)
	}
	if params.AfterID != "" {
		values.Set("after_id", params.AfterID)
	}

	path := "/im/api/conversations/" + url.PathEscape(conversationID) + "/messages"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []Message
	if err := s.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
