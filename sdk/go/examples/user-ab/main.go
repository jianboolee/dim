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
)

var USER_A = dim.User{
	ID:       "user_a",
	Nickname: "User A",
	Avatar:   "https://oss.21rv.com/mock/images/1.jpg?x-oss-process=image/resize,m_fill,w_200,h_200",
}

var USER_B = dim.User{
	ID:       "user_b",
	Nickname: "User B",
	Avatar:   "https://oss.21rv.com/mock/images/2.jpg?x-oss-process=image/resize,m_fill,w_200,h_200",
}

type output struct {
	Token          string `json:"token"`
	ExpiresIn      int    `json:"expires_in"`
	ConversationID string `json:"conversation_id"`
	RedirectURL    string `json:"redirect_url"`
	// Conversation   dim.Conversation `json:"conversation"`
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
		FromUser: USER_A,
		ToUser:   USER_B,
	})
	if err != nil {
		log.Fatal(err)
	}

	// userClient := dim.NewUserClient(dim.Config{
	// 	BaseURL: apiBase,
	// 	Token:   session.Token,
	// })

	// conversation, err := userClient.OpenConversation(ctx, session.ConversationID)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{
		Token:          session.Token,
		ExpiresIn:      session.ExpiresIn,
		ConversationID: session.ConversationID,
		RedirectURL:    session.RedirectURL,
		// Conversation:   *conversation,
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
