---
title: "OpenTelemetry計装"
status: implemented
priority: medium
created: 2026-04-11
updated: 2026-04-11
related_files:
  - cmd/server/main.go
  - internal/handler/logs.go
  - internal/store/sql_store.go
---

# OpenTelemetry計装

## 概要

workflow_lens_serverにOpenTelemetryのトレースとメトリクスを導入する。HTTP受信からDB保存までの処理をスパンで記録し、挿入件数等をメトリクスで計測する。

## 背景・目的

サーバーはログの最終保存先であり、HTTP受信→パース→DB保存の各段階を可観測にすることで、パフォーマンスのボトルネックやエラーの発生箇所を特定できるようにする。middlewareから引き継いだトレースコンテキストと接続し、クライアント→middleware→サーバーの一貫したトレースを実現する。

## 要件

- [ ] OTEL_EXPORTER_OTLP_ENDPOINT未設定時はno-op（パフォーマンス影響なし）
- [ ] HTTP受信を自動計装する（otelhttp）
- [ ] ハンドラ内のパース・DB保存をスパンとして記録する
- [ ] DB操作にセマンティック属性を付与する
- [ ] ログ挿入件数をメトリクスとして記録する

## 設計

### OTel SDK初期化

middlewareと同様の方式。`OTEL_EXPORTER_OTLP_ENDPOINT`未設定時はno-opプロバイダを使用する。

```go
// initOTel はOpenTelemetry SDKを初期化する。
// OTEL_EXPORTER_OTLP_ENDPOINT未設定時はno-opプロバイダを返す。
func initOTel(ctx context.Context) (shutdown func(context.Context) error, err error)
```

### HTTP計装

`otelhttp.NewMiddleware`でServeMuxをラップし、全HTTPリクエストのトレース・メトリクスを自動収集する。

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /logs", handler.HandlePostLogs(store))
mux.HandleFunc("GET /health", handler.HandleHealth())

wrappedHandler := otelhttp.NewMiddleware("server")(mux)
http.ListenAndServe(addr, wrappedHandler)
```

otelhttp が自動的に以下を記録する:
- HTTPリクエスト/レスポンスのスパン
- `http.server.request.duration` ヒストグラム
- `http.server.request.body.size` ヒストグラム
- `http.server.response.body.size` ヒストグラム

### トレース

ハンドラとDB層で子スパンを生成する。

| スパン名 | 種類 | 生成場所 | 説明 |
|---------|------|---------|------|
| `handler.parse_request` | Internal | handler/logs.go | リクエストボディのJSONパース・バリデーション |
| `handler.insert_logs` | Internal | handler/logs.go | ストアへのログ挿入呼び出し |
| `db.insert_logs` | Client | store/sql_store.go | DB INSERT処理 |

### DB計装

`db.insert_logs`スパンにOpenTelemetryのセマンティック規約に準じた属性を付与する。

| 属性 | 型 | 値 | 説明 |
|------|-----|-----|------|
| `db.system` | string | `sqlite` / `postgresql` / `mysql` | DBシステム名 |
| `db.operation` | string | `INSERT` | SQL操作 |
| `db.sql.table` | string | `logs` | 対象テーブル名 |
| `db.statement` | string | (省略可) | 実行SQL（デバッグ用、デフォルトは付与しない） |

### メトリクス

| メトリクス名 | 種類 | 単位 | 説明 |
|-------------|------|------|------|
| `server.logs.inserted` | Counter | `{log}` | DBに挿入されたログの総数 |

HTTPメトリクス（リクエスト時間、ボディサイズ等）はotelhttpが自動で記録するため、手動での計装は不要。

### 環境変数

| 変数名 | 必須 | デフォルト | 説明 |
|--------|------|-----------|------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | いいえ | (未設定=no-op) | OTLPエクスポーター先のgRPCエンドポイント |
| `OTEL_SERVICE_NAME` | いいえ | `workflow_lens_server` | サービス名 |
| `OTEL_SDK_DISABLED` | いいえ | `false` | `true`でSDKを無効化（no-op） |

### 依存パッケージ

| パッケージ | 用途 |
|-----------|------|
| `go.opentelemetry.io/otel` | OTel APIコア |
| `go.opentelemetry.io/otel/sdk` | TracerProvider |
| `go.opentelemetry.io/otel/sdk/metric` | MeterProvider |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` | トレースのgRPCエクスポーター |
| `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc` | メトリクスのgRPCエクスポーター |
| `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | HTTP自動計装 |

## テスト方針

- [ ] 単体テスト: OTEL_EXPORTER_OTLP_ENDPOINT未設定時にno-opで起動すること
- [ ] 単体テスト: POST /logsでhandler.parse_request, handler.insert_logsスパンが生成されること
- [ ] 単体テスト: db.insert_logsスパンにDB属性が付与されること
- [ ] 単体テスト: server.logs.insertedカウンターが正しくインクリメントされること
- [ ] 結合テスト: インメモリエクスポーターでHTTPリクエストからDB保存までのスパンツリーを検証

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
