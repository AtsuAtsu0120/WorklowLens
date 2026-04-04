# tool_logger_server

ログ保存・閲覧のWebサーバー（Go + PostgreSQL）。

## ビルド・テスト

```bash
go build ./cmd/server         # ビルド
go run ./cmd/server           # 実行（要 DATABASE_URL）
go test ./...                 # 全テスト
go vet ./...                  # 静的解析
```

## プロジェクト構成

```
cmd/server/main.go             エントリポイント
internal/
  handler/logs.go               POST /logs ハンドラ
  model/log_message.go          LogMessage, EventType
  store/postgres.go             DB接続・INSERT
Document/                       仕様書 (SDD: 仕様書駆動設計)
```

## 仕様書駆動設計 (SDD)

ルートの `CLAUDE.md` を参照。機能の実装・修正前に必ず `Document/features/` の該当仕様書を確認すること。

## コーディング規約

- `gofmt` + `go vet` 準拠
- コメント・ドキュメントは日本語
- DBドライバは `pgx` を使用
- テストは各パッケージ内の `_test.go` に配置
