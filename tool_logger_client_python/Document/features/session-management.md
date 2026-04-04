---
title: "セッション管理"
status: implemented
priority: medium
created: 2026-04-04
updated: 2026-04-04
related_files:
  - src/tool_logger_client/client.py
---

# セッション管理

## 概要

ツールの起動から終了までをセッションとして管理し、session_idで紐付ける機能。

## 背景・目的

同一セッション内のイベント（開始→操作→エラー→終了）を紐付けることで、セッション時間やクラッシュ検知が可能になる。

## 要件

- [ ] start_session()でsession_idを自動生成し、session_startイベントを送信
- [ ] end_session()でsession_endイベントを送信
- [ ] session_idはstart_session()呼出しごとに新規生成
- [ ] session_idはsend()に自動的に付与される
- [ ] start_session()を呼ばずにsend()した場合もsession_idが付与される（コンストラクタで初期生成）
- [ ] session_idプロパティで現在のsession_idを参照できる
- [ ] コンテキストマネージャ（with文）でstart_session/end_session/closeを自動化

## 設計

### セッションライフサイクル

```
コンストラクタ → session_id初期生成
     ↓
start_session() → 新しいsession_id生成 + session_startイベント送信
     ↓
send() / log_usage() / log_error() 等 → 現在のsession_idを自動付与
     ↓
end_session() → session_endイベント送信
     ↓
（再度start_session()で新セッション開始可能）
```

### コンテキストマネージャ

```python
with ToolLogger("my_tool", "1.0.0") as logger:
    logger.log_usage("ボタン押下")
# __enter__でstart_session()、__exit__でend_session() + close()
```

### 公開API

```python
class ToolLogger:
    @property
    def session_id(self) -> str: ...
    def start_session(self, message="Session started", details=None): ...
    def end_session(self, message="Session ended", details=None): ...
    def __enter__(self): ...
    def __exit__(self, *args): ...
```

### session_id生成

`uuid.uuid4().hex[:8]` — 8文字の16進文字列。

## テスト方針

- [ ] コンストラクタでsession_idが生成されること
- [ ] start_session()でsession_startイベントが送信されること
- [ ] start_session()で新しいsession_idが生成されること
- [ ] end_session()でsession_endイベントが送信されること
- [ ] send()にsession_idが自動付与されること
- [ ] 2回start_session()すると異なるsession_idになること
- [ ] コンテキストマネージャでstart/endが自動送信されること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
