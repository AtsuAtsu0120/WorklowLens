# workflow_lens_server

ログ保存・閲覧のWebサーバー（Go + database/sql）。

## ビルド・テスト

```bash
go build ./cmd/server         # ビルド（SQLiteデフォルト）
go run ./cmd/server           # 実行（ゼロ設定でSQLite使用）
go test ./...                 # 全テスト
go vet ./...                  # 静的解析
```

## プロジェクト構成

```
cmd/server/main.go             エントリポイント
internal/
  handler/logs.go               POST /logs ハンドラ
  model/log_message.go          LogMessage, Category
  store/sql_store.go            DB接続・INSERT（database/sql汎用実装）
  store/dialect.go              SQL方言定義（SQLite/PostgreSQL/MySQL）
  store/driver_default.go       デフォルトSQLiteドライバ
  store/driver_postgres.go      PostgreSQLドライバ（ビルドタグ）
  store/driver_mysql.go         MySQLドライバ（ビルドタグ）
Document/                       仕様書 (SDD: 仕様書駆動設計)
```

## 仕様書駆動設計 (SDD)

ルートの `CLAUDE.md` を参照。機能の実装・修正前に必ず `Document/features/` の該当仕様書を確認すること。

## コーディング規約

- `gofmt` + `go vet` 準拠
- コメント・ドキュメントは日本語
- DBドライバは `database/sql` 標準インターフェースを使用（ドライバはビルドタグで切り替え）
- テストは各パッケージ内の `_test.go` に配置
