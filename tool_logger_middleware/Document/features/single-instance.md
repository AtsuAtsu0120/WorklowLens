---
title: "多重起動防止"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - internal/lock/instance_lock.go
  - cmd/middleware/main.go
---

# 多重起動防止

## 概要

tool_loggerの多重起動を防止し、同一マシンで常に1インスタンスのみ動作することを保証する。

## 背景・目的

tool_loggerはローカルの中継ツールであり、同じマシンで複数起動する必要がない。誤って2つ目を起動すると、同じポートの取り合いやログの分散が発生する。

## 要件

- [x] 同じマシンで2つ目のtool_loggerを起動しようとしたとき、分かりやすいエラーメッセージを表示して終了する
- [x] メインポートを変えても多重起動を検出できる（例: 1つ目が59100、2つ目が59200）
- [x] プロセスが異常終了（クラッシュ、kill）した場合、ロックが自動解放される（手動クリーンアップ不要）
- [x] 標準ライブラリのみで実装する

## 設計

### 方式: ロックポート方式

メインポートとは別に、ロック専用のTCPポート（デフォルト: 59099）をバインドする。2つ目のインスタンスが同じポートをバインドしようとすると失敗するため、「既に起動中」と判定できる。

GoではC#の名前付きMutexに相当する機能がないため、「ポートのバインド」を排他制御の手段として代用する。ポートはプロセス終了時にOSが自動解放するため、PIDファイルのような「クラッシュ後にゴミが残る」問題がない。

### モジュール構成

```
internal/lock/instance_lock.go  → InstanceLock構造体、ロック取得/解放
cmd/middleware/main.go           → サーバー起動前にロック取得
```

### InstanceLock (`internal/lock/instance_lock.go`)

```go
// InstanceLock はインスタンスロック。内部でTCPリスナーを保持し、
// Close()されるまでポートを占有する。
type InstanceLock struct {
    listener net.Listener
}

// Acquire はデフォルトポート（59099）でロックを取得する。
func Acquire() (*InstanceLock, error)

// AcquireOnPort は指定ポートでロックを取得する（テスト用）。
func AcquireOnPort(port int) (*InstanceLock, error)

// Port はロックに使用しているポート番号を返す。
func (l *InstanceLock) Port() int

// Close はロックを解放する。
func (l *InstanceLock) Close() error
```

Goでは`defer lock.Close()`で関数終了時に自動解放する:

```go
func main() {
    lock, err := lock.Acquire()
    if err != nil {
        slog.Error("起動エラー", "error", err)
        os.Exit(1)
    }
    defer lock.Close()
    // ... サーバー起動
}
```

### エラー

```go
// ErrAlreadyRunning は既に別インスタンスが起動中であることを示すエラー。
var ErrAlreadyRunning = errors.New(
    "tool_loggerは既に起動中です。複数のインスタンスを同時に実行することはできません。",
)
```

## テスト方針

- [x] ユニットテスト: ロックを取得できる
- [x] ユニットテスト: 二重ロックはエラーになる
- [x] ユニットテスト: ロック解放後に再取得できる

テストではポート0（OS自動割り当て）を使用し、テストの並列実行で競合しないようにする。

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
| 2026-04-04 | Go版に更新 |
