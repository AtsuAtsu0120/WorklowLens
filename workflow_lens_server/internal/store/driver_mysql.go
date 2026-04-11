//go:build mysql

package store

// MySQLドライバ。
// ビルド時に -tags mysql を指定すると有効になる。
import _ "github.com/go-sql-driver/mysql"
