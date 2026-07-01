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

var User = demo.USER_SYSTEM_NOTICE
var PeerUser = demo.USER_A

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
	if err := imClient.EnsureUsers(ctx, User, PeerUser); err != nil {
		log.Fatal(err)
	}
	session, err := imClient.Login(ctx, User.ID)
	if err != nil {
		log.Fatal(err)
	}
	conversation, err := session.Conversations().GetOrCreatePrivate(ctx, PeerUser.ID)
	if err != nil {
		log.Fatal(err)
	}

	content := fmt.Sprintf("你好，欢迎使用消息系统，当前时间是: %s", time.Now().Format("2006-01-02 15:04:05"))
	message, err := session.Messages().Send(ctx, conversation.ID, dim.NewMessage(dim.TextMessage(content)))
	if err != nil {
		log.Fatal(err)
	}

	// PeerUser 默认消息免打扰
	muted := true
	peerSession, err := imClient.Login(ctx, PeerUser.ID)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := peerSession.Conversations().UpdateSettings(ctx, conversation.ID, dim.ConversationSettingsPatch{Muted: &muted}); err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		ConversationID: conversation.ID,
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
