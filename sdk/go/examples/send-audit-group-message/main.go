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

var AuditBot = dim.USER_SYSTEM_AUDIT

type output struct {
	ConversationID string      `json:"conversation_id"`
	Message        dim.Message `json:"message"`
}

func main() {
	apiBase := envOr("IM_API_BASE", "http://localhost:8901")
	integrationKey := envOr("INTEGRATION_API_KEY", "")
	groupName := envOr("AUDIT_GROUP_NAME", "内容审核群")

	flag.StringVar(&apiBase, "api-base", apiBase, "IM API base URL")
	flag.StringVar(&integrationKey, "key", integrationKey, "X-Integration-Key")
	flag.StringVar(&groupName, "group-name", groupName, "audit group name")
	flag.Parse()

	if integrationKey == "" {
		log.Fatal("INTEGRATION_API_KEY is required (flag -key or env)")
	}

	apiBase = strings.TrimRight(apiBase, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imClient := dim.New(dim.WithBaseURL(apiBase), dim.WithAPIKey(integrationKey))
	conv, err := imClient.Services().GetOrCreateGroupConversation(ctx, AuditBot, dim.GroupTarget{
		UniqueKey: "content_audit",
		Name:      groupName,
		MemberUsers: []dim.UserInput{
			demo.USER_CUSTOMER_SERVICE,
			dim.USER_SYSTEM_AUDIT,
			demo.USER_A,
			demo.USER_B,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	message, err := conv.SendTextMessage(ctx, auditMessageContent())
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

func auditMessageContent() string {
	return fmt.Sprintf(
		"【审核群提醒】\n\n📋 标题：这是一篇测试内容\n👤 发布者：%s\n👥 审核成员：%s、%s\n🔖 状态：待审核\n\n请及时登录后台完成审核：\n%s\n\n发送时间：%s",
		demo.USER_A.Nickname,
		demo.USER_A.Nickname,
		demo.USER_B.Nickname,
		"https://example.com/content/123",
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
