# d-im Go SDK

Go SDK for business systems integrating with d-im over HTTP.

This module is intentionally outside `im-backend`. It must not import backend
`internal` packages or write directly to the database.

## High-level Services

Use `Services()` when the business system wants to perform a complete workflow
without manually composing user upsert, login, conversation creation, and send.

```go
imClient := dim.New(
    dim.WithBaseURL("http://localhost:8901"),
    dim.WithAPIKey("change-me-integration-key"),
)

result, err := imClient.Services().SendTextMessage(
    ctx,
    dim.UserInput{ID: "system_notice", Nickname: "系统通知", Type: dim.UserTypeSystem},
    dim.UserInput{ID: "user_a", Nickname: "Alice"},
    "hello",
    dim.WithInitialPeerMuted(true),
)
```

## Low-level APIs

Use the low-level APIs when you need to control each step.

```go
if err := imClient.EnsureUsers(ctx,
    dim.UserInput{ID: "user_a", Nickname: "Alice"},
    dim.UserInput{ID: "user_b", Nickname: "Bob"},
); err != nil {
    return err
}

session, err := imClient.Login(ctx, "user_a")
if err != nil {
    return err
}

conv, err := session.Conversations().GetOrCreatePrivate(ctx, "user_b")
if err != nil {
    return err
}

_, err = session.Messages().Send(ctx, conv.ID, dim.NewMessage(dim.TextMessage("hello")))
```

## Seed Example

Login as `user_a` and print the redirect URL for the IM home page:

```bash
cd sdk/go
go run ./examples/login-a \
  -api-base http://localhost:8901 \
  -key change-me-integration-key
```

Create the fixed `user_a` / `user_b` demo conversation and print the returned
conversation info:

```bash
cd sdk/go
go run ./examples/user-ab \
  -api-base http://localhost:8901 \
  -key change-me-integration-key
```

Send a text message from `user_a` to `user_b`:

```bash
cd sdk/go
go run ./examples/user-a-send-b \
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
