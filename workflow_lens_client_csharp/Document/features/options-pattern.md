---
title: "Optionsパターン"
status: implemented
priority: high
created: 2026-04-11
updated: 2026-04-11
related_files:
  - src/WorkflowLensClient/WorkflowLensOptions.cs
  - src/WorkflowLensClient/WorkflowLens.cs
---

# Optionsパターン

## 概要

`WorkflowLens`コンストラクタの7パラメータを`WorkflowLensOptions`クラスに集約し、`Action<WorkflowLensOptions>`コールバックで設定できるようにする。

## 背景・目的

現在のコンストラクタは7パラメータあり、ツール開発者にとって見通しが悪い。デフォルト値で十分なケースが大半なので、必要なものだけを明示的に設定できるOptionsパターンに移行する。

## 要件

- [ ] `WorkflowLensOptions`クラスを新設し、`toolName`以外の全パラメータをプロパティとして持つ
- [ ] `WorkflowLens`に`Action<WorkflowLensOptions>`を受け取るコンストラクタオーバーロードを追加する
- [ ] 既存の7パラメータコンストラクタは互換性のため残し、内部でOptionsに委譲する
- [ ] `toolName`はコンストラクタの第1引数として必須のまま残す（Optionsに入れない）
- [ ] AutoSession設定もOptionsに含める（auto-session.md参照）

## 設計

### WorkflowLensOptions

```csharp
public class WorkflowLensOptions
{
    public string? ToolVersion { get; set; }
    public string? UserId { get; set; }
    public string Host { get; set; } = "127.0.0.1";
    public int Port { get; set; } = 59100;
    public string? MiddlewarePath { get; set; }
    public bool AutoStartMiddleware { get; set; } = true;
    public bool AutoSession { get; set; } = true;
}
```

### 公開API

```csharp
public class WorkflowLens : IDisposable
{
    // 新: Optionsパターン
    public WorkflowLens(string toolName, Action<WorkflowLensOptions>? configure = null);

    // 既存: 互換性のため残す（内部でOptionsに委譲）
    public WorkflowLens(string toolName, string? toolVersion = null,
                        string? userId = null,
                        string host = "127.0.0.1", int port = 59100,
                        string? middlewarePath = null,
                        bool autoStartMiddleware = true);
}
```

### 使用例

```csharp
// 最小形（デフォルト値で十分な場合）
using var logger = new WorkflowLens("my_tool");

// 一部のオプションだけ変更
using var logger = new WorkflowLens("my_tool", o =>
{
    o.ToolVersion = "1.0.0";
    o.Port = 59200;
});
```

### 内部実装

既存コンストラクタはOptionsコンストラクタに委譲する:

```csharp
public WorkflowLens(string toolName, string? toolVersion, ...)
    : this(toolName, o =>
    {
        o.ToolVersion = toolVersion;
        o.UserId = userId;
        o.Host = host;
        o.Port = port;
        o.MiddlewarePath = middlewarePath;
        o.AutoStartMiddleware = autoStartMiddleware;
    })
{ }
```

### 依存

追加の依存なし。

## テスト方針

- [ ] デフォルトOptionsでインスタンス生成できること
- [ ] Optionsで各プロパティが正しく反映されること
- [ ] 既存の7パラメータコンストラクタが従来通り動作すること
- [ ] configure引数がnullでもデフォルト値で動作すること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
