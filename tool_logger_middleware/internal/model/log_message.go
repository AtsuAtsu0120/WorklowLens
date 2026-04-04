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

// LogMessage はクライアントから送られるログメッセージ。
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

// Parse はJSONバイト列からLogMessageをパースし、バリデーションを行う。
func Parse(data []byte) (LogMessage, error) {
	var msg LogMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return LogMessage{}, fmt.Errorf("invalid JSON: %w", err)
	}

	if msg.ToolName == "" {
		return LogMessage{}, fmt.Errorf("missing required field: tool_name")
	}
	if msg.EventType == "" {
		return LogMessage{}, fmt.Errorf("missing required field: event_type")
	}
	if !ValidateEventType(msg.EventType) {
		return LogMessage{}, fmt.Errorf("invalid event_type: '%s'", msg.EventType)
	}
	if msg.Timestamp.IsZero() {
		return LogMessage{}, fmt.Errorf("missing required field: timestamp")
	}
	if msg.Message == "" {
		return LogMessage{}, fmt.Errorf("missing required field: message")
	}

	return msg, nil
}
