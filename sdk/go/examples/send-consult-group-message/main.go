package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	dim "d-im-go-sdk"
	"d-im-go-sdk/examples/demo"
)

type output struct {
	ConversationID string      `json:"conversation_id"`
	Message        dim.Message `json:"message"`
}

var Bot = demo.USER_CUSTOMER_SERVICE
var User = demo.USER_H
var PeerUser = demo.USER_SELLER_A

var Card = demo.CARD_INPUT_A

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

	UniqueKey := PeerUser.ID + "_" + User.ID

	groupName := fmt.Sprintf("[咨询] %s", Card.Title)

	createdConv, err := imClient.Services().GetOrCreateGroupConversation(ctx, Bot, dim.GroupTarget{
		UniqueKey: UniqueKey,
		Name:      groupName,
		MemberUsers: []dim.UserInput{
			PeerUser,
			User,
			demo.USER_A,
			demo.USER_B,
			demo.USER_C,
			demo.USER_D,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	senderConv, err := imClient.Services().GetConversation(ctx, User, createdConv.ID())
	if err != nil {
		log.Fatal(err)
	}
	message, err := senderConv.SendCardMessage(ctx, Card)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		ConversationID: senderConv.ID(),
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
