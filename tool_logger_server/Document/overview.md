# tool_logger_server

## 概要

tool_logger_middleware（Rustローカル中継）からHTTPでログを受け取り、PostgreSQLに保存するオンラインサーバー。
保存されたログはGrafana Cloudから直接SQLクエリして可視化する。

## アーキテクチャ

```
┌──────────────────┐       ┌──────────────────────┐       ┌────────────┐
│  tool_logger     │       │  tool_logger_server   │       │ PostgreSQL │
│  (middleware)    │─HTTP──▶  (このプロジェクト)     │──SQL──▶ (Supabase)  │
│  Rust / ローカル  │  POST  │  Go / Cloud Run      │       │            │
└──────────────────┘       └──────────────────────┘       └─────┬──────┘
                                                                │
                                                          ┌─────▼──────┐
                                                          │  Grafana   │
                                                          │  Cloud     │
                                                          └────────────┘
```

- **通信プロトコル**: HTTP（JSON）
- **ホスティング**: Google Cloud Run（リクエストゼロ時スケールtoゼロ）
- **データベース**: PostgreSQL（Supabase無料枠）
- **可視化**: Grafana Cloud（PostgreSQLに直接接続）

## モジュール一覧

| パッケージ | ディレクトリ | 役割 |
|-----------|-------------|------|
| main | `cmd/server/` | エントリポイント、サーバー起動 |
| handler | `internal/handler/` | HTTPハンドラ（`POST /logs`） |
| model | `internal/model/` | LogMessage構造体、EventType定義 |
| store | `internal/store/` | PostgreSQL接続、INSERT処理 |

## プロジェクト構成

```
tool_logger_server/
├── Document/
│   ├── overview.md          # この文書
│   └── features/
│       ├── log-receiver.md  # ログ受信API
│       └── log-storage.md   # ログ保存
├── cmd/
│   └── server/
│       └── main.go          # エントリポイント
├── internal/
│   ├── handler/
│   │   └── logs.go          # POST /logs ハンドラ
│   ├── model/
│   │   └── log_message.go   # LogMessage, EventType
│   └── store/
│       └── postgres.go      # DB接続・INSERT
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## 設計判断

| 判断 | 選択 | 理由 |
|------|------|------|
| 言語 | Go | HTTPサーバー+DB接続が標準ライブラリ+最小依存で書ける。保守コスト最小 |
| フレームワーク | `net/http`（標準のみ） | Go 1.22以降のServeMuxでメソッドルーティング対応済み。外部フレームワーク不要 |
| DBドライバ | `pgx` | Go最速のPostgreSQLドライバ。プリペアドステートメント対応 |
| バッチINSERT | middlewareがログをバッファリングして配列で送信 | HTTPリクエスト数削減、DB負荷軽減 |
| Grafana直結 | GrafanaからPostgreSQLにSQLクエリ | ダッシュボード用APIの開発が不要になる |
| middleware側の転送方式 | middleware→server はHTTP POST | 将来的にmiddlewareがオフライン時のリトライ対応もしやすい |

## API仕様

### `POST /logs`

middlewareからのログを一括受信する。

**リクエスト**:
```http
POST /logs HTTP/1.1
Content-Type: application/json

[
  {
    "tool_name": "UnityTerrainEditor",
    "event_type": "session_start",
    "timestamp": "2026-04-04T10:00:00Z",
    "message": "Tool opened",
    "session_id": "ute-a1b2c3d4",
    "tool_version": "2.1.0"
  },
  {
    "tool_name": "UnityTerrainEditor",
    "event_type": "usage",
    "timestamp": "2026-04-04T10:05:00Z",
    "message": "Brush applied",
    "session_id": "ute-a1b2c3d4",
    "tool_version": "2.1.0",
    "details": {"feature": "paint_height", "brush_size": 5}
  }
]
```

**レスポンス**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"inserted": 2}
```

**エラー**:
| ステータス | 条件 |
|-----------|------|
| 400 | JSONパースエラー、必須フィールド欠落 |
| 500 | DB接続エラー、INSERT失敗 |

### `GET /health`

Cloud Runのヘルスチェック用。

```http
HTTP/1.1 200 OK

ok
```

## データベーススキーマ

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

`received_at` はサーバー受信時刻。`timestamp` はクライアントが送った時刻。クライアントの時計がズレている場合の調査に使う。

## 環境変数

| 変数名 | 必須 | 説明 | 例 |
|--------|------|------|----|
| `DATABASE_URL` | はい | PostgreSQL接続文字列 | `postgres://user:pass@host:5432/dbname?sslmode=require` |
| `PORT` | いいえ | リッスンポート（デフォルト: 8080） | `8080` |

Cloud Runでは `PORT` が自動設定される。

## 機能仕様インデックス

| 機能名 | ファイル | status |
|--------|---------|--------|
| ログ受信API | [log-receiver.md](features/log-receiver.md) | implemented |
| ログ保存 | [log-storage.md](features/log-storage.md) | implemented |
