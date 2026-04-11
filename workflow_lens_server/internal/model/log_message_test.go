package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidateCategory(t *testing.T) {
	valid := []string{"asset", "build", "edit", "error", "session"}
	for _, c := range valid {
		if !ValidateCategory(c) {
			t.Errorf("expected %q to be valid", c)
		}
	}

	invalid := []string{"unknown", "usage", "", "Asset", "SESSION"}
	for _, c := range invalid {
		if ValidateCategory(c) {
			t.Errorf("expected %q to be invalid", c)
		}
	}
}

func TestValidateLogs_Valid(t *testing.T) {
	logs := []LogMessage{
		{
			ToolName:  "TestTool",
			Category:  "edit",
			Action:    "brush_apply",
			Timestamp: time.Now(),
		},
		{
			ToolName:  "TestTool",
			Category:  "session",
			Action:    "start",
			Timestamp: time.Now(),
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
		{Category: "edit", Action: "brush_apply", Timestamp: time.Now()},
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

func TestValidateLogs_MissingCategory(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", Action: "brush_apply", Timestamp: time.Now()},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestValidateLogs_InvalidCategory(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", Category: "unknown", Action: "brush_apply", Timestamp: time.Now()},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
	expected := "invalid category: 'unknown' at index 0"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidateLogs_MissingAction(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", Category: "edit", Timestamp: time.Now()},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for missing action")
	}
}

func TestValidateLogs_MissingTimestamp(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", Category: "edit", Action: "brush_apply"},
	}
	err := ValidateLogs(logs)
	if err == nil {
		t.Fatal("expected error for missing timestamp")
	}
}

func TestValidateLogs_ErrorAtSecondIndex(t *testing.T) {
	logs := []LogMessage{
		{ToolName: "TestTool", Category: "edit", Action: "brush_apply", Timestamp: time.Now()},
		{ToolName: "", Category: "edit", Action: "brush_apply", Timestamp: time.Now()},
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
		"category": "edit",
		"action": "brush_apply",
		"timestamp": "2026-04-04T10:00:00Z",
		"session_id": "abc123",
		"tool_version": "1.0.0",
		"user_id": "tanaka",
		"duration_ms": 120
	}`

	var msg LogMessage
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg.ToolName != "UnityTool" {
		t.Errorf("expected ToolName 'UnityTool', got %q", msg.ToolName)
	}
	if msg.Category != "edit" {
		t.Errorf("expected Category 'edit', got %q", msg.Category)
	}
	if msg.Action != "brush_apply" {
		t.Errorf("expected Action 'brush_apply', got %q", msg.Action)
	}
	if msg.SessionID == nil || *msg.SessionID != "abc123" {
		t.Errorf("expected SessionID 'abc123', got %v", msg.SessionID)
	}
	if msg.ToolVersion == nil || *msg.ToolVersion != "1.0.0" {
		t.Errorf("expected ToolVersion '1.0.0', got %v", msg.ToolVersion)
	}
	if msg.UserID == nil || *msg.UserID != "tanaka" {
		t.Errorf("expected UserID 'tanaka', got %v", msg.UserID)
	}
	if msg.DurationMs == nil || *msg.DurationMs != 120 {
		t.Errorf("expected DurationMs 120, got %v", msg.DurationMs)
	}
}

func TestLogMessage_JSONUnmarshal_OptionalFieldsOmitted(t *testing.T) {
	input := `{
		"tool_name": "MayaTool",
		"category": "error",
		"action": "shader_compile",
		"timestamp": "2026-04-04T10:00:00Z"
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
	if msg.UserID != nil {
		t.Errorf("expected UserID nil, got %v", msg.UserID)
	}
	if msg.DurationMs != nil {
		t.Errorf("expected DurationMs nil, got %v", msg.DurationMs)
	}
}
