---
title: "ログメッセージ"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - internal/model/log_message.go
---

# ログメッセージ

## 概要

クライアント（Unity, Maya等のツール）からUDPで送られるログメッセージのデータ構造を定義する。

## 背景・目的

ゲーム開発で使うツールの使用率やエラー情報を統一的に収集するため、言語に依存しないJSONベースのメッセージフォーマットが必要。

## 要件

- [x] JSON形式でメッセージを表現できる
- [x] ツール名、イベント種別（使用/エラー）、タイムスタンプ、メッセージ本文を必須フィールドとする
- [x] 任意の追加情報（details）をフリーフォームで添付できる
- [x] 不正なJSONやフィールド不足時にわかりやすいエラーを返す
- [x] session_id/tool_versionをオプショナルフィールドとしてサポートする

## 設計

### データ構造

`LogMessage` — 1つのログメッセージを表す構造体。tool_logger_serverの`model.LogMessage`と同じフィールドを持つ。

```go
// validEventTypes は許可するイベント種別。
var validEventTypes = map[string]bool{
    "usage":         true,
    "error":         true,
    "session_start": true,
    "session_end":   true,
    "cancellation":  true,
}

// LogMessage はクライアントから送られるログメッセージ。
type LogMessage struct {
    ToolName    string          `json:"tool_name"`
    EventType   string          `json:"event_type"`
    Timestamp   time.Time       `json:"timestamp"`
    Message     string          `json:"message"`
    SessionID   *string         `json:"session_id,omitempty"`
    ToolVersion *string         `json:"tool_version,omitempty"`
    Details     json.RawMessage `json:"details,omitempty"`
}
```

**イベント種別の使い分け**:

| イベント | 用途 | 得られる指標 |
|---------|------|-------------|
| `usage` | ツール機能の実行（ボタン押下など） | 機能別利用分布、操作回数 |
| `error` | エラー発生 | エラー率、バージョン別改善傾向 |
| `session_start` | ツール起動/ウィンドウ表示 | セッション数、DAU、利用頻度 |
| `session_end` | ツール終了/ウィンドウ閉じ | セッション時間、クラッシュ検知（session_endがない場合） |
| `cancellation` | 操作のキャンセル | キャンセル率、UX問題の発見 |

**フィールド説明**:

- `session_id` — ツール起動ごとにユニークなIDを生成し、そのセッション内の全イベントに付与する。これにより同一セッション内のイベント（開始→操作→エラー→終了）を紐付けられる。
- `tool_version` — ツールのバージョン文字列。バージョンごとのエラー率比較やアップデート後の改善確認に使う。
- `details` — `json.RawMessage`で任意のJSONデータを保持する。`omitempty`によりフィールドが省略された場合はnilになる。

### 公開API

```go
// ValidateEventType はevent_typeが有効な値かどうかを検証する。
func ValidateEventType(eventType string) bool

// Parse はJSONバイト列からLogMessageをパースし、バリデーションを行う。
// 必須フィールドの欠落やevent_typeの不正を検出する。
func Parse(data []byte) (LogMessage, error)
```

### エラーハンドリング

| エラーケース | 対処 |
|-------------|------|
| JSONとしてパースできない | エラーを返す |
| 必須フィールドが欠落 | `"missing required field: xxx"` エラーを返す |
| event_typeが未知の値 | `"invalid event_type: 'xxx'"` エラーを返す |
| timestampが不正な形式 | json.Unmarshalのエラーを返す |

### 依存パッケージ

標準ライブラリのみ:

| パッケージ | 用途 |
|-----------|------|
| `encoding/json` | JSONパース |
| `time` | 日時型（time.Time） |
| `fmt` | エラーメッセージ |

## テスト方針

- [x] 単体テスト: 正常なJSONのパース
- [x] 単体テスト: 全イベント種別のパース（usage, error, session_start, session_end, cancellation）
- [x] 単体テスト: details省略時にnilになること
- [x] 単体テスト: detailsにネストされたオブジェクトを含む場合
- [x] 単体テスト: 不正なevent_typeの拒否
- [x] 単体テスト: 必須フィールド欠落の拒否（tool_name, event_type, timestamp, message）
- [x] 単体テスト: 不正なtimestamp形式の拒否
- [x] 単体テスト: 不正なJSON文字列の拒否
- [x] 単体テスト: 空JSON・空文字列の拒否
- [x] 単体テスト: session_id/tool_version省略時の後方互換性
- [x] 単体テスト: シリアライズのラウンドトリップ

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-04 | Go版に更新 |
