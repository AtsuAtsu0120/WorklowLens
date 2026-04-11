---
title: "OpenTelemetry対応"
status: implemented
priority: medium
created: 2026-04-11
updated: 2026-04-11
related_files:
  - src/WorkflowLensClient/WorkflowLens.cs
  - src/WorkflowLensClient/LogMessage.cs
---

# OpenTelemetry対応

## 概要

C#クライアントライブラリにOpenTelemetryのトレーシング対応を追加する。`System.Diagnostics.Activity` APIを使用し、OTel SDK自体への依存は持たない。ホストアプリ側でOTel SDKを構成するとスパンが自動的にエクスポートされる。

## 背景・目的

クライアント側でログ送信にスパンを生成し、traceparentをJSONペイロードに埋め込むことで、クライアント→middleware→サーバーの一貫したトレースを実現する。`System.Diagnostics.Activity`は.NETランタイムに組み込まれているため、外部依存なしで計装できる。

## 要件

- [ ] `Send()`呼び出し時にActivityを開始する
- [ ] ActivityのTraceId/SpanIdからtraceparentを生成し、JSONペイロードに注入する
- [ ] OTel SDKが未構成の場合でも正常に動作する（Activityはリスナーなしでno-op相当）
- [ ] 外部NuGetパッケージへの依存を追加しない

## 設計

### ActivitySource

ライブラリ内でstaticな`ActivitySource`を公開する。名前は`"WorkflowLensClient"`とする。

```csharp
public static class WorkflowLensDiagnostics
{
    public static readonly ActivitySource Source = new("WorkflowLensClient");
}
```

### Send()でのActivity開始

`Send()`メソッド内でActivityを開始する。

```csharp
public void Send(string eventType, string message, string? details = null)
{
    using var activity = WorkflowLensDiagnostics.Source.StartActivity(
        "workflowlens.send",
        ActivityKind.Producer);

    // 属性を設定
    activity?.SetTag("tool.name", _toolName);
    activity?.SetTag("event.type", eventType);
    activity?.SetTag("session.id", _sessionId);

    // traceparentを生成
    string? traceparent = null;
    if (activity != null)
    {
        traceparent = $"00-{activity.TraceId}-{activity.SpanId}-{(activity.Recorded ? "01" : "00")}";
    }

    // JSONペイロード組み立て（traceparent含む）
    var payload = LogMessage.Build(
        _toolName, eventType, message, _sessionId,
        _toolVersion, details, traceparent);

    var bytes = Encoding.UTF8.GetBytes(payload);
    activity?.SetTag("messaging.message.payload_size_bytes", bytes.Length);

    // UDP送信
    // ...
}
```

### 属性

| 属性 | 型 | 説明 |
|------|-----|------|
| `tool.name` | string | ツール名 |
| `event.type` | string | イベント種別 |
| `session.id` | string | セッションID |
| `messaging.message.payload_size_bytes` | int | JSONペイロードのバイト数 |

### traceparent注入

Activityが存在する場合（リスナーが登録されている場合）、W3C Trace Context形式のtraceparentをJSONペイロードに埋め込む。

**形式**: `00-{TraceId}-{SpanId}-{TraceFlags}`

- `TraceId`: 32文字の16進数
- `SpanId`: 16文字の16進数
- `TraceFlags`: `01`（Recorded）または `00`

**JSONへの埋め込み**:

```json
{
  "tool_name": "UnityTerrainEditor",
  "event_type": "usage",
  "timestamp": "2026-04-11T10:00:00Z",
  "message": "Brush applied",
  "session_id": "ute-a1b2c3d4",
  "traceparent": "00-4bf92f3577b86cd53f044612e066a272-00f067aa0ba902b7-01"
}
```

Activityがnull（リスナー未登録）の場合、`traceparent`フィールドは省略する。

### LogMessageへの変更

`LogMessage.Build()`に`traceparent`パラメータを追加する。

```csharp
public static string Build(
    string toolName, string eventType, string message,
    string? sessionId, string? toolVersion, string? details,
    string? traceparent)
```

`traceparent`がnullでない場合、JSONに`"traceparent":"..."`フィールドを追加する。

### ホストアプリでのOTel SDK設定例

ライブラリ自体はOTel SDKに依存しないが、ホストアプリ側で以下のようにSDKを構成するとスパンがエクスポートされる。

```csharp
// NuGetパッケージ:
// - OpenTelemetry
// - OpenTelemetry.Exporter.OpenTelemetryProtocol

using var tracerProvider = Sdk.CreateTracerProviderBuilder()
    .AddSource("WorkflowLensClient")
    .AddOtlpExporter()
    .Build();
```

### 外部依存

なし。`System.Diagnostics.Activity`および`System.Diagnostics.ActivitySource`は.NET Standard 2.1に含まれる。

## テスト方針

- [ ] 単体テスト: ActivitySourceにリスナーが登録されていない場合でもSend()が正常に動作すること
- [ ] 単体テスト: リスナー登録時にActivityが生成され、属性が正しく設定されること
- [ ] 単体テスト: traceparentがW3C形式でJSONに埋め込まれること
- [ ] 単体テスト: リスナー未登録時にtraceparentフィールドがJSONに含まれないこと
- [ ] 単体テスト: LogMessage.Build()にtraceparent=nullを渡した場合の後方互換性

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
