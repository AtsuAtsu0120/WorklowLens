package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/kaido-atsuya/workflow_lens_server/internal/model"
)

func newTestStore(t *testing.T) *SQLStore {
	t.Helper()
	ctx := context.Background()
	s, err := NewSQLStore(ctx, "sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewSQLStore failed: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestNewSQLStore_CreatesTable(t *testing.T) {
	s := newTestStore(t)

	// テーブルが存在することを確認
	var name string
	err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='logs'").Scan(&name)
	if err != nil {
		t.Fatalf("logs table not found: %v", err)
	}
	if name != "logs" {
		t.Errorf("expected table 'logs', got %q", name)
	}
}

func TestNewSQLStore_UnsupportedDriver(t *testing.T) {
	ctx := context.Background()
	_, err := NewSQLStore(ctx, "unknown_driver", "test.db")
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestInsertLogs_Empty(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	count, err := s.InsertLogs(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestInsertLogs_SingleLog(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ts := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	logs := []model.LogMessage{
		{
			ToolName:  "TestTool",
			Category:  "edit",
			Timestamp: ts,
			Action:    "hello",
		},
	}

	count, err := s.InsertLogs(ctx, logs)
	if err != nil {
		t.Fatalf("InsertLogs failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	// DBから取得して検証
	var toolName, category, action string
	err = s.db.QueryRow("SELECT tool_name, category, action FROM logs").Scan(&toolName, &category, &action)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if toolName != "TestTool" || category != "edit" || action != "hello" {
		t.Errorf("unexpected values: tool_name=%q category=%q action=%q", toolName, category, action)
	}
}

func TestInsertLogs_WithOptionalFields(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sessionID := "abc123"
	toolVersion := "1.0.0"
	userID := "user-42"
	var durationMs int64 = 1500
	ts := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)

	logs := []model.LogMessage{
		{
			ToolName:    "TestTool",
			Category:    "session",
			Timestamp:   ts,
			Action:      "start",
			SessionID:   &sessionID,
			ToolVersion: &toolVersion,
			UserID:      &userID,
			DurationMs:  &durationMs,
		},
	}

	count, err := s.InsertLogs(ctx, logs)
	if err != nil {
		t.Fatalf("InsertLogs failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	var sid, ver, uid sql.NullString
	var dur sql.NullInt64
	err = s.db.QueryRow("SELECT session_id, tool_version, user_id, duration_ms FROM logs").Scan(&sid, &ver, &uid, &dur)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if !sid.Valid || sid.String != "abc123" {
		t.Errorf("expected session_id 'abc123', got %v", sid)
	}
	if !ver.Valid || ver.String != "1.0.0" {
		t.Errorf("expected tool_version '1.0.0', got %v", ver)
	}
	if !uid.Valid || uid.String != "user-42" {
		t.Errorf("expected user_id 'user-42', got %v", uid)
	}
	if !dur.Valid || dur.Int64 != 1500 {
		t.Errorf("expected duration_ms 1500, got %v", dur)
	}
}

func TestInsertLogs_NilOptionalFields(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ts := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	logs := []model.LogMessage{
		{
			ToolName:  "TestTool",
			Category:  "edit",
			Timestamp: ts,
			Action:    "hello",
		},
	}

	_, err := s.InsertLogs(ctx, logs)
	if err != nil {
		t.Fatalf("InsertLogs failed: %v", err)
	}

	var sid, ver, uid sql.NullString
	var dur sql.NullInt64
	err = s.db.QueryRow("SELECT session_id, tool_version, user_id, duration_ms FROM logs").Scan(&sid, &ver, &uid, &dur)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if sid.Valid {
		t.Error("expected session_id to be NULL")
	}
	if ver.Valid {
		t.Error("expected tool_version to be NULL")
	}
	if uid.Valid {
		t.Error("expected user_id to be NULL")
	}
	if dur.Valid {
		t.Error("expected duration_ms to be NULL")
	}
}

func TestInsertLogs_MultipleLogs(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ts := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	logs := make([]model.LogMessage, 5)
	for i := range logs {
		logs[i] = model.LogMessage{
			ToolName:  "TestTool",
			Category:  "edit",
			Timestamp: ts,
			Action:    "msg",
		}
	}

	count, err := s.InsertLogs(ctx, logs)
	if err != nil {
		t.Fatalf("InsertLogs failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}

	var total int
	s.db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&total)
	if total != 5 {
		t.Errorf("expected 5 rows in DB, got %d", total)
	}
}
