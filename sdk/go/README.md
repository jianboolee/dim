# d-im Go SDK

Go SDK for business systems integrating with d-im over HTTP.

The module path is:

```go
github.com/jianboolee/dim/sdk/go
```

This SDK is intended for business systems. It should call d-im through HTTP APIs
and must not import backend `internal` packages or write directly to the
database.

## Installation

```bash
go get github.com/jianboolee/dim/sdk/go@v0.1.1
go mod tidy
```

Import it in your code:

```go
import dim "github.com/jianboolee/dim/sdk/go"
```

Do not import `github.com/jianboolee/dim/sdk/go/examples/demo` from production
business code. It is only sample data for SDK examples.

If the repository is private, configure the caller machine first:

```bash
go env -w GOPRIVATE=github.com/jianboolee/dim
```

## Client

Create a client with the d-im API base URL and integration API key.

```go
imClient := dim.New(
    dim.WithBaseURL("https://your-dim-api-host"),
    dim.WithAPIKey("your-integration-api-key"),
)
```

Available client options:

| Option | Description |
| --- | --- |
| `dim.WithBaseURL(baseURL)` | d-im API server address, for example `http://localhost:8901`. |
| `dim.WithAPIKey(apiKey)` | Integration key sent as `X-Integration-Key`. |
| `dim.WithHTTPClient(httpClient)` | Custom `*http.Client`. The default timeout is 30 seconds. |

## Recommended: Services API

Use `client.Services()` for normal business integrations. It wraps common server
side flows such as ensuring users exist, logging in as an operator, creating or
getting conversations, and returning a conversation object that can send
messages directly.

### Send a Private Message

```go
package main

import (
    "context"
    "log"
    "time"

    dim "github.com/jianboolee/dim/sdk/go"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    imClient := dim.New(
        dim.WithBaseURL("https://your-dim-api-host"),
        dim.WithAPIKey("your-integration-api-key"),
    )

    conv, err := imClient.Services().GetOrCreatePrivateConversation(
        ctx,
        dim.UserInput{
            ID:       "system_notice",
            Nickname: "系统通知",
            Type:     dim.UserTypeSystem,
        },
        dim.UserInput{
            ID:       "user_123",
            Nickname: "张三",
            AvatarURL: "https://example.com/avatar.png",
        },
        dim.WithInitialPeerMuted(true),
    )
    if err != nil {
        log.Fatal(err)
    }

    message, err := conv.SendTextMessage(ctx, "你的订单已发货")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("conversation:", conv.ID(), "message:", message.ID)
}
```

### Send a Group Message

When `GroupTarget.UniqueKey` is provided, the SDK calls get-or-create group
conversation. Use a stable `UniqueKey` when the same business object should map
to the same group, for example an audit task, order, or ticket.

```go
conv, err := imClient.Services().GetOrCreateGroupConversation(
    ctx,
    dim.UserInput{
        ID:       "audit_bot",
        Nickname: "审核助手",
        Type:     dim.UserTypeBot,
    },
    dim.GroupTarget{
        Name:      "内容审核群",
        UniqueKey: "content_audit_123",
        MemberUsers: []dim.UserInput{
            {ID: "auditor_1", Nickname: "审核员 A"},
            {ID: "auditor_2", Nickname: "审核员 B"},
        },
    },
)
if err != nil {
    return err
}

_, err = conv.SendTextMessage(ctx, "有一条内容需要审核")
```

### Send a Card Message

```go
_, err = conv.SendCardMessage(ctx, dim.CardInput{
    Title:       "订单 #202607010001",
    Description: "待支付订单",
    URL:         "https://example.com/orders/202607010001",
    ImageURL:    "https://example.com/order-cover.png",
    PriceText:   "¥199.00",
})
```

### Continue an Existing Conversation

Use `GetConversation` when your business system already has a conversation ID
and wants to send as a specific user.

```go
conv, err := imClient.Services().GetConversation(
    ctx,
    dim.UserInput{ID: "system_notice"},
    "conversation_id",
)
if err != nil {
    return err
}

_, err = conv.SendTextMessage(ctx, "这是一条后续通知")
```

### Ensure Group Members

Use `EnsureGroupMembers` to invite users into an existing group conversation.

```go
err := imClient.Services().EnsureGroupMembers(
    ctx,
    dim.UserInput{ID: "audit_bot", Nickname: "审核助手", Type: dim.UserTypeBot},
    "group_id",
    []dim.UserInput{
        {ID: "auditor_3", Nickname: "审核员 C"},
    },
    nil,
)
```

### Services Methods

| Method | Description |
| --- | --- |
| `Services().LoginUser(ctx, user)` | Ensure a user exists, then log in as that user. |
| `Services().GetOrCreatePrivateConversation(ctx, user, peerUser, options...)` | Ensure both users exist, log in as `user`, get or create a private conversation with `peerUser`. |
| `Services().GetOrCreateGroupConversation(ctx, owner, target, options...)` | Ensure users exist, log in as `owner`, create or get a group conversation. |
| `Services().EnsureGroupMembers(ctx, operator, groupID, members, memberIDs, options...)` | Ensure users exist, log in as `operator`, invite members to a group. |
| `Services().GetConversation(ctx, user, conversationID)` | Log in as `user`, load an existing conversation, and return a send-capable conversation session. |

Available service options:

| Option | Applies to | Description |
| --- | --- | --- |
| `dim.WithoutEnsureUsers()` | private conversation, group conversation, group members | Skip automatic user upsert when the caller already knows users exist. |
| `dim.WithInitialPeerMuted(muted)` | private conversation | Set initial mute state for the peer user when creating the conversation. |
| `dim.WithInitialSenderMuted(muted)` | private conversation | Set initial mute state for the sender user when creating the conversation. |
| `dim.WithInitialMemberMuted(userID, muted)` | private conversation and low-level private conversation creation | Set initial mute state for a specific member. |

## ConversationSession

`Services()` returns `*dim.ConversationSession` for conversation-oriented flows.

| Method / Field | Description |
| --- | --- |
| `conv.ID()` | Conversation ID. |
| `conv.Conversation` | Raw `*dim.Conversation` object returned by the API. |
| `conv.SendTextMessage(ctx, content)` | Send a text message. |
| `conv.SendCardMessage(ctx, card)` | Send a card message. |
| `conv.SendMessage(ctx, message)` | Send any `dim.MessageInput`, including image, video, audio, link, or custom body. |

## Users

Use `dim.UserInput` when passing users to the SDK.

```go
user := dim.UserInput{
    ID:        "user_123",
    Nickname:  "张三",
    AvatarURL: "https://example.com/avatar.png",
    Type:      dim.UserTypeNormal,
}
```

Fields:

| Field | Description |
| --- | --- |
| `ID` | Required. Stable business user ID. |
| `Nickname` | Display name. |
| `Avatar` / `AvatarURL` | Avatar address. Prefer `AvatarURL` for new code. |
| `Type` | User type: `dim.UserTypeNormal`, `dim.UserTypeSystem`, or `dim.UserTypeBot`. |

## Messages

High-level send helpers:

```go
_, err := conv.SendTextMessage(ctx, "hello")

_, err = conv.SendCardMessage(ctx, dim.CardInput{
    Title:       "卡片标题",
    Description: "卡片描述",
    URL:         "https://example.com",
    ImageURL:    "https://example.com/image.png",
    PriceText:   "¥99.00",
})
```

Low-level message builders:

| Builder | Message type |
| --- | --- |
| `dim.TextMessage(content)` | `text` |
| `dim.CardMessage(card)` | `card` |
| `dim.ImageMessage(url)` | `image` |
| `dim.VideoMessage(url)` | `video` |
| `dim.AudioMessage(url)` | `audio` |
| `dim.LinkMessage(card)` | `link` |
| `dim.NewMessage(body)` | Wrap a `MessageBody` as `MessageInput`. |

Example:

```go
msg := dim.NewMessage(dim.ImageMessage("https://example.com/image.png"))
msg.ClientMessageID = "business-message-id-001"

_, err := conv.SendMessage(ctx, msg)
```

## Low-level APIs

Use low-level APIs when you need to control each step yourself, such as explicit
login, listing conversations, reading messages, or managing groups.

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

### Client and Session Methods

| Method | Description |
| --- | --- |
| `client.EnsureUser(ctx, user)` | Upsert one user. |
| `client.EnsureUsers(ctx, users...)` | Upsert multiple users. |
| `client.Login(ctx, userID)` | Log in as a user and return `*dim.Session`. |
| `session.Token()` | Access token. |
| `session.UserID()` | Logged-in user ID. |
| `session.RefreshToken()` | Refresh token returned by the API. |
| `session.SessionID()` | Session ID returned by the API. |
| `session.ExpiresIn()` | Access token lifetime in seconds. |
| `session.RedirectURL(options...)` | Frontend redirect URL. Use `dim.WithConversationID(conversationID)` to open a conversation. |
| `session.Conversations()` | Low-level conversation service. |
| `session.Messages()` | Low-level message service. |
| `session.Groups()` | Low-level group service. |
| `session.Users()` | Low-level user service. |

### ConversationService

```go
conversations := session.Conversations()
```

| Method | Description |
| --- | --- |
| `GetOrCreatePrivate(ctx, peerID, options...)` | Get or create a private conversation with `peerID`. |
| `Get(ctx, conversationID)` | Get one conversation. |
| `Activate(ctx, conversationID)` | Activate/open a conversation for the current user. |
| `MarkRead(ctx, conversationID)` | Mark a conversation as read. |
| `UpdateSettings(ctx, conversationID, patch)` | Update current user's conversation settings, such as muted or pinned. |
| `List(ctx, params)` | List conversations. |

Examples:

```go
muted := true
state, err := session.Conversations().UpdateSettings(ctx, conv.ID, dim.ConversationSettingsPatch{
    Muted: &muted,
})

page, err := session.Conversations().List(ctx, dim.ListConversationsParams{
    Limit: 20,
    Query: "订单",
})
```

### MessageService

```go
messages := session.Messages()
```

| Method | Description |
| --- | --- |
| `Send(ctx, conversationID, msg)` | Send a message. |
| `List(ctx, conversationID, params)` | List messages in a conversation. |
| `Search(ctx, conversationID, params)` | Search messages in a conversation. |

Examples:

```go
items, err := session.Messages().List(ctx, conv.ID, dim.ListMessagesParams{
    Limit: 20,
})

result, err := session.Messages().Search(ctx, conv.ID, dim.SearchMessagesParams{
    Query: "订单",
    Limit: 20,
})
```

### GroupService

```go
groups := session.Groups()
```

| Method | Description |
| --- | --- |
| `CreateConversation(ctx, req)` | Create a group conversation. |
| `GetOrCreateConversation(ctx, req)` | Get or create a group conversation by `UniqueKey`. |
| `Detail(ctx, groupID)` | Get group detail and members. |
| `ListMembers(ctx, groupID, limit, cursor)` | List group members. |
| `Invite(ctx, groupID, userIDs)` | Invite users to a group. |
| `Kick(ctx, groupID, userID)` | Remove a user from a group. |
| `Leave(ctx, groupID)` | Current user leaves a group. |

Example:

```go
groupConv, err := session.Groups().GetOrCreateConversation(ctx, dim.GetOrCreateGroupParams{
    Name:      "订单协作群",
    UniqueKey: "order_202607010001",
    MemberIDs: []string{"user_a", "user_b"},
})
if err != nil {
    return err
}

detail, err := session.Groups().Detail(ctx, groupConv.Group.ID)
```

### UserService

```go
users := session.Users()
```

| Method | Description |
| --- | --- |
| `Me(ctx)` | Get current logged-in user. |
| `Get(ctx, userID)` | Get one user by ID. |

## Common Data Types

### GroupTarget

Used by `Services().GetOrCreateGroupConversation`.

| Field | Description |
| --- | --- |
| `Name` | Group name. |
| `UniqueKey` | Stable business key. If present, the SDK uses get-or-create semantics. |
| `ScopeUserID` | Optional scope user ID for server-side grouping semantics. |
| `MemberUsers` | Member user objects. These users are ensured automatically unless `WithoutEnsureUsers()` is used. |
| `MemberIDs` | Member IDs when user profiles are already known by d-im. |

### CreateGroupRequest / GetOrCreateGroupParams

Used by low-level `GroupService`.

| Field | Description |
| --- | --- |
| `Name` | Group name. |
| `MemberIDs` | Initial group member IDs. |
| `UniqueKey` | Stable business key for get-or-create. |
| `ScopeUserID` | Optional scope user ID. |

### ListConversationsParams

| Field | Description |
| --- | --- |
| `Limit` | Page size. |
| `Cursor` | Pagination cursor. |
| `Query` | Search keyword. |
| `ActiveConversationID` | Conversation ID that should be treated as active. |

### ListMessagesParams

| Field | Description |
| --- | --- |
| `Limit` | Page size. |
| `BeforeID` | Load messages before this message ID. |
| `AfterID` | Load messages after this message ID. |

### SearchMessagesParams

| Field | Description |
| --- | --- |
| `Query` | Search keyword. |
| `Limit` | Page size. |
| `Cursor` | Pagination cursor. |

## Examples

High-level service examples:

```bash
cd sdk/go
go run ./examples/send-message \
  -api-base http://localhost:8901 \
  -key change-me-integration-key

go run ./examples/system-notice \
  -api-base http://localhost:8901 \
  -key change-me-integration-key

go run ./examples/send-order-message \
  -api-base http://localhost:8901 \
  -key change-me-integration-key
```

Low-level examples:

```bash
cd sdk/go
go run ./examples/login \
  -api-base http://localhost:8901 \
  -key change-me-integration-key

go run ./examples/user-ab \
  -api-base http://localhost:8901 \
  -key change-me-integration-key
```

The seed example simulates a business system creating conversations and sending
messages.

```bash
cd sdk/go
go run ./examples/seed \
  -api-base http://localhost:8901 \
  -key change-me-integration-key \
  -count 20
```

## Version Tags

This SDK is a Go module in the `sdk/go` subdirectory of the large repository.
When publishing a new SDK version, create tags with the subdirectory prefix:

```bash
git tag sdk/go/v0.1.1
git push origin sdk/go/v0.1.1
```

Callers still install the version without the prefix:

```bash
go get github.com/jianboolee/dim/sdk/go@v0.1.1
```
