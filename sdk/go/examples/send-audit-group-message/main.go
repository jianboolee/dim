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

var AuditBot = demo.USER_SYSTEM_AUDIT

type output struct {
	Group          *dim.Group  `json:"group"`
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
	if err := imClient.EnsureUsers(ctx, AuditBot, demo.USER_A, demo.USER_B); err != nil {
		log.Fatal(err)
	}
	botSession, err := imClient.Login(ctx, AuditBot.ID)
	if err != nil {
		log.Fatalf("login audit bot: %v", err)
	}

	group, err := botSession.Groups().GetOrCreate(ctx, dim.GetOrCreateGroupParams{
		UniqueKey: "content_audit",
		Name:      groupName,
		MemberIDs: []string{
			demo.USER_A.ID,
			demo.USER_B.ID,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if group.Group == nil {
		log.Fatal("create group response missing group")
	}

	message, err := botSession.Messages().Send(ctx, group.Group.ConversationID, dim.NewMessage(dim.TextMessage(auditMessageContent())))
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
