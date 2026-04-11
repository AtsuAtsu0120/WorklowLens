package store

import "fmt"

// dialect はDBドライバごとのSQL方言を定義する。
type dialect struct {
	createTableSQL string
	placeholder    func(n int) string // n番目（1始まり）のプレースホルダを返す
}

// getDialect はドライバ名に対応するdialectを返す。
func getDialect(driverName string) (dialect, error) {
	switch driverName {
	case "sqlite", "sqlite3":
		return sqliteDialect(), nil
	case "postgres", "pgx":
		return postgresDialect(), nil
	case "mysql":
		return mysqlDialect(), nil
	default:
		return dialect{}, fmt.Errorf("unsupported driver: %s (supported: sqlite, postgres, mysql)", driverName)
	}
}

func sqliteDialect() dialect {
	return dialect{
		createTableSQL: `CREATE TABLE IF NOT EXISTS logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	tool_name TEXT NOT NULL,
	event_type TEXT NOT NULL,
	timestamp TEXT NOT NULL,
	message TEXT NOT NULL,
	session_id TEXT,
	tool_version TEXT,
	details TEXT,
	received_at TEXT NOT NULL DEFAULT (datetime('now'))
)`,
		placeholder: func(n int) string { return "?" },
	}
}

func postgresDialect() dialect {
	return dialect{
		createTableSQL: `CREATE TABLE IF NOT EXISTS logs (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tool_name TEXT NOT NULL,
	event_type TEXT NOT NULL,
	timestamp TIMESTAMPTZ NOT NULL,
	message TEXT NOT NULL,
	session_id TEXT,
	tool_version TEXT,
	details JSONB,
	received_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		placeholder: func(n int) string { return fmt.Sprintf("$%d", n) },
	}
}

func mysqlDialect() dialect {
	return dialect{
		createTableSQL: `CREATE TABLE IF NOT EXISTS logs (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	tool_name TEXT NOT NULL,
	event_type VARCHAR(255) NOT NULL,
	timestamp DATETIME(6) NOT NULL,
	message TEXT NOT NULL,
	session_id VARCHAR(255),
	tool_version VARCHAR(255),
	details JSON,
	received_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
)`,
		placeholder: func(n int) string { return "?" },
	}
}

// buildInsertSQL はINSERT文を組み立てる。
func (d *dialect) buildInsertSQL() string {
	return fmt.Sprintf(
		"INSERT INTO logs (tool_name, event_type, timestamp, message, session_id, tool_version, details) VALUES (%s, %s, %s, %s, %s, %s, %s)",
		d.placeholder(1), d.placeholder(2), d.placeholder(3), d.placeholder(4),
		d.placeholder(5), d.placeholder(6), d.placeholder(7),
	)
}
