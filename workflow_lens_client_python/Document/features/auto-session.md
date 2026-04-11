# Auto-Session（自動セッション管理）

## 概要

`auto_session=True` の場合、コンストラクタで `session/start` を自動送信し、`close()` 時に `session/end` を自動送信する。重複送信防止フラグにより、手動送信やコンテキストマネージャとの併用時も安全に動作する。

## 動機

- ツール制作者が `log(Category.SESSION, "start"/"end")` を手動で呼ぶ必要をなくす
- コンテキストマネージャと手動管理の混在による二重送信を防ぐ
- C# クライアントの `AutoSession` 機能に対応

## 仕様

### auto_session フラグ

- `WorkflowLensOptions.auto_session` で制御（デフォルト: `True`）
- 従来の位置引数コンストラクタでは `auto_session=False`（後方互換）

### 自動送信タイミング

| タイミング | 送信内容 | 条件 |
|-----------|---------|------|
| コンストラクタ完了時 | `session/start` | `auto_session=True` |
| `close()` 呼び出し時 | `session/end` | `auto_session=True` かつ未送信 |

### 重複送信防止

内部フラグ `_session_start_sent` / `_session_end_sent` で管理。

- `log(Category.SESSION, "start")` 呼び出し時、`_session_start_sent=True` なら送信をスキップ
- `log(Category.SESSION, "end")` 呼び出し時、`_session_end_sent=True` なら送信をスキップ
- フラグは送信成功時に `True` に設定

### コンテキストマネージャとの併用

```python
# auto_session=True のとき
# __enter__ で session/start → 既に送信済みなのでスキップ
# __exit__ で session/end → 1回だけ送信される
with WorkflowLens("my_tool", options=WorkflowLensOptions()) as logger:
    logger.log(Category.EDIT, "brush_apply")
```

### close() での自動送信

```python
# コンテキストマネージャを使わない場合でも、close()で session/end が送信される
logger = WorkflowLens("my_tool", options=WorkflowLensOptions())
# → session/start が自動送信される
logger.log(Category.EDIT, "brush_apply")
logger.close()
# → session/end が自動送信される
```

## 後方互換性

- 従来の位置引数コンストラクタでは `auto_session=False` なので挙動は変わらない
- `with` 文での利用（`__enter__`/`__exit__`）も重複防止フラグにより正しく動作
