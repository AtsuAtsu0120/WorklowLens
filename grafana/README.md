# WorkflowLens Grafana ダッシュボード

OpenTelemetryメトリクスによるゲーム開発ツールの利用状況可視化ダッシュボード。

## アーキテクチャ

```
workflow_lens_server → OTel metrics → OTel Collector → Prometheus → Grafana
```

サーバーがログ受信時にOTelメトリクスを記録し、OTel Collector経由でPrometheusに蓄積、GrafanaがPromQLで可視化する。

## Docker Compose（ローカル環境）

### 起動

プロジェクトルートから monitoring プロファイルで実行する。

```bash
docker compose --profile monitoring up
```

### アクセス

| サービス | URL | 備考 |
|----------|-----|------|
| Grafana | http://localhost:3000 | 初期ログイン: admin / admin |
| Prometheus | http://localhost:9090 | メトリクスクエリ |
| Server (PostgreSQL) | http://localhost:8081 | POST /logs, GET /health |
| OTel Collector | localhost:4317 (gRPC), :4318 (HTTP) | OTLP受信 |
| PostgreSQL | localhost:5432 | DB: workflowlens |

起動後、Grafanaの「WorkflowLens」ダッシュボードが自動で表示される。

### 停止

```bash
docker compose --profile monitoring down        # コンテナ停止（データ保持）
docker compose --profile monitoring down -v     # コンテナ停止 + データ削除
```

## ダッシュボード概要

| セクション | パネル | PromQLメトリクス |
|-----------|--------|-----------------|
| 概要 | Total Events, Active Tools, Sessions Started, Error Rate | `app_events_total`, `app_sessions_started_total` |
| イベント推移 | Event Volume Over Time, Events by Tool | `rate(app_events_total[5m])` |
| エラー分析 | Error Rate by Tool, Errors Over Time | `app_events_total{event_type="error"}` |
| セッション・バージョン | Session Count, Tool Version Distribution | `app_sessions_started_total` |

## カスタム環境でのセットアップ

1. OTel Collector → Prometheus パイプラインを構成（`otel-collector/config.yaml` 参照）
2. GrafanaにPrometheusデータソースを追加（UID: `workflowlens-prometheus`）
3. `dashboards/workflowlens.json` をインポート
