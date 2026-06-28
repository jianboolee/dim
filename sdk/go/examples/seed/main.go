package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	dim "d-im-go-sdk"
	"d-im-go-sdk/examples/demo"
)

var PeerUser = demo.USER_B

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

type config struct {
	apiBase        string
	integrationKey string
	peerUserID     string
	peerNickname   string
	peerAvatar     string
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

	integrationClient := dim.NewIntegrationClient(dim.Config{
		BaseURL: cfg.apiBase,
		APIKey:  cfg.integrationKey,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var okCount, failCount int
	for i := 0; i < cfg.count; i++ {
		seq := cfg.startIndex + i
		fromID := fmt.Sprintf("%s%03d", cfg.userIDPrefix, seq)
		fromName := englishName(seq - 1)
		fromAvatar := mockAvatarURL(seq)

		log.Printf("[%d/%d] %s (%s) -> %s", i+1, cfg.count, fromID, fromName, cfg.peerUserID)

		if cfg.dryRun {
			log.Printf("  dry-run: create conversation + send message")
			okCount++
			continue
		}

		session, err := integrationClient.CreateConversation(ctx, dim.IntegrationPrivateConversationRequest{
			User: dim.User{
				ID:       fromID,
				Nickname: fromName,
				Avatar:   fromAvatar,
			},
			PeerUser: dim.User{
				ID:       cfg.peerUserID,
				Nickname: cfg.peerNickname,
				Avatar:   cfg.peerAvatar,
			},
		})
		if err != nil {
			log.Printf("  create failed: %v", err)
			failCount++
			continue
		}

		userClient := dim.NewUserClient(dim.Config{
			BaseURL: cfg.apiBase,
			Token:   session.Token,
		})
		content := fmt.Sprintf("Hi %s, this is %s. Nice to meet you!", cfg.peerNickname, fromName)
		if _, err := userClient.SendTextMessage(ctx, session.ConversationID, content); err != nil {
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
	peerUserID := envOr("PEER_USER_ID", PeerUser.ID)
	peerNickname := envOr("PEER_USER_NICKNAME", PeerUser.Nickname)
	peerAvatar := envOr("PEER_USER_AVATAR", PeerUser.Avatar)

	var (
		flagAPIBase    string
		flagKey        string
		flagPeerUser   string
		flagPeerNick   string
		flagPeerAvatar string
		flagCount      int
		flagStart      int
		flagPrefix     string
		flagDryRun     bool
	)

	flag.StringVar(&flagAPIBase, "api-base", apiBase, "IM API base URL")
	flag.StringVar(&flagKey, "key", integrationKey, "X-Integration-Key")
	flag.StringVar(&flagPeerUser, "peer-user", peerUserID, "会话对方用户 ID")
	flag.StringVar(&flagPeerNick, "peer-nickname", peerNickname, "会话对方昵称")
	flag.StringVar(&flagPeerAvatar, "peer-avatar", peerAvatar, "会话对方头像 URL")
	flag.IntVar(&flagCount, "count", 100, "创建会话数量")
	flag.IntVar(&flagStart, "start", 1, "起始序号")
	flag.StringVar(&flagPrefix, "prefix", "seed_user_", "用户 ID 前缀")
	flag.BoolVar(&flagDryRun, "dry-run", false, "只打印计划，不请求 API")
	flag.Parse()

	return config{
		apiBase:        strings.TrimRight(flagAPIBase, "/"),
		integrationKey: flagKey,
		peerUserID:     flagPeerUser,
		peerNickname:   flagPeerNick,
		peerAvatar:     flagPeerAvatar,
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
	return fmt.Sprintf("https://oss.21rv.com/mock/images/%d.jpg%s", index, demo.AvatarSuffix)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
