---
title: "セッション管理"
status: implemented
priority: medium
created: 2026-04-04
updated: 2026-04-04
related_files:
  - src/WorkflowLensClient/WorkflowLens.cs
---

# セッション管理

## 概要

ツールの起動から終了までをセッションとして管理し、session_idで紐付ける機能。

## 背景・目的

同一セッション内のイベント（開始→操作→エラー→終了）を紐付けることで、セッション時間やクラッシュ検知（session_endがないセッション）が可能になる。

## 要件

- [ ] StartSession()でsession_idを自動生成し、session_startイベントを送信
- [ ] EndSession()でsession_endイベントを送信
- [ ] session_idはStartSession()呼出しごとに新規生成
- [ ] session_idはSend()に自動的に付与される
- [ ] StartSession()を呼ばずにSend()した場合もsession_idが付与される（コンストラクタで初期生成）
- [ ] SessionIdプロパティで現在のsession_idを参照できる

## 設計

### セッションライフサイクル

```
コンストラクタ → session_id初期生成
     ↓
StartSession() → 新しいsession_id生成 + session_startイベント送信
     ↓
Send() / LogUsage() / LogError() 等 → 現在のsession_idを自動付与
     ↓
EndSession() → session_endイベント送信
     ↓
（再度StartSession()で新セッション開始可能）
```

### 公開API

```csharp
public class WorkflowLens
{
    public string SessionId { get; }

    public void StartSession(string message = "Session started", string? details = null);
    public void EndSession(string message = "Session ended", string? details = null);
}
```

### session_id生成

`Guid.NewGuid().ToString("N").Substring(0, 8)` — 8文字の16進文字列。

## テスト方針

- [ ] コンストラクタでsession_idが生成されること
- [ ] StartSession()でsession_startイベントが送信されること
- [ ] StartSession()で新しいsession_idが生成されること
- [ ] EndSession()でsession_endイベントが送信されること
- [ ] Send()にsession_idが自動付与されること
- [ ] 2回StartSession()すると異なるsession_idになること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
