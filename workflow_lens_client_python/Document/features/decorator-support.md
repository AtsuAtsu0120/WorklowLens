# Decorator Support（デコレータ対応）

## 概要

`measure()` メソッドおよび `CategoryLogger` をデコレータとしても使用可能にする。既存のコンテキストマネージャとしての使用方法はそのまま維持。

## 動機

- Pythonツール制作者はデコレータによる宣言的なログ記録を期待する
- `with logger.measure(...)` をメソッド内部に書く必要がなくなり、関心の分離が改善される

## measure() のデコレータ対応

### 変更内容

`measure()` の戻り値を `@contextmanager` ジェネレータから `_MeasureDecorator` クラスに変更。
`_MeasureDecorator` は `__enter__`/`__exit__`（コンテキストマネージャ）と `__call__`（デコレータ）の両方を実装。

```python
class _MeasureDecorator:
    def __enter__(self) -> _MeasureDecorator: ...  # 計測開始
    def __exit__(self, ...) -> None: ...            # duration_ms 付き送信
    def __call__(self, func) -> Callable: ...       # デコレータとして関数ラップ
```

### 使用例

```python
# コンテキストマネージャ（従来通り）
with logger.measure(Category.BUILD, "compile"):
    run_compile()

# デコレータ（新規）
@logger.measure(Category.BUILD, "compile")
def run_compile():
    ...
```

## CategoryLogger のデコレータ対応

`CategoryLogger` も `__call__` を実装し、デコレータとして使用可能。

```python
@logger.build("compile")
def run_compile():
    ...
```

action 未指定時にデコレータとして使用すると `ValueError` を送出する。

## 後方互換性

- `with logger.measure(Category.BUILD, "compile"):` は従来通り動作
- 戻り値の型が `Iterator[None]` → `_MeasureDecorator` に変更されるが、コンテキストマネージャプロトコル（`__enter__`/`__exit__`）は同一
