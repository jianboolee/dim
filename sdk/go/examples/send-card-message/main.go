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

var User = demo.USER_A
var PeerUser = demo.USER_JIANBO

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

	var payload = dim.Payload{
		Title:       "[租车]暑期特惠，租车低至899元/天起",
		Description: "暑期出行，租车低至899元/天起，支持全国取还车，随时随地畅享自驾乐趣。",
		ImageURL:    "https://oss.21rv.com/uploads/im/images/20260628/bqks3c09hnp9lms7z264xahx533m7fhr.jpg?x-oss-process=image/resize,w_960,m_lfit",
		URL:         "https://code.visualstudio.com/docs/agents/overview",
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
