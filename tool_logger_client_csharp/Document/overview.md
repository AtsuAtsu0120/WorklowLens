# tool_logger_client_csharp

## 概要

tool_logger_middlewareへUDPでログを送信するC#クライアントライブラリ。
主にUnity上のゲーム開発ツールからの利用を想定する。

## アーキテクチャ

```
┌─────────────────────┐     ┌──────────────────┐
│ Unity ツール (C#)    │     │                  │
│                     │──UDP──▶ tool_logger      │
│ ToolLogger クラス    │     │  middleware       │
└─────────────────────┘     └──────────────────┘
  このライブラリ              ローカル中継
```

- **通信プロトコル**: UDP（1データグラム = 1 JSONメッセージ）
- **デフォルト送信先**: `127.0.0.1:59100`
- **fire-and-forget**: 送信失敗時は例外を投げない

## モジュール一覧

| ファイル | 役割 |
|---------|------|
| `src/ToolLoggerClient/ToolLogger.cs` | メインクラス。UdpClientでJSON送信 |
| `src/ToolLoggerClient/LogMessage.cs` | JSONペイロードの組み立て |
| `src/ToolLoggerClient/EventType.cs` | イベント種別の定数定義 |

## 設計判断

| 判断 | 選択 | 理由 |
|------|------|------|
| ターゲット | .NET Standard 2.1 | Unity 2021.2+対応。Nullable参照型が使える |
| details型 | `string`（生JSON） | System.Text.Json非依存。呼び出し側が好きなシリアライザを使える |
| JSON組み立て | 文字列補間 | 固定構造なのでシリアライザ不要。外部依存ゼロ |
| スレッドセーフ | ロック不要 | UdpClient.Send()自体がスレッドセーフ |
| エラーハンドリング | サイレントキャッチ | ログ送信がツール本体を壊してはならない |
| session_id | Guid先頭8文字 | 短くて実用上十分 |

## 機能仕様インデックス

| 機能名 | ファイル | status |
|--------|---------|--------|
| UDP送信 | [udp-sender.md](features/udp-sender.md) | implemented |
| セッション管理 | [session-management.md](features/session-management.md) | implemented |
