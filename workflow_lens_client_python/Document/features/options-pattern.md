# Options Pattern（設定パターン）

## 概要

`WorkflowLensOptions` データクラスによる設定の一元管理パターン。
7引数コンストラクタを整理し、`configure` コールバックまたは `options` オブジェクトで設定を渡せるようにする。

## 動機

- コンストラクタの引数が多く、可読性が低い
- `auto_session` など新しい設定項目の追加が難しい
- C# クライアントの `WorkflowLensOptions` + `Action<WorkflowLensOptions>` パターンに合わせる

## WorkflowLensOptions

```python
@dataclass
class WorkflowLensOptions:
    tool_version: Optional[str] = None
    user_id: Optional[str] = None
    host: str = "127.0.0.1"
    port: int = 59100
    middleware_path: Optional[str] = None
    auto_start_middleware: bool = True
    auto_session: bool = True
```

### フィールド一覧

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `tool_version` | `Optional[str]` | `None` | ツールバージョン |
| `user_id` | `Optional[str]` | `None` | ユーザーID（未指定時はOSユーザー名） |
| `host` | `str` | `"127.0.0.1"` | UDP送信先ホスト |
| `port` | `int` | `59100` | UDP送信先ポート |
| `middleware_path` | `Optional[str]` | `None` | ミドルウェアバイナリのパス |
| `auto_start_middleware` | `bool` | `True` | ミドルウェアの自動起動 |
| `auto_session` | `bool` | `True` | セッション自動管理（auto-session参照） |

## コンストラクタ拡張

```python
WorkflowLens(
    tool_name: str,
    tool_version=None, user_id=None, host="127.0.0.1", port=59100,
    middleware_path=None, auto_start_middleware=True,
    *,
    configure: Optional[Callable[[WorkflowLensOptions], None]] = None,
    options: Optional[WorkflowLensOptions] = None,
)
```

### 設定の優先順位

1. `options` が指定されている場合 → そのまま使用
2. `configure` が指定されている場合 → デフォルトの `WorkflowLensOptions` を作成し、`configure` コールバックで変更
3. どちらも未指定 → 従来の位置引数から `WorkflowLensOptions` を構築（`auto_session=False`）

### 後方互換性

- 従来の位置引数によるコンストラクタはそのまま動作
- 従来パスでは `auto_session=False` となり、挙動は完全に同一

## 使用例

```python
# 従来通り（後方互換）
logger = WorkflowLens("my_tool", "1.0.0", port=59100)

# options オブジェクト
opts = WorkflowLensOptions(tool_version="1.0.0", auto_session=True)
logger = WorkflowLens("my_tool", options=opts)

# configure コールバック
def setup(o):
    o.tool_version = "1.0.0"
    o.auto_session = True

logger = WorkflowLens("my_tool", configure=setup)

# configure ラムダ
logger = WorkflowLens("my_tool", configure=lambda o: setattr(o, 'tool_version', '1.0.0'))
```
