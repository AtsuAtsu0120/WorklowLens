# tool_logger

## 概要

ゲーム開発ツール（Unity, Maya等）の使用率やエラーログを収集するローカル中継ツール。
各ツールからUDP経由でJSONログを受信し、将来的にはオンラインサーバーへ転送する。

## アーキテクチャ

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────┐
│ Unity (C#)  │     │                  │     │              │
│ Maya (Python)│──UDP──▶  tool_logger    │──HTTP──▶ オンライン   │
│ 他ツール     │     │  (このプロジェクト) │     │  サーバー     │
└─────────────┘     └──────────────────┘     └──────────────┘
  クライアント         ローカル中継            別プロジェクト
```

- **通信プロトコル**: UDP（1データグラム = 1 JSONメッセージ）
- **同期モデル**: `std::net::UdpSocket`（シングルスレッド `recv_from` ループ）
- **デフォルトポート**: 59100

## モジュール一覧

| モジュール | ファイル | 役割 |
|-----------|---------|------|
| main | `src/main.rs` | エントリポイント、ポート設定 |
| server | `src/server.rs` | UDPソケット、recv_fromループ、データグラム処理 |
| message | `src/message.rs` | LogMessage構造体、EventType列挙体、JSONパース |
| instance_lock | `src/instance_lock.rs` | 多重起動防止（ポートベースのロック） |
| error | `src/error.rs` | アプリケーション独自のエラー型 |

## 設計判断

| 判断 | 選択 | 理由 |
|------|------|------|
| TCP vs UDP | UDP | ログ送信はfire-and-forget。接続管理不要でクライアント・サーバーともにシンプル |
| 同期モデル | `std::net::UdpSocket`（シングルスレッド） | コネクションレスなのでスレッド不要。Rust初心者にも読みやすい |
| メッセージ形式 | 1データグラム = 1 JSON | 最もシンプル。C#/Pythonから簡単に送れる |
| detailsの型 | `serde_json::Value` | ツールごとに異なるデータを柔軟に送れる |
| 多重起動防止の方法 | ポートバインド（59099） | PIDファイルはクラッシュ時にゴミが残る。ポートはOS終了時に自動解放。stdのみで実装可能 |

## 機能仕様インデックス

| 機能名 | ファイル | status |
|--------|---------|--------|
| UDPサーバー基盤 | [udp-server.md](features/udp-server.md) | implemented |
| ログメッセージ | [log-message.md](features/log-message.md) | implemented |
| 多重起動防止 | [single-instance.md](features/single-instance.md) | implemented |
