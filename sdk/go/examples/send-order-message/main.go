package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	dim "d-im-go-sdk"
	"d-im-go-sdk/examples/demo"
)

var User = demo.USER_SYSTEM_NOTICE
var PeerUser = demo.USER_B

type output struct {
	ConversationID string      `json:"conversation_id"`
	Message        dim.Message `json:"message"`
}

func main() {
	apiBase := envOr("IM_API_BASE", "http://localhost:8901")
	integrationKey := envOr("INTEGRATION_API_KEY", "")

	flag.StringVar(&apiBase, "api-base", apiBase, "IM API base URL")
	flag.StringVar(&integrationKey, "key", integrationKey, "X-Integration-Key")
	flag.Parse()

	if integrationKey == "" {
		log.Fatal("INTEGRATION_API_KEY is required (flag -key or env)")
	}

	apiBase = strings.TrimRight(apiBase, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	integrationClient := dim.NewIntegrationClient(dim.Config{
		BaseURL: apiBase,
		APIKey:  integrationKey,
	})

	session, err := integrationClient.CreateConversation(ctx, dim.IntegrationPrivateConversationRequest{
		User:     User,
		PeerUser: PeerUser,
	})
	if err != nil {
		log.Fatal(err)
	}

	userClient := dim.NewUserClient(dim.Config{
		BaseURL: apiBase,
		Token:   session.Token,
	})

	// var payload = dim.Payload{
	// 	Title:       "租车订单已提交，等待商家确认",
	// 	Description: "您的租车订单已成功提交，请耐心等待商家确认。商家确认后我们将第一时间通知您。",
	// 	ImageURL:    "https://img02.wanfangche.com/uploads/im/images/20260628/bqks3c09hnp9lms7z264xahx533m7fhr.jpg?x-oss-process=image/resize,w_960,m_lfit",
	// 	URL:         "https://www.example.com/order/detail",
	// 	Price:       "待确认",
	// }

	var payload = dim.Payload{
		Title:       "[租车]商家已确认",
		Description: "您的租车订单商家已确认。",
		ImageURL:    "https://img02.wanfangche.com/uploads/im/images/20260629/qsj3rve45p0rqcckj0x7byyinywolyw2.jpg?x-oss-process=image/resize,w_960,m_lfit",
		URL:         "https://www.example.com/order/detail",
		Price:       "￥899元/天起",
	}

	message, err := userClient.SendCardMessage(ctx, session.ConversationID, payload)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		ConversationID: session.ConversationID,
		Message:        *message,
	}); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
