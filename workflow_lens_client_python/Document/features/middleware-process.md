---
title: "Middlewareプロセス管理"
status: implemented
priority: high
created: 2026-04-06
updated: 2026-04-06
related_files:
  - src/workflow_lens_client/client.py
  - tests/test_client.py
---

# Middlewareプロセス管理

## 概要

`WorkflowLens`クラスにmiddlewareバイナリのプロセス起動・停止機能を統合する。ユーザーがバイナリパスを指定すると、コンストラクタでプロセスを起動し、`close()`/コンテキストマネージャの`__exit__`で停止する。

## 背景・目的

現状、middlewareの起動・停止はユーザーが`subprocess.Popen`を自前で書く必要がある。クライアントライブラリに統合することで、ログ送信とプロセスライフサイクルをワンストップで管理できるようにする。

## 要件

### プロセス管理

- [ ] `WorkflowLens.__init__`に`middleware_path`パラメータ（任意）を追加する
- [ ] `middleware_path`が指定された場合、コンストラクタでmiddlewareプロセスを起動する
- [ ] `middleware_path`が`None`かつ`auto_start_middleware`が`False`の場合、従来通りプロセス管理なし（UDP送信のみ）
- [ ] 起動時にポート番号をコマンドライン引数として渡す（`port`パラメータと連動）
- [ ] `close()`でプロセスを`terminate()`→`wait()`で停止する
- [ ] コンテキストマネージャ(`__exit__`)経由でも正しく停止する
- [ ] プロセスの起動に失敗した場合は例外を投げる（`FileNotFoundError`等をそのまま伝播）
- [ ] プロセスの停止に失敗した場合は例外を握りつぶす
- [ ] プロセスの標準出力・標準エラー出力は`subprocess.DEVNULL`にリダイレクトする

### バイナリ自動探索

- [ ] `auto_start_middleware`パラメータ（bool, デフォルト`True`）を追加する
- [ ] `auto_start_middleware=True`の場合、以下の優先順位でバイナリを探索する:
  1. `middleware_path`（明示指定、最優先）
  2. 環境変数`WORKFLOW_LENS_MIDDLEWARE_PATH`
  3. PATH探索（バイナリ名: `workflow_lens_middleware`、`shutil.which()`を使用）
- [ ] どの方法でも見つからない場合、`FileNotFoundError`で明確なエラーメッセージを投げる
- [ ] エラーメッセージにバイナリ名と環境変数名を含め、対処方法がわかるようにする

## 設計

### コンストラクタ

```python
def __init__(
    self,
    tool_name: str,
    tool_version: Optional[str] = None,
    host: str = "127.0.0.1",
    port: int = 59100,
    middleware_path: Optional[str] = None,
    auto_start_middleware: bool = True,
) -> None:
```

### バイナリパス解決

```python
@classmethod
def _resolve_middleware_path(
    cls, middleware_path: Optional[str], auto_start_middleware: bool
) -> Optional[str]:
```

優先順位:
1. `middleware_path` が非Noneならそのまま返す
2. `auto_start_middleware` が `False` なら `None` を返す（プロセス起動なし）
3. 環境変数 `WORKFLOW_LENS_MIDDLEWARE_PATH` を参照
4. `shutil.which("workflow_lens_middleware")` でPATH探索
5. 見つからなければ `FileNotFoundError` をスロー

### プロセス起動

```python
resolved_path = self._resolve_middleware_path(middleware_path, auto_start_middleware)
if resolved_path is not None:
    self._process = subprocess.Popen(
        [resolved_path, str(port)],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
```

### close

```python
def close(self) -> None:
    with self._lock:
        if self._sock is not None:
            self._sock.close()
            self._sock = None

    if self._process is not None:
        try:
            self._process.terminate()
            self._process.wait(timeout=5)
        except Exception:
            pass
        self._process = None
```

### エラーハンドリング

| ケース | 対処 |
|--------|------|
| バイナリが見つからない（明示パス） | `FileNotFoundError`をそのまま伝播 |
| バイナリが見つからない（自動探索） | `FileNotFoundError`（対処方法を含むメッセージ） |
| プロセスterminate/wait失敗 | 例外を握りつぶす |
| middleware_pathが空文字 | `ValueError`をスロー |

### 依存

追加の依存なし。`subprocess`, `os`, `shutil`はPython標準ライブラリ。

## テスト方針

- [ ] 単体テスト: `middleware_path`が`None`の場合、プロセスが起動されないこと
- [ ] 単体テスト: `auto_start_middleware=False`（デフォルト）でプロセスが起動されないこと
- [ ] 単体テスト: `middleware_path`が指定されていればそのまま返すこと
- [ ] 単体テスト: `middleware_path`と`auto_start_middleware`両方指定時、`middleware_path`が優先すること
- [ ] 単体テスト: 環境変数が設定されていればそれを使うこと
- [ ] 単体テスト: バイナリが見つからない場合、わかりやすいエラーメッセージが出ること
- [ ] 単体テスト: `close()`を複数回呼んでも例外にならないこと
- [ ] 結合テスト: 実際のバイナリを使ったプロセス起動・停止の動作確認（手動）

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-06 | 初版作成 |
| 2026-04-07 | バイナリ自動探索機能（auto_start_middleware）を追加 |
