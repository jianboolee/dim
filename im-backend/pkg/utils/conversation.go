package utils

import (
	"crypto/sha1"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateConversationID 根据参与者生成唯一、稳定的 ObjectID（顺序无关 + 去重）
func GenerateConversationHashID(participants []string) primitive.ObjectID {
	// 去重
	uniq := make(map[string]struct{})
	for _, p := range participants {
		uniq[p] = struct{}{}
	}

	if len(uniq) == 0 {
		// 参与者为空，生成一个默认的 ObjectID
		return primitive.NewObjectID()
	}

	deduped := make([]string, 0, len(uniq))
	for p := range uniq {
		deduped = append(deduped, p)
	}

	// 排序
	sort.Strings(deduped)

	// 连接为字符串
	raw := strings.Join(deduped, "|")

	// 计算 SHA1 hash
	hash := sha1.Sum([]byte(raw)) // 20 字节

	// 用前 12 字节构造 ObjectID（满足 MongoDB 要求）
	var b [12]byte
	copy(b[:], hash[:12])

	return primitive.ObjectID(b)
}
