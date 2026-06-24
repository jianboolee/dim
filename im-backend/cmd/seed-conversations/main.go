// 批量创建与 user_b 的单聊会话，并为每个买家发送一条首消息。
//
// 用法（在 im-backend 目录）:
//
//	go run ./cmd/seed-conversations/
//	go run ./cmd/seed-conversations/ -count 10 -dry-run
//
// 环境变量: IM_API_BASE, INTEGRATION_API_KEY, TO_USER_ID, TO_USER_NICKNAME
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const ossAvatarSuffix = "?x-oss-process=image/resize,m_fill,w_200,h_200"

var (
	firstNames = []string{
		"James", "Emma", "Liam", "Olivia", "Noah",
		"Ava", "Oliver", "Sophia", "Elijah", "Isabella",
	}
	lastNames = []string{
		"Wilson", "Thompson", "Moore", "Taylor", "Anderson",
		"Thomas", "Jackson", "White", "Harris", "Martin",
	}
)

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type createConversationData struct {
	Token          string `json:"token"`
	ConversationID string `json:"conversation_id"`
	RedirectURL    string `json:"redirect_url"`
}

type integrationUser struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type createConversationRequest struct {
	FromUser integrationUser `json:"from_user"`
	ToUser   integrationUser `json:"to_user"`
}

type sendMessageRequest struct {
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
	Type       string `json:"type"`
}

type config struct {
	apiBase        string
	integrationKey string
	toUserID       string
	toNickname     string
	toAvatar       string
	count          int
	startIndex     int
	userIDPrefix   string
	dryRun         bool
}

func main() {
	cfg := parseConfig()

	if cfg.integrationKey == "" {
		log.Fatal("INTEGRATION_API_KEY is required (flag -key or env)")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	var okCount, failCount int
	for i := 0; i < cfg.count; i++ {
		seq := cfg.startIndex + i
		fromID := fmt.Sprintf("%s%03d", cfg.userIDPrefix, seq)
		fromName := englishName(seq - 1)
		fromAvatar := mockAvatarURL(seq)

		log.Printf("[%d/%d] %s (%s) -> %s", i+1, cfg.count, fromID, fromName, cfg.toUserID)

		if cfg.dryRun {
			log.Printf("  dry-run: create conversation + send message")
			okCount++
			continue
		}

		token, err := createConversation(client, cfg, fromID, fromName, fromAvatar)
		if err != nil {
			log.Printf("  create failed: %v", err)
			failCount++
			continue
		}

		content := fmt.Sprintf("Hi %s, this is %s. Nice to meet you!", cfg.toNickname, fromName)
		if err := sendMessage(client, cfg.apiBase, token, cfg.toUserID, content); err != nil {
			log.Printf("  message failed: %v", err)
			failCount++
			continue
		}

		okCount++
	}

	log.Printf("done: success=%d failed=%d", okCount, failCount)
	if failCount > 0 {
		os.Exit(1)
	}
}

func parseConfig() config {
	apiBase := envOr("IM_API_BASE", "http://localhost:8901")
	integrationKey := envOr("INTEGRATION_API_KEY", "")
	toUserID := envOr("TO_USER_ID", "user_b")
	toNickname := envOr("TO_USER_NICKNAME", "Brock")
	toAvatar := envOr("TO_USER_AVATAR", "https://oss.21rv.com/uploads/images/20260624/mznepycjzlgsonoph9fi4lw6d82mezb1.jpg"+ossAvatarSuffix)

	var (
		flagAPIBase  string
		flagKey      string
		flagToUser   string
		flagToNick   string
		flagToAvatar string
		flagCount    int
		flagStart    int
		flagPrefix   string
		flagDryRun   bool
	)

	flag.StringVar(&flagAPIBase, "api-base", apiBase, "IM API base URL")
	flag.StringVar(&flagKey, "key", integrationKey, "X-Integration-Key")
	flag.StringVar(&flagToUser, "to-user", toUserID, "会话对方用户 ID")
	flag.StringVar(&flagToNick, "to-nickname", toNickname, "会话对方昵称（用于消息文案）")
	flag.StringVar(&flagToAvatar, "to-avatar", toAvatar, "会话对方头像 URL")
	flag.IntVar(&flagCount, "count", 100, "创建会话数量")
	flag.IntVar(&flagStart, "start", 1, "起始序号（影响用户 ID 与 mock 头像编号）")
	flag.StringVar(&flagPrefix, "prefix", "seed_user_", "from_user ID 前缀")
	flag.BoolVar(&flagDryRun, "dry-run", false, "只打印计划，不请求 API")
	flag.Parse()

	return config{
		apiBase:        strings.TrimRight(flagAPIBase, "/"),
		integrationKey: flagKey,
		toUserID:       flagToUser,
		toNickname:     flagToNick,
		toAvatar:       flagToAvatar,
		count:          flagCount,
		startIndex:     flagStart,
		userIDPrefix:   flagPrefix,
		dryRun:         flagDryRun,
	}
}

func englishName(index int) string {
	if index < 0 {
		index = 0
	}
	first := firstNames[index%len(firstNames)]
	last := lastNames[(index/len(firstNames))%len(lastNames)]
	return fmt.Sprintf("%s %s", first, last)
}

func mockAvatarURL(index int) string {
	if index <= 0 {
		index = 1
	}
	if index > 100 {
		index = ((index - 1) % 100) + 1
	}
	return fmt.Sprintf("https://oss.21rv.com/mock/images/%d.jpg%s", index, ossAvatarSuffix)
}

func createConversation(client *http.Client, cfg config, fromID, fromName, fromAvatar string) (string, error) {
	body := createConversationRequest{
		FromUser: integrationUser{
			ID:       fromID,
			Nickname: fromName,
			Avatar:   fromAvatar,
		},
		ToUser: integrationUser{
			ID:       cfg.toUserID,
			Nickname: cfg.toNickname,
			Avatar:   cfg.toAvatar,
		},
	}

	var resp apiResponse
	if err := postJSON(client, cfg.apiBase+"/im/api/integration/conversations", cfg.integrationKey, "", body, &resp); err != nil {
		return "", err
	}
	if resp.Code != 200 {
		return "", fmt.Errorf("api code=%d message=%s", resp.Code, resp.Message)
	}

	var data createConversationData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("decode conversation data: %w", err)
	}
	if data.Token == "" {
		return "", fmt.Errorf("empty token in response")
	}

	return data.Token, nil
}

func sendMessage(client *http.Client, apiBase, token, receiverID, content string) error {
	body := sendMessageRequest{
		ReceiverID: receiverID,
		Content:    content,
		Type:       "text",
	}

	req, err := http.NewRequest(http.MethodPost, apiBase+"/im/api/messages", mustJSON(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

func postJSON(client *http.Client, url, integrationKey, bearer string, payload any, out *apiResponse) error {
	req, err := http.NewRequest(http.MethodPost, url, mustJSON(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if integrationKey != "" {
		req.Header.Set("X-Integration-Key", integrationKey)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func mustJSON(v any) io.Reader {
	buf, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("json marshal: %v", err)
	}
	return bytes.NewReader(buf)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
