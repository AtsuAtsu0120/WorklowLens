---
title: "UDP送信"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-11
related_files:
  - src/WorkflowLensClient/WorkflowLens.cs
  - src/WorkflowLensClient/LogMessage.cs
  - src/WorkflowLensClient/Category.cs
---

# UDP送信

## 概要

workflow_lens_middlewareへUDPデータグラムでJSONログメッセージを送信する機能。

## 背景・目的

UnityツールからC#で簡単にログを送信できるようにする。fire-and-forgetでツール本体のパフォーマンスに影響を与えない。

## 要件

- [x] UDPデータグラムでJSONメッセージを送信できる
- [x] 1データグラム = 1 JSONメッセージ
- [x] UTF-8エンコード
- [x] 送信先ホスト・ポートをコンストラクタで指定可能（デフォルト: 127.0.0.1:59100）
- [x] 送信失敗時（middleware未起動等）に例外を投げない
- [x] IDisposableでUdpClientを解放できる
- [x] スレッドセーフ
- [x] 外部依存なし（標準ライブラリのみ）

## 設計

### メッセージJSON構造

```json
{
  "tool_name": "string (必須, コンストラクタで設定)",
  "category": "string (必須, enum) - asset/build/edit/error/session",
  "action": "string (必須)",
  "timestamp": "ISO 8601 (自動)",
  "session_id": "string (自動)",
  "user_id": "string (自動 or コンストラクタ)",
  "tool_version": "string (任意, コンストラクタ)",
  "duration_ms": "number (任意)"
}
```

### カテゴリ

```csharp
public enum Category
{
    Asset,
    Build,
    Edit,
    Error,
    Session,
}
```

### 公開API

```csharp
public class WorkflowLens : IDisposable
{
    public WorkflowLens(string toolName, string? toolVersion = null,
                      string? userId = null,
                      string host = "127.0.0.1", int port = 59100);

    public void Log(Category category, string action, long? durationMs = null);
    public IDisposable MeasureScope(Category category, string action);
    public void Dispose();
}
```

- `Log()` — ログを1件送信する。ツール制作者が考えるのはcategoryとactionの2つだけ。
- `MeasureScope()` — `using` ブロックで囲むとブロック内の所要時間を自動計測して `duration_ms` に設定する。

### JSON組み立て

`LogMessage` 内部クラスで文字列補間によりJSON文字列を構築する。

### エラーハンドリング

| ケース | 対処 |
|--------|------|
| middleware未起動 | `SocketException` をキャッチして無視 |
| Dispose済み | `ObjectDisposedException` をキャッチして無視 |

## テスト方針

- [x] UDPソケットをbindして送信メッセージを受信・検証
- [x] 必須フィールドのみのJSON構造検証
- [x] 全フィールドありのJSON構造検証
- [x] duration_ms省略時にキーが省略されること
- [x] タイムスタンプがISO 8601形式であること
- [x] middleware未起動時にLog()が例外を投げないこと
- [x] Dispose後のLog()が例外を投げないこと
- [x] MeasureScope()でduration_msが自動設定されること
- [x] user_id未指定時にOSユーザー名が自動設定されること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-11 | v2: category+action 2層構造に移行。Log/MeasureScope API、user_id/duration_ms追加 |
