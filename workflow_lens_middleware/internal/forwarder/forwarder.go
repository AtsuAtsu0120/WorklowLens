// Package forwarder はバリデーション済みログメッセージをバッファリングし、
// サーバーへHTTP POSTでバッチ転送する。
package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	defaultMaxBatch      = 100
	defaultFlushInterval = 5 * time.Second
	httpTimeout          = 10 * time.Second
)

var tracer = otel.Tracer("forwarder")

// Forwarder はログメッセージをバッファリングし、サーバーへバッチ転送する。
type Forwarder struct {
	serverURL     string
	client        *http.Client
	mu            sync.Mutex
	buffer        []json.RawMessage
	maxBatch      int
	flushInterval time.Duration
}

// New は新しいForwarderを生成する。
func New(serverURL string) *Forwarder {
	return &Forwarder{
		serverURL:     serverURL,
		client:        &http.Client{Timeout: httpTimeout},
		buffer:        make([]json.RawMessage, 0, defaultMaxBatch),
		maxBatch:      defaultMaxBatch,
		flushInterval: defaultFlushInterval,
	}
}

// Add はバリデーション済みメッセージ（生JSON）をバッファに追加する。
// バッファがmaxBatchに達した場合は即座にフラッシュする。
func (f *Forwarder) Add(ctx context.Context, raw []byte) {
	f.mu.Lock()
	// rawのコピーを保持（元バッファは再利用される可能性がある）
	cp := make(json.RawMessage, len(raw))
	copy(cp, raw)
	f.buffer = append(f.buffer, cp)

	if len(f.buffer) >= f.maxBatch {
		batch := f.buffer
		f.buffer = make([]json.RawMessage, 0, f.maxBatch)
		f.mu.Unlock()
		f.send(ctx, batch)
		return
	}
	f.mu.Unlock()
}

// Run は定期フラッシュgoroutineを開始する。ctxがキャンセルされると終了する。
func (f *Forwarder) Run(ctx context.Context) {
	ticker := time.NewTicker(f.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f.Flush(ctx)
		}
	}
}

// Flush はバッファ内のメッセージをサーバーへ転送する。
func (f *Forwarder) Flush(ctx context.Context) {
	f.mu.Lock()
	if len(f.buffer) == 0 {
		f.mu.Unlock()
		return
	}
	batch := f.buffer
	f.buffer = make([]json.RawMessage, 0, f.maxBatch)
	f.mu.Unlock()

	f.send(ctx, batch)
}

// send はバッチをHTTP POSTでサーバーへ送信する。
func (f *Forwarder) send(ctx context.Context, batch []json.RawMessage) {
	ctx, span := tracer.Start(ctx, "middleware.forward_http",
		)
	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.url", f.serverURL+"/logs"),
		attribute.Int("batch.size", len(batch)),
	)
	defer span.End()

	body, err := json.Marshal(batch)
	if err != nil {
		slog.Warn("バッチJSON生成エラー", "error", err)
		span.SetStatus(codes.Error, "marshal failed")
		span.RecordError(err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.serverURL+"/logs", bytes.NewReader(body))
	if err != nil {
		slog.Warn("HTTPリクエスト生成エラー", "error", err)
		span.SetStatus(codes.Error, "request creation failed")
		span.RecordError(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// W3C Trace Context ヘッダーを注入
	otel.GetTextMapPropagator().Inject(ctx, propagationCarrier(req.Header))

	resp, err := f.client.Do(req)
	if err != nil {
		slog.Warn("HTTP転送エラー", "error", err, "url", f.serverURL, "batch_size", len(batch))
		span.SetStatus(codes.Error, "http request failed")
		span.RecordError(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Warn("HTTP転送エラーレスポンス", "status", resp.StatusCode, "batch_size", len(batch))
		span.SetStatus(codes.Error, "http error response")
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		return
	}

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	slog.Debug("バッチ転送完了", "batch_size", len(batch), "status", resp.StatusCode)
}

// propagationCarrier はhttp.HeaderをOTel TextMapCarrierとして使うアダプタ。
type propagationCarrier http.Header

func (c propagationCarrier) Get(key string) string {
	return http.Header(c).Get(key)
}

func (c propagationCarrier) Set(key, value string) {
	http.Header(c).Set(key, value)
}

func (c propagationCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
