package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kaido-atsuya/workflow_lens_server/internal/model"
)

// mockStore はテスト用のStore実装。
type mockStore struct {
	insertFunc func(ctx context.Context, logs []model.LogMessage) (int, error)
}

func (m *mockStore) InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error) {
	return m.insertFunc(ctx, logs)
}

func (m *mockStore) Close() error {
	return nil
}

func newMockStore(inserted int, err error) *mockStore {
	return &mockStore{
		insertFunc: func(_ context.Context, logs []model.LogMessage) (int, error) {
			if err != nil {
				return 0, err
			}
			return inserted, nil
		},
	}
}

func TestHandlePostLogs_Valid(t *testing.T) {
	body := `[{"tool_name":"TestTool","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"hello"}]`
	req := httptest.NewRequest("POST", "/logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	store := newMockStore(1, nil)
	HandlePostLogs(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]int
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["inserted"] != 1 {
		t.Errorf("expected inserted=1, got %d", resp["inserted"])
	}
}

func TestHandlePostLogs_EmptyArray(t *testing.T) {
	req := httptest.NewRequest("POST", "/logs", strings.NewReader("[]"))
	rec := httptest.NewRecorder()

	store := newMockStore(0, nil)
	HandlePostLogs(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandlePostLogs_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/logs", strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	store := newMockStore(0, nil)
	HandlePostLogs(store)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandlePostLogs_MissingField(t *testing.T) {
	body := `[{"event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"hello"}]`
	req := httptest.NewRequest("POST", "/logs", strings.NewReader(body))
	rec := httptest.NewRecorder()

	store := newMockStore(0, nil)
	HandlePostLogs(store)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp["error"], "tool_name") {
		t.Errorf("expected error about tool_name, got %q", resp["error"])
	}
}

func TestHandlePostLogs_InvalidEventType(t *testing.T) {
	body := `[{"tool_name":"TestTool","event_type":"unknown","timestamp":"2026-04-04T10:00:00Z","message":"hello"}]`
	req := httptest.NewRequest("POST", "/logs", strings.NewReader(body))
	rec := httptest.NewRecorder()

	store := newMockStore(0, nil)
	HandlePostLogs(store)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandlePostLogs_StoreError(t *testing.T) {
	body := `[{"tool_name":"TestTool","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"hello"}]`
	req := httptest.NewRequest("POST", "/logs", strings.NewReader(body))
	rec := httptest.NewRecorder()

	store := newMockStore(0, errors.New("db error"))
	HandlePostLogs(store)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestHandlePostLogs_BodyTooLarge(t *testing.T) {
	// 1MiB + 1バイトのボディを生成
	largeBody := bytes.Repeat([]byte("a"), maxBodySize+1)
	req := httptest.NewRequest("POST", "/logs", bytes.NewReader(largeBody))
	rec := httptest.NewRecorder()

	store := newMockStore(0, nil)
	HandlePostLogs(store)(rec, req)

	// 413 or 400 — MaxBytesReaderのエラー
	if rec.Code != http.StatusRequestEntityTooLarge && rec.Code != http.StatusBadRequest {
		t.Errorf("expected 413 or 400, got %d", rec.Code)
	}
}

func TestHandlePostLogs_WithOptionalFields(t *testing.T) {
	body := `[{
		"tool_name":"TestTool",
		"event_type":"session_start",
		"timestamp":"2026-04-04T10:00:00Z",
		"message":"started",
		"session_id":"abc123",
		"tool_version":"1.0.0",
		"details":{"feature":"paint"}
	}]`
	req := httptest.NewRequest("POST", "/logs", strings.NewReader(body))
	rec := httptest.NewRecorder()

	// InsertLogsに渡されたデータを検証
	var captured []model.LogMessage
	store := &mockStore{
		insertFunc: func(_ context.Context, logs []model.LogMessage) (int, error) {
			captured = logs
			return len(logs), nil
		},
	}
	HandlePostLogs(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(captured) != 1 {
		t.Fatalf("expected 1 log, got %d", len(captured))
	}
	if captured[0].SessionID == nil || *captured[0].SessionID != "abc123" {
		t.Errorf("expected session_id 'abc123'")
	}
	if captured[0].ToolVersion == nil || *captured[0].ToolVersion != "1.0.0" {
		t.Errorf("expected tool_version '1.0.0'")
	}
	if captured[0].Details == nil {
		t.Error("expected details to be non-nil")
	}
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	HandleHealth()(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected 'ok', got %q", rec.Body.String())
	}
}
