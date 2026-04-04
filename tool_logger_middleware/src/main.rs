use std::net::SocketAddr;

/// デフォルトのリッスンポート
const DEFAULT_PORT: u16 = 59100;

fn main() {
    // 多重起動防止: ロックを取得（C#の named Mutex に相当）
    //
    // 重要: `_lock` と `_` は全く異なる振る舞いをする！
    //   - `let _lock = ...` → main()終了まで変数が生き続け、ポートを占有し続ける
    //   - `let _ = ...`     → 即座にDropされ、ポートが解放されてしまう
    //
    // C#でいうと:
    //   - `using var mutex = new Mutex(...)` → スコープ終了まで保持（_lock）
    //   - `new Mutex(...)` を変数に入れない  → GC対象になる（_）
    //
    // Rustでは `_` は「パターンマッチの捨て値」であり、値の所有権を受け取らない。
    // 一方 `_lock` は普通の変数なので、スコープ終了時にDropされる。
    let _lock = match tool_logger::instance_lock::InstanceLock::acquire() {
        Ok(lock) => lock,
        Err(e) => {
            eprintln!("起動エラー: {e}");
            std::process::exit(1);
        }
    };

    // コマンドライン引数からポートを取得（省略時はデフォルト値を使用）
    // C#での `args[0]` に相当。Rustでは `std::env::args()` でイテレータとして取得する。
    // `nth(1)` は0始まりの2番目の要素（プログラム名の次）を取得する。
    let port = std::env::args()
        .nth(1)
        .and_then(|s| s.parse::<u16>().ok())
        .unwrap_or(DEFAULT_PORT);

    let addr = SocketAddr::from(([127, 0, 0, 1], port));

    // `if let Err(e)` はC#の `try { ... } catch (Exception e) { ... }` に近いパターン。
    // `server::run` が `Err` を返した場合のみブロック内が実行される。
    if let Err(e) = tool_logger::server::run(addr) {
        eprintln!("サーバーエラー: {e}");
        std::process::exit(1);
    }
}
