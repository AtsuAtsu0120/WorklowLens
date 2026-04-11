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

func makeUsageJSON(toolName, message string) string {
	return fmt.Sprintf(`{"tool_name":%q,"event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":%q}`, toolName, message)
}

func makeErrorJSON(toolName, message string) string {
	return fmt.Sprintf(`{"tool_name":%q,"event_type":"error","timestamp":"2026-04-04T10:00:00Z","message":%q}`, toolName, message)
}

func makeSessionJSON(toolName, eventType, message, sessionID string) string {
	return fmt.Sprintf(`{"tool_name":%q,"event_type":%q,"timestamp":"2026-04-04T10:00:00Z","message":%q,"session_id":%q}`, toolName, eventType, message, sessionID)
}

func TestServer_ValidDatagram(t *testing.T) {
	addr, _ := startTestServer(t)
	sendDatagram(t, addr, []byte(makeUsageJSON("TestTool", "test message")))
	time.Sleep(50 * time.Millisecond)
	// サーバーがパニックしなければ成功
}

func TestServer_MultipleMessages(t *testing.T) {
	addr, _ := startTestServer(t)
	for i := range 5 {
		sendDatagram(t, addr, []byte(makeUsageJSON("TestTool", fmt.Sprintf("message %d", i))))
	}
	time.Sleep(100 * time.Millisecond)
}

func TestServer_InvalidJSON(t *testing.T) {
	addr, _ := startTestServer(t)

	// 不正なJSONを送信
	sendDatagram(t, addr, []byte(`{invalid json}`))
	time.Sleep(50 * time.Millisecond)

	// その後の有効なメッセージも処理できる
	sendDatagram(t, addr, []byte(makeUsageJSON("TestTool", "after invalid")))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_EmptyDatagram(t *testing.T) {
	addr, _ := startTestServer(t)
	sendDatagram(t, addr, []byte(""))
	sendDatagram(t, addr, []byte("  \n\t  "))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_DetailsPayload(t *testing.T) {
	addr, _ := startTestServer(t)
	data := `{"tool_name":"TestTool","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"test","details":{"brush_size":5,"color":"red"}}`
	sendDatagram(t, addr, []byte(data))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_MultipleClients(t *testing.T) {
	addr, _ := startTestServer(t)

	// 複数クライアントから送信
	for i := range 3 {
		toolName := fmt.Sprintf("Tool%d", i)
		sendDatagram(t, addr, []byte(makeUsageJSON(toolName, "concurrent message")))
	}
	time.Sleep(100 * time.Millisecond)
}

func TestServer_SessionLifecycle(t *testing.T) {
	addr, _ := startTestServer(t)
	sessionID := "test-session-001"

	sendDatagram(t, addr, []byte(makeSessionJSON("TestTool", "session_start", "Tool opened", sessionID)))
	time.Sleep(20 * time.Millisecond)
	sendDatagram(t, addr, []byte(makeSessionJSON("TestTool", "usage", "Feature used", sessionID)))
	time.Sleep(20 * time.Millisecond)
	sendDatagram(t, addr, []byte(makeSessionJSON("TestTool", "session_end", "Tool closed", sessionID)))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_CancellationEvent(t *testing.T) {
	addr, _ := startTestServer(t)
	data := `{"tool_name":"TestTool","event_type":"cancellation","timestamp":"2026-04-04T10:00:00Z","message":"Operation cancelled"}`
	sendDatagram(t, addr, []byte(data))
	time.Sleep(50 * time.Millisecond)
}

func TestServer_ErrorEvent(t *testing.T) {
	addr, _ := startTestServer(t)
	sendDatagram(t, addr, []byte(makeErrorJSON("TestTool", "NullReferenceException")))
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

	// メッセージ送信後にシャットダウン
	sendDatagram(t, addr, []byte(makeUsageJSON("TestTool", "before shutdown")))
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
	// 不正なUTF-8バイト列を送信
	sendDatagram(t, addr, []byte{0xff, 0xfe, 0xfd})
	time.Sleep(50 * time.Millisecond)

	// その後の有効なメッセージも処理できる
	sendDatagram(t, addr, []byte(makeUsageJSON("TestTool", "after invalid utf8")))
	time.Sleep(50 * time.Millisecond)
}
