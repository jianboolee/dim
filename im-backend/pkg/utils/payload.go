package utils

import (
	"encoding/json"
	"fmt"
)

// ParsePayloadByType 尝试将 any 类型的 payload 解码为指定的结构体类型 T
func ParsePayloadByType[T any](payload any) (*T, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	switch v := payload.(type) {
	case T: // 直接是目标类型
		return &v, nil
	case *T: // 已是指针
		return v, nil
	case map[string]interface{}:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal failed: %w", err)
		}
		var result T
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("unmarshal failed: %w", err)
		}
		return &result, nil
	default:
		// 再尝试用 JSON 编码/解码处理其他兼容结构
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal failed: %w", err)
		}
		var result T
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("unmarshal failed: %w", err)
		}
		return &result, nil
	}
}
