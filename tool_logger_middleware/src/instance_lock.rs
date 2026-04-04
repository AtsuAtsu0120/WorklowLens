use std::fmt;
use std::net::{SocketAddr, TcpListener};

/// 多重起動防止に使用するデフォルトポート
///
/// メインサーバーのポート（59100）とは別に、このポートを「ロック」として使う。
/// ポートがバインドできれば自分が最初のインスタンス、できなければ既に起動中と判定する。
const DEFAULT_LOCK_PORT: u16 = 59099;

/// インスタンスロック — 多重起動を防止する仕組み
///
/// C#での多重起動防止は `Mutex` を使うのが一般的:
/// ```csharp
/// using var mutex = new Mutex(true, "tool_logger", out bool createdNew);
/// if (!createdNew) {
///     Console.Error.WriteLine("既に起動中です");
///     return;
/// }
/// ```
///
/// Rustの標準ライブラリには名前付きMutexがないため、
/// TCPポートのバインドを排他制御の手段として代用する。
///
/// ## 重要: `_lock` vs `_`
///
/// ```rust,ignore
/// let _lock = InstanceLock::acquire().unwrap();  // main()終了まで保持される
/// let _ = InstanceLock::acquire().unwrap();       // 即座にDropされ、ロック解放！
/// ```
///
/// `_lock`（名前付き変数）はスコープ終了まで生き続けるが、
/// `_`（アンダースコアのみ）は即座にDropされる。
/// C#でいうと `using var m = ...` と `new Mutex(...)` を変数に入れない違いに近い。
pub struct InstanceLock {
    /// リスナーを保持し続けることでポートを占有する。
    /// フィールド名の `_` プレフィックスは「直接使わないが保持が必要」という慣習。
    /// C#では `[SuppressMessage]` や `#pragma warning disable` で未使用警告を抑制するが、
    /// Rustでは `_` プレフィックスが同じ役割を果たす。
    _listener: TcpListener,
}

impl InstanceLock {
    /// デフォルトポート（59099）でロックを取得する
    ///
    /// 成功すれば `InstanceLock` を返す。
    /// 既に別のインスタンスが起動中なら `AlreadyRunningError` を返す。
    pub fn acquire() -> Result<Self, AlreadyRunningError> {
        Self::acquire_on_port(DEFAULT_LOCK_PORT)
    }

    /// 指定ポートでロックを取得する（テスト用）
    ///
    /// `acquire()` と同じだが、ポートを指定できる。
    /// テストでは port 0（OS自動割り当て）を使うことで、
    /// テストの並列実行時にポートが衝突しないようにする。
    pub fn acquire_on_port(port: u16) -> Result<Self, AlreadyRunningError> {
        let addr = SocketAddr::from(([127, 0, 0, 1], port));
        match TcpListener::bind(addr) {
            Ok(listener) => Ok(InstanceLock {
                _listener: listener,
            }),
            Err(_) => Err(AlreadyRunningError),
        }
    }

    /// ロックに使用しているポート番号を返す
    ///
    /// テストで `acquire_on_port(0)` を使った場合、
    /// OSが実際に割り当てたポートを知るために使う。
    pub fn port(&self) -> u16 {
        self._listener
            .local_addr()
            .expect("ロックポートのアドレス取得に失敗")
            .port()
    }
}

/// 既に別のインスタンスが起動中であることを示すエラー
///
/// C#のカスタム例外クラスに相当:
/// ```csharp
/// public class AlreadyRunningException : Exception
/// {
///     public override string Message => "既に起動中です";
/// }
/// ```
#[derive(Debug)]
pub struct AlreadyRunningError;

impl fmt::Display for AlreadyRunningError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "tool_loggerは既に起動中です。複数のインスタンスを同時に実行することはできません。"
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn ロックを取得できる() {
        // port 0 を指定すると、OSが空いているポートを自動割り当てする。
        // テストの並列実行でポートが衝突しないようにするテクニック。
        let lock = InstanceLock::acquire_on_port(0).expect("ロック取得に失敗");
        // OSが割り当てたポートは0ではないはず
        assert_ne!(lock.port(), 0);
    }

    #[test]
    fn 二重ロックはエラーになる() {
        let lock1 = InstanceLock::acquire_on_port(0).expect("1つ目のロック取得に失敗");
        let port = lock1.port();

        // 同じポートで2つ目のロックを取得しようとするとエラーになる
        let result = InstanceLock::acquire_on_port(port);
        assert!(
            result.is_err(),
            "同じポートで2つ目のロックが取得できてしまった"
        );
    }

    #[test]
    fn ロック解放後に再取得できる() {
        let lock = InstanceLock::acquire_on_port(0).expect("ロック取得に失敗");
        let port = lock.port();

        // ロックを解放（C#の `Dispose()` に相当）
        // `drop()` は明示的にオブジェクトを破棄する関数。
        // 通常はスコープ終了時に自動でDropされるが、テストでは明示的に呼ぶ。
        drop(lock);

        // 解放後は同じポートで再取得できる
        let _lock2 = InstanceLock::acquire_on_port(port).expect("ロック解放後の再取得に失敗");
    }
}
