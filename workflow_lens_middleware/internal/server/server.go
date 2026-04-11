package server

import (
	"context"
	"log/slog"
	"net"
	"unicode/utf8"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/kaido-atsuya/workflow_lens_middleware/internal/forwarder"
	"github.com/kaido-atsuya/workflow_lens_middleware/internal/model"
)

// MaxDatagramSize はデータグラムの最大サイズ（64 KiB）。
const MaxDatagramSize = 64 * 1024

var (
	tracer = otel.Tracer("middleware")
	meter  = otel.Meter("middleware")
)

// Run はUDPサーバーを起動し、データグラムを受信する。
// fwdがnilの場合はログ出力のみ。非nilの場合はバリデーション通過メッセージを転送する。
// ctxがキャンセルされるとソケットを閉じて終了する。
func Run(ctx context.Context, addr string, fwd *forwarder.Forwarder) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	// contextキャンセル時にソケットを閉じる
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// メトリクス初期化
	datagramsReceived, _ := meter.Int64Counter("middleware.datagrams.received",
		metric.WithDescription("受信したデータグラム数"))
	datagramsInvalid, _ := meter.Int64Counter("middleware.datagrams.invalid",
		metric.WithDescription("無効なデータグラム数"))
	datagramSize, _ := meter.Int64Histogram("middleware.datagram.size_bytes",
		metric.WithDescription("データグラムのペイロードサイズ"))

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
		processDatagram(ctx, data, src, fwd, datagramsReceived, datagramsInvalid, datagramSize)
	}
}

// processDatagram は1つのデータグラムを処理する。
func processDatagram(
	ctx context.Context,
	data []byte,
	src net.Addr,
	fwd *forwarder.Forwarder,
	datagramsReceived metric.Int64Counter,
	datagramsInvalid metric.Int64Counter,
	datagramSize metric.Int64Histogram,
) {
	datagramsReceived.Add(ctx, 1)
	datagramSize.Record(ctx, int64(len(data)))

	// UTF-8バリデーション
	if !utf8.Valid(data) {
		slog.Warn("不正なUTF-8データグラム", "src", src.String())
		datagramsInvalid.Add(ctx, 1, metric.WithAttributes(
			attribute.String("reason", "utf8")))
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
		datagramsInvalid.Add(ctx, 1, metric.WithAttributes(
			attribute.String("reason", "json_parse")))
		return
	}

	// traceparentがあれば親コンテキストを抽出
	spanCtx := ctx
	if msg.Traceparent != nil {
		carrier := propagation.MapCarrier{"traceparent": *msg.Traceparent}
		spanCtx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	_, span := tracer.Start(spanCtx, "middleware.process_datagram",
		trace.WithAttributes(
			attribute.String("net.transport", "udp"),
			attribute.String("net.peer.ip", src.String()),
			attribute.String("tool.name", msg.ToolName),
			attribute.String("category", msg.Category),
			attribute.String("action", msg.Action),
			attribute.Int("messaging.message.payload_size_bytes", len(data)),
		),
	)
	defer span.End()

	if msg.SessionID != nil {
		span.SetAttributes(attribute.String("session.id", *msg.SessionID))
	}

	slog.Info("メッセージ受信",
		"tool_name", msg.ToolName,
		"category", msg.Category,
		"action", msg.Action,
		"timestamp", msg.Timestamp,
		"src", src.String(),
	)

	// バリデーション通過メッセージをforwarderに渡す
	if fwd != nil {
		fwd.Add(spanCtx, trimmed)
	}
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
