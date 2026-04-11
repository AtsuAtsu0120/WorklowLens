---
title: "セッション自動化"
status: implemented
priority: high
created: 2026-04-11
updated: 2026-04-11
related_files:
  - src/WorkflowLensClient/WorkflowLens.cs
  - src/WorkflowLensClient/WorkflowLensOptions.cs
---

# セッション自動化

## 概要

`WorkflowLens`のコンストラクタで`Session/start`を、`Dispose()`で`Session/end`を自動送信する。ツール開発者が手動でセッション開始・終了を呼ぶ必要をなくす。

## 背景・目的

現在、ツール開発者は`Log(Category.Session, "start")`と`Log(Category.Session, "end")`を自分で呼ぶ必要があり、忘れやすい。特にendの呼び忘れはクラッシュ検知の精度に影響する。コンストラクタ/Disposeに紐付けることで確実にセッションライフサイクルを管理できる。

## 要件

- [ ] `AutoSession`オプション（デフォルト`true`）を`WorkflowLensOptions`に追加する
- [ ] `AutoSession=true`の場合、コンストラクタ末尾で`Log(Category.Session, "start")`を自動送信する
- [ ] `AutoSession=true`の場合、`Dispose()`のUdpClient破棄前に`Log(Category.Session, "end")`を自動送信する
- [ ] 手動で`Log(Category.Session, "start")`が呼ばれた場合、重複送信を防止する（内部フラグで制御）
- [ ] 手動で`Log(Category.Session, "end")`が呼ばれた場合も同様に重複を防止する
- [ ] `AutoSession=false`の場合、従来通り手動管理（既存互換）

## 設計

### 内部状態

```csharp
private bool _sessionStartSent;
private bool _sessionEndSent;
```

### コンストラクタ末尾

```csharp
if (options.AutoSession)
{
    Log(Category.Session, "start");
}
```

### Dispose（UdpClient破棄前）

```csharp
if (_autoSession && !_sessionEndSent)
{
    Log(Category.Session, "end");
}
```

### Log() の重複防止

```csharp
public void Log(Category category, string action, long? durationMs = null)
{
    if (category == Category.Session)
    {
        if (action == "start")
        {
            if (_sessionStartSent) return;
            _sessionStartSent = true;
        }
        else if (action == "end")
        {
            if (_sessionEndSent) return;
            _sessionEndSent = true;
        }
    }
    // ... 既存の送信処理
}
```

### 既存コンストラクタとの互換

既存の7パラメータコンストラクタは`AutoSession`パラメータを持たないため、Optionsコンストラクタへの委譲時に`AutoSession = false`を設定する。これにより、既存コードの動作は一切変わらない。

### セッション管理仕様との関係

既存の`session-management.md`で定義されたセッションライフサイクルの自動化版。手動呼び出しも引き続き可能だが、AutoSessionが有効な場合は重複防止ロジックが働く。

## テスト方針

- [ ] AutoSession=true（デフォルト）でSession/startが自動送信されること
- [ ] AutoSession=trueでDispose()時にSession/endが自動送信されること
- [ ] AutoSession=falseで自動送信されないこと（既存互換）
- [ ] AutoSession=trueで手動start後、重複startが送信されないこと
- [ ] AutoSession=trueで手動end後、Dispose()で重複endが送信されないこと
- [ ] 既存の7パラメータコンストラクタでAutoSessionが無効であること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
