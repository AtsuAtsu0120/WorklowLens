# tool_logger

ゲーム開発ツール（Unity、Maya等）の利用ログ・エラーログを収集するローカルTCPサーバーです。
ツールからJSON形式のログをTCPで受信し、標準出力に表示します。

## システム構成

```
┌─────────────────────┐
│ ゲーム開発ツール     │
│ (Unity, Maya, etc.) │
└──────────┬──────────┘
           │ TCP + NDJSON (localhost:59100)
           ▼
┌──────────────────────┐
│   tool_logger        │  ← このプロジェクト
│ (ローカルサーバー)    │
└──────────┬───────────┘
           │ HTTP (将来実装予定)
           ▼
┌──────────────────────┐
│  オンラインサーバー   │
└──────────────────────┘
```

## ビルド

```bash
cargo build --release
```

リリースバイナリは `target/release/tool_logger` に生成されます。

## 起動方法

```bash
# デフォルトポート (59100) で起動
./target/release/tool_logger

# ポートを指定して起動
./target/release/tool_logger 8080
```

起動すると `127.0.0.1:59100` でTCP接続を待ち受けます。

## ログメッセージ仕様

TCP接続でNDJSON（改行区切りJSON）を送信します。1行が1つのログメッセージです。

### メッセージ形式

```json
{
  "tool_name": "UnityTerrainEditor",
  "event_type": "usage",
  "timestamp": "2026-04-04T10:30:00Z",
  "message": "Terrain brush applied",
  "details": { "brush_size": 5, "brush_type": "raise" }
}
```

### フィールド

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `tool_name` | string | Yes | ツール名（例: `"UnityTerrainEditor"`） |
| `event_type` | string | Yes | `"usage"` または `"error"` |
| `timestamp` | string | Yes | ISO 8601形式（例: `"2026-04-04T10:30:00Z"`） |
| `message` | string | Yes | ログメッセージ本文 |
| `details` | object/null | No | 任意の追加情報。省略可。 |

### 制約

- 1行の最大サイズ: **64 KiB**
- エンコーディング: **UTF-8**
- 行末は `\n`（改行）で区切る

## ツールへの組み込み

ツールへの同伴方法やクライアント実装例は [tool_logger_middleware/README.md](./tool_logger_middleware/README.md) を参照してください。

## 開発

```bash
cargo build          # ビルド
cargo run            # 実行（デフォルトポート）
cargo run -- 8080    # ポート指定で実行
cargo test           # テスト実行
cargo clippy         # lint
cargo fmt            # フォーマット
```

詳細な仕様は [Document/](./tool_logger_middleware/Document/) ディレクトリを参照してください。
