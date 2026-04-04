package store

import (
	"context"

	"github.com/kaido-atsuya/tool_logger_server/internal/model"
)

// Store はログの永続化を抽象化するインターフェース。
// テスト時にモック実装に差し替えられる。
type Store interface {
	InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error)
	Close() error
}
