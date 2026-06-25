# d-im Go SDK

Go SDK for business systems integrating with d-im over HTTP.

This module is intentionally outside `im-backend`. It must not import backend
`internal` packages or write directly to the database.

## Example

```go
integrationClient := dim.NewIntegrationClient(dim.Config{
    BaseURL: "http://localhost:8901",
    APIKey:  "change-me-integration-key",
})

session, err := integrationClient.CreateConversation(ctx, dim.CreateConversationRequest{
    FromUser: dim.User{ID: "user_a", Nickname: "Alice"},
    ToUser:   dim.User{ID: "user_b", Nickname: "Bob"},
})
if err != nil {
    return err
}

userClient := dim.NewUserClient(dim.Config{
    BaseURL: "http://localhost:8901",
    Token:   session.Token,
})

_, err = userClient.SendTextMessage(ctx, session.ConversationID, "hello")
```

## Seed Example

Create the fixed `user_a` / `user_b` demo conversation and print the returned
conversation info:

```bash
cd sdk/go
go run ./examples/user-ab \
  -api-base http://localhost:8901 \
  -key change-me-integration-key
```

The seed example simulates a business system creating conversations and sending
messages through the SDK.

```bash
cd sdk/go
go run ./examples/seed \
  -api-base http://localhost:8901 \
  -key change-me-integration-key \
  -count 20
```
