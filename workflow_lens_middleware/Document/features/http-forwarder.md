---
title: "HTTP転送"
status: implemented
priority: high
created: 2026-04-11
updated: 2026-04-11
related_files:
  - internal/forwarder/forwarder.go
  - internal/server/server.go
  - cmd/middleware/main.go
---

# HTTP転送

## 概要

UDPサーバーが受信・バリデーション済みのJSONメッセージをバッファリングし、workflow_lens_serverへバッチPOSTで転送する機能。

## 背景・目的

middlewareはUDPで受信したログをstdoutに出力するだけだったが、workflow_lens_serverへ自動転送することでログの永続化・閲覧を可能にする。UDPの fire-and-forget 思想を維持し、転送失敗時もメッセージを破棄してサーバーの動作を妨げない。

## 要件

- [x] バリデーション済みJSONメッセージをバッファリングする
- [x] バッファが100件に達した時点でフラッシュする
- [x] 最後のフラッシュから5秒経過した時点でフラッシュする（バッファが空でない場合）
- [x] フラッシュ条件はバッファ件数・経過時間のいずれか早い方
- [x] 転送先URLは環境変数 `WORKFLOW_LENS_SERVER_URL` で設定する
- [x] 環境変数が未設定の場合、log-onlyモード（従来動作）で動作する
- [x] HTTP POST `{serverURL}/logs` にJSON配列をボディとして送信する
- [x] Content-Type: `application/json`
- [x] HTTPタイムアウト: 10秒
- [x] 転送失敗時は警告ログを出力し、メッセージを破棄する（fire-and-forget）
- [x] OpenTelemetryトレース: span `middleware.forward_http` を作成し、W3C Trace Contextヘッダーを注入する
- [x] graceful shutdown時にバッファに残っているメッセージをフラッシュする

## 設計

### モジュール構成

```
cmd/middleware/main.go              → 環境変数読み取り、Forwarder生成、server.Run()呼び出し
internal/forwarder/forwarder.go     → バッファリング、バッチPOST転送
internal/server/server.go           → UDPサーバー、Forwarderへのメッセージ引き渡し
```

### Forwarder (`internal/forwarder/forwarder.go`)

```go
// Forwarder はバリデーション済みメッセージをバッファリングし、
// サーバーへバッチPOSTで転送する。
type Forwarder struct { ... }

// New はForwarderを生成する。serverURLは転送先のベースURL。
func New(serverURL string) *Forwarder

// Send はメッセージをバッファに追加する。
// バッファが上限に達した場合、即座にフラッシュする。
func (f *Forwarder) Send(msg []byte)

// Start はタイマーによる定期フラッシュを開始する。
func (f *Forwarder) Start(ctx context.Context)

// Shutdown はバッファに残っているメッセージをフラッシュして終了する。
func (f *Forwarder) Shutdown()
```

### サーバー連携 (`internal/server/server.go`)

```go
// Run はUDPサーバーを起動する。
// fwdがnilの場合はlog-onlyモード（従来動作）。
func Run(ctx context.Context, addr string, fwd *forwarder.Forwarder) error
```

`server.Run()` は `*forwarder.Forwarder` を受け取る。nilの場合はlog-onlyモードとして動作し、従来どおりstdout出力のみを行う。

### フラッシュ条件

| 条件 | 閾値 |
|------|------|
| バッファ件数 | 100件 |
| 経過時間 | 5秒 |

いずれか早い方でフラッシュが発動する。

### HTTPリクエスト

| 項目 | 値 |
|------|-----|
| メソッド | POST |
| URL | `{WORKFLOW_LENS_SERVER_URL}/logs` |
| Content-Type | `application/json` |
| ボディ | JSON配列（バッファ内メッセージの配列） |
| タイムアウト | 10秒 |

### エラーハンドリング

UDPの fire-and-forget 思想に合わせ、転送失敗時はメッセージを破棄する:

1. HTTPリクエスト失敗 → `slog.Warn` で警告出力、メッセージ破棄
2. サーバーが非2xxを返した → `slog.Warn` で警告出力、メッセージ破棄
3. リトライは行わない

### OpenTelemetryトレース

- フラッシュ時に span `middleware.forward_http` を作成する
- W3C Trace Contextヘッダー（`traceparent`, `tracestate`）をHTTPリクエストに注入する

### 設定

| 環境変数 | 説明 | デフォルト |
|---------|------|-----------|
| `WORKFLOW_LENS_SERVER_URL` | 転送先サーバーのベースURL | 未設定（log-onlyモード） |

## テスト方針

- [ ] Forwarderのバッファリング: Send()でメッセージが蓄積される
- [ ] バッファ件数によるフラッシュ: 100件でフラッシュが発動する
- [ ] タイマーによるフラッシュ: 5秒経過でフラッシュが発動する
- [ ] HTTPリクエストの形式: POST、JSON配列ボディ、Content-Type
- [ ] 転送失敗時にクラッシュしない
- [ ] fwd=nilでlog-onlyモードとして動作する
- [ ] Shutdownで残バッファがフラッシュされる

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
