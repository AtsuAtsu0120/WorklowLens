package store

import (
	"context"
	"database/sql"
	"encoding/json"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/kaido-atsuya/workflow_lens_server/internal/model"
)

var (
	storeTracer = otel.Tracer("store")
	storeMeter  = otel.Meter("store")
)

// SQLStore は database/sql を使った汎用Store実装。
// SQLite, PostgreSQL, MySQL など database/sql 対応のドライバで動作する。
type SQLStore struct {
	db           *sql.DB
	d            dialect
	insertSQL    string
	driverName   string
	logsInserted metric.Int64Counter
}

// NewSQLStore はDB接続を開き、テーブルを自動作成して SQLStore を返す。
func NewSQLStore(ctx context.Context, driverName, dsn string) (*SQLStore, error) {
	d, err := getDialect(driverName)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	// テーブルの自動作成
	if _, err := db.ExecContext(ctx, d.createTableSQL); err != nil {
		db.Close()
		return nil, err
	}

	logsInserted, _ := storeMeter.Int64Counter("server.logs.inserted",
		metric.WithDescription("挿入されたログ行数"))

	return &SQLStore{
		db:           db,
		d:            d,
		insertSQL:    d.buildInsertSQL(),
		driverName:   driverName,
		logsInserted: logsInserted,
	}, nil
}

// InsertLogs はログをトランザクション内でバッチINSERTする。挿入件数を返す。
func (s *SQLStore) InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error) {
	if len(logs) == 0 {
		return 0, nil
	}

	ctx, span := storeTracer.Start(ctx, "db.insert_logs",
		trace.WithAttributes(
			attribute.String("db.system", s.driverName),
			attribute.String("db.operation", "INSERT"),
			attribute.String("db.sql.table", "logs"),
			attribute.Int("db.batch_size", len(logs)),
		),
	)
	defer span.End()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		span.SetStatus(codes.Error, "begin tx failed")
		span.RecordError(err)
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, s.insertSQL)
	if err != nil {
		span.SetStatus(codes.Error, "prepare failed")
		span.RecordError(err)
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, log := range logs {
		var sessionID, toolVersion any
		if log.SessionID != nil {
			sessionID = *log.SessionID
		}
		if log.ToolVersion != nil {
			toolVersion = *log.ToolVersion
		}

		var details any
		if log.Details != nil {
			details = string(json.RawMessage(log.Details))
		}

		if _, err := stmt.ExecContext(ctx,
			log.ToolName, log.EventType, log.Timestamp, log.Message,
			sessionID, toolVersion, details,
		); err != nil {
			span.SetStatus(codes.Error, "exec failed")
			span.RecordError(err)
			return 0, err
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		span.SetStatus(codes.Error, "commit failed")
		span.RecordError(err)
		return 0, err
	}

	span.SetAttributes(attribute.Int("db.rows_affected", count))
	s.logsInserted.Add(ctx, int64(count))

	return count, nil
}

// Close はDB接続を閉じる。
func (s *SQLStore) Close() error {
	return s.db.Close()
}
