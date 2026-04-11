---
title: "ログ受信API"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-11
related_files:
  - internal/handler/logs.go
  - internal/model/log_message.go
  - internal/handler/logs_test.go
  - internal/model/log_message_test.go
---

# ログ受信API

## 概要

middlewareからHTTP POSTで送られるログメッセージ（JSON配列）を受信し、バリデーションしてストアに渡す。

## 背景・目的

workflow_lens_middleware はローカルでUDPログを受信してバッファリングし、定期的にこのサーバーへHTTPで転送する。サーバー側は受け取ったJSONを検証してDBに渡す責務を持つ。

## 要件

- [x] `POST /logs` でJSON配列を受信できる
- [x] 必須フィールド（tool_name, category, action, timestamp）のバリデーション
- [x] categoryが定義済みの値（asset, build, edit, error, session）であることを検証
- [x] timestampがISO 8601形式であることを検証
- [x] オプションフィールド（session_id, tool_version, user_id, duration_ms）が欠落してもエラーにならない
- [x] 不正なJSONに対して400エラーとエラーメッセージを返す
- [x] 正常時は挿入件数を返す
- [x] `GET /health` でヘルスチェックに応答する
- [x] リクエストボディの上限を1MiBに制限する

## 設計

### データ構造

```go
// LogMessage はmiddlewareから受信するログメッセージ。
// middleware側のLogMessage構造体と同じフィールドを持つ。
type LogMessage struct {
    ToolName    string    `json:"tool_name"`
    Category    string    `json:"category"`
    Action      string    `json:"action"`
    Timestamp   time.Time `json:"timestamp"`
    SessionID   *string   `json:"session_id,omitempty"`
    ToolVersion *string   `json:"tool_version,omitempty"`
    UserID      *string   `json:"user_id,omitempty"`
    DurationMs  *int64    `json:"duration_ms,omitempty"`
}
```

**middleware側との対応**:

| middleware (Go) | server (Go) | 備考 |
|-----------------|-------------|------|
| `string` | `string` | tool_name, category, action |
| `string` + バリデーション | `string` + バリデーション | categoryの許可値はmap[string]boolで定義 |
| `time.Time` | `time.Time` | `encoding/json` がISO 8601を自動パース |
| `*string` | `*string` | nilで省略を表現（session_id, tool_version, user_id） |
| `*int64` | `*int64` | nilで省略を表現（duration_ms） |

### 公開API

```go
// HandlePostLogs は POST /logs のハンドラ。
// JSON配列を受け取り、バリデーション後にstoreへ渡す。
func HandlePostLogs(store Store) http.HandlerFunc

// HandleHealth は GET /health のハンドラ。
func HandleHealth() http.HandlerFunc
```

### ルーティング

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /logs", handler.HandlePostLogs(store))
mux.HandleFunc("GET /health", handler.HandleHealth())
```

Go 1.22以降の `ServeMux` はメソッド指定ルーティングをサポートしているため、外部ルーターは不要。

### バリデーション

```go
// validCategories は許可するカテゴリ。
var validCategories = map[string]bool{
    "asset":   true,
    "build":   true,
    "edit":    true,
    "error":   true,
    "session": true,
}
```

バリデーションエラー時は最初に見つかったエラーを返す（全件チェックはしない）。

### エラーレスポンス

```json
{"error": "invalid category: 'unknown' at index 2"}
```

### エラーハンドリング

| エラーケース | HTTPステータス | メッセージ例 |
|-------------|--------------|-------------|
| JSONパースエラー | 400 | `invalid JSON: unexpected EOF` |
| 必須フィールド欠落 | 400 | `missing required field: tool_name at index 0` |
| 未知のcategory | 400 | `invalid category: 'unknown' at index 2` |
| ボディサイズ超過 | 413 | `request body too large` |
| DB INSERT失敗 | 500 | `internal server error` |

500エラー時は詳細をクライアントに返さない（ログに出力する）。

### 依存パッケージ

| パッケージ | 用途 |
|-----------|------|
| `net/http` (標準) | HTTPサーバー、ルーティング |
| `encoding/json` (標準) | JSONパース |
| `log/slog` (標準) | 構造化ログ |

ハンドラ層では外部依存なし。

## テスト方針

- [x] 単体テスト: 正常なJSON配列の受信
- [x] 単体テスト: 空配列の受信（200、inserted: 0）
- [x] 単体テスト: 不正なJSONで400
- [x] 単体テスト: 必須フィールド欠落で400
- [x] 単体テスト: 未知のcategoryで400
- [x] 単体テスト: session_id/tool_version/user_id/duration_ms省略でも正常受信
- [x] 単体テスト: ボディサイズ制限のテスト
- [x] 単体テスト: GET /health の応答
- [ ] 結合テスト: httptest.Serverを使ったE2Eテスト

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-11 | v2: category+action 2層構造に移行。event_type/message/details を削除、user_id/duration_ms を追加 |
