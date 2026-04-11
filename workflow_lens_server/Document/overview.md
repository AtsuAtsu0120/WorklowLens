# workflow_lens_server

## 概要

workflow_lens_middleware（Goローカル中継）からHTTPでログを受け取り、データベースに保存するオンラインサーバー。
保存されたログはOpenTelemetryメトリクスとしてエクスポートし、Grafana等で可視化する。

## アーキテクチャ

```
┌──────────────────┐       ┌──���───────────────────┐       ┌────────────┐
│  workflow_lens   │       ��  workflow_lens_server   │       │  Database  │
│  (middleware)    │─HTTP──▶  (このプロジェクト)     │──SQL──▶ SQLite /   │
│  Go / ローカル    │  POST  │  Go                  │       │ PostgreSQL │
└─────────���────────┘       └──────────────────────┘       │ MySQL      │
                                                           └─────┬──────┘
                                                                 │
                                                           ┌─────▼──────���
                                                           │  Grafana   │
                                                           │  (任意)    │
                                                           └──────────��─┘
```

- **通信プロトコル**: HTTP（JSON）
- **データベース**: `database/sql` 対応の任意のDB（デフォルト: SQLite）
- **可視化**: OpenTelemetryメトリクス → Prometheus → Grafana等
- **テレメトリ**: OpenTelemetry（OTLP gRPCエクスポート、環境変数で有効化）

## モジュール一覧

| パッケージ | ディレク���リ | 役�� |
|-----------|-------------|------|
| main | `cmd/server/` | エントリポイント、サーバー起動 |
| handler | `internal/handler/` | HTTPハンドラ（`POST /logs`）、アプリケーションメトリクス |
| model | `internal/model/` | LogMessage構造体、Category定義 |
| store | `internal/store/` | DB接続、INSERT処理（プラガブル） |
| telemetry | `internal/telemetry/` | OpenTelemetry初期化・終了 |

## プロジェクト構成

```
workflow_lens_server/
├── Document/
│   ├── overview.md          # この文書
│   └── features/
│       ├── log-receiver.md  # ログ受信API
��       └── log-storage.md   # ログ保存
├── cmd/
│   └── server/
│       └── main.go          # エントリポイント
├── internal/
���   ├── handler/
│   │   └── logs.go          # POST /logs ハンドラ
│   ├── model/
│   │   └── log_message.go   # LogMessage, Category
│   ├── store/
│   │   ├── store.go         # Storeインターフェース
│   │   ���── sql_store.go     # database/sql汎用Store実装
│   │   ├── dialect.go       # SQL方言定義
│   │   ├── driver_default.go    # SQLiteドライバ（デフ���ルト）
│   │   ├── driver_postgres.go   # PostgreSQLドライバ（ビルド��グ）
│   │   └── driver_mysql.go      # MySQLドライバ（ビルドタグ）
│   └── telemetry/
│       └── telemetry.go     # OpenTelemetry初期化・終了
├── go.mod
├── go.sum
└── README.md
```

## 設計判断

| 判断 | 選択 | 理由 |
|------|------|------|
| 言語 | Go | HTTPサーバー+DB接続が標準ライブラリ+最小依存で書ける。保守コスト最小 |
| フレームワーク | `net/http`（標準のみ） | Go 1.22以降のServeMuxでメソッドルーティング対応済み。外部フレームワーク不要 |
| DBインター��ェース | `database/sql`（標準） | Go標準のDB抽象化。ドライバを差し替えるだけでSQLite/PostgreSQL/MySQL等に対応 |
| デフォルトDB | SQLite（`modernc.org/sqlite`） | pure Go実装でCGO不要。ゼロ設定で動作し、クロスコンパイルも容易 |
| ドライバ切り替え | ビルドタグ | 不要なドライバをバイナリに含めない。`-tags postgres` で追加 |
| バッチINSERT | トランザクション内で1行ずつINSERT | `database/sql` 標準APIで全DB共通。期待されるバッチサイズ（数件〜数百件）では十分な性能 |
| テーブル作成 | 起動時に `CREATE TABLE IF NOT EXISTS` | 手動でのスキーマ作成が不要。初回起動で自動的にテーブルを作成 |
| 可視化 | OpenTelemetryメトリクス → Prometheus | プラットフォーム非依存。Grafana以外のバックエンドにも対応可能 |

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
    "category": "session",
    "action": "start",
    "timestamp": "2026-04-04T10:00:00Z",
    "session_id": "ute-a1b2c3d4",
    "tool_version": "2.1.0",
    "user_id": "tanaka"
  },
  {
    "tool_name": "UnityTerrainEditor",
    "category": "edit",
    "action": "brush_apply",
    "timestamp": "2026-04-04T10:05:00Z",
    "session_id": "ute-a1b2c3d4",
    "tool_version": "2.1.0",
    "user_id": "tanaka",
    "duration_ms": 120
  }
]
```

**レス��ンス**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"inserted": 2}
```

**エラー**:
| ステータス | 条件 |
|-----------|------|
| 400 | JSONパースエラー、必須フィールド欠落 |
| 413 | リクエストボディが1MiB超 |
| 500 | DB接続エラー、INSERT失�� |

### `GET /health`

ヘルスチェック用。

```http
HTTP/1.1 200 OK

ok
```

## データベーススキーマ

テーブルはサーバー起動時に自動作成される。以下はPostgreSQLでの例:

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

SQLiteの場合は型が `TEXT` / `INTEGER` に置き換わる。詳細は `internal/store/dialect.go` を参照。

`received_at` はサ��バー受信時刻。`timestamp` はクライアントが送った時刻。クライアントの時計がズレている場合の調査��使う。

## 環境変数

| 変数名 | 必須 | 説明 | 例 |
|--------|------|------|----|
| `STORE_DRIVER` | いいえ | DBドライバ名（デフォルト: `sqlite`） | `sqlite`, `postgres`, `mysql` |
| `DATABASE_URL` | SQLite以外はYes | DB接続文字列 | `postgres://user:pass@host:5432/dbname?sslmode=require` |
| `PORT` | いいえ | リッスンポート（デフォルト: 8080） | `8080` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | いいえ | OTLPエンドポイント（未設定でOTel無効） | `http://localhost:4317` |
| `OTEL_SERVICE_NAME` | いいえ | OTelサービ���名 | `workflow_lens_server` |

SQLiteの場合、`DATABASE_URL` 未指定時は `workflowlens.db` がデフォルトで使用される。

## 機能仕様インデックス

| 機能名 | ファイル | status |
|--------|---------|--------|
| ログ受信API | [log-receiver.md](features/log-receiver.md) | implemented |
| ログ保存 | [log-storage.md](features/log-storage.md) | implemented |
| OpenTelemetry | [opentelemetry.md](features/opentelemetry.md) | implemented |
