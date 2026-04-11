package model

import (
	"fmt"
	"time"
)

// validCategories は許可するカテゴリ。
var validCategories = map[string]bool{
	"asset":   true,
	"build":   true,
	"edit":    true,
	"error":   true,
	"session": true,
}

// LogMessage はmiddlewareから受信するログメッセージ。
// middleware側のLogMessage構造体と同じフィールドを持つ。
type LogMessage struct {
	ToolName    string    `json:"tool_name"`
	Category    string    `json:"category"`
	Action      string    `json:"action"`
	Timestamp   time.Time `json:"timestamp"`
	SessionID   *string   `json:"session_id,omitempty"`
	ToolVersion *string   `json:"tool_version,omitempty"`
	UserID      *string   `json:"user_id,omitempty"`
	DurationMs  *int64    `json:"duration_ms,omitempty"`
}

// ValidateCategory はcategoryが有効な値かどうかを検証する。
func ValidateCategory(category string) bool {
	return validCategories[category]
}

// ValidateLogs はLogMessage配列のバリデーションを行う���
// エラーがあれば最初に見つかったエラーを返す。
func ValidateLogs(logs []LogMessage) error {
	for i, log := range logs {
		if log.ToolName == "" {
			return fmt.Errorf("missing required field: tool_name at index %d", i)
		}
		if log.Category == "" {
			return fmt.Errorf("missing required field: category at index %d", i)
		}
		if !ValidateCategory(log.Category) {
			return fmt.Errorf("invalid category: '%s' at index %d", log.Category, i)
		}
		if log.Action == "" {
			return fmt.Errorf("missing required field: action at index %d", i)
		}
		if log.Timestamp.IsZero() {
			return fmt.Errorf("missing required field: timestamp at index %d", i)
		}
	}
	return nil
}
