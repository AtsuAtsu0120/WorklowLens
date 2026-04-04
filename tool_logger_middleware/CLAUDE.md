# tool_logger

ツール利用のログを記録するRustプロジェクト。

## ビルド・実行

```bash
cargo build          # ビルド
cargo run            # 実行
cargo test           # テスト実行
cargo clippy         # lint
cargo fmt            # フォーマット
```

## プロジェクト構成

- `src/main.rs` — エントリポイント
- Rust Edition: 2024

## コーディング規約

- `cargo fmt` のデフォルト設定に従う
- `cargo clippy` の警告をすべて解消する
- コメントやドキュメントは日本語で記述する
- テストコードはteste/以下に記載してください

## ユーザーコンテキスト

このプロジェクトのユーザーはC#に精通しているがRustは初心者である。コード説明時は以下を意識すること:

- C#との対比で説明すると理解しやすい（例: `trait` = C#の`interface`、`enum` = C#の判別共用体に近い、`Result<T,E>` = 例外の代わり、`Option<T>` = `Nullable<T>`に近い）
- 所有権・借用・ライフタイムなどRust固有の概念は丁寧に補足する
- `unsafe`、マクロ、高度なジェネリクスなどは必要最小限に留め、初学者でも読みやすいコードを優先する
- このプロジェクトを進めつつRustの学習も深めたいので、細かい単位でいろいろ教えて
