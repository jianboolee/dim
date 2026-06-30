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

var LOGIN_USER = demo.USER_CUSTOMER_SERVICE

func main() {
	apiBase := envOr("IM_API_BASE", "http://localhost:8901")
	integrationKey := envOr("INTEGRATION_API_KEY", "")

	flag.StringVar(&apiBase, "api-base", apiBase, "IM API base URL")
	flag.StringVar(&integrationKey, "key", integrationKey, "X-Integration-Key")
	flag.Parse()

	if integrationKey == "" {
		log.Fatal("INTEGRATION_API_KEY is required (flag -key or env)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imClient := dim.New(
		dim.WithBaseURL(strings.TrimRight(apiBase, "/")),
		dim.WithAPIKey(integrationKey),
	)

	if err := imClient.EnsureUser(ctx, LOGIN_USER); err != nil {
		log.Fatal(err)
	}
	session, err := imClient.Login(ctx, LOGIN_USER.ID)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(map[string]any{
		"token":         session.Token(),
		"user_id":       session.UserID(),
		"session_id":    session.SessionID(),
		"expires_in":    session.ExpiresIn(),
		"redirect_url":  session.RedirectURL(),
		"refresh_token": session.RefreshToken(),
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
