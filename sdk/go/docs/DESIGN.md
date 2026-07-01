# d-im Go SDK 重设计

## 定位

Go SDK 面向业务后端服务使用，提供一组稳定、类型安全、可测试的 HTTP 操作封装。它是“业务服务端内部操作 IM”的能力层，不负责前端登录态、多端在线、refresh token 自动续期。

SDK 同时提供两层能力：底层 API 保持清晰、可组合；高层 `Services()` 提供通用的 conversation-first 工作流，用于完成“以某个用户身份获取/创建会话，再围绕 conversation_id 发送消息”。SDK 不内置业务专用 facade，例如 `SendOrderMessage`、`SendAuditMessage`。

---

## 设计原则

| 原则 | 说明 |
|------|------|
| 职责单一 | `Client` 管配置与 integration 能力，`Session` 表示“以某个用户身份操作”，领域 Service 只管自己的资源 |
| 显式流转 | `Client.EnsureUser(ctx, user)` 写资料，`Client.Login(ctx, userID)` 创建 `Session`，所有用户态操作都从 `Session` 发起 |
| 会话优先 | 高层 `Services()` 先获取/创建 `ConversationSession`，消息发送只围绕 conversation_id 进行 |
| 上下文优先 | 所有网络方法第一个参数都是 `context.Context` |
| 值构造器 | 消息 body 用 `TextMessage()` / `CardMessage()` 等纯函数构造，不为每种消息膨胀发送方法 |
| 参数结构化 | 复杂查询优先使用 Params struct，避免过多函数式 option 导致调用含义分散 |
| 可测试 | 对外类型职责稳定，业务侧可围绕小接口自行 mock |

---

## 非目标

以下能力不放进第一版 SDK 重构范围：

- 不管理 refresh token、自动续期、多设备登录生命周期。
- 不做 Web/APP/小程序登录态协调。
- 不提供设备管理 UI 或设备列表能力。
- 不内置业务专用 facade，例如 `SendOrderMessage`、`SendAuditMessage`；通用会话工作流由 `Services()` 提供。
- 不直接暴露内部系统事件消息的构造与发送，例如“xxx 创建了群聊”“xxx 加入了群聊”。

SDK 可以接收后端返回的 `refresh_token` / `session_id` 字段以保持响应结构完整，但不主动使用这些字段做续期逻辑。

---

## 类型职责总览

```text
Client        — 持有 BaseURL、APIKey、HTTPClient；负责 integration 用户 upsert、登录等服务端集成能力
Session       — 持有 access token；代表“以某个 IM 用户身份操作”
  ├── Conversations() — 会话创建、查询、列表、激活、已读
  ├── Messages()      — 消息发送、消息列表
  ├── Users()         — 当前用户、用户资料查询
  └── Groups()        — 群组创建、详情、成员管理

MessageInput  — 消息请求 envelope，承载幂等字段和具体消息 body
```

| 类型 | 职责 | 不做什么 |
|------|------|---------|
| `Client` | 初始化配置、integration API、用户资料 upsert、创建 `Session` | 不直接发送用户消息、不操作用户态会话 |
| `Session` | 保存 access token，提供领域 Service 入口 | 不管理 refresh token，不做自动重登 |
| `ConversationService` | 会话 CRUD、列表、搜索、激活、已读 | 不发送消息 |
| `MessageService` | 消息发送、消息列表 | 不创建会话、不构造业务流程 |
| `GroupService` | 群创建、详情、成员邀请/踢出/退出 | 不发送消息 |
| `UserService` | 当前用户、用户资料查询 | 不做 integration upsert |

---

## API 设计

### 1. 初始化

业务启动时创建一次 `Client`，复用底层 `http.Client`。

```go
imClient := im.New(
    im.WithBaseURL("http://localhost:8901"),
    im.WithAPIKey("secret-xxx"),
    im.WithHTTPClient(httpClient),
)
```

### 2. Integration 用户写入

后端需要新增独立 integration user upsert endpoint，SDK 暴露显式 upsert 方法。`EnsureUser(s)` 是用户资料写入语义，必须幂等，重复调用只更新资料，不创建登录态。

```go
err := imClient.EnsureUser(ctx, im.UserInput{
    ID:       "user_b",
    Nickname: "Brock",
    Avatar:   "https://example.com/avatar.png",
    Type:     im.UserTypeNormal,
})

err = imClient.EnsureUsers(ctx,
    im.UserInput{ID: "user_a", Nickname: "Alice"},
    im.UserInput{ID: "user_b", Nickname: "Brock"},
)
```

`UserInput.Type` 应保留。它决定账号的业务能力和交互语义，例如普通用户、系统账号、可交互机器人。SDK 不应该只靠 ID 前缀推断类型；ID 前缀最多作为服务端兼容兜底，不作为长期主设计。

upsert 语义：

- `ID` 必填，`Type` 只允许 `normal` / `system` / `bot`，不传默认为 `normal`。
- 空 `Nickname`、空 `Avatar` 不覆盖旧值，避免局部更新时误清资料。
- 非空字段按新值覆盖旧值，重复 ID 以请求中的最后一个非空值为准。
- 批量 `EnsureUsers` 建议服务端按单次请求原子处理；若任一用户非法，整批返回错误。

### 3. 创建 Session

`Login` 只根据已有用户 ID 换取用户态 access token，不写入或更新用户资料。SDK 不负责 refresh token 续期；业务方如果需要长期会话，应重新调用 integration 登录或自行管理后端返回字段。

```go
session, err := imClient.Login(ctx, "audit_bot")
if err != nil {
    return err
}
```

如果用户不存在，`Login` 应返回明确错误，例如 `ErrUserNotFound`。不要让 `Login` 隐式创建用户，否则“认证”和“资料写入”会重新耦合，后续用户类型、头像昵称、机器人能力都会变得不可控。

`Session` 可以暴露基础元信息：

```go
token := session.Token()
userID := session.UserID()
redirectURL := session.RedirectURL() // 跳转到 IM 首页
conversationURL := session.RedirectURL(im.WithConversationID(conversationID)) // 跳转到指定会话
sessionID := session.SessionID()
expiresIn := session.ExpiresIn()
```

### 4. 用户态分步操作

所有用户态操作都从 `Session` 下的领域 Service 进入。

```go
conv, err := session.Conversations().GetOrCreatePrivate(ctx, "user_b")
if err != nil {
    return err
}

msg, err := session.Messages().Send(ctx, conv.ID, im.TextMessage("你好"))
if err != nil {
    return err
}
```

列表类接口使用 Params struct：

```go
messages, err := session.Messages().List(ctx, conv.ID, im.ListMessagesParams{
    Limit:    50,
    BeforeID: "msg-id",
})

page, err := session.Conversations().List(ctx, im.ListConversationsParams{
    Limit:                20,
    Query:                "brock",
    ActiveConversationID: conv.ID,
})
```

### 5. 群组能力

群聊不是 peer-user 模型，SDK 应明确提供群组 Service，而不是把群聊塞进私聊快捷方法。创建或获取群时优先返回轻量的群会话结果，不默认携带成员列表；需要展示群成员时再调用详情或成员分页接口。

```go
group, err := session.Groups().CreateConversation(ctx, im.CreateGroupParams{
    Name:      "项目讨论",
    MemberIDs: []string{"user_c", "user_d"},
})
if err != nil {
    return err
}

err = session.Groups().Invite(ctx, group.Group.ID, []string{"user_e"})
err = session.Groups().Kick(ctx, group.Group.ID, []string{"user_d"})

msg, err := session.Messages().Send(ctx, group.Group.ConversationID, im.TextMessage("大家好"))
```

对于“内容审核群”这类全局或业务唯一群，不应通过“某个机器人只能创建一个群”的隐式规则实现，而应显式建模为带业务唯一键的 get-or-create：

```go
group, err := session.Groups().GetOrCreateConversation(ctx, im.GetOrCreateGroupParams{
    UniqueKey: "content_audit",
    Name:      "内容审核群",
    MemberIDs: nil,
})
```

建议语义：

- `scope_user_id + unique_key + status(active)` 唯一；同一机器人、同一业务键只返回一个活跃群。
- `scope_user_id` 默认为创建者 ID，创建后不可变；即使未来发生群主转让，也不影响唯一键语义。
- `MemberIDs` 允许为空；创建者自动成为 owner/member，所以群内至少有机器人自己。
- 不限制机器人只能创建一个群；机器人可以为不同业务键创建多个群。
- 群解散后允许使用相同 `scope_user_id + unique_key` 重建新群。
- 后端需要为活跃群建立 `scope_user_id + unique_key` 的唯一约束；MongoDB 可使用 partial unique index，例如只约束 `status=active`。

审核机器人仍然可以给具体用户发私聊通知：

```go
conv, err := session.Conversations().GetOrCreatePrivate(ctx, "user_a")
if err != nil {
    return err
}

_, err = session.Messages().Send(ctx, conv.ID, im.TextMessage("内容已审核通过"))
```

用户也可以按机器人 ID 拉取与该机器人的通知会话。

### 6. 用户类型

沿用当前 `normal` / `system` / `bot` 三种类型，不引入新命名和能力矩阵：

| 类型 | 语义 |
|------|------|
| `normal` | 普通用户，可以正常收发消息；默认不传即为该类型 |
| `system` | 系统账号，只能主动发消息；用户不能回复，不能主动向它发起会话 |
| `bot` | 可交互机器人，可以主动发起会话，也可以被用户发起会话/回复，可以创建群 |

`UserInput.Type` 由业务方显式传入，不靠 ID 前缀推断。`system` 和 `bot` 都允许通过 `Login(ctx, userID)` 创建服务端操作 session，因为它们需要以自身身份发送消息、创建会话或创建群。

落地时需要同步修正后端语义：`bot` 不应被当作只读系统账号处理，用户可以向 `bot` 发起会话并回复；`system` 才禁止用户发起会话和回复。

---

## 消息类型

统一用 `MessageService.Send` 发送，用纯函数构造 `MessageInput`。

```go
func (s *MessageService) Send(ctx context.Context, conversationID string, msg MessageInput) (*Message, error)

func NewMessage(body MessageBody) MessageInput
func TextMessage(content string) MessageBody
func CardMessage(input CardInput) MessageBody
func ImageMessage(url string) MessageBody
func VideoMessage(url string) MessageBody
func AudioMessage(url string) MessageBody
func LinkMessage(input LinkInput) MessageBody
```

`MessageInput` 应支持幂等字段，便于业务方避免重复发送。`ClientMessageID` 属于消息 envelope，不属于卡片、图片、文本等 payload：

```go
msg := im.NewMessage(im.CardMessage(im.CardInput{
    Title:       "租车订单已提交",
    Description: "等待商家确认",
    URL:         "https://example.com/order/123",
}))
msg.ClientMessageID = "order-123-submitted"
```

卡片消息的价格字段是展示文案，不参与金额计算。为避免 `price` 被误解为数值金额，接口字段建议命名为 `price_text`，Go 字段命名为 `PriceText`：

```go
type CardInput struct {
    Title       string
    Description string
    URL         string
    ImageURL    string
    PriceText   string // JSON: price_text，例如 "￥899元/天起"、"待确认"
}
```

落地时后端、前端、SDK 统一从 `price` 调整为 `price_text`，不保留双字段兼容，避免新旧字段并存造成歧义。

系统事件消息由后端内部流程产生，不作为普通 SDK 构造器暴露。

---

## 推荐业务封装方式

SDK 不提供 `SendOrderMessage`、`SendAuditMessage` 这类业务专用 facade；业务语义仍然留在业务系统。但 SDK 提供通用的 conversation-first `Services()` 工作流，业务系统应优先通过“获取/创建会话，再围绕会话发送消息”的方式封装自己的业务动作：

```go
func SendOrderMessage(ctx context.Context, imClient *im.Client, order Order) error {
    conv, err := imClient.Services().GetOrCreatePrivateConversation(
        ctx,
        im.UserInput{
            ID:       "order_notification",
            Nickname: "订单消息",
            Type:     im.UserTypeSystem,
        },
        im.UserInput{
            ID:       order.UserID,
            Nickname: order.UserNickname,
        },
        im.WithInitialPeerMuted(true),
    )
    if err != nil {
        return err
    }

    _, err = conv.SendMessage(ctx, im.MessageInput{
        ClientMessageID: "order-" + order.ID + "-submitted",
        Body: im.CardMessage(im.CardInput{
            Title:       "订单已提交",
            Description: "等待商家确认",
            URL:         order.URL,
            PriceText:   order.PriceText,
        }),
    })
    return err
}
```

这样 SDK 保持稳定的通用会话模型，业务语义留在业务系统。

---

## 示例

```go
imClient := im.New(
    im.WithBaseURL(apiBase),
    im.WithAPIKey(integrationKey),
)

// 1. 以订单消息身份打开/创建与 user_b 的私聊
conv, err := imClient.Services().GetOrCreatePrivateConversation(
    ctx,
    im.UserInput{ID: "system_order", Nickname: "订单消息", Type: im.UserTypeSystem},
    im.UserInput{ID: "user_b", Nickname: "Brock"},
    im.WithInitialPeerMuted(true),
)
if err != nil {
    return err
}

// 2. 围绕 conversation_id 发送卡片消息
_, err = conv.SendMessage(ctx, im.MessageInput{
    ClientMessageID: "order-123-submitted",
    Body: im.CardMessage(im.CardInput{
        Title:       "租车订单已提交",
        Description: "您的租车订单已成功提交，请耐心等待商家确认。",
        URL:         "https://www.example.com/order/detail",
        PriceText:   "待确认",
    }),
})
```

---

## 已确认决策

1. 新增独立 integration 用户 upsert endpoint（后端需新增 `POST /im/api/integration/users`），SDK 提供 `EnsureUser(s)`，并支持显式传入用户类型。
2. `Login` 只接收用户 ID，不写用户资料；用户不存在时返回明确错误。
3. 私聊会话方法命名为 `GetOrCreatePrivate`，避免 `CreatePrivate` 造成“必定新建”的误解。
4. SDK 不提供 `SendOrderMessage` 这类业务专用 facade；通用高层流程使用 `Services().GetOrCreatePrivateConversation` / `Services().GetOrCreateGroupConversation` 获取 `ConversationSession` 后发送。
5. SDK 不管理 refresh token、多设备登录和自动续期；相关字段只作为响应数据保留。
6. 群成员管理第一版提供 `CreateConversation`、`Detail`、`Invite`、`Kick`、`Leave`。
7. 业务唯一群使用 `Groups().GetOrCreateConversation(ctx, params)`，通过 `scope_user_id + unique_key + active` 唯一约束实现；群解散后允许重建。
8. `MessageInput.ClientMessageID` 放在消息 envelope，作为构造参数直接传入，不使用链式方法。
9. 用户类型沿用 `normal` / `system` / `bot`，不引入新命名和能力矩阵。
10. 不保留 `CreatePrivateConversationSession`；前端跳转地址通过 `Session.RedirectURL()` 获取，进入指定会话使用 `Session.RedirectURL(im.WithConversationID(conversationID))`。
11. 卡片消息价格字段使用展示语义命名 `price_text` / `PriceText`，不使用容易被误解为数值金额的 `price`。
