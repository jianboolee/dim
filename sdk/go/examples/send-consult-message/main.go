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
var PeerUser = demo.USER_CUSTOMER_SERVICE
var card = demo.CARD_INPUT_A

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

	result, err := imClient.Services().SendCardMessage(ctx, User, PeerUser, card)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		ConversationID: result.Conversation.ID,
		Message:        *result.Message,
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
