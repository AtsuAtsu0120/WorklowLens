use std::net::{SocketAddr, UdpSocket};

use crate::error::AppError;
use crate::message::LogMessage;

/// 受信バッファサイズ（64 KiB）
const MAX_DATAGRAM_SIZE: usize = 64 * 1024;

/// UDPサーバーを起動し、データグラムを受信する
///
/// C#での以下のコードに相当:
/// ```csharp
/// var udpClient = new UdpClient(port);
/// while (true) {
///     var remoteEP = new IPEndPoint(IPAddress.Any, 0);
///     byte[] data = udpClient.Receive(ref remoteEP);
///     ProcessDatagram(data, remoteEP);
/// }
/// ```
///
/// TCPと違い、コネクションの概念がないためスレッド生成は不要。
/// 1つのソケットで全クライアントからのデータグラムを受信する。
pub fn run(addr: SocketAddr) -> Result<(), AppError> {
    let socket = UdpSocket::bind(addr)?;
    println!("サーバー起動: {addr} (UDP)");
    println!("Ctrl+C で終了");

    let mut buf = [0u8; MAX_DATAGRAM_SIZE];

    // `recv_from()` はデータグラムが届くまでブロックする。
    // C#の `UdpClient.Receive()` と同じ。
    // TCPの `accept()` + `read_line()` の2段階が、UDPでは `recv_from()` の1段階で済む。
    loop {
        match socket.recv_from(&mut buf) {
            Ok((len, src)) => {
                process_datagram(&buf[..len], src);
            }
            Err(err) => {
                eprintln!("[エラー] データグラム受信失敗: {err}");
            }
        }
    }
}

/// 1つのデータグラムを処理する
///
/// TCPの `connection::handle_connection()` が行っていた処理を簡略化したもの。
/// TCPでは BufReader + read_line で行単位に読み取っていたが、
/// UDPでは1データグラム = 1メッセージなのでバッファリング不要。
fn process_datagram(data: &[u8], src: SocketAddr) {
    // UTF-8バリデーション
    //
    // TCPの `read_line()` は `String` に直接読み込むため暗黙的にUTF-8前提だったが、
    // UDPの `recv_from()` は `&[u8]`（バイト列）を返すため、明示的な変換が必要。
    //
    // C#では `Encoding.UTF8.GetString(data)` に相当。
    // ただしC#の `string` はUTF-16なので変換は常に行われるが、
    // Rustの `&str` はUTF-8でなければならないため、不正なバイト列を検出できる。
    let text = match std::str::from_utf8(data) {
        Ok(s) => s,
        Err(err) => {
            eprintln!("[警告] {src}: 不正なUTF-8データグラム: {err}");
            return;
        }
    };

    let trimmed = text.trim();
    if trimmed.is_empty() {
        return;
    }

    // JSONパース
    match LogMessage::parse(trimmed) {
        Ok(msg) => {
            println!("[受信] {src}: {msg:?}");
        }
        Err(err) => {
            let preview = if trimmed.len() > 100 {
                &trimmed[..100]
            } else {
                trimmed
            };
            eprintln!("[警告] {src}: JSONパースエラー: {err} | 受信データ: {preview}");
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 正常なjsonデータグラムを処理できる() {
        // process_datagram はstdout/stderrに出力するだけなので、
        // パニックしないことを確認するテスト。
        let json = r#"{"tool_name":"TestTool","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"hello"}"#;
        let addr: SocketAddr = "127.0.0.1:12345".parse().unwrap();
        process_datagram(json.as_bytes(), addr);
    }

    #[test]
    fn 不正なjsonでもパニックしない() {
        let addr: SocketAddr = "127.0.0.1:12345".parse().unwrap();
        process_datagram(b"this is not json", addr);
    }

    #[test]
    fn 空データグラムを無視する() {
        let addr: SocketAddr = "127.0.0.1:12345".parse().unwrap();
        process_datagram(b"", addr);
        process_datagram(b"   \n  ", addr);
    }

    #[test]
    fn 不正なutf8を警告してスキップする() {
        let addr: SocketAddr = "127.0.0.1:12345".parse().unwrap();
        // 不正なUTF-8バイト列
        let invalid_utf8 = &[0xff, 0xfe, 0xfd];
        process_datagram(invalid_utf8, addr);
        // パニックしなければ成功
    }
}
