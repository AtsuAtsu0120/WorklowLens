# WorkflowLens - Analytics for Game Development Tools

ゲーム開発ツール（Unity、Maya等）の利用ログ・エラーログを収集・蓄積しそれを分析することでワークフローを最適化していくシステムです。
ローカルの中継サーバー（middleware）がツールからUDPでログを受信し、オンラインサーバーへHTTPで転送、データベースに保存します。

## システム構成

```
┌─────────────────────┐
│ ゲーム開発ツール     │
│ (Unity, Maya, etc.) │
└──────────┬──────────┘
           │ UDP + JSON (localhost:59100)
           ▼
┌──────────────────────┐
│ workflow_lens_middleware│  ← ローカル中継サーバー (Go)
│ (UDP → HTTP転送)     │
└──────────┬───────────┘
           │ HTTP POST /logs (バッチ送信)
           ▼
┌──────────────────────┐
│ workflow_lens_server   │  ← オンラインサーバー (Go)
│ (HTTP → DB保存)      │
└──────────┬───────────┘
           │ SQL
           ▼
┌──────────────────────┐
│ Database             │  ← SQLite / PostgreSQL / MySQL
└──────────────────────┘
```

## ビルド・起動

### workflow_lens_middleware（ローカル中継サーバー）

```bash
cd workflow_lens_middleware

# ビルド
go build -o middleware ./cmd/middleware

# デフォルトポート (59100) で起動（ログ出力のみ）
./middleware

# サーバーへHTTP転送する場合
WORKFLOW_LENS_SERVER_URL=http://localhost:8080 ./middleware

# ポートを指定して起動
./middleware 59200
```

起動すると `127.0.0.1:59100` でUDP接続を待ち受けます。
`WORKFLOW_LENS_SERVER_URL` を設定すると、受信したログをサーバーへバッチ転送します（100件 or 5秒ごと）。
多重起動防止のため、ポート59099をロックとして使用します。

### workflow_lens_server（オンラインサーバー）

```bash
cd workflow_lens_server

# ビルド（デフォルト: SQLiteドライバ組み込み）
go build -o server ./cmd/server

# SQLiteで起動（ゼロ設定）
./server

# PostgreSQLで起動
STORE_DRIVER=postgres DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=require" ./server

# ポートを指定して起動（デフォルト: 8080）
PORT=9000 ./server
```

#### ビルドタグによるドライバ選択

```bash
go build ./cmd/server                        # SQLiteのみ（デフォルト）
go build -tags postgres ./cmd/server         # SQLite + PostgreSQL
go build -tags "postgres,mysql" ./cmd/server # 全ドライバ
```

| 環境変数 | 必須 | デフォルト | 説明 |
|----------|------|-----------|------|
| `STORE_DRIVER` | No | `sqlite` | データベースドライバ名（`sqlite`, `postgres`, `mysql`） |
| `DATABASE_URL` | SQLite以外はYes | `workflowlens.db` | データベース接続文字列（DSN） |
| `PORT` | No | `8080` | HTTPリッスンポート |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | （未設定でOTel無効） | OpenTelemetry OTLPエンドポイント |
| `OTEL_SERVICE_NAME` | No | `workflow_lens_server` | OTelサービス名 |

#### workflow_lens_middleware 環境変数

| 環境変数 | 必須 | デフォルト | 説明 |
|----------|------|-----------|------|
| `WORKFLOW_LENS_SERVER_URL` | No | （未設定でログ出力のみ） | サーバーURL（例: `http://localhost:8080`） |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | （未設定でOTel無効） | OpenTelemetry OTLPエンドポイント |
| `OTEL_SERVICE_NAME` | No | `workflow_lens_middleware` | OTelサービス名 |

## ログメッセージ仕様

ツールからmiddlewareへ、UDPで1データグラム1メッセージのJSONを送信します。

### メッセージ形式

```json
{
  "tool_name": "UnityTerrainEditor",
  "event_type": "usage",
  "timestamp": "2026-04-04T10:30:00Z",
  "message": "Terrain brush applied",
  "session_id": "ute-a1b2c3d4",
  "tool_version": "2.1.0",
  "details": { "feature": "paint_height", "brush_size": 5 }
}
```

### フィールド

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `tool_name` | string | Yes | ツール名（例: `"UnityTerrainEditor"`） |
| `event_type` | string | Yes | 下記5種類のいずれか |
| `timestamp` | string | Yes | ISO 8601形式（例: `"2026-04-04T10:30:00Z"`） |
| `message` | string | Yes | ログメッセージ本文 |
| `session_id` | string | No | セッションID。同一起動内のイベントを紐付ける |
| `tool_version` | string | No | ツールのバージョン。バージョン別比較に使用 |
| `details` | object/null | No | 任意の追加情報（feature, error_code, severity等） |

### イベントタイプ

| event_type | 用途 |
|-----------|------|
| `session_start` | ツール起動時（ウィンドウを開いた、プラグインロード等） |
| `usage` | 機能を実行したとき（ボタン押下、ブラシ適用、エクスポート等） |
| `error` | エラー発生時（例外、処理失敗等） |
| `cancellation` | 操作をキャンセルしたとき（ダイアログキャンセル、処理中断等） |
| `session_end` | ツール終了時（ウィンドウを閉じた、プラグインアンロード等） |

### 制約

- 1データグラムの最大サイズ: **64 KiB**
- エンコーディング: **UTF-8**

## ツールへの組み込み

ツールへの同伴方法やクライアント実装例は [workflow_lens_middleware/README.md](./workflow_lens_middleware/README.md) を参照してください。

## 開発

### workflow_lens_middleware

```bash
cd workflow_lens_middleware
go build ./cmd/middleware    # ビルド
go run ./cmd/middleware      # 実行（デフォルトポート）
go run ./cmd/middleware 8080 # ポート指定で実行
go test ./...               # テスト実行
go vet ./...                # 静的解析
```

### workflow_lens_server

```bash
cd workflow_lens_server
go build ./cmd/server        # ビルド
go test ./...                # テスト実行
go vet ./...                 # 静的解析
```

### Docker Compose

```bash
# SQLiteで server のみ起動（デフォルト）
docker compose up

# PostgreSQL + OTel Collector + Prometheus + Grafana で起動
docker compose --profile monitoring up
```

## 仕様書

- middleware: [workflow_lens_middleware/Document/](./workflow_lens_middleware/Document/)
- server: [workflow_lens_server/Document/](./workflow_lens_server/Document/)

## セットアップ

### 前提条件

- Go 1.24 以上

### 1. リポジトリのクローン

```bash
git clone https://github.com/AtsuAtsu0120/WorkflowLens.git
cd WorkflowLens
```

### 2. workflow_lens_server のビルド・起動

```bash
cd workflow_lens_server

# SQLiteで起動（ゼロ設定、テーブル自動作成）
go build -o server ./cmd/server
./server
```

PostgreSQLを使う場合:

```bash
# PostgreSQLドライバ付きでビルド
go build -tags postgres -o server ./cmd/server

# 起動
STORE_DRIVER=postgres DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=require" ./server
```

### 3. workflow_lens_middleware のビルド

```bash
cd workflow_lens_middleware
go build -o middleware ./cmd/middleware
```

ビルドしたバイナリをゲーム開発ツールのプロジェクトに配置する。
配置例やクライアント実装については [workflow_lens_middleware/README.md](./workflow_lens_middleware/README.md) を参照。

### 4. モニタリング（任意）

`docker compose --profile monitoring up` で以下が起動する:
- **PostgreSQL** — ログデータの保存
- **OTel Collector** — OpenTelemetryテレメトリの収集
- **Prometheus** — メトリクスの蓄積
- **Grafana** (http://localhost:3000) — サンプルダッシュボード付き

サーバーは `OTEL_EXPORTER_OTLP_ENDPOINT` が設定されると、トレース・メトリクスをOTLPで送信する。
Grafana以外のバックエンド（Jaeger, Datadog等）にも `otel-collector/config.yaml` の設定変更で対応可能。
