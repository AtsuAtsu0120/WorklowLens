---
title: "ログ保存"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - internal/store/store.go
  - internal/store/postgres.go
---

# ログ保存

## 概要

受信したログメッセージをPostgreSQLにバッチINSERTで保存する。

## 背景・目的

middlewareから一括送信されるログ（数件〜数百件）を効率的にDBへ保存する。個別INSERTではなくバッチINSERTにすることで、DBラウンドトリップを1回に抑える。

## 要件

- [x] LogMessage配列をPostgreSQLの `logs` テーブルにバッチINSERTする
- [x] 1リクエスト内のログは1トランザクションで保存する（全件成功 or 全件ロールバック）
- [x] `received_at` カラムにサーバー受信時刻を自動設定する
- [x] `details` フィールドをJSONBとして保存する
- [x] 接続文字列は環境変数 `DATABASE_URL` から取得する
- [x] コネクションプールを使う

## 設計

### Storeインターフェース

```go
// Store はログの永続化を抽象化するインターフェース。
// テスト時にモック実装に差し替えられる。
type Store interface {
    InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error)
    Close() error
}
```

### PostgreSQL実装

```go
// PostgresStore はPostgreSQLを使ったStore実装。
type PostgresStore struct {
    pool *pgxpool.Pool
}

// NewPostgresStore は接続プールを作成してPostgresStoreを返す。
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error)

// InsertLogs はログをバッチINSERTする。挿入件数を返す。
func (s *PostgresStore) InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error)

// Close はコネクションプールを閉じる。
func (s *PostgresStore) Close() error
```

### バッチINSERTの実装方針

pgxの `CopyFrom` を使ってバルクINSERTする。これはPostgreSQLの `COPY` プロトコルを使い、大量行の挿入に最も効率的。

```go
// CopyFrom は PostgreSQL の COPY プロトコルを使ったバルクINSERT。
// 通常の INSERT INTO ... VALUES ... を繰り返すより高速。
func (s *PostgresStore) InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error) {
    columns := []string{
        "tool_name", "event_type", "timestamp", "message",
        "session_id", "tool_version", "details",
    }
    rows := make([][]any, len(logs))
    for i, log := range logs {
        rows[i] = []any{
            log.ToolName, log.EventType, log.Timestamp, log.Message,
            log.SessionID, log.ToolVersion, log.Details,
        }
    }
    copyCount, err := s.pool.CopyFrom(ctx,
        pgx.Identifier{"logs"},
        columns,
        pgx.CopyFromRows(rows),
    )
    return int(copyCount), err
}
```

### テーブル定義

```sql
CREATE TABLE logs (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tool_name    TEXT        NOT NULL,
    event_type   TEXT        NOT NULL,
    timestamp    TIMESTAMPTZ NOT NULL,
    message      TEXT        NOT NULL,
    session_id   TEXT,
    tool_version TEXT,
    details      JSONB,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_logs_tool_name  ON logs (tool_name);
CREATE INDEX idx_logs_event_type ON logs (event_type);
CREATE INDEX idx_logs_timestamp  ON logs (timestamp);
CREATE INDEX idx_logs_session_id ON logs (session_id);
```

### 依存パッケージ

| パッケージ | 用途 |
|-----------|------|
| `github.com/jackc/pgx/v5` | PostgreSQLドライバ（Go最速） |
| `github.com/jackc/pgx/v5/pgxpool` | コネクションプール |

### エラーハンドリング

| エラーケース | 対処 |
|-------------|------|
| 接続失敗 | サーバー起動時にFatal終了 |
| INSERT失敗 | エラーを返す → ハンドラが500を返す |
| タイムアウト | contextのタイムアウトで制御 |

## テスト方針

- [ ] 単体テスト: Store インターフェースのモック実装
- [ ] 結合テスト: テスト用PostgreSQLへの実INSERT（CI環境で実施）
- [ ] 結合テスト: 空配列のINSERT（0件挿入、エラーなし）
- [ ] 結合テスト: details=nullの行が正しく保存される
- [ ] 結合テスト: detailsにネストしたJSONが保存・取得できる

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
