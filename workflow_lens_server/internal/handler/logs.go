package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/kaido-atsuya/workflow_lens_server/internal/model"
	"github.com/kaido-atsuya/workflow_lens_server/internal/store"
)

// maxBodySize はリクエストボディの上限（1MiB）。
const maxBodySize = 1 << 20

var (
	tracer = otel.Tracer("handler")
	meter  = otel.Meter("handler")
)

// HandlePostLogs は POST /logs のハンドラ。
// JSON配列を受け取り、バリデーション後にstoreへ渡す。
// アプリケーションメトリクス（イベント数、セッション数）もOTelで記録する。
func HandlePostLogs(s store.Store) http.HandlerFunc {
	// アプリケーションメトリクス
	eventsTotal, _ := meter.Int64Counter("app.events.total",
		metric.WithDescription("受信イベント総数"))
	sessionsStarted, _ := meter.Int64Counter("app.sessions.started",
		metric.WithDescription("開始されたセッション数"))
	sessionsEnded, _ := meter.Int64Counter("app.sessions.ended",
		metric.WithDescription("終了したセッション数"))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// ボディサイズ制限
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

		// リクエストパース
		var logs []model.LogMessage
		func() {
			_, span := tracer.Start(ctx, "handler.parse_request")
			defer span.End()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				if err.Error() == "http: request body too large" {
					span.SetStatus(codes.Error, "request body too large")
					writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
					return
				}
				span.SetStatus(codes.Error, "failed to read request body")
				writeError(w, http.StatusBadRequest, "failed to read request body")
				return
			}

			if err := json.Unmarshal(body, &logs); err != nil {
				span.SetStatus(codes.Error, "invalid JSON")
				writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
				return
			}

			if err := model.ValidateLogs(logs); err != nil {
				span.SetStatus(codes.Error, "validation error")
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}

			span.SetAttributes(attribute.Int("logs.count", len(logs)))
		}()

		if logs == nil {
			return
		}

		// DB挿入
		_, insertSpan := tracer.Start(ctx, "handler.insert_logs",
			trace.WithAttributes(attribute.Int("logs.count", len(logs))))
		inserted, err := s.InsertLogs(ctx, logs)
		if err != nil {
			insertSpan.SetStatus(codes.Error, "insert failed")
			insertSpan.End()
			slog.Error("failed to insert logs", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		insertSpan.SetAttributes(attribute.Int("logs.inserted", inserted))
		insertSpan.End()

		// アプリケーションメトリクス記録
		for _, log := range logs {
			toolVersion := "unknown"
			if log.ToolVersion != nil {
				toolVersion = *log.ToolVersion
			}

			eventsTotal.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("tool_name", log.ToolName),
					attribute.String("event_type", log.EventType),
				))

			switch log.EventType {
			case "session_start":
				sessionsStarted.Add(ctx, 1,
					metric.WithAttributes(
						attribute.String("tool_name", log.ToolName),
						attribute.String("tool_version", toolVersion),
					))
			case "session_end":
				sessionsEnded.Add(ctx, 1,
					metric.WithAttributes(
						attribute.String("tool_name", log.ToolName),
					))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"inserted": inserted})
	}
}

// HandleHealth は GET /health のハンドラ。
func HandleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
