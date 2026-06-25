package demo

import dim "d-im-go-sdk"

const AvatarSuffix = "?x-oss-process=image/resize,m_fill,w_200,h_200"

var USER_SYSTEM_NOTICE = dim.User{
	ID:       "system_notice",
	Nickname: "系统通知",
	Avatar:   "https://oss.21rv.com/uploads/avatar/21.jpg" + AvatarSuffix,
}

var USER_SYSTEM_OPS = dim.User{
	ID:       "system_ops",
	Nickname: "系统运维",
	Avatar:   "https://oss.21rv.com/uploads/avatar/23.jpg" + AvatarSuffix,
}

var USER_A = dim.User{
	ID:       "user_a",
	Nickname: "User A",
	Avatar:   "https://oss.21rv.com/mock/images/1.jpg" + AvatarSuffix,
}

var USER_B = dim.User{
	ID:       "user_b",
	Nickname: "User B",
	Avatar:   "https://oss.21rv.com/mock/images/2.jpg" + AvatarSuffix,
}
