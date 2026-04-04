// ライブラリクレートのルート
//
// Rustでは `main.rs`（バイナリクレート）の中身は外部から参照できない。
// テストや将来のクライアントライブラリから使えるように、
// モジュールを `lib.rs`（ライブラリクレート）で公開する。
//
// C#でいうと、`Program.cs` から共通ロジックをクラスライブラリに分離するのと同じ発想。

pub mod error;
pub mod instance_lock;
pub mod message;
pub mod server;
