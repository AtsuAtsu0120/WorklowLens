package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// validCategories は許可するカ���ゴリ。
var validCategories = map[string]bool{
	"asset":   true,
	"build":   true,
	"edit":    true,
	"error":   true,
	"session": true,
}

// LogMessage はクライアントから送られるログメッセージ。
type LogMessage struct {
	ToolName    string    `json:"tool_name"`
	Category    string    `json:"category"`
	Action      string    `json:"action"`
	Timestamp   time.Time `json:"timestamp"`
	SessionID   *string   `json:"session_id,omitempty"`
	ToolVersion *string   `json:"tool_version,omitempty"`
	UserID      *string   `json:"user_id,omitempty"`
	DurationMs  *int64    `json:"duration_ms,omitempty"`
	Traceparent *string   `json:"traceparent,omitempty"`
}

// ValidateCategory はcategoryが有効な値かどうかを検証する。
func ValidateCategory(category string) bool {
	return validCategories[category]
}

// Parse はJSONバイト列からLogMessageをパースし、バリデーションを行う。
func Parse(data []byte) (LogMessage, error) {
	var msg LogMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return LogMessage{}, fmt.Errorf("invalid JSON: %w", err)
	}

	if msg.ToolName == "" {
		return LogMessage{}, fmt.Errorf("missing required field: tool_name")
	}
	if msg.Category == "" {
		return LogMessage{}, fmt.Errorf("missing required field: category")
	}
	if !ValidateCategory(msg.Category) {
		return LogMessage{}, fmt.Errorf("invalid category: '%s'", msg.Category)
	}
	if msg.Action == "" {
		return LogMessage{}, fmt.Errorf("missing required field: action")
	}
	if msg.Timestamp.IsZero() {
		return LogMessage{}, fmt.Errorf("missing required field: timestamp")
	}

	return msg, nil
}
