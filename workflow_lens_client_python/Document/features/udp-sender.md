---
title: "UDP送信"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - src/workflow_lens_client/client.py
  - src/workflow_lens_client/log_message.py
  - src/workflow_lens_client/event_type.py
---

# UDP送信

## 概要

workflow_lens_middlewareへUDPデータグラムでJSONログメッセージを送信する機能。

## 背景・目的

MayaツールからPythonで簡単にログを送信できるようにする。fire-and-forgetでツール本体のパフォーマンスに影響を与えない。

## 要件

- [ ] UDPデータグラムでJSONメッセージを送信できる
- [ ] 1データグラム = 1 JSONメッセージ
- [ ] UTF-8エンコード
- [ ] 送信先ホスト・ポートをコンストラクタで指定可能（デフォルト: 127.0.0.1:59100）
- [ ] 送信失敗時（middleware未起動等）に例外を投げない
- [ ] close()でソケットを解放できる
- [ ] threading.Lockでスレッドセーフ
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

```python
USAGE = "usage"
ERROR = "error"
SESSION_START = "session_start"
SESSION_END = "session_end"
CANCELLATION = "cancellation"
```

### 公開API

```python
class WorkflowLens:
    def __init__(self, tool_name, tool_version=None, host="127.0.0.1", port=59100): ...
    def send(self, event_type, message, details=None): ...
    def log_usage(self, message, details=None): ...
    def log_error(self, message, details=None): ...
    def log_cancellation(self, message, details=None): ...
    def close(self): ...
```

### エラーハンドリング

| ケース | 対処 |
|--------|------|
| middleware未起動 | `OSError` をキャッチして無視 |
| close済み | `OSError` をキャッチして無視 |

## テスト方針

- [ ] UDPソケットをbindして送信メッセージを受信・検証
- [ ] 必須フィールドのみのJSON構造検証
- [ ] 全フィールドありのJSON構造検証
- [ ] details=Noneのときキーが省略されること
- [ ] タイムスタンプがISO 8601形式であること
- [ ] middleware未起動時にsend()が例外を投げないこと
- [ ] close後のsend()が例外を投げないこと

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
