---
title: "ログメッセージ"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-11
related_files:
  - internal/model/log_message.go
---

# ログメッセージ

## 概要

クライアント（Unity, Maya等のツール）からUDPで送られるログメッセージのデータ構造を定義する。

## 背景・目的

ゲーム開発で使うツールの操作ログを統一的に収集するため、言語に依存しないJSONベースのメッセージフォーマットが必要。

v2ではツール制作者の認知負荷を下げるため、`event_type` + `message` + `details` を `category`（enum）+ `action`（string）の2層構造に統合した。設計判断の詳細は [ADR-0001](../../../Document/adr/0001-log-schema-v2.md) を参照。

## 要件

- [x] JSON形式でメッセージを表現できる
- [x] ツール名、カテゴリ、アクション、タイムスタンプを必須フィールドとする
- [x] カテゴリはenum（定義済みの値のみ許可）
- [x] 不正なJSONやフィールド不足時にわかりやすいエラーを返す
- [x] session_id/tool_version/user_id/duration_msをオプショナルフィールドとしてサポートする

## 設計

### データ構造

`LogMessage` — 1つのログメッセージを表す構造体。workflow_lens_serverの`model.LogMessage`と同じフィールドを持つ。

```go
// validCategories は許可するカテゴリ。
var validCategories = map[string]bool{
    "asset":   true,
    "build":   true,
    "edit":    true,
    "error":   true,
    "session": true,
}

// LogMessage はクライアントから送られるログメッセージ。
type LogMessage struct {
    ToolName    string    `json:"tool_name"`
    Category    string    `json:"category"`
    Action      string    `json:"action"`
    Timestamp   time.Time `json:"timestamp"`
    SessionID   *string   `json:"session_id,omitempty"`
    ToolVersion *string   `json:"tool_version,omitempty"`
    UserID      *string   `json:"user_id,omitempty"`
    DurationMs  *int64    `json:"duration_ms,omitempty"`
    Traceparent *string   `json:"traceparent,omitempty"`
}
```

**カテゴリの使い分け**:

| カテゴリ | 用途 | action例 |
|---------|------|----------|
| `asset` | アセット関連操作 | `import`, `export`, `convert` |
| `build` | ビルド・コンパイル関連 | `compile`, `bundle`, `bake` |
| `edit` | 編集操作 | `brush_apply`, `parameter_change`, `place` |
| `error` | エラー発生 | `shader_compile`, `asset_load`, `network` |
| `session` | セッション管理 | `start`, `end` |

**フィールド説明**:

- `category` — 操作の大分類。enumで定義された値のみ許可。
- `action` — カテゴリ内の具体的操作。ツール制作者が自由に定義する文字列。
- `session_id` — ツール起動ごとにユニークなIDを生成し、そのセッション内の全イベントに付与する。
- `tool_version` — ツールのバージョン文字列。バージョンごとの比較に使う。
- `user_id` — ユーザー識別子。未指定時はクライアントがOSユーザー名を自動設定する。
- `duration_ms` — 操作の所要時間（ミリ秒）。手動設定またはスコープAPIで自動計測。

### 公開API

```go
// ValidateCategory はcategoryが有効な値かどうかを検証する。
func ValidateCategory(category string) bool

// Parse はJSONバイト列からLogMessageをパースし、バリデーションを行う。
// 必須フィールドの欠落やcategoryの不正を検出する。
func Parse(data []byte) (LogMessage, error)
```

### エラーハンドリング

| エラーケース | 対処 |
|-------------|------|
| JSONとしてパースできない | エラーを返す |
| 必須フィールドが欠落 | `"missing required field: xxx"` エラーを返す |
| categoryが未知の値 | `"invalid category: 'xxx'"` エラーを返す |
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
- [x] 単体テスト: 全カテゴリのパース（asset, build, edit, error, session）
- [x] 単体テスト: オプションフィールド省略時にnilになること
- [x] 単体テスト: user_id/duration_ms付きのパース
- [x] 単体テスト: 不正なcategoryの拒否
- [x] 単体テスト: 必須フィールド欠落の拒否（tool_name, category, action, timestamp）
- [x] 単体テスト: 不正なtimestamp形式の拒否
- [x] 単体テスト: 不正なJSON文字列の拒否
- [x] 単体テスト: 空JSON・空文字列の拒否
- [x] 単体テスト: session_id/tool_version省略時の後方互換性
- [x] 単体テスト: シリアライズのラウンドトリップ

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-04 | Go版に更新 |
| 2026-04-11 | v2: category+action 2層構造に移行（ADR-0001）。event_type/message/details を削除、user_id/duration_ms を追加 |
