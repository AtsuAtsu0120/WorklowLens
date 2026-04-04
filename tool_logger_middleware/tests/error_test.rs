// AppError の単体テスト

use tool_logger::error::AppError;

#[test]
fn io_errorからapperrorに変換できる() {
    // `From` トレイトのテスト。
    // C# の暗黙的型変換（implicit operator）が正しく動くか確認するのと同じ。
    let io_err = std::io::Error::new(std::io::ErrorKind::ConnectionRefused, "接続拒否");
    let app_err = AppError::from(io_err);

    // `matches!` マクロは C# の `is` パターンマッチに相当。
    // `app_err is AppError.Io` のようなイメージ。
    assert!(matches!(app_err, AppError::Io(_)));
}

#[test]
fn json_errorからapperrorに変換できる() {
    // 不正なJSONをパースしてエラーを意図的に生成する
    let json_err = serde_json::from_str::<serde_json::Value>("invalid").unwrap_err();
    let app_err = AppError::from(json_err);

    assert!(matches!(app_err, AppError::Json(_)));
}

#[test]
fn io_errorのdisplayメッセージ() {
    // `Display` トレイトのテスト。C# の `ToString()` のテストに相当。
    // `format!("{err}")` が `err.ToString()` と同じ。
    let io_err = std::io::Error::new(std::io::ErrorKind::AddrInUse, "ポート使用中");
    let app_err = AppError::Io(io_err);
    let msg = format!("{app_err}");

    assert!(msg.contains("I/Oエラー"));
    assert!(msg.contains("ポート使用中"));
}

#[test]
fn json_errorのdisplayメッセージ() {
    let json_err = serde_json::from_str::<serde_json::Value>("{invalid}").unwrap_err();
    let app_err = AppError::Json(json_err);
    let msg = format!("{app_err}");

    assert!(msg.contains("JSONパースエラー"));
}

#[test]
fn datagram_too_largeのdisplayメッセージ() {
    let app_err = AppError::DatagramTooLarge {
        length: 100_000,
        max: 65_536,
    };
    let msg = format!("{app_err}");

    assert!(msg.contains("100000"));
    assert!(msg.contains("65536"));
    assert!(msg.contains("データグラムが大きすぎます"));
}

#[test]
fn question演算子でio_errorを自動変換できる() {
    // `?` 演算子（C# にはない概念）のテスト。
    // 関数内で `io::Error` が発生したとき、`?` を付けると自動的に `AppError::Io` に変換される。
    // これは `From` トレイトの実装があるから動く仕組み。
    fn might_fail() -> Result<(), AppError> {
        let result: Result<(), std::io::Error> =
            Err(std::io::Error::new(std::io::ErrorKind::NotFound, "テスト"));
        result?; // ここで io::Error → AppError::Io に自動変換
        Ok(())
    }

    let err = might_fail().unwrap_err();
    assert!(matches!(err, AppError::Io(_)));
}

#[test]
fn question演算子でjson_errorを自動変換できる() {
    fn parse_something() -> Result<(), AppError> {
        let _value: serde_json::Value = serde_json::from_str("invalid")?;
        Ok(())
    }

    let err = parse_something().unwrap_err();
    assert!(matches!(err, AppError::Json(_)));
}
