# tool_logger_middleware

ゲーム開発ツール（Unity, Maya等）のログをUDPで受信するローカル中継サーバー。

## ビルド・テスト

```bash
go build ./cmd/middleware      # ビルド
go run ./cmd/middleware        # 実行（デフォルト: 59100）
go run ./cmd/middleware 59200  # ポート指定
go test ./...                  # 全テスト
go vet ./...                   # 静的解析
```

## プロジェクト構成

```
cmd/middleware/main.go          エントリポイント
internal/
  model/log_message.go          LogMessage, Parse(), ValidateEventType()
  server/server.go              UDPサーバー (ReadFromループ, graceful shutdown)
  lock/instance_lock.go         多重起動防止 (TCPポートバインド方式)
Document/                       仕様書 (SDD: 仕様書駆動設計)
```

## 仕様書駆動設計 (SDD)

ルートの `CLAUDE.md` を参照。機能の実装・修正前に必ず `Document/features/` の該当仕様書を確認すること。

## コーディング規約

- `gofmt` + `go vet` 準拠
- コメント・ドキュメントは日本語
- 外部依存なし（標準ライブラリのみ）
- テストは各パッケージ内の `_test.go` に配置

## ユーザーコンテキスト

C#に精通、Goもある程度わかる。冗長な言語基礎の説明は不要。
