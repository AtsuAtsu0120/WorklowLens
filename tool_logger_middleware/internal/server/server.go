package server

import (
	"context"
	"log/slog"
	"net"
	"unicode/utf8"

	"github.com/kaido-atsuya/tool_logger_middleware/internal/model"
)

// MaxDatagramSize はデータグラムの最大サイズ（64 KiB）。
const MaxDatagramSize = 64 * 1024

// Run はUDPサーバーを起動し、データグラムを受信する。
// ctxがキャンセルされるとソケットを閉じて終了する。
func Run(ctx context.Context, addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	// contextキャンセル時にソケットを閉じる
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	slog.Info("UDPサーバー起動", "addr", conn.LocalAddr().String())

	buf := make([]byte, MaxDatagramSize)
	for {
		n, src, err := conn.ReadFrom(buf)
		if err != nil {
			// contextキャンセルによるクローズ
			if ctx.Err() != nil {
				slog.Info("サーバー停止")
				return nil
			}
			return err
		}

		data := make([]byte, n)
		copy(data, buf[:n])
		processDatagram(data, src)
	}
}

// processDatagram は1つのデータグラムを処理する。
func processDatagram(data []byte, src net.Addr) {
	// UTF-8バリデーション
	if !utf8.Valid(data) {
		slog.Warn("不正なUTF-8���ータグラム", "src", src.String())
		return
	}

	// 空データグラムは無視
	trimmed := trimBytes(data)
	if len(trimmed) == 0 {
		return
	}

	// JSONパース
	msg, err := model.Parse(trimmed)
	if err != nil {
		slog.Warn("JSONパースエラー", "error", err, "src", src.String())
		return
	}

	slog.Info("メッセージ受信",
		"tool_name", msg.ToolName,
		"event_type", msg.EventType,
		"timestamp", msg.Timestamp,
		"message", msg.Message,
		"src", src.String(),
	)
}

// trimBytes はバイト列の前後の空白を除去する。
func trimBytes(data []byte) []byte {
	start := 0
	for start < len(data) && isWhitespace(data[start]) {
		start++
	}
	end := len(data)
	for end > start && isWhitespace(data[end-1]) {
		end--
	}
	return data[start:end]
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
