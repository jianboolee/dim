# d-im Go SDK

Go SDK for business systems integrating with d-im over HTTP.

This module is intentionally outside `im-backend`. It must not import backend
`internal` packages or write directly to the database.

## Example

```go
imClient := dim.New(
    dim.WithBaseURL("http://localhost:8901"),
    dim.WithAPIKey("change-me-integration-key"),
)

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
