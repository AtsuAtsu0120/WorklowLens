---
title: "UDPサーバー基盤"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - cmd/middleware/main.go
  - internal/server/server.go
---

# UDPサーバー基盤

## 概要

localhostでUDPデータグラムを待ち受け、クライアントからのJSONメッセージを受信・パースしてstdoutに出力するサーバー。

## 背景・目的

ゲーム開発ツール（Unity, Maya等）からのログ送信は fire-and-forget（送りっぱなし）の性質であり、TCPのコネクション管理はオーバーヘッドとなる。UDPに変更することで:

- クライアント側: `UdpClient.Send()` 一発で送信完了（接続・切断の管理不要）
- サーバー側: シングルgoroutineの `ReadFrom` ループでシンプルに実装可能

localhost通信ではUDPのパケットロスは実質ゼロなので、信頼性の問題もない。

## 要件

- [x] 指定ポート（デフォルト: 59100）でUDPデータグラムを待ち受ける
- [x] 1データグラム = 1 JSONメッセージとして処理する
- [x] 受信したメッセージをパースしてstdoutに表示する
- [x] 受信バッファサイズは64KiB（UDPデータグラムの実質上限）
- [x] 不正なJSON受信時もサーバーがクラッシュしない
- [x] 不正なUTF-8データグラムを警告してスキップする
- [x] コマンドライン引数でポートを変更できる
- [x] context cancelによるgraceful shutdownをサポートする

## 設計

### モジュール構成

```
cmd/middleware/main.go        → ポート設定、server.Run()呼び出し
internal/server/server.go     → UDPソケット、ReadFromループ、データグラム処理
internal/model/log_message.go → LogMessage、イベント種別（別仕様: log-message.md）
internal/lock/instance_lock.go → 多重起動防止（別仕様: single-instance.md）
```

### サーバー (`internal/server/server.go`)

```go
const MaxDatagramSize = 64 * 1024 // 64 KiB

// Run はUDPサーバーを起動し、データグラムを受信する。
// ctxがキャンセルされるとソケットを閉じて終了する。
func Run(ctx context.Context, addr string) error
```

処理フロー:
1. `net.ListenPacket("udp", addr)` でソケットをバインド
2. goroutineでctx.Done()を監視し、キャンセル時にconn.Close()
3. `ReadFrom()` でデータグラムをループ受信
4. 各データグラムを `processDatagram()` で処理

```go
// processDatagram は1つのデータグラムを処理する。
func processDatagram(data []byte, src net.Addr)
```

処理フロー:
1. `utf8.Valid(data)` でUTF-8バリデーション
2. `bytes.TrimSpace()` して空なら無視
3. `model.Parse()` でパース
4. 成功 → slog.Info出力、失敗 → slog.Warn出力

### 依存パッケージ

標準ライブラリのみ:

| パッケージ | 用途 |
|-----------|------|
| `net` | UDPソケット |
| `context` | graceful shutdown |
| `log/slog` | 構造化ログ |
| `unicode/utf8` | UTF-8バリデーション |

## テスト方針

- [x] 結合テスト: UDPでデータグラム送信 → サーバーが正常に処理
- [x] 正常なJSONデータグラムの送受信
- [x] 複数メッセージの連続送信
- [x] 不正なJSONでもクラッシュしない
- [x] 空データグラムを無視する
- [x] 複数クライアントからの送信
- [x] detailsを含むメッセージの送受信
- [x] セッションライフサイクル（start → usage → end）
- [x] キャンセルイベントの送受信
- [x] context cancelによるgraceful shutdown

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-04 | Go版に更新 |
