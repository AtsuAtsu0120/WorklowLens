---
title: "ログ保存"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-11
related_files:
  - internal/store/store.go
  - internal/store/sql_store.go
  - internal/store/dialect.go
---

# ログ保存

## 概要

受信したログメッセージをデータベースにバッチINSERTで保存する。
`database/sql` 標準インターフェースを使い、SQLite / PostgreSQL / MySQL に対応する。

## 背景・目的

middlewareから一括送信されるログ（数件〜数百件）を効率的にDBへ保存する。トランザクション内でINSERTすることで、全件成功 or 全件ロールバックを保証する。

`database/sql` を使うことで、ユーザーが用途に応じてDBを自由に選択できる。

## 要件

- [x] LogMessage配列を `logs` テーブルにバッチINSERTする
- [x] 1リクエスト内のログは1トランザクションで保存する（全件成功 or 全件ロールバック）
- [x] `received_at` カラムにサーバー受信時刻を自動設定する（DBのDEFAULT値）
- [x] 接続文字列は環境変数 `DATABASE_URL` から取得する（SQLiteはデフォルト値あり）
- [x] ドライバは環境変数 `STORE_DRIVER` で切り替える（デフォルト: `sqlite`）
- [x] テーブルはサーバー起動時に自動作成する（`CREATE TABLE IF NOT EXISTS`）

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

### SQLStore実装

```go
// SQLStore は database/sql を使った汎用Store実装。
type SQLStore struct {
    db        *sql.DB
    d         dialect
    insertSQL string
}

// NewSQLStore はDB接続を開き、テーブルを自動作成して SQLStore を返す。
func NewSQLStore(ctx context.Context, driverName, dsn string) (*SQLStore, error)

// InsertLogs はログをトランザクション内でバッチINSERTする。挿入件数を返す。
func (s *SQLStore) InsertLogs(ctx context.Context, logs []model.LogMessage) (int, error)

// Close はDB接続を閉じる。
func (s *SQLStore) Close() error
```

### SQL方言（dialect）

ドライバ名に応じて DDL とプレースホルダスタイルを切り替える。

| ドライバ | プレースホルダ | ID列 |
|---------|-------------|------|
| `sqlite` | `?` | INTEGER PRIMARY KEY AUTOINCREMENT |
| `postgres` / `pgx` | `$N` | BIGINT GENERATED ALWAYS AS IDENTITY |
| `mysql` | `?` | BIGINT AUTO_INCREMENT |

### ドライバのビルドタグ

| ファイル | ビルドタグ | ドライバ |
|---------|----------|---------|
| `driver_default.go` | （なし） | `modernc.org/sqlite`（pure Go） |
| `driver_postgres.go` | `postgres` | `github.com/jackc/pgx/v5/stdlib` |
| `driver_mysql.go` | `mysql` | `github.com/go-sql-driver/mysql` |

```bash
go build ./cmd/server                        # SQLiteのみ
go build -tags postgres ./cmd/server         # SQLite + PostgreSQL
go build -tags "postgres,mysql" ./cmd/server # 全ドライバ
```

### テーブル定義（PostgreSQLの例）

```sql
CREATE TABLE IF NOT EXISTS logs (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tool_name    TEXT        NOT NULL,
    category     TEXT        NOT NULL,
    action       TEXT        NOT NULL,
    timestamp    TIMESTAMPTZ NOT NULL,
    session_id   TEXT,
    tool_version TEXT,
    user_id      TEXT,
    duration_ms  BIGINT,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### エラーハンドリング

| エラーケース | 対処 |
|-------------|------|
| 接続失敗 | サーバー起動時にFatal終了 |
| テーブル作成失敗 | サーバー起動時にFatal終了 |
| INSERT失敗 | エラーを返す → ハンドラが500を返す |
| タイムアウト | contextのタイムアウトで制御 |

## テスト方針

- [x] 統合テスト: SQLite in-memory でテーブル自動作成を確認
- [x] 統合テスト: 単一ログのINSERT・SELECTで値を検証
- [x] 統合テスト: オプションフィールド付きログの保存（user_id, duration_ms）
- [x] 統合テスト: NULL値（オプションフィールド未指定）の保存
- [x] 統合テスト: 複数ログのバッチINSERT
- [x] 統合テスト: 空配列のINSERT（0件挿入、エラーなし）
- [x] 単体テスト: 未サポートドライバ名でエラー

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成（PostgreSQL + pgx） |
| 2026-04-11 | database/sql ベースのプラガブルストレージに移行。SQLite/PostgreSQL/MySQL対応 |
| 2026-04-11 | v2: category/action/user_id/duration_ms列に変更。event_type/message/details列を削除 |
