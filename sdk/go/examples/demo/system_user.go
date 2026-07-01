// Package demo 提供 IM SDK 的示例代码。
//
// 本文件定义了系统默认用户（System User），用于业务系统向指定用户
// 发送系统级消息（如通知、提醒、客服等场景）。
// 系统用户通过 EnsureUser 写入资料，再通过 Login 获取服务端操作会话。
// 用户类型必须显式传入，不能依赖 "system_" ID 前缀推断。
//
// 使用方法：
//
//	// 以系统通知身份向用户 Alice 发送文本消息
//	imClient.EnsureUsers(ctx, demo.USER_SYSTEM_NOTICE, demo.USER_A)
//	session, _ := imClient.Login(ctx, demo.USER_SYSTEM_NOTICE.ID)
//	conv, _ := session.Conversations().GetOrCreatePrivate(ctx, demo.USER_A.ID)
//	session.Messages().Send(ctx, conv.ID, dim.NewMessage(dim.TextMessage("通知内容")))
//
//	// 以系统订单身份向用户 Alice 发送卡片消息
//	imClient.EnsureUsers(ctx, demo.USER_SYSTEM_ORDER, demo.USER_A)
//	session, _ := imClient.Login(ctx, demo.USER_SYSTEM_ORDER.ID)
//	conv, _ := session.Conversations().GetOrCreatePrivate(ctx, demo.USER_A.ID)
//	session.Messages().Send(ctx, conv.ID, dim.NewMessage(dim.CardMessage(dim.CardInput{
//	    Title:       "订单已发货",
//	    Description: "您的订单已由顺丰快递发出，运单号：SF1234567890",
//	    ImageURL:    "https://img.example.com/order.jpg",
//	    URL:         "https://example.com/order/123",
//	})))
//
// 系统用户 ID 列表及用途说明：
//
//	system_notice       系统通知 —— 通用系统级通知消息
//	system_ops          系统运维 —— 运维公告、升级维护通知
//	system_order        订单消息 —— 订单状态变更（已下单/已发货/已完成等）
//	system_payment      支付消息 —— 支付成功、退款、分期审核等
//	system_logistics    物流消息 —— 发货、运输中、签收等物流动态
//	system_interaction  互动消息 —— 点赞、评论、关注等社交互动提醒
//	system_service      客服消息 —— 工单处理、投诉反馈、咨询回复
//	system_reminder     提醒通知 —— 预约提醒、到期提醒、活动截止等
//	system_audit        审核消息 —— 内容审核结果、实名认证结果等
//	system_report       数据报告 —— 周报、月报、销售数据等统计报告
package demo

import dim "github.com/jianboolee/dim-sdk"

const AvatarSuffix = "?x-oss-process=image/resize,m_fill,w_200,h_200"

// ---- 通用系统 ----

// USER_SYSTEM_NOTICE 系统通知，用于通用系统级通知消息
var USER_SYSTEM_NOTICE = dim.User{
	ID:       "system_notice",
	Nickname: "系统通知",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/21.jpg" + AvatarSuffix,
	Type:     "system",
}

// USER_SYSTEM_OPS 系统运维，用于运维公告、升级维护通知等
var USER_SYSTEM_OPS = dim.User{
	ID:       "system_ops",
	Nickname: "系统运维",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/23.jpg" + AvatarSuffix,
	Type:     "system",
}

// ---- 交易相关 ----

// USER_SYSTEM_ORDER 订单消息，用于订单状态变更通知
var USER_SYSTEM_ORDER = dim.User{
	ID:       "system_order",
	Nickname: "订单消息",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/22.jpg" + AvatarSuffix,
	Type:     "system",
}

// USER_SYSTEM_PAYMENT 支付消息，用于支付成功/失败、退款到账等
var USER_SYSTEM_PAYMENT = dim.User{
	ID:       "system_payment",
	Nickname: "支付消息",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/27.jpg" + AvatarSuffix,
	Type:     "system",
}

// USER_SYSTEM_LOGISTICS 物流消息，用于发货、运输中、签收等物流动态
var USER_SYSTEM_LOGISTICS = dim.User{
	ID:       "system_logistics",
	Nickname: "物流消息",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/28.jpg" + AvatarSuffix,
	Type:     "system",
}

// ---- 社交互动 ----

// USER_SYSTEM_INTERACTION 互动消息，用于点赞、评论、关注等社交提醒
var USER_SYSTEM_INTERACTION = dim.User{
	ID:       "system_interaction",
	Nickname: "互动消息",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/24.jpg" + AvatarSuffix,
	Type:     "system",
}

// ---- 运营服务 ----

// USER_SYSTEM_SERVICE 客服消息（bot 类型，用户可回复），用于工单处理、投诉反馈、咨询回复等
var USER_SYSTEM_SERVICE = dim.User{
	ID:       "system_service",
	Nickname: "客服消息",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/29.jpg" + AvatarSuffix,
	Type:     "bot",
}

// USER_SYSTEM_REMINDER 提醒通知，用于预约提醒、到期提醒、活动截止等
var USER_SYSTEM_REMINDER = dim.User{
	ID:       "system_reminder",
	Nickname: "提醒通知",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/30.jpg" + AvatarSuffix,
	Type:     "system",
}

// USER_SYSTEM_AUDIT 审核消息，用于内容审核、实名认证等审核结果通知
var USER_SYSTEM_AUDIT = dim.User{
	ID:       "system_audit",
	Nickname: "审核消息",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/23.jpg" + AvatarSuffix,
	Type:     "bot",
}

// ---- 数据分析 ----

// USER_SYSTEM_REPORT 数据报告，用于周报、月报、销售数据统计等
var USER_SYSTEM_REPORT = dim.User{
	ID:       "system_report",
	Nickname: "数据报告",
	Avatar:   "https://img01.wanfangche.com/uploads/avatar/32.jpg" + AvatarSuffix,
	Type:     "system",
}
