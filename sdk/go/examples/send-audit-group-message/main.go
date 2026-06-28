package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	dim "d-im-go-sdk"
	"d-im-go-sdk/examples/demo"
)

var AuditBot = demo.USER_SYSTEM_AUDIT

type createGroupRequest struct {
	Name      string   `json:"name"`
	AvatarURL string   `json:"avatar_url,omitempty"`
	MemberIDs []string `json:"member_ids,omitempty"`
}

type groupDetailResponse struct {
	Group groupDTO `json:"group"`
}

type groupDTO struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	Name           string `json:"name"`
	MemberCount    int    `json:"member_count"`
	Status         string `json:"status"`
}

type apiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type output struct {
	Group          groupDTO    `json:"group"`
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

	integrationClient := dim.NewIntegrationClient(dim.Config{
		BaseURL: apiBase,
		APIKey:  integrationKey,
	})

	// integration login 会 upsert 用户；这里先确保审核机器人、USER_A、USER_B 都存在。
	botSession, err := integrationClient.Login(ctx, dim.LoginRequest{User: AuditBot})
	if err != nil {
		log.Fatalf("login audit bot: %v", err)
	}
	if _, err := integrationClient.Login(ctx, dim.LoginRequest{User: demo.USER_A}); err != nil {
		log.Fatalf("upsert user A: %v", err)
	}
	if _, err := integrationClient.Login(ctx, dim.LoginRequest{User: demo.USER_B}); err != nil {
		log.Fatalf("upsert user B: %v", err)
	}

	group, err := createAuditGroup(ctx, apiBase, botSession.Token, createGroupRequest{
		Name: groupName,
		MemberIDs: []string{
			demo.USER_A.ID,
			demo.USER_B.ID,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	userClient := dim.NewUserClient(dim.Config{
		BaseURL: apiBase,
		Token:   botSession.Token,
	})

	message, err := userClient.SendTextMessage(ctx, group.Group.ConversationID, auditMessageContent())
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

func createAuditGroup(ctx context.Context, apiBase, token string, reqBody createGroupRequest) (*groupDetailResponse, error) {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/im/api/groups", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("create group status %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}

	var wrapped apiResponse[groupDetailResponse]
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("decode create group response: %w", err)
	}
	if wrapped.Code != 0 {
		return nil, fmt.Errorf("create group code %d: %s", wrapped.Code, wrapped.Message)
	}
	return &wrapped.Data, nil
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
