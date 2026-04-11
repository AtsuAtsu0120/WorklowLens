---
title: "セッション管理"
status: implemented
priority: medium
created: 2026-04-04
updated: 2026-04-11
related_files:
  - src/workflow_lens_client/client.py
---

# セッション管理

## 概要

ツールの起動から終了までをセッションとして管理し、session_idで紐付ける機能。

## 背景・目的

同一セッション内のイベント（開始→操作→エラー→終了）を紐付けることで、セッション時間やクラッシュ検知が可能になる。

## 要件

- [x] コンストラクタでsession_idを自動生成する
- [x] session_idは全てのlog()に自動的に付与される
- [x] session_idプロパティで現在のsession_idを参照できる
- [x] セッション開始/終了は `log(Category.SESSION, "start")` / `log(Category.SESSION, "end")` で送信する
- [x] コンテキストマネージャ（with文）でセッション開始/終了/closeを自動化

## 設計

### セッションライフサイクル

```
コンストラクタ → session_id初期生成
     ↓
log(Category.SESSION, "start") → セッション開始ログ送信
     ↓
log(Category.EDIT, "brush_apply") 等 → 現在のsession_idを自動付与
     ↓
log(Category.SESSION, "end") → セッション終了ログ送信
```

### コンテキストマネージャ

```python
with WorkflowLens("my_tool", "1.0.0") as logger:
    logger.log(Category.EDIT, "brush_apply")
# __enter__でlog(Category.SESSION, "start")、__exit__でlog(Category.SESSION, "end") + close()
```

### 公開API

```python
class WorkflowLens:
    @property
    def session_id(self) -> str: ...
    def __enter__(self): ...
    def __exit__(self, *args): ...
```

専用のstart_session()/end_session()メソッドは廃止し、`log(Category.SESSION, "start"/"end")` に統一。APIの表面積を最小限に保つ。

### session_id生成

`uuid.uuid4().hex[:8]` — 8文字の16進文字列。

## テスト方針

- [x] コンストラクタでsession_idが生成されること
- [x] log(Category.SESSION, "start")でセッション開始ログが送信されること
- [x] log(Category.SESSION, "end")でセッション終了ログが送信されること
- [x] log()にsession_idが自動付与されること
- [x] コンテキストマネージャでstart/endが自動送信されること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-11 | v2: start_session/end_session廃止。log(Category.SESSION, "start"/"end")に統一 |
