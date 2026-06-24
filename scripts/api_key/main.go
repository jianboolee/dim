package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func generateSignature(appID, secret, timestamp string) string {
	data := appID + "." + timestamp

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GenerateAPIKey(appID, secret string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	sign := generateSignature(appID, secret, timestamp)

	raw := strings.Join([]string{
		appID,
		timestamp,
		sign,
	}, ".")

	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func ParseAPIKey(apiKey string, secret string) (string, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(apiKey)
	if err != nil {
		return "", fmt.Errorf("invalid base64 key")
	}

	parts := strings.Split(string(rawBytes), ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid key format")
	}

	appID := parts[0]
	timestamp := parts[1]
	sign := parts[2]

	expectedSign := generateSignature(appID, secret, timestamp)

	if !hmac.Equal([]byte(sign), []byte(expectedSign)) {
		return "", fmt.Errorf("signature mismatch")
	}

	return appID, nil
}

func main() {
	appID := "dbe77e578b022317"
	secret := "yTdSoukh5TXe4K3793krMgPZ1XBEEStXNygLzAZEQ"

	apiKey := GenerateAPIKey(appID, secret)
	fmt.Println("API KEY:", apiKey)

	parsedAppID, err := ParseAPIKey(apiKey, secret)
	if err != nil {
		panic(err)
	}

	fmt.Println("Parsed APPID:", parsedAppID)
}
