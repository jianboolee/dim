package utils

import (
	"crypto/sha1"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NormalizeParticipantIDs 去重并按字典序排序参与者 ID
func NormalizeParticipantIDs(participants []string) []string {
	uniq := make(map[string]struct{}, len(participants))
	for _, participant := range participants {
		participant = strings.TrimSpace(participant)
		if participant == "" {
			continue
		}
		uniq[participant] = struct{}{}
	}

	result := make([]string, 0, len(uniq))
	for participant := range uniq {
		result = append(result, participant)
	}
	sort.Strings(result)
	return result
}

// GenerateConversationHashID 根据参与者生成唯一、稳定的 ObjectID（顺序无关 + 去重）
func GenerateConversationHashID(participants []string) primitive.ObjectID {
	deduped := NormalizeParticipantIDs(participants)

	if len(deduped) == 0 {
		return primitive.NewObjectID()
	}

	raw := strings.Join(deduped, "|")
	hash := sha1.Sum([]byte(raw))

	var b [12]byte
	copy(b[:], hash[:12])

	return primitive.ObjectID(b)
}
