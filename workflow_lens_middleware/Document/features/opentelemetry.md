---
title: "OpenTelemetry計装"
status: implemented
priority: medium
created: 2026-04-11
updated: 2026-04-11
related_files:
  - internal/server/server.go
  - internal/model/log_message.go
---

# OpenTelemetry計装

## 概要

workflow_lens_middlewareにOpenTelemetryのトレースとメトリクスを導入する。UDPデータグラムの受信・処理をスパンとして記録し、処理量やエラー率をメトリクスで計測する。

## 背景・目的

middlewareはクライアントとサーバーの間に位置する中継コンポーネントであり、ここを計装することでエンドツーエンドの可観測性を実現する。クライアントが送信したtraceparentを受け取り、サーバーへのHTTP転送まで一貫したトレースを構成できるようにする。

## 要件

- [ ] OTEL_EXPORTER_OTLP_ENDPOINT未設定時はno-op（パフォーマンス影響なし）
- [ ] データグラム処理ごとにスパンを生成する
- [ ] クライアントから送られたtraceparentフィールドから親コンテキストを抽出する
- [ ] データグラムの受信数・不正数・サイズをメトリクスとして記録する
- [ ] 環境変数で動作を制御できる

## 設計

### OTel SDK初期化

アプリケーション起動時にTracerProviderとMeterProviderを初期化する。`OTEL_EXPORTER_OTLP_ENDPOINT`が未設定の場合はno-opのプロバイダを使用し、計装コードは残しつつも一切のオーバーヘッドを発生させない。

```go
// initOTel はOpenTelemetry SDKを初期化する。
// OTEL_EXPORTER_OTLP_ENDPOINT未設定時はno-opプロバイダを返す。
// 返却されるshutdown関数はmain()のdeferで呼ぶこと。
func initOTel(ctx context.Context) (shutdown func(context.Context) error, err error)
```

### トレース

`processDatagram` の処理にスパンを生成する。

| スパン名 | 種類 | 説明 |
|---------|------|------|
| `middleware.process_datagram` | Internal | 1データグラムの受信・パース・転送処理全体 |

**属性**:

| 属性 | 型 | 説明 |
|------|-----|------|
| `tool.name` | string | パース後のLogMessage.ToolName |
| `event.type` | string | パース後のLogMessage.EventType |
| `session.id` | string | パース後のLogMessage.SessionID（存在する場合） |
| `messaging.message.payload_size_bytes` | int | データグラムのバイト数 |

### traceparentフィールド

クライアントライブラリがJSONペイロードに`traceparent`フィールドを埋め込む。middlewareはこのフィールドからW3C Trace Context形式のトレースコンテキストを抽出し、スパンの親コンテキストとして設定する。

**LogMessageへの新フィールド追加**:

```go
type LogMessage struct {
    // 既存フィールド...
    Traceparent *string `json:"traceparent,omitempty"`
}
```

**traceparentの形式**: W3C Trace Context準拠

```
00-<trace-id(32hex)>-<parent-id(16hex)>-<trace-flags(2hex)>
```

例: `00-4bf92f3577b86cd53f044612e066a272-00f067aa0ba902b7-01`

**抽出処理**:

1. JSONパース後に`traceparent`フィールドの有無を確認
2. 存在する場合、`go.opentelemetry.io/otel/propagation.TraceContext`を使ってコンテキストを抽出
3. 抽出したコンテキストを`middleware.process_datagram`スパンの親に設定
4. `traceparent`フィールドが存在しない場合は新規トレースとしてスパンを開始

### メトリクス

| メトリクス名 | 種類 | 単位 | 説明 |
|-------------|------|------|------|
| `middleware.datagrams.received` | Counter | `{datagram}` | 受信したデータグラムの総数 |
| `middleware.datagrams.invalid` | Counter | `{datagram}` | パース/バリデーション失敗のデータグラム数 |
| `middleware.datagram.size_bytes` | Histogram | `By` | データグラムのサイズ分布 |

**`middleware.datagrams.invalid` の属性**:

| 属性 | 値の例 | 説明 |
|------|--------|------|
| `reason` | `json_parse_error`, `validation_error` | 不正の理由 |

### 環境変数

| 変数名 | 必須 | デフォルト | 説明 |
|--------|------|-----------|------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | いいえ | (未設定=no-op) | OTLPエクスポーター先のgRPCエンドポイント |
| `OTEL_SERVICE_NAME` | いいえ | `workflow_lens_middleware` | サービス名 |
| `OTEL_SDK_DISABLED` | いいえ | `false` | `true`でSDKを無効化（no-op） |

### 依存パッケージ

| パッケージ | 用途 |
|-----------|------|
| `go.opentelemetry.io/otel` | OTel APIコア |
| `go.opentelemetry.io/otel/sdk` | TracerProvider |
| `go.opentelemetry.io/otel/sdk/metric` | MeterProvider |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` | トレースのgRPCエクスポーター |
| `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc` | メトリクスのgRPCエクスポーター |
| `go.opentelemetry.io/otel/propagation` | traceparentの抽出 |

## テスト方針

- [ ] 単体テスト: OTEL_EXPORTER_OTLP_ENDPOINT未設定時にno-opで起動すること
- [ ] 単体テスト: traceparentフィールドから親コンテキストを正しく抽出すること
- [ ] 単体テスト: traceparentが不正な形式の場合に新規トレースとして処理すること
- [ ] 単体テスト: traceparentフィールドが存在しない場合に新規トレースとして処理すること
- [ ] 単体テスト: メトリクスのカウンターが正しくインクリメントされること
- [ ] 結合テスト: インメモリエクスポーターでスパンとメトリクスを検証

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
