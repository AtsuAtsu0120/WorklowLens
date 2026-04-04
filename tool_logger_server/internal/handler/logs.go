package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/kaido-atsuya/tool_logger_server/internal/model"
	"github.com/kaido-atsuya/tool_logger_server/internal/store"
)

// maxBodySize はリクエストボディの上限（1MiB）。
const maxBodySize = 1 << 20

// HandlePostLogs は POST /logs のハンドラ。
// JSON配列を受け取り、バリデーション後にstoreへ渡す。
func HandlePostLogs(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ボディサイズ制限
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			if err.Error() == "http: request body too large" {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		var logs []model.LogMessage
		if err := json.Unmarshal(body, &logs); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if err := model.ValidateLogs(logs); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		inserted, err := s.InsertLogs(r.Context(), logs)
		if err != nil {
			slog.Error("failed to insert logs", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
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
