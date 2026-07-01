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

	imClient := dim.New(dim.WithBaseURL(apiBase), dim.WithAPIKey(integrationKey))
	conv, err := imClient.Services().GetOrCreatePrivateConversation(
		ctx,
		User,
		PeerUser,
		dim.WithInitialPeerMuted(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	// var payload = dim.Payload{
	// 	Title:       "租车订单已提交，等待商家确认",
	// 	Description: "您的租车订单已成功提交，请耐心等待商家确认。商家确认后我们将第一时间通知您。",
	// 	ImageURL:    "https://img01.wanfangche.com/uploads/im/images/20260628/bqks3c09hnp9lms7z264xahx533m7fhr.jpg?x-oss-process=image/resize,w_960,m_lfit",
	// 	URL:         "https://www.example.com/order/detail",
	// 	PriceText:   "待确认",
	// }

	card := dim.CardInput{
		Title:       "[租车]商家已确认",
		Description: "您的租车订单商家已确认。",
		ImageURL:    "https://img01.wanfangche.com/uploads/im/images/20260629/qsj3rve45p0rqcckj0x7byyinywolyw2.jpg?x-oss-process=image/resize,w_960,m_lfit",
		URL:         "https://www.example.com/order/detail",
		PriceText:   "￥899元/天起",
	}

	message, err := conv.SendCardMessage(ctx, card)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		ConversationID: conv.ID(),
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
