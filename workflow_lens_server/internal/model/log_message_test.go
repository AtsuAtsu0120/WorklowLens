package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidateEventType(t *testing.T) {
	valid := []string{"usage", "error", "session_start", "session_end", "cancellation"}
	for _, et := range valid {
		if !ValidateEventType(et) {
			t.Errorf("expected %q to be valid", et)
		}
	}

	invalid := []string{"unknown", "start", "end", "", "Usage"}
	for _, et := range invalid {
		if ValidateEventType(et) {
			t.Errorf("expected %q to be invalid", et)
		}
	}
}

func TestValidateLogs_Valid(t *testing.T) {
	logs := []LogMessage{
		{
			ToolName:  "TestTool",
			EventType: "usage",
			Timestamp: time.Now(),
			Message:   "hello",
		},
		{
			ToolName:  "TestTool",
			EventType: "session_start",
			Timestamp: time.Now(),
			Message:   "started",
		},
	}
	if err := ValidateLogs(logs); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateLogs_Empty(t *testing.T) {
	if err := ValidateLogs([]LogMessage{}); err != nil {
		t.Errorf("expected no error for empty slice, got %v", err)
	}
}

func TestValidateLogs_MissingToolName(t *testing.T) {
	logs := []LogMessage{
		{EventType: "usage", Timestamp: time.Now(), Message: "hello"},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for missing tool_name")
	}
	expected := "missing required field: tool_name at index 0"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidateLogs_MissingEventType(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", Timestamp: time.Now(), Message: "hello"},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for missing event_type")
	}
}

func TestValidateLogs_InvalidEventType(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", EventType: "unknown", Timestamp: time.Now(), Message: "hello"},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for invalid event_type")
	}
	expected := "invalid event_type: 'unknown' at index 0"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidateLogs_MissingTimestamp(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", EventType: "usage", Message: "hello"},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for missing timestamp")
	}
}

func TestValidateLogs_MissingMessage(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", EventType: "usage", Timestamp: time.Now()},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for missing message")
	}
}

func TestValidateLogs_ErrorAtSecondIndex(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", EventType: "usage", Timestamp: time.Now(), Message: "ok"},
		{ToolName: "", EventType: "usage", Timestamp: time.Now(), Message: "bad"},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error")
	}
	expected := "missing required field: tool_name at index 1"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestLogMessage_JSONUnmarshal(t *testing.T) {
	input := `{
		"tool_name": "UnityTool",
		"event_type": "usage",
		"timestamp": "2026-04-04T10:00:00Z",
		"message": "hello",
		"session_id": "abc123",
		"tool_version": "1.0.0",
		"details": {"feature": "paint"}
	}`

	var msg LogMessage
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg.ToolName != "UnityTool" {
		t.Errorf("expected ToolName 'UnityTool', got %q", msg.ToolName)
	}
	if msg.EventType != "usage" {
		t.Errorf("expected EventType 'usage', got %q", msg.EventType)
	}
	if msg.SessionID == nil || *msg.SessionID != "abc123" {
		t.Errorf("expected SessionID 'abc123', got %v", msg.SessionID)
	}
	if msg.ToolVersion == nil || *msg.ToolVersion != "1.0.0" {
		t.Errorf("expected ToolVersion '1.0.0', got %v", msg.ToolVersion)
	}
	if msg.Details == nil {
		t.Error("expected Details to be non-nil")
	}
}

func TestLogMessage_JSONUnmarshal_OptionalFieldsOmitted(t *testing.T) {
	input := `{
		"tool_name": "MayaTool",
		"event_type": "error",
		"timestamp": "2026-04-04T10:00:00Z",
		"message": "something failed"
	}`

	var msg LogMessage
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg.SessionID != nil {
		t.Errorf("expected SessionID nil, got %v", msg.SessionID)
	}
	if msg.ToolVersion != nil {
		t.Errorf("expected ToolVersion nil, got %v", msg.ToolVersion)
	}
	if msg.Details != nil {
		t.Errorf("expected Details nil, got %v", msg.Details)
	}
}
