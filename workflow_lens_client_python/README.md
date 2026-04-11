# workflow-lens-client (Python)

workflow_lens_middlewareへUDPでログを送信するPythonクライアントライブラリ。

## インストール

```bash
pip install workflow-lens-client
```

またはソースコードを直接プロジェクトに含める。

## 使い方

### 基本

```python
from workflow_lens_client import WorkflowLens

# クライアント作成
logger = WorkflowLens("my_tool", tool_version="1.0.0")

# セッション開始
logger.start_session()

# 使用ログ
logger.log_usage("ボタンAを押下")

# details付き
logger.log_usage("エクスポート完了", {"format": "fbx", "count": 42})

# エラーログ
logger.log_error("ファイルが見つかりません", {"path": "/tmp/missing.txt"})

# キャンセルログ
logger.log_cancellation("エクスポートをキャンセル")

# セッション終了
logger.end_session()
logger.close()
```

### コンテキストマネージャ

```python
with WorkflowLens("my_tool", "1.0.0") as logger:
    logger.log_usage("操作")
# __enter__でstart_session、__exit__でend_session + closeが自動実行
```

### 汎用送信

```python
from workflow_lens_client import WorkflowLens, USAGE

logger = WorkflowLens("my_tool")
logger.send(USAGE, "カスタムイベント", {"key": "value"})
logger.close()
```

### イベント種別

| 定数 | 値 | 用途 |
|------|---|------|
| `USAGE` | `"usage"` | ツール機能の実行 |
| `ERROR` | `"error"` | エラー発生 |
| `SESSION_START` | `"session_start"` | ツール起動 |
| `SESSION_END` | `"session_end"` | ツール終了 |
| `CANCELLATION` | `"cancellation"` | 操作キャンセル |

## 設計

- **通信**: UDP fire-and-forget（middleware未起動でも例外を投げない）
- **スレッドセーフ**: `threading.Lock` で `sendto()` をガード
- **details**: `dict` を渡す（`json.dumps()` でシリアライズ）
- **session_id**: `start_session()` で自動生成（UUID先頭8文字）
- **依存**: なし（Python標準ライブラリのみ）
