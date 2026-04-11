---
title: "セッション管理"
status: implemented
priority: medium
created: 2026-04-04
updated: 2026-04-11
related_files:
  - src/WorkflowLensClient/WorkflowLens.cs
---

# セッション管理

## 概要

ツールの起動から終了までをセッションとして管理し、session_idで紐付ける機能。

## 背景・目的

同一セッション内のイベント（開始→操作→エラー→終了）を紐付けることで、セッション時間やクラッシュ検知（session/endがないセッション）が可能になる。

## 要件

- [x] コンストラクタでsession_idを自動生成する
- [x] session_idは全てのLog()に自動的に付与される
- [x] SessionIdプロパティで現在のsession_idを参照できる
- [x] セッション開始/終了は `Log(Category.Session, "start")` / `Log(Category.Session, "end")` で送信する

## 設計

### セッションライフサイクル

```
コンストラクタ → session_id初期生成
     ↓
Log(Category.Session, "start") → セッション開始ログ送信
     ↓
Log(Category.Edit, "brush_apply") 等 → 現在のsession_idを自動付与
     ↓
Log(Category.Session, "end") → セッション終了ログ送信
```

### 公開API

```csharp
public class WorkflowLens
{
    public string SessionId { get; }
}
```

専用のStartSession()/EndSession()メソッドは廃止し、`Log(Category.Session, "start"/"end")` に統一。APIの表面積を最小限に保つ。

### session_id生成

`Guid.NewGuid().ToString("N").Substring(0, 8)` — 8文字の16進文字列。

## テスト方針

- [x] コンストラクタでsession_idが生成されること
- [x] Log(Category.Session, "start")でセッション開始ログが送信されること
- [x] Log(Category.Session, "end")でセッション終了ログが送信されること
- [x] Log()にsession_idが自動付与されること

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-11 | v2: StartSession/EndSession廃止。Log(Category.Session, "start"/"end")に統一 |
