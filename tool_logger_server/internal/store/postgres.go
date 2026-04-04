package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaido-atsuya/tool_logger_server/internal/model"
)

// PostgresStore はPostgreSQLを使ったStore実装。
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore は接続プールを作成してPostgresStoreを返す。
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

// InsertLogs はログをバッチINSERTする。挿入件数を返す。
// pgxのCopyFromを使い、PostgreSQLのCOPYプロトコルでバルクINSERTする。
func (s *PostgresStore) InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error) {
	if len(logs) == 0 {
		return 0, nil
	}

	columns := []string{
		"tool_name", "event_type", "timestamp", "message",
		"session_id", "tool_version", "details",
	}

	rows := make([][]any, len(logs))
	for i, log := range logs {
		var sessionID, toolVersion any
		if log.SessionID != nil {
			sessionID = *log.SessionID
		}
		if log.ToolVersion != nil {
			toolVersion = *log.ToolVersion
		}

		var details any
		if log.Details != nil {
			details = []byte(log.Details)
		}

		rows[i] = []any{
			log.ToolName, log.EventType, log.Timestamp, log.Message,
			sessionID, toolVersion, details,
		}
	}

	copyCount, err := s.pool.CopyFrom(
		ctx,
		pgx.Identifier{"logs"},
		columns,
		pgx.CopyFromRows(rows),
	)
	return int(copyCount), err
}

// Close はコネクションプールを閉じる。
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}
