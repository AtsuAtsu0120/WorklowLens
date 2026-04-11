# workflow_lens_middleware

## 概要

ゲーム開発ツール（Unity, Maya等）の使用率やエラーログを収集するローカル中継ツール。
各ツールからUDP経由でJSONログを受信し、オンラインサーバーへHTTPでバッチ転送する。

## アーキテクチャ

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────┐
│ Unity (C#)  │     │                  │     │              │
│ Maya (Python)│──UDP──▶  workflow_lens  │──HTTP──▶ オンライン   │
│ 他ツール     │     │  (このプロジェクト) │     │  サーバー     │
└─────────────┘     └──────────────────┘     └──────────────┘
  クライアント         ローカル中継            別プロジェクト
```

- **通信プロトコル**: UDP（1データグラム = 1 JSONメッセージ）
- **同期モデル**: `net.PacketConn`（シングルgoroutine `ReadFrom` ループ）
- **デフォルトポート**: 59100

## モジュール一覧

| モジュール | ファイル | 役割 |
|-----------|---------|------|
| main | `cmd/middleware/main.go` | エントリポイント、ポート設定、forwarder初期化 |
| server | `internal/server/server.go` | UDPソケット、ReadFromループ、データグラム処理 |
| model | `internal/model/log_message.go` | LogMessage構造体、イベント種別、JSONパース |
| forwarder | `internal/forwarder/forwarder.go` | バッファリング + サーバーへのHTTPバッチ転送 |
| lock | `internal/lock/instance_lock.go` | 多重起動防止（ポートベースのロック） |
| telemetry | `internal/telemetry/telemetry.go` | OpenTelemetry初期化・終了 |

## 設計判断

| 判断 | 選択 | 理由 |
|------|------|------|
| TCP vs UDP | UDP | ログ送信はfire-and-forget。接続管理不要でクライアント・サーバーともにシンプル |
| 同期モデル | `net.ListenPacket`（シングルgoroutine） | コネクションレスなのでgoroutine不要。シンプルなループで十分 |
| メッセージ形式 | 1データグラム = 1 JSON | 最もシンプル。C#/Pythonから簡単に送れる |
| detailsの型 | `json.RawMessage` | ツールごとに異なるデータを柔軟に送れる |
| 多重起動防止の方法 | ポートバインド（59099） | PIDファイルはクラッシュ時にゴミが残る。ポートはOS終了時に自動解放 |
| 言語 | Go | workflow_lens_serverと言語を統一。標準ライブラリのみで実装可能 |

## 環境変数

| 変数名 | 必須 | 説明 | デフォルト |
|--------|------|------|-----------|
| `WORKFLOW_LENS_SERVER_URL` | いいえ | 転送先サーバーURL。未設定時はログ出力のみ | （なし） |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | いいえ | OTLPエンドポイント。未設定時はOTel無効 | （なし） |
| `OTEL_SERVICE_NAME` | いいえ | OTelサービス名 | `workflow_lens_middleware` |

## 機能仕様インデックス

| 機能名 | ファイル | status |
|--------|---------|--------|
| UDPサーバー基盤 | [udp-server.md](features/udp-server.md) | implemented |
| ログメッセージ | [log-message.md](features/log-message.md) | implemented |
| 多重起動防止 | [single-instance.md](features/single-instance.md) | implemented |
| HTTP転送 | [http-forwarder.md](features/http-forwarder.md) | implemented |
| OpenTelemetry | [opentelemetry.md](features/opentelemetry.md) | implemented |
