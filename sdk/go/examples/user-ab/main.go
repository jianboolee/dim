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
	Token          string           `json:"token"`
	ExpiresIn      int              `json:"expires_in"`
	ConversationID string           `json:"conversation_id"`
	RedirectURL    string           `json:"redirect_url"`
	Conversation   dim.Conversation `json:"conversation"`
}

var User = demo.USER_A
var PeerUser = demo.USER_B

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
	conversation, err = session.Conversations().Activate(ctx, conversation.ID)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		Token:          session.Token(),
		ExpiresIn:      session.ExpiresIn(),
		ConversationID: conversation.ID,
		RedirectURL:    session.RedirectURL(dim.WithConversationID(conversation.ID)),
		Conversation:   *conversation,
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
