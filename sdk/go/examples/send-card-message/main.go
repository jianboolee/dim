package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	dim "github.com/jianboolee/dim-sdk"
	"github.com/jianboolee/dim-sdk/examples/demo"
)

var User = demo.USER_A
var PeerUser = demo.USER_B

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
	conv, err := imClient.Services().GetOrCreatePrivateConversation(ctx, User, PeerUser)
	if err != nil {
		log.Fatal(err)
	}

	card := dim.CardInput{
		Title:       "[租车]暑期特惠，租车低至899元/天起",
		Description: "暑期出行，租车低至899元/天起，支持全国取还车，随时随地畅享自驾乐趣。",
		ImageURL:    "https://img01.wanfangche.com/uploads/im/images/20260628/bqks3c09hnp9lms7z264xahx533m7fhr.jpg?x-oss-process=image/resize,w_960,m_lfit",
		URL:         "https://code.visualstudio.com/docs/agents/overview",
		PriceText:   "￥899元/天起",
	}

	message, err := conv.SendCardMessage(ctx, card)
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
