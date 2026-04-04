---
title: "UDPサーバー基盤"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - src/main.rs
  - src/server.rs
  - src/error.rs
---

# UDPサーバー基盤

## 概要

localhostでUDPデータグラムを待ち受け、クライアントからのJSONメッセージを受信・パースしてstdoutに出力するサーバー。

## 背景・目的

ゲーム開発ツール（Unity, Maya等）からのログ送信は fire-and-forget（送りっぱなし）の性質であり、TCPのコネクション管理はオーバーヘッドとなる。UDPに変更することで:

- クライアント側: `UdpClient.Send()` 一発で送信完了（接続・切断の管理不要）
- サーバー側: シングルスレッドの `recv_from` ループでシンプルに実装可能
- スレッド生成のコストがゼロ

localhost通信ではUDPのパケットロスは実質ゼロなので、信頼性の問題もない。

## 要件

- [x] 指定ポート（デフォルト: 59100）でUDPデータグラムを待ち受ける
- [x] 1データグラム = 1 JSONメッセージとして処理する
- [x] 受信したメッセージをパースしてstdoutに表示する
- [x] 64KiBを超えるデータグラムは警告を出す
- [x] 不正なJSON受信時もサーバーがクラッシュしない
- [x] 不正なUTF-8データグラムを警告してスキップする
- [x] コマンドライン引数でポートを変更できる

## 設計

### モジュール構成

```
src/main.rs        → ポート設定、server::run()呼び出し
src/server.rs      → UdpSocket、recv_fromループ、データグラム処理
src/message.rs     → LogMessage、EventType（別仕様: log-message.md）
src/error.rs       → AppError型
```

TCPの `connection.rs`（接続ハンドラ）はUDPでは不要。UDPはコネクションレスなので、1つのソケットで全クライアントからのデータグラムを受信する。

### サーバー (`src/server.rs`)

```rust
/// UDPサーバーを起動し、データグラムを受信する
pub fn run(addr: SocketAddr) -> Result<(), AppError>
```

処理フロー:
1. `UdpSocket::bind(addr)` でソケットをバインド
2. `recv_from()` でデータグラムをループ受信
3. 各データグラムを `process_datagram()` で処理

**C#との対比**: `new UdpClient(port)` → `UdpClient.Receive(ref remoteEP)` ループに相当。TCPと違いスレッド生成不要。C#でも `UdpClient` はシングルスレッドで使える。

```rust
/// 1つのデータグラムを処理する
fn process_datagram(data: &[u8], src: SocketAddr)
```

処理フロー:
1. `std::str::from_utf8(data)` でUTF-8バリデーション（C#では不要な手順。C#の `string` は常にUTF-16だが、Rustの `&str` はUTF-8でなければならない）
2. `trim()` して空なら無視
3. `LogMessage::parse()` でパース
4. 成功 → stdout出力、失敗 → 警告出力

### エラー型 (`src/error.rs`)

```rust
pub enum AppError {
    Io(std::io::Error),
    Json(serde_json::Error),
    /// データグラムが大きすぎる（旧: LineTooLong）
    DatagramTooLarge { length: usize, max: usize },
    AlreadyRunning,
}
```

### 依存クレート

| クレート名 | 用途 | バージョン |
|-----------|------|-----------|
| serde | 構造体のシリアライズ/デシリアライズ | 1 (features = ["derive"]) |
| serde_json | JSONパース | 1 |
| chrono | 日時型 | 0.4 (features = ["serde"]) |

追加の依存クレートは不要。`std::net::UdpSocket` は標準ライブラリに含まれる。

## テスト方針

- [ ] 結合テスト: UdpSocketでデータグラム送信 → サーバーが正常に処理
- [ ] 正常なJSONデータグラムの送受信
- [ ] 不正なJSONでもクラッシュしない
- [ ] 空データグラムを無視する
- [ ] 複数クライアントからの同時受信
- [ ] detailsを含むメッセージの送受信
- [ ] 不正なUTF-8データグラムの無視

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成（TCP版から移行） |
