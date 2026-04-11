---
title: "カテゴリロガー"
status: implemented
priority: medium
created: 2026-04-11
updated: 2026-04-11
related_files:
  - src/WorkflowLensClient/CategoryLogger.cs
  - src/WorkflowLensClient/WorkflowLens.cs
---

# カテゴリロガー

## 概要

`CategoryLogger`クラスを導入し、カテゴリを固定した`using`ブロックでのログ送信を提供する。action指定時はブロック全体の所要時間を自動計測してDispose時に送信し、action省略時はグルーピングとして複数のログを内部で送信できる。既存の`MeasureScope()`を自然に置き換える。

## 背景・目的

- `MeasureScope(Category.Build, "compile")`は機能的には問題ないが、「Scope」という概念がやや抽象的
- `logger.Build("compile")`のusingブロックの方が、「このブロックの処理時間を計測してBuild/compileとして記録する」という意図が直感的に伝わる
- 同一カテゴリの複数ログを論理的にグルーピングするユースケースにも対応する

## 要件

- [ ] `CategoryLogger`クラスを新設する（`IDisposable`実装）
- [ ] action指定ありの場合: usingブロック全体の所要時間を自動計測し、Dispose時にログ送信する
- [ ] action省略の場合: グルーピング用。Dispose時にログは送信しない。中の`Log()`呼び出しは即時送信
- [ ] `CategoryLogger.Log(string action, long? durationMs = null)`で個別ログを送信できる
- [ ] `WorkflowLens`に各カテゴリのファクトリメソッドを追加する: `Asset()`, `Build()`, `Edit()`, `Error()`
- [ ] 各ファクトリメソッドはオプションのaction引数を受け取る
- [ ] 既存の`MeasureScope()`は互換性のため残すが、内部的にCategoryLoggerに委譲する
- [ ] Source Generatorで`CategoryLogger`への拡張メソッドを生成する土台となる

## 設計

### CategoryLogger

```csharp
public sealed class CategoryLogger : IDisposable
{
    private readonly WorkflowLens _logger;
    private readonly Category _category;
    private readonly string? _action;
    private readonly Stopwatch? _stopwatch;
    private bool _disposed;

    internal CategoryLogger(WorkflowLens logger, Category category, string? action = null)
    {
        _logger = logger;
        _category = category;
        _action = action;

        if (action != null)
        {
            _stopwatch = Stopwatch.StartNew();
        }
    }

    /// <summary>指定actionでログを即時送信する</summary>
    public void Log(string action, long? durationMs = null)
        => _logger.Log(_category, action, durationMs);

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;

        // action指定ありの場合のみ、Dispose時に自動計測ログを送信
        if (_action != null && _stopwatch != null)
        {
            _stopwatch.Stop();
            _logger.Log(_category, _action, _stopwatch.ElapsedMilliseconds);
        }
    }
}
```

### WorkflowLens への追加

```csharp
public class WorkflowLens : IDisposable
{
    public CategoryLogger Asset(string? action = null) => new CategoryLogger(this, Category.Asset, action);
    public CategoryLogger Build(string? action = null) => new CategoryLogger(this, Category.Build, action);
    public CategoryLogger Edit(string? action = null)  => new CategoryLogger(this, Category.Edit, action);
    public CategoryLogger Error(string? action = null) => new CategoryLogger(this, Category.Error, action);
}
```

### 使用例

```csharp
using var logger = new WorkflowLens("my_tool");

// 自動計測: usingブロックの所要時間をBuild/compileとして送信
using (logger.Build("compile"))
{
    RunCompiler();
}
// → Dispose時に {"category":"build", "action":"compile", "duration_ms": 1234} を送信

// グルーピング: カテゴリを固定して複数ログを送信
using (var edit = logger.Edit())
{
    edit.Log("brush_apply");
    edit.Log("layer_change");
}
// → 各Log()が即時送信。Disposeでは何も送信しない

// 自動計測 + 内部ログ（両方可能）
using (var build = logger.Build("full_build"))
{
    build.Log("compile_start");
    RunCompiler();
    build.Log("link_start");
    RunLinker();
}
// → compile_start, link_startは即時送信。full_buildは全体の所要時間でDispose時に送信

// 既存APIも引き続き使える
logger.Log(Category.Edit, "brush_apply");

// Source Generatorと組み合わせると:
using (logger.Build("compile")) { RunCompiler(); }
// ↓ 型安全版
using (logger.Build.MeasureCompile()) { RunCompiler(); }
```

### MeasureScopeとの関係

既存の`MeasureScope()`は内部的にCategoryLoggerを使うよう書き換えることも可能だが、API互換性のため当面はそのまま残す。

```csharp
// 既存（引き続き動作する）
using (logger.MeasureScope(Category.Build, "compile")) { ... }

// 新（推奨）
using (logger.Build("compile")) { ... }
```

### 依存

追加の依存なし。`System.Diagnostics.Stopwatch`は.NET Standard 2.1に含まれる。

## テスト方針

- [ ] action指定ありの場合、Dispose時に自動計測ログが送信されること
- [ ] action指定ありの場合、duration_msが実行時間を反映していること
- [ ] action省略の場合、Dispose時にログが送信されないこと
- [ ] グルーピング内のLog()呼び出しが正しいカテゴリで即時送信されること
- [ ] 各ファクトリメソッド（Asset/Build/Edit/Error）が正しいカテゴリのCategoryLoggerを返すこと
- [ ] Dispose多重呼び出しで例外が出ないこと
- [ ] 自動計測 + 内部ログの組み合わせが正しく動作すること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成。CategoryScope → CategoryLogger + IDisposable自動計測に変更 |
