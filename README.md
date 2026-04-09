# WorkflowLens - Analytics for Game Development Tools

ゲーム開発ツール（Unity、Maya等）の利用ログ・エラーログを収集・蓄積しそれを分析することでワークフローを最適化していくシステムです。
ローカルの中継サーバー（middleware）がツールからUDPでログを受信し、オンラインサーバーへHTTPで転送、PostgreSQLに保存します。

## システム構成

```
┌─────────────────────┐
│ ゲーム開発ツール     │
│ (Unity, Maya, etc.) │
└──────────┬──────────┘
           │ UDP + JSON (localhost:59100)
           ▼
┌──────────────────────┐
│ tool_logger_middleware│  ← ローカル中継サーバー (Go)
│ (UDP → HTTP転送)     │
└──────────┬───────────┘
           │ HTTP POST /logs (バッチ送信)
           ▼
┌──────────────────────┐
│ tool_logger_server   │  ← オンラインサーバー (Go, Cloud Run)
│ (HTTP → PostgreSQL)  │
└──────────┬───────────┘
           │ SQL
           ▼
┌──────────────────────┐
│ PostgreSQL (Supabase)│
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Grafana Cloud        │  ← 可視化
└──────────────────────┘
```

## ビルド・起動

### tool_logger_middleware（ローカル中継サーバー）

```bash
cd tool_logger_middleware

# ビルド
go build -o middleware ./cmd/middleware

# デフォルトポート (59100) で起動
./middleware

# ポートを指定して起動
./middleware 59200
```

起動すると `127.0.0.1:59100` でUDP接続を待ち受けます。
多重起動防止のため、ポート59099をロックとして使用します。

### tool_logger_server（オンラインサーバー）

```bash
cd tool_logger_server

# ビルド
go build -o server ./cmd/server

# 起動（DATABASE_URLは必須）
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=require" ./server

# ポートを指定して起動（デフォルト: 8080）
DATABASE_URL="postgres://..." PORT=9000 ./server
```

| 環境変数 | 必須 | デフォルト | 説明 |
|----------|------|-----------|------|
| `DATABASE_URL` | Yes | — | PostgreSQL接続文字列 |
| `PORT` | No | `8080` | HTTPリッスンポート（Cloud Runが自動設定） |

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

ツールへの同伴方法やクライアント実装例は [tool_logger_middleware/README.md](./tool_logger_middleware/README.md) を参照してください。

## 開発

### tool_logger_middleware

```bash
cd tool_logger_middleware
go build ./cmd/middleware    # ビルド
go run ./cmd/middleware      # 実行（デフォルトポート）
go run ./cmd/middleware 8080 # ポート指定で実行
go test ./...               # テスト実行
go vet ./...                # 静的解析
```

### tool_logger_server

```bash
cd tool_logger_server
go build ./cmd/server        # ビルド
go test ./...                # テスト実行
go vet ./...                 # 静的解析
```

## 仕様書

- middleware: [tool_logger_middleware/Document/](./tool_logger_middleware/Document/)
- server: [tool_logger_server/Document/](./tool_logger_server/Document/)

## セットアップ

### 前提条件

- Go 1.24 以上

### 1. リポジトリのクローン

```bash
git clone https://github.com/AtsuAtsu0120/ToolLogger.git
cd ToolLogger
```

### 2. PostgreSQL（Supabase）のセットアップ

[Supabase](https://supabase.com/) でプロジェ��トを作成し、SQL Editorで以下を実行してテーブルとインデックスを作成する。

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

接続文字列はSupabaseダッシュボードの **Settings > Database > Connection string > URI** から取得できる。

### 3. tool_logger_server のデプロイ

```bash
cd tool_logger_server
go build -o server ./cmd/server

# ローカルで動作確認
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=require" ./server
```

Cloud Runへのデプロイ時は `DATABASE_URL` をシークレットとして設定する。

### 4. tool_logger_middleware のビルド

```bash
cd tool_logger_middleware
go build -o middleware ./cmd/middleware
```

ビルドしたバイナリをゲーム開発ツールのプロジェクトに配置する。
配置例やクライアント実装については [tool_logger_middleware/README.md](./tool_logger_middleware/README.md) を参照。

### 5. Grafana Cloud の接続（任意）

[Grafana Cloud](https://grafana.com/products/cloud/) でアカウントを作成し、データソースにSupabaseのPostgreSQLを追加する。`logs` テーブルに対してSQLクエリでダッシュボードを構築できる。
