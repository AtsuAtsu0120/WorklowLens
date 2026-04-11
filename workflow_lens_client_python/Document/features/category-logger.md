# Category-Scoped Logger（カテゴリスコープロガー）

## 概要

カテゴリを固定した `CategoryLogger` オブジェクトを返すファクトリメソッド。コンテキストマネージャとして使用すると操作時間を自動計測し、デコレータとしても使用可能。

## 動機

- `log(Category.BUILD, "compile")` の繰り返しが冗長
- C# クライアントの `CategoryLogger` + ファクトリメソッド（`.Build()`, `.Edit()` 等）に対応
- コンテキストマネージャ / デコレータの両方で使える統一的な API を提供

## CategoryLogger クラス

```python
class CategoryLogger:
    def __init__(self, logger: WorkflowLens, category: Category, action: Optional[str] = None): ...
    def log(self, action: str, duration_ms: Optional[int] = None) -> None: ...
    def __enter__(self) -> CategoryLogger: ...
    def __exit__(self, ...) -> None: ...
    def __call__(self, func: Callable) -> Callable: ...
```

### 動作モード

| action | コンテキストマネージャ | デコレータ | log() メソッド |
|--------|---------------------|-----------|---------------|
| 指定あり | enter で計測開始、exit で duration_ms 付き送信 | 関数全体の実行時間を計測して送信 | 即時送信（action は log() の引数を使用） |
| `None` | enter/exit でログ送信しない | 使用不可（ValueError） | 即時送信 |

### ファクトリメソッド

`WorkflowLens` に以下のメソッドを追加:

```python
def asset(self, action: Optional[str] = None) -> CategoryLogger: ...
def build(self, action: Optional[str] = None) -> CategoryLogger: ...
def edit(self, action: Optional[str] = None) -> CategoryLogger: ...
def error(self, action: Optional[str] = None) -> CategoryLogger: ...
```

## 使用例

### コンテキストマネージャ（自動計測）

```python
with logger.build("compile"):
    run_compile()
# → category=build, action=compile, duration_ms=計測値 が送信される
```

### デコレータ

```python
@logger.build("compile")
def run_compile():
    ...

run_compile()
# → category=build, action=compile, duration_ms=計測値 が送信される
```

### グルーピング（action なし）

```python
with logger.build() as b:
    b.log("shader_compile", duration_ms=100)
    b.log("texture_bake", duration_ms=200)
# exit 時にログは送信されない（個別の log() で送信済み）
```

### 即時送信

```python
logger.build().log("compile", duration_ms=3200)
```

## エクスポート

`CategoryLogger` は `workflow_lens_client` パッケージからエクスポートする。
