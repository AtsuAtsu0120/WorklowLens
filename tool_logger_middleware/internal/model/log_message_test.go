package model

import (
	"encoding/json"
	"testing"
	"time"
)

// テスト用ヘルパー: 基本的な有効JSONを生成する。
func makeValidJSON() string {
	return `{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"Terrain brush applied"}`
}

func TestParse_ValidJSON(t *testing.T) {
	msg, err := Parse([]byte(makeValidJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ToolName != "UnityTerrainEditor" {
		t.Errorf("ToolName = %q, want %q", msg.ToolName, "UnityTerrainEditor")
	}
	if msg.EventType != "usage" {
		t.Errorf("EventType = %q, want %q", msg.EventType, "usage")
	}
	if msg.Message != "Terrain brush applied" {
		t.Errorf("Message = %q, want %q", msg.Message, "Terrain brush applied")
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestParse_ErrorEventType(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"error","timestamp":"2026-04-04T10:00:00Z","message":"NullReferenceException"}`
	msg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.EventType != "error" {
		t.Errorf("EventType = %q, want %q", msg.EventType, "error")
	}
}

func TestParse_SessionStartEventType(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"session_start","timestamp":"2026-04-04T10:00:00Z","message":"Tool opened"}`
	msg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.EventType != "session_start" {
		t.Errorf("EventType = %q, want %q", msg.EventType, "session_start")
	}
}

func TestParse_SessionEndEventType(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"session_end","timestamp":"2026-04-04T10:00:00Z","message":"Tool closed"}`
	msg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.EventType != "session_end" {
		t.Errorf("EventType = %q, want %q", msg.EventType, "session_end")
	}
}

func TestParse_CancellationEventType(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"cancellation","timestamp":"2026-04-04T10:00:00Z","message":"Operation cancelled"}`
	msg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.EventType != "cancellation" {
		t.Errorf("EventType = %q, want %q", msg.EventType, "cancellation")
	}
}

func TestParse_DetailsOptional_Omitted(t *testing.T) {
	msg, err := Parse([]byte(makeValidJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Details != nil {
		t.Errorf("Details should be nil when omitted, got %s", string(msg.Details))
	}
}

func TestParse_DetailsOptional_Null(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"test","details":null}`
	msg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// json.RawMessage with null contains the bytes "null"
	if msg.Details != nil && string(msg.Details) != "null" {
		t.Errorf("Details = %s, want nil or null", string(msg.Details))
	}
}

func TestParse_DetailsWithNestedObject(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"test","details":{"brush_size":5,"nested":{"key":"value"}}}`
	msg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Details == nil {
		t.Fatal("Details should not be nil")
	}
	var details map[string]interface{}
	if err := json.Unmarshal(msg.Details, &details); err != nil {
		t.Fatalf("failed to unmarshal details: %v", err)
	}
	if details["brush_size"] != float64(5) {
		t.Errorf("brush_size = %v, want 5", details["brush_size"])
	}
}

func TestParse_WithSessionIDAndToolVersion(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"test","session_id":"ute-a1b2c3d4","tool_version":"2.1.0"}`
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
}

func TestParse_SessionIDAndToolVersion_Omitted(t *testing.T) {
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
}

func TestParse_InvalidEventType(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"unknown","timestamp":"2026-04-04T10:00:00Z","message":"test"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for invalid event_type")
	}
}

func TestParse_MissingToolName(t *testing.T) {
	data := `{"event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"test"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing tool_name")
	}
}

func TestParse_MissingEventType(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","timestamp":"2026-04-04T10:00:00Z","message":"test"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing event_type")
	}
}

func TestParse_MissingTimestamp(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"usage","message":"test"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing timestamp")
	}
}

func TestParse_MissingMessage(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"2026-04-04T10:00:00Z"}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing message")
	}
}

func TestParse_InvalidTimestamp(t *testing.T) {
	data := `{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"not-a-date","message":"test"}`
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

func TestSerializeRoundTrip(t *testing.T) {
	sessionID := "ute-a1b2c3d4"
	toolVersion := "2.1.0"
	original := LogMessage{
		ToolName:    "UnityTerrainEditor",
		EventType:   "usage",
		Timestamp:   time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
		Message:     "Terrain brush applied",
		SessionID:   &sessionID,
		ToolVersion: &toolVersion,
		Details:     json.RawMessage(`{"brush_size":5}`),
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
	if parsed.EventType != original.EventType {
		t.Errorf("EventType = %q, want %q", parsed.EventType, original.EventType)
	}
	if !parsed.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", parsed.Timestamp, original.Timestamp)
	}
	if parsed.Message != original.Message {
		t.Errorf("Message = %q, want %q", parsed.Message, original.Message)
	}
	if *parsed.SessionID != *original.SessionID {
		t.Errorf("SessionID = %q, want %q", *parsed.SessionID, *original.SessionID)
	}
	if *parsed.ToolVersion != *original.ToolVersion {
		t.Errorf("ToolVersion = %q, want %q", *parsed.ToolVersion, *original.ToolVersion)
	}
}

func TestParse_AllEventTypes_SnakeCase(t *testing.T) {
	eventTypes := []string{"usage", "error", "session_start", "session_end", "cancellation"}
	for _, et := range eventTypes {
		data := `{"tool_name":"Test","event_type":"` + et + `","timestamp":"2026-04-04T10:00:00Z","message":"test"}`
		msg, err := Parse([]byte(data))
		if err != nil {
			t.Errorf("event_type %q: unexpected error: %v", et, err)
			continue
		}
		if msg.EventType != et {
			t.Errorf("EventType = %q, want %q", msg.EventType, et)
		}
	}
}

func TestValidateEventType(t *testing.T) {
	valid := []string{"usage", "error", "session_start", "session_end", "cancellation"}
	for _, et := range valid {
		if !ValidateEventType(et) {
			t.Errorf("ValidateEventType(%q) = false, want true", et)
		}
	}

	invalid := []string{"", "unknown", "Usage", "ERROR", "session-start"}
	for _, et := range invalid {
		if ValidateEventType(et) {
			t.Errorf("ValidateEventType(%q) = true, want false", et)
		}
	}
}

func TestParse_SerializeOptionalFields(t *testing.T) {
	// session_id/tool_versionなしでシリアライズした場合、JSONにフィールドが含まれないことを確認
	msg := LogMessage{
		ToolName:  "Test",
		EventType: "usage",
		Timestamp: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
		Message:   "test",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, ok := raw["session_id"]; ok {
		t.Error("session_id should not be present when nil")
	}
	if _, ok := raw["tool_version"]; ok {
		t.Error("tool_version should not be present when nil")
	}
	if _, ok := raw["details"]; ok {
		t.Error("details should not be present when nil")
	}
}
