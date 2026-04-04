---
title: "ログメッセージ"
status: implemented
priority: high
created: 2026-04-04
updated: 2026-04-04
related_files:
  - src/message.rs
---

# ログメッセージ

## 概要

クライアント（Unity, Maya等のツール）からTCPで送られるログメッセージのデータ構造とプロトコルを定義する。

## 背景・目的

ゲーム開発で使うツールの使用率やエラー情報を統一的に収集するため、言語に依存しないJSONベースのメッセージフォーマットが必要。

## 要件

- [x] JSON形式でメッセージを表現できる
- [x] ツール名、イベント種別（使用/エラー）、タイムスタンプ、メッセージ本文を必須フィールドとする
- [x] 任意の追加情報（details）をフリーフォームで添付できる
- [x] 不正なJSONやフィールド不足時にわかりやすいエラーを返す
- [x] 改行区切り（NDJSON）でフレーミングする

## 設計

### 通信プロトコル

```
トランスポート: TCP（localhost）
フレーミング:   1メッセージ = 1行（改行 '\n' で区切り。'\r\n' も許容）
エンコーディング: UTF-8
最大行長:      64 KiB（超過時はスキップ）
```

クライアントからの送信例:
```json
{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"2026-04-04T10:30:00Z","message":"Terrain brush applied","details":{"brush_size":5}}
```

### データ構造

`EventType` — イベントの種類を表すenum。C#の `enum` に近いが、Rustの `enum` は各バリアントにデータを持てる（今回は持たないシンプルな形）。

```rust
/// ログイベントの種類
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum EventType {
    /// ツールの通常利用
    Usage,
    /// エラー発生
    Error,
}
```

`LogMessage` — 1つのログメッセージを表す構造体。C#でいう `record` や `class` に相当。

```rust
/// クライアントから送られるログメッセージ
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LogMessage {
    /// ツール名（例: "UnityTerrainEditor"）
    pub tool_name: String,
    /// イベント種別
    pub event_type: EventType,
    /// ISO 8601形式のタイムスタンプ
    pub timestamp: DateTime<Utc>,
    /// 人間が読めるメッセージ
    pub message: String,
    /// 任意の追加情報（C#でいうJTokenのようなフリーフォーム）
    #[serde(default)]
    pub details: Option<serde_json::Value>,
}
```

**C#との対比**:
- `#[derive(Deserialize)]` → C#では `JsonSerializer.Deserialize<T>()` で自動的に行われる。Rustではderiveマクロで明示する
- `#[serde(rename_all = "snake_case")]` → C#の `[JsonPropertyName("...")]` に相当。JSONのキー名をスネークケースに統一
- `Option<serde_json::Value>` → C#の `JsonElement?` や `JToken?` に近い。nullまたは任意のJSON値

### 公開API

```rust
impl LogMessage {
    /// JSON文字列からLogMessageをパースする
    pub fn parse(json_str: &str) -> Result<Self, serde_json::Error>;
}
```

### エラーハンドリング

| エラーケース | 対処 |
|-------------|------|
| JSONとしてパースできない | `serde_json::Error` を返す |
| 必須フィールドが欠落 | `serde_json::Error`（missing field）を返す |
| event_typeが未知の値 | `serde_json::Error`（unknown variant）を返す |
| timestampが不正な形式 | `chrono` のパースエラーが `serde_json::Error` として返る |

### 依存クレート

| クレート名 | 用途 | バージョン |
|-----------|------|-----------|
| serde | 構造体のシリアライズ/デシリアライズ | 1 (features = ["derive"]) |
| serde_json | JSONパース | 1 |
| chrono | 日時型（DateTime<Utc>） | 0.4 (features = ["serde"]) |

## テスト方針

- [x] 単体テスト: 正常なJSONのパース
- [x] 単体テスト: details省略時にNoneになること
- [x] 単体テスト: 不正なevent_typeの拒否
- [x] 単体テスト: 必須フィールド欠落の拒否

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-04 | 初版作成 |
