//go:build postgres

package store

// PostgreSQLドライバ（pgx経由のdatabase/sqlインターフェース）。
// ビルド時に -tags postgres を指定すると有効になる。
import _ "github.com/jackc/pgx/v5/stdlib"
