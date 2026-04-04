use std::fmt;

/// アプリケーションのエラー型
///
/// C#ではカスタム例外クラス（`IOException`, `JsonException` など）を作るが、
/// Rustでは `enum` の各バリアントがそれぞれの例外に対応する。
/// `match` 式で分岐するのが、C#の `catch (IOException)` / `catch (JsonException)` に相当。
#[derive(Debug)]
#[allow(dead_code)]
pub enum AppError {
    /// ネットワークI/Oエラー（C#の `IOException` に相当）
    Io(std::io::Error),
    /// JSONパースエラー（C#の `JsonException` に相当）
    Json(serde_json::Error),
    /// データグラムが大きすぎる
    DatagramTooLarge { length: usize, max: usize },
    /// 多重起動エラー（別のインスタンスが既に起動中）
    AlreadyRunning,
}

/// `Display` トレイトの実装
///
/// C#の `ToString()` オーバーライドに相当。
/// エラーメッセージを人間が読める形式で表示するために必要。
impl fmt::Display for AppError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AppError::Io(err) => write!(f, "I/Oエラー: {err}"),
            AppError::Json(err) => write!(f, "JSONパースエラー: {err}"),
            AppError::DatagramTooLarge { length, max } => {
                write!(f, "データグラムが大きすぎます: {length}バイト（上限: {max}バイト）")
            }
            AppError::AlreadyRunning => {
                write!(
                    f,
                    "tool_loggerは既に起動中です。複数のインスタンスを同時に実行することはできません。"
                )
            }
        }
    }
}

/// `From<std::io::Error>` の実装
///
/// これにより `?` 演算子で `io::Error` を自動的に `AppError::Io` に変換できる。
/// C#では暗黙的な型変換（implicit operator）に近い概念。
/// 例: `let data = stream.read(...)?;` で、I/Oエラーが自動的に `AppError` に変換される。
impl From<std::io::Error> for AppError {
    fn from(err: std::io::Error) -> Self {
        AppError::Io(err)
    }
}

/// `From<serde_json::Error>` の実装
impl From<serde_json::Error> for AppError {
    fn from(err: serde_json::Error) -> Self {
        AppError::Json(err)
    }
}

/// `From<AlreadyRunningError>` の実装
///
/// これにより `?` 演算子で `AlreadyRunningError` を `AppError::AlreadyRunning` に変換できる。
impl From<crate::instance_lock::AlreadyRunningError> for AppError {
    fn from(_: crate::instance_lock::AlreadyRunningError) -> Self {
        AppError::AlreadyRunning
    }
}
