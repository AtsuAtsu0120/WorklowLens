---
title: "UDP送信"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-11
related_files:
  - src/workflow_lens_client/client.py
  - src/workflow_lens_client/log_message.py
  - src/workflow_lens_client/category.py
---

# UDP送信

## 概要

workflow_lens_middlewareへUDPデータグラムでJSONログメッセージを送信する機能。

## 背景・目的

MayaツールからPythonで簡単にログを送信できるようにする。fire-and-forgetでツール本体のパフォーマンスに影響を与えない。

## 要件

- [x] UDPデータグラムでJSONメッセージを送信できる
- [x] 1データグラム = 1 JSONメッセージ
- [x] UTF-8エンコード
- [x] 送信先ホスト・ポートをコンストラクタで指定可能（デフォルト: 127.0.0.1:59100）
- [x] 送信失敗時（middleware未起動等）に例外を投げない
- [x] close()でソケットを解放できる
- [x] threading.Lockでスレッドセーフ
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

```python
from enum import Enum

class Category(Enum):
    ASSET = "asset"
    BUILD = "build"
    EDIT = "edit"
    ERROR = "error"
    SESSION = "session"
```

### 公開API

```python
class WorkflowLens:
    def __init__(self, tool_name, tool_version=None, user_id=None,
                 host="127.0.0.1", port=59100): ...
    def log(self, category: Category, action: str, duration_ms: int = None): ...
    def measure(self, category: Category, action: str): ...  # context manager
    def close(self): ...
```

- `log()` — ログを1件送信する。ツール制作者が考えるのはcategoryとactionの2つだけ。
- `measure()` — `with` ブロックで囲むとブロック内の所要時間を自動計測して `duration_ms` に設定する。

### エラーハンドリング

| ケース | 対処 |
|--------|------|
| middleware未起動 | `OSError` をキャッチして無視 |
| close済み | `OSError` をキャッチして無視 |

## テスト方針

- [x] UDPソケットをbindして送信メッセージを受信・検証
- [x] 必須フィールドのみのJSON構造検証
- [x] 全フィールドありのJSON構造検証
- [x] duration_ms=Noneのときキーが省略されること
- [x] タイムスタンプがISO 8601形式であること
- [x] middleware未起動時にlog()が例外を投げないこと
- [x] close後のlog()が例外を投げないこと
- [x] measure()でduration_msが自動設定されること
- [x] user_id未指定時にOSユーザー名が自動設定されること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-11 | v2: category+action 2層構造に移行。log/measure API、user_id/duration_ms追加 |
