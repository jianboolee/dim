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

type output struct {
	Group          *dim.Group  `json:"group"`
	ConversationID string      `json:"conversation_id"`
	Message        dim.Message `json:"message"`
}

func main() {
	apiBase := envOr("IM_API_BASE", "http://localhost:8901")
	integrationKey := envOr("INTEGRATION_API_KEY", "")
	groupName := envOr("GROUP_NAME", "咨询群")

	flag.StringVar(&apiBase, "api-base", apiBase, "IM API base URL")
	flag.StringVar(&integrationKey, "key", integrationKey, "X-Integration-Key")
	flag.StringVar(&groupName, "group-name", groupName, "group name")
	flag.Parse()

	if integrationKey == "" {
		log.Fatal("INTEGRATION_API_KEY is required (flag -key or env)")
	}

	apiBase = strings.TrimRight(apiBase, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imClient := dim.New(dim.WithBaseURL(apiBase), dim.WithAPIKey(integrationKey))

	// 1. 确保客服、卖家、买家用户的资料存在
	if err := imClient.EnsureUsers(ctx, demo.USER_CUSTOMER_SERVICE, demo.USER_SELLER_A, demo.USER_A); err != nil {
		log.Fatal(err)
	}

	// 2. 以客服身份登录并创建/获取咨询群
	csSession, err := imClient.Login(ctx, demo.USER_CUSTOMER_SERVICE.ID)
	if err != nil {
		log.Fatalf("login customer service: %v", err)
	}

	group, err := csSession.Groups().GetOrCreate(ctx, dim.GetOrCreateGroupParams{
		UniqueKey: "customer_service_seller_a_user_a",
		Name:      groupName,
		MemberIDs: []string{
			demo.USER_SELLER_A.ID,
			demo.USER_A.ID,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if group.Group == nil {
		log.Fatal("create group response missing group")
	}

	// 3. 以买家 USER_A 身份登录并发送卡片消息到群内
	userASession, err := imClient.Login(ctx, demo.USER_A.ID)
	if err != nil {
		log.Fatalf("login user_a: %v", err)
	}

	// card := dim.CardInput{
	// 	Title:       "瑞弗房车 B型房车 2026款 启界R900-六座航空座椅版",
	// 	Description: "",
	// 	ImageURL:    "https://img01.wanfangche.com/public/upload/202604/30/69f2f4529ed7d.png?x-oss-process=style/16x9",
	// 	URL:         "https://www.21rv.com/auto/model/3075",
	// 	PriceText:   "¥ 35.00万元起",
	// }

	// card := dim.CardInput{
	// 	Title:       "瑞弗房车 B型房车 2026款 启界R900-星耀版",
	// 	Description: "",
	// 	ImageURL:    "https://img01.wanfangche.com/public/upload/202604/30/69f2ef68e39ac.png?x-oss-process=style/16x9",
	// 	URL:         "https://www.21rv.com/auto/model/3074",
	// 	PriceText:   "¥ 50.80万元",
	// }

	card := dim.CardInput{
		Title:       "瑞弗房车 C型房车 2026款 V820-6.5T",
		Description: "",
		ImageURL:    "https://img01.wanfangche.com/public/upload/202604/30/69f3154534fa3.png?x-oss-process=style/16x9",
		URL:         "https://www.21rv.com/auto/model/3082",
		PriceText:   "¥ 60.8万",
	}

	message, err := userASession.Messages().Send(ctx, group.Group.ConversationID, dim.NewMessage(dim.CardMessage(card)))
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
