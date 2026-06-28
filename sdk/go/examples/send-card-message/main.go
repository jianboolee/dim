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

var FromUser = demo.USER_A
var ToUser = demo.USER_JIANBO

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

	integrationClient := dim.NewIntegrationClient(dim.Config{
		BaseURL: apiBase,
		APIKey:  integrationKey,
	})

	session, err := integrationClient.CreateConversation(ctx, dim.CreateConversationRequest{
		FromUser: FromUser,
		ToUser:   ToUser,
	})
	if err != nil {
		log.Fatal(err)
	}

	userClient := dim.NewUserClient(dim.Config{
		BaseURL: apiBase,
		Token:   session.Token,
	})

	var payload = dim.Payload{
		// Title:       "[电影] 玩具总动员 5",
		Description: "自第三部开始，每部《玩具总动员》都给人一种大结局的感觉，皮克斯又总能找到全新的叙事角度，再次赚足全球观众眼泪。胡迪和巴斯光年的故事告一段落后，本作的第一主角落到了翠丝肩上，在成长之余化解过去的心结。功能日新月异的电子产品，以及和大大小小的屏幕一起成长的新一代孩子，则成了玩具们面临的最大危机。",
		ImageURL:    "https://oss.21rv.com/uploads/im/images/20260628/udyrq36sbafix342l8l2yhui9p9f8jyk.webp?x-oss-process=image/resize,w_960,m_lfit",
		// URL:   "https://sspai.com/post/111562",
		// Price: "¥99.99",
	}

	message, err := userClient.SendCardMessage(ctx, session.ConversationID, payload)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		ConversationID: session.ConversationID,
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
