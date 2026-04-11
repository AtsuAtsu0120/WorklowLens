package model

import (
	"encoding/json"
	"testing"
	"time"
)

// テスト用ヘルパー: 基本的な有効JSONを生成する。
func makeValidJSON() string {
	return `{"tool_name":"UnityTerrainEditor","category":"edit","action":"brush_apply","timestamp":"2026-04-04T10:00:00Z"}`
}

func TestParse_ValidJSON(t *testing.T) {
	msg, err := Parse([]byte(makeValidJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ToolName != "UnityTerrainEditor" {
		t.Errorf("ToolName = %q, want %q", msg.ToolName, "UnityTerrainEditor")
	}
	if msg.Category != "edit" {
		t.Errorf("Category = %q, want %q", msg.Category, "edit")
	}
	if msg.Action != "brush_apply" {
		t.Errorf("Action = %q, want %q", msg.Action, "brush_apply")
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestParse_AllCategories(t *testing.T) {
	categories := []string{"asset", "build", "edit", "error", "session"}
	for _, c := range categories {
		data := `{"tool_name":"Test","category":"` + c + `","action":"test","timestamp":"2026-04-04T10:00:00Z"}`
		msg, err := Parse([]byte(data))
		if err != nil {
			t.Errorf("category %q: unexpected error: %v", c, err)
			continue
		}
		if msg.Category != c {
			t.Errorf("Category = %q, want %q", msg.Category, c)
		}
	}
}

func TestParse_WithOptionalFields(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","category":"edit","action":"brush_apply","timestamp":"2026-04-04T10:00:00Z","session_id":"ute-a1b2c3d4","tool_version":"2.1.0","user_id":"tanaka","duration_ms":120}`
	msg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.SessionID == nil || *msg.SessionID != "ute-a1b2c3d4" {
		t.Errorf("SessionID = %v, want %q", msg.SessionID, "ute-a1b2c3d4")
	}
	if msg.ToolVersion == nil || *msg.ToolVersion != "2.1.0" {
		t.Errorf("ToolVersion = %v, want %q", msg.ToolVersion, "2.1.0")
	}
	if msg.UserID == nil || *msg.UserID != "tanaka" {
		t.Errorf("UserID = %v, want %q", msg.UserID, "tanaka")
	}
	if msg.DurationMs == nil || *msg.DurationMs != 120 {
		t.Errorf("DurationMs = %v, want 120", msg.DurationMs)
	}
}

func TestParse_OptionalFields_Omitted(t *testing.T) {
	msg, err := Parse([]byte(makeValidJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.SessionID != nil {
		t.Errorf("SessionID should be nil when omitted")
	}
	if msg.ToolVersion != nil {
		t.Errorf("ToolVersion should be nil when omitted")
	}
	if msg.UserID != nil {
		t.Errorf("UserID should be nil when omitted")
	}
	if msg.DurationMs != nil {
		t.Errorf("DurationMs should be nil when omitted")
	}
}

func TestParse_InvalidCategory(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","category":"unknown","action":"test","timestamp":"2026-04-04T10:00:00Z"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestParse_MissingToolName(t *testing.T) {
	data := `{"category":"edit","action":"test","timestamp":"2026-04-04T10:00:00Z"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing tool_name")
	}
}

func TestParse_MissingCategory(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","action":"test","timestamp":"2026-04-04T10:00:00Z"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestParse_MissingAction(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","category":"edit","timestamp":"2026-04-04T10:00:00Z"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing action")
	}
}

func TestParse_MissingTimestamp(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","category":"edit","action":"test"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing timestamp")
	}
}

func TestParse_InvalidTimestamp(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","category":"edit","action":"test","timestamp":"not-a-date"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse([]byte(`{invalid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParse_EmptyObject(t *testing.T) {
	_, err := Parse([]byte(`{}`))
	if err == nil {
		t.Fatal("expected error for empty JSON object")
	}
}

func TestParse_EmptyString(t *testing.T) {
	_, err := Parse([]byte(``))
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestValidateCategory(t *testing.T) {
	valid := []string{"asset", "build", "edit", "error", "session"}
	for _, c := range valid {
		if !ValidateCategory(c) {
			t.Errorf("ValidateCategory(%q) = false, want true", c)
		}
	}

	invalid := []string{"", "unknown", "Edit", "ERROR", "usage"}
	for _, c := range invalid {
		if ValidateCategory(c) {
			t.Errorf("ValidateCategory(%q) = true, want false", c)
		}
	}
}

func TestSerializeRoundTrip(t *testing.T) {
	sessionID := "ute-a1b2c3d4"
	toolVersion := "2.1.0"
	userID := "tanaka"
	var durationMs int64 = 120
	original := LogMessage{
		ToolName:    "UnityTerrainEditor",
		Category:    "edit",
		Action:      "brush_apply",
		Timestamp:   time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
		SessionID:   &sessionID,
		ToolVersion: &toolVersion,
		UserID:      &userID,
		DurationMs:  &durationMs,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed.ToolName != original.ToolName {
		t.Errorf("ToolName = %q, want %q", parsed.ToolName, original.ToolName)
	}
	if parsed.Category != original.Category {
		t.Errorf("Category = %q, want %q", parsed.Category, original.Category)
	}
	if parsed.Action != original.Action {
		t.Errorf("Action = %q, want %q", parsed.Action, original.Action)
	}
	if !parsed.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", parsed.Timestamp, original.Timestamp)
	}
	if *parsed.SessionID != *original.SessionID {
		t.Errorf("SessionID = %q, want %q", *parsed.SessionID, *original.SessionID)
	}
	if *parsed.ToolVersion != *original.ToolVersion {
		t.Errorf("ToolVersion = %q, want %q", *parsed.ToolVersion, *original.ToolVersion)
	}
	if *parsed.UserID != *original.UserID {
		t.Errorf("UserID = %q, want %q", *parsed.UserID, *original.UserID)
	}
	if *parsed.DurationMs != *original.DurationMs {
		t.Errorf("DurationMs = %d, want %d", *parsed.DurationMs, *original.DurationMs)
	}
}

func TestParse_SerializeOptionalFields(t *testing.T) {
	msg := LogMessage{
		ToolName:  "Test",
		Category:  "edit",
		Action:    "test",
		Timestamp: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	for _, key := range []string{"session_id", "tool_version", "user_id", "duration_ms"} {
		if _, ok := raw[key]; ok {
			t.Errorf("%s should not be present when nil", key)
		}
	}
}
