package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// validEventTypes は許可するイベント種別。
var validEventTypes = map[string]bool{
	"usage":         true,
	"error":         true,
	"session_start": true,
	"session_end":   true,
	"cancellation":  true,
}

// LogMessage はmiddlewareから受信するログメ���セージ。
// middleware側のRust LogMessage構造体と同じフィールドを持��。
type LogMessage struct {
	ToolName    string          `json:"tool_name"`
	EventType   string          `json:"event_type"`
	Timestamp   time.Time       `json:"timestamp"`
	Message     string          `json:"message"`
	SessionID   *string         `json:"session_id,omitempty"`
	ToolVersion *string         `json:"tool_version,omitempty"`
	Details     json.RawMessage `json:"details,omitempty"`
}

// ValidateEventType はevent_typeが有効な値かどうかを検証する。
func ValidateEventType(eventType string) bool {
	return validEventTypes[eventType]
}

// ValidateLogs はLogMessage配列のバリデーションを行う。
// エラーがあれば最初に見つかったエラーを返す。
func ValidateLogs(logs []LogMessage) error {
	for i, log := range logs {
		if log.ToolName == "" {
			return fmt.Errorf("missing required field: tool_name at index %d", i)
		}
		if log.EventType == "" {
			return fmt.Errorf("missing required field: event_type at index %d", i)
		}
		if !ValidateEventType(log.EventType) {
			return fmt.Errorf("invalid event_type: '%s' at index %d", log.EventType, i)
		}
		if log.Timestamp.IsZero() {
			return fmt.Errorf("missing required field: timestamp at index %d", i)
		}
		if log.Message == "" {
			return fmt.Errorf("missing required field: message at index %d", i)
		}
	}
	return nil
}
