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
	Group          *dim.Group  `json:"group"`
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

	// 1. 确保客服、卖家、买家用户的资料存在
	if err := imClient.EnsureUsers(ctx, Bot, PeerUser, User, demo.USER_A, demo.USER_B, demo.USER_C, demo.USER_D); err != nil {
		log.Fatal(err)
	}

	// 2. 以客服身份登录并创建/获取咨询群
	csSession, err := imClient.Login(ctx, Bot.ID)
	if err != nil {
		log.Fatalf("login customer service: %v", err)
	}

	UniqueKey := PeerUser.ID + "_" + User.ID

	groupName := fmt.Sprintf("[咨询] %s", Card.Title)

	group, err := csSession.Groups().GetOrCreate(ctx, dim.GetOrCreateGroupParams{
		UniqueKey: UniqueKey,
		Name:      groupName,
		MemberIDs: []string{
			PeerUser.ID,
			User.ID,
			demo.USER_A.ID,
			demo.USER_B.ID,
			demo.USER_C.ID,
			demo.USER_D.ID,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if group.Group == nil {
		log.Fatal("create group response missing group")
	}

	// 3. 以买家 USER_A 身份登录并发送卡片消息到群内
	userASession, err := imClient.Login(ctx, User.ID)
	if err != nil {
		log.Fatalf("login %s: %v", User.ID, err)
	}

	message, err := userASession.Messages().Send(ctx, group.Group.ConversationID, dim.NewMessage(dim.CardMessage(Card)))
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		Group:          group.Group,
		ConversationID: group.Group.ConversationID,
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
