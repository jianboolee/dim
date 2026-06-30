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

var User = demo.USER_A
var PeerUser = demo.USER_CUSTOMER_SERVICE

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

	// card := dim.CardInput{
	// 	Title:       "览众房车 山海炮 2025款 山海炮侧拓",
	// 	Description: "",
	// 	ImageURL:    "https://img01.wanfangche.com/public/upload/202507/09/686ddeeb32cd8.jpg?x-oss-process=style/16x9",
	// 	URL:         "https://www.21rv.com/auto/model/3001",
	// 	PriceText:   "¥ 59.80万元",
	// }

	card := dim.CardInput{
		Title:       "瑞弗房车 B型房车 2026款 启界R900-六座航空座椅版",
		Description: "",
		ImageURL:    "https://img01.wanfangche.com/public/upload/202604/30/69f2f4529ed7d.png?x-oss-process=style/16x9",
		URL:         "https://www.21rv.com/auto/model/3075",
		PriceText:   "¥ 35.00万元起",
	}

	message, err := session.Messages().Send(ctx, conversation.ID, dim.NewMessage(dim.CardMessage(card)))
	if err != nil {
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
