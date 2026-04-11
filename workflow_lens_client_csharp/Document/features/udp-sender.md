---
title: "UDP送信"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - src/WorkflowLensClient/WorkflowLens.cs
  - src/WorkflowLensClient/LogMessage.cs
  - src/WorkflowLensClient/EventType.cs
---

# UDP送信

## 概要

workflow_lens_middlewareへUDPデータグラムでJSONログメッセージを送信する機能。

## 背景・目的

UnityツールからC#で簡単にログを送信できるようにする。fire-and-forgetでツール本体のパフォーマンスに影響を与えない。

## 要件

- [ ] UDPデータグラムでJSONメッセージを送信できる
- [ ] 1データグラム = 1 JSONメッセージ
- [ ] UTF-8エンコード
- [ ] 送信先ホスト・ポートをコンストラクタで指定可能（デフォルト: 127.0.0.1:59100）
- [ ] 送信失敗時（middleware未起動等）に例外を投げない
- [ ] IDisposableでUdpClientを解放できる
- [ ] スレッドセーフ
- [ ] 外部依存なし（標準ライブラリのみ）

## 設計

### メッセージJSON構造

```json
{
  "tool_name": "string (必須)",
  "event_type": "string (必須) - usage/error/session_start/session_end/cancellation",
  "timestamp": "ISO 8601 (必須)",
  "message": "string (必須)",
  "session_id": "string (任意)",
  "tool_version": "string (任意)",
  "details": "JSON object (任意)"
}
```

### イベント種別

```csharp
public static class EventType
{
    public const string Usage = "usage";
    public const string Error = "error";
    public const string SessionStart = "session_start";
    public const string SessionEnd = "session_end";
    public const string Cancellation = "cancellation";
}
```

### 公開API

```csharp
public class WorkflowLens : IDisposable
{
    public WorkflowLens(string toolName, string? toolVersion = null,
                      string host = "127.0.0.1", int port = 59100);

    public void Send(string eventType, string message, string? details = null);
    public void LogUsage(string message, string? details = null);
    public void LogError(string message, string? details = null);
    public void LogCancellation(string message, string? details = null);
    public void Dispose();
}
```

### JSON組み立て

`LogMessage` 内部クラスで文字列補間によりJSON文字列を構築する。`details` は生JSON文字列をそのまま埋め込む。

### エラーハンドリング

| ケース | 対処 |
|--------|------|
| middleware未起動 | `SocketException` をキャッチして無視 |
| Dispose済み | `ObjectDisposedException` をキャッチして無視 |

## テスト方針

- [ ] UDPソケットをbindして送信メッセージを受信・検証
- [ ] 必須フィールドのみのJSON構造検証
- [ ] 全フィールドありのJSON構造検証
- [ ] details省略時にキーが省略されること
- [ ] タイムスタンプがISO 8601形式であること
- [ ] middleware未起動時にSend()が例外を投げないこと
- [ ] Dispose後のSend()が例外を投げないこと

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
