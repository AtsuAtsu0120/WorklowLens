package store

// デフォルトでSQLiteドライバを組み込む。
// pure Go実装のため、CGO不要でクロスコンパイルが容易。
import _ "modernc.org/sqlite"
