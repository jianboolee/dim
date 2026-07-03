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

	dim "github.com/jianboolee/dim/sdk/go"
	"github.com/jianboolee/dim/sdk/go/examples/demo"
)

var User = dim.USER_SYSTEM_AUDIT
var PeerUser = demo.USER_AUDITOR

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

	title := "这是一篇测试内容"
	publisher := "Alice"
	status := "待审核"
	url := "https://example.com/content/123"

	content := fmt.Sprintf(
		"【内容审核提醒】\n\n📋 标题：%s\n👤 发布者：%s\n🔖 状态：%s\n\n请及时登录后台完成审核：\n%s\n\n请在24小时内完成审核。",
		title, publisher, status, url,
	)
	message, err := conv.SendTextMessage(ctx, content)
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
