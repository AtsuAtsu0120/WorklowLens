package server

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// startTestServer はテスト用UDPサーバーを起動し、アドレスとキャンセル関数を返す。
func startTestServer(t *testing.T) (addr string, cancel context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	// ポート0でOS自動割り当て
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr = conn.LocalAddr().String()
	conn.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, addr, nil)
	}()

	// サーバーが起動するのを少し待つ
	time.Sleep(50 * time.Millisecond)

	t.Cleanup(func() {
		cancel()
		// サーバーが終了するのを待つ
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("server error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop in time")
		}
	})

	return addr, cancel
}

// sendDatagram はテスト用にUDPデータグラムを送信する。
func sendDatagram(t *testing.T, addr string, data []byte) {
	t.Helper()
	conn, err := net.Dial("udp", addr)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	if _, err := conn.Write(data); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
}

func makeCategoryJSON(toolName, category, action string) string {
	return fmt.Sprintf(`{"tool_name":%q,"category":%q,"action":%q,"timestamp":"2026-04-04T10:00:00Z"}`, toolName, category, action)
}

func makeSessionCategoryJSON(toolName, action, sessionID string) string {
	return fmt.Sprintf(`{"tool_name":%q,"category":"session","action":%q,"timestamp":"2026-04-04T10:00:00Z","session_id":%q}`, toolName, action, sessionID)
}

func TestServer_ValidDatagram(t *testing.T) {
	addr, _ := startTestServer(t)
	sendDatagram(t, addr, []byte(makeCategoryJSON("TestTool", "edit", "brush_apply")))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_MultipleMessages(t *testing.T) {
	addr, _ := startTestServer(t)
	for i := range 5 {
		sendDatagram(t, addr, []byte(makeCategoryJSON("TestTool", "edit", fmt.Sprintf("action_%d", i))))
	}
	time.Sleep(100 * time.Millisecond)
}

func TestServer_InvalidJSON(t *testing.T) {
	addr, _ := startTestServer(t)

	sendDatagram(t, addr, []byte(`{invalid json}`))
	time.Sleep(50 * time.Millisecond)

	sendDatagram(t, addr, []byte(makeCategoryJSON("TestTool", "edit", "after_invalid")))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_EmptyDatagram(t *testing.T) {
	addr, _ := startTestServer(t)
	sendDatagram(t, addr, []byte(""))
	sendDatagram(t, addr, []byte("  \n\t  "))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_OptionalFields(t *testing.T) {
	addr, _ := startTestServer(t)
	data := `{"tool_name":"TestTool","category":"edit","action":"brush_apply","timestamp":"2026-04-04T10:00:00Z","user_id":"tanaka","duration_ms":120}`
	sendDatagram(t, addr, []byte(data))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_MultipleClients(t *testing.T) {
	addr, _ := startTestServer(t)

	for i := range 3 {
		toolName := fmt.Sprintf("Tool%d", i)
		sendDatagram(t, addr, []byte(makeCategoryJSON(toolName, "edit", "concurrent")))
	}
	time.Sleep(100 * time.Millisecond)
}

func TestServer_SessionLifecycle(t *testing.T) {
	addr, _ := startTestServer(t)
	sessionID := "test-session-001"

	sendDatagram(t, addr, []byte(makeSessionCategoryJSON("TestTool", "start", sessionID)))
	time.Sleep(20 * time.Millisecond)
	sendDatagram(t, addr, []byte(makeCategoryJSON("TestTool", "edit", "brush_apply")))
	time.Sleep(20 * time.Millisecond)
	sendDatagram(t, addr, []byte(makeSessionCategoryJSON("TestTool", "end", sessionID)))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_ErrorCategory(t *testing.T) {
	addr, _ := startTestServer(t)
	sendDatagram(t, addr, []byte(makeCategoryJSON("TestTool", "error", "shader_compile")))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := conn.LocalAddr().String()
	conn.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, addr, nil)
	}()
	time.Sleep(50 * time.Millisecond)

	sendDatagram(t, addr, []byte(makeCategoryJSON("TestTool", "edit", "before_shutdown")))
	time.Sleep(50 * time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServer_InvalidUTF8(t *testing.T) {
	addr, _ := startTestServer(t)
	sendDatagram(t, addr, []byte{0xff, 0xfe, 0xfd})
	time.Sleep(50 * time.Millisecond)

	sendDatagram(t, addr, []byte(makeCategoryJSON("TestTool", "edit", "after_invalid_utf8")))
	time.Sleep(50 * time.Millisecond)
}
