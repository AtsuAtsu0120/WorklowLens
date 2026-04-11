---
title: "Middlewareプロセス管理"
status: implemented
priority: high
created: 2026-04-06
updated: 2026-04-06
related_files:
  - src/WorkflowLensClient/WorkflowLens.cs
  - tests/WorkflowLensClient.Tests/WorkflowLensTests.cs
---

# Middlewareプロセス管理

## 概要

`WorkflowLens`クラスにmiddlewareバイナリのプロセス起動・停止機能を統合する。ユーザーがバイナリパスを指定すると、コンストラクタでプロセスを起動し、`Dispose()`で停止する。

## 背景・目的

現状、middlewareの起動・停止はユーザーが`System.Diagnostics.Process`を自前で書く必要がある。クライアントライブラリに統合することで、ログ送信とプロセスライフサイクルをワンストップで管理できるようにする。

## 要件

### プロセス管理

- [ ] `WorkflowLens`のコンストラクタに`middlewarePath`パラメータ（任意）を追加する
- [ ] `middlewarePath`が指定された場合、コンストラクタでmiddlewareプロセスを起動する
- [ ] `middlewarePath`が`null`かつ`autoStartMiddleware`が`false`の場合、従来通りプロセス管理なし（UDP送信のみ）
- [ ] 起動時にポート番号をコマンドライン引数として渡す（`port`パラメータと連動）
- [ ] `Dispose()`でプロセスを停止（Kill）し、リソースを解放する
- [ ] プロセスの起動に失敗した場合は例外を投げる（バイナリが見つからない等は利用者のミス）
- [ ] プロセスの停止に失敗した場合は例外を握りつぶす（既に終了済み等）
- [ ] プロセスの標準出力・標準エラー出力はリダイレクトし、コンソールに表示しない
- [ ] ウィンドウを表示しない（`CreateNoWindow = true`）

### バイナリ自動探索

- [ ] `autoStartMiddleware`パラメータ（bool, デフォルト`true`）を追加する
- [ ] `autoStartMiddleware: true`の場合、以下の優先順位でバイナリを探索する:
  1. `middlewarePath`（明示指定、最優先）
  2. 環境変数`WORKFLOW_LENS_MIDDLEWARE_PATH`
  3. PATH探索（バイナリ名: `workflow_lens_middleware`）
- [ ] どの方法でも見つからない場合、`InvalidOperationException`で明確なエラーメッセージを投げる
- [ ] エラーメッセージにバイナリ名と環境変数名を含め、対処方法がわかるようにする

## 設計

### コンストラクタ

```csharp
public WorkflowLens(
    string toolName,
    string? toolVersion = null,
    string host = "127.0.0.1",
    int port = 59100,
    string? middlewarePath = null,
    bool autoStartMiddleware = true)
```

### バイナリパス解決

```csharp
internal static string? ResolveMiddlewarePath(string? middlewarePath, bool autoStartMiddleware)
```

優先順位:
1. `middlewarePath` が非nullならそのまま返す
2. `autoStartMiddleware` が `false` なら `null` を返す（プロセス起動なし）
3. 環境変数 `WORKFLOW_LENS_MIDDLEWARE_PATH` を参照
4. PATH環境変数を分割して `workflow_lens_middleware` を探索（Windowsでは `.exe` も対象）
5. 見つからなければ `InvalidOperationException` をスロー

### プロセス起動

```csharp
var resolvedPath = ResolveMiddlewarePath(middlewarePath, autoStartMiddleware);
if (resolvedPath != null)
{
    _process = new Process
    {
        StartInfo = new ProcessStartInfo
        {
            FileName = resolvedPath,
            Arguments = port.ToString(),
            UseShellExecute = false,
            CreateNoWindow = true,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
        }
    };
    _process.Start();
}
```

### Dispose

```csharp
public void Dispose()
{
    if (_disposed) return;
    _disposed = true;

    _client?.Dispose();
    _client = null;

    if (_process is { HasExited: false })
    {
        try { _process.Kill(); } catch { }
    }
    _process?.Dispose();
    _process = null;
}
```

### エラーハンドリング

| ケース | 対処 |
|--------|------|
| バイナリが見つからない（明示パス） | `Process.Start()`が投げる例外をそのまま伝播 |
| バイナリが見つからない（自動探索） | `InvalidOperationException`（対処方法を含むメッセージ） |
| プロセスKill失敗（既に終了済み等） | 例外を握りつぶす |
| middlewarePathが空文字 | `ArgumentException`をスロー |

### 依存

追加の依存なし。`System.Diagnostics.Process`, `System.IO`, `System.Runtime.InteropServices`は.NET Standard 2.1に含まれる。

## テスト方針

- [ ] 単体テスト: `middlewarePath`が`null`の場合、プロセスが起動されないこと
- [ ] 単体テスト: `autoStartMiddleware=false`（デフォルト）でプロセスが起動されないこと
- [ ] 単体テスト: `middlewarePath`が指定されていればそのまま返すこと
- [ ] 単体テスト: `middlewarePath`と`autoStartMiddleware`両方指定時、`middlewarePath`が優先すること
- [ ] 単体テスト: 環境変数が設定されていればそれを使うこと
- [ ] 単体テスト: バイナリが見つからない場合、わかりやすいエラーメッセージが出ること
- [ ] 単体テスト: `Dispose()`を複数回呼んでも例外にならないこと
- [ ] 結合テスト: 実際のバイナリを使ったプロセス起動・停止の動作確認（手動）

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-06 | 初版作成 |
| 2026-04-07 | バイナリ自動探索機能（autoStartMiddleware）を追加 |
