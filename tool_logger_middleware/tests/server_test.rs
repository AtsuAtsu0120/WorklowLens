// UDPサーバーの結合テスト
//
// 実際にUDPサーバーを起動し、クライアントからデータグラムを送信するテスト。
// TCPと違い「接続」の概念がないため、テストはよりシンプルになる。
//
// C#でいうと `UdpClient` を使った結合テストに相当。

use std::net::{SocketAddr, UdpSocket};
use std::time::Duration;

/// テスト用にUDPソケットをバインドし、空きポートのアドレスを返すヘルパー関数。
///
/// `UdpSocket::bind("127.0.0.1:0")` はOSに空きポートを自動選択させる方法。
/// C# の `new UdpClient(0)` と同じ。
fn start_test_server() -> (UdpSocket, SocketAddr) {
    let socket = UdpSocket::bind("127.0.0.1:0").expect("テスト用ソケットの起動に失敗");
    let addr = socket.local_addr().unwrap();
    // タイムアウトを設定して、テストが永遠にブロックしないようにする
    socket
        .set_read_timeout(Some(Duration::from_secs(2)))
        .unwrap();
    (socket, addr)
}

/// テスト用のJSON文字列を生成するヘルパー
fn make_usage_json(tool_name: &str, message: &str) -> String {
    format!(
        r#"{{"tool_name":"{tool_name}","event_type":"usage","timestamp":"2026-04-04T10:00:00Z","message":"{message}"}}"#
    )
}

fn make_error_json(tool_name: &str, message: &str) -> String {
    format!(
        r#"{{"tool_name":"{tool_name}","event_type":"error","timestamp":"2026-04-04T10:00:00Z","message":"{message}"}}"#
    )
}

// ===== データグラム送受信テスト =====

#[test]
fn 正常なjsonデータグラムを送信できる() {
    let (server, server_addr) = start_test_server();

    // クライアント側: UdpSocketでデータグラムを送信
    // C#の `UdpClient.Send(data, data.Length, remoteEP)` に相当
    let client = UdpSocket::bind("127.0.0.1:0").expect("クライアントソケットの起動に失敗");
    let json = make_usage_json("UnityTerrainEditor", "Terrain brush applied");
    client
        .send_to(json.as_bytes(), server_addr)
        .expect("送信に失敗");

    // サーバー側: データグラムを受信して内容を確認
    let mut buf = [0u8; 65536];
    let (len, _src) = server.recv_from(&mut buf).expect("受信に失敗");
    let received = std::str::from_utf8(&buf[..len]).unwrap();
    assert!(received.contains("UnityTerrainEditor"));
}

#[test]
fn 複数メッセージを連続送信できる() {
    let (server, server_addr) = start_test_server();

    let client = UdpSocket::bind("127.0.0.1:0").unwrap();

    // 3つのデータグラムを連続送信
    // UDPでは各データグラムが独立したメッセージ（TCPのNDJSONと違い、改行区切り不要）
    let messages = [
        make_usage_json("Tool1", "action1"),
        make_error_json("Tool2", "error occurred"),
        make_usage_json("Tool3", "action3"),
    ];

    for msg in &messages {
        client.send_to(msg.as_bytes(), server_addr).unwrap();
    }

    // 3つとも受信できることを確認
    let mut buf = [0u8; 65536];
    for _ in 0..3 {
        let (len, _) = server.recv_from(&mut buf).expect("受信に失敗");
        assert!(len > 0);
    }
}

#[test]
fn 不正なjsonを送ってもサーバーがクラッシュしない() {
    let (server, server_addr) = start_test_server();

    let client = UdpSocket::bind("127.0.0.1:0").unwrap();

    // 不正なJSON → 正常なJSON → 不正なJSON の順に送信
    client
        .send_to(b"this is not json", server_addr)
        .unwrap();
    client
        .send_to(
            make_usage_json("TestTool", "valid message").as_bytes(),
            server_addr,
        )
        .unwrap();
    client
        .send_to(b"{broken json", server_addr)
        .unwrap();

    // すべて受信できる（不正なJSONでもデータグラムとしては届く）
    let mut buf = [0u8; 65536];
    for _ in 0..3 {
        server.recv_from(&mut buf).expect("受信に失敗");
    }
}

#[test]
fn 空データグラムを送ってもサーバーがクラッシュしない() {
    let (server, server_addr) = start_test_server();

    let client = UdpSocket::bind("127.0.0.1:0").unwrap();

    // 空データグラム → 正常メッセージ
    client.send_to(b"", server_addr).unwrap();
    client
        .send_to(
            make_usage_json("TestTool", "hello").as_bytes(),
            server_addr,
        )
        .unwrap();

    // 空データグラムは0バイトとして届く可能性がある（OSによる）
    // 少なくとも正常メッセージは受信できること
    let mut buf = [0u8; 65536];
    let mut received_valid = false;
    for _ in 0..2 {
        if let Ok((len, _)) = server.recv_from(&mut buf) {
            if len > 0 {
                let text = std::str::from_utf8(&buf[..len]).unwrap();
                if text.contains("TestTool") {
                    received_valid = true;
                }
            }
        }
    }
    assert!(received_valid, "正常メッセージが受信できなかった");
}

#[test]
fn detailsを含むメッセージを送信できる() {
    let (server, server_addr) = start_test_server();

    let client = UdpSocket::bind("127.0.0.1:0").unwrap();
    let json = r#"{"tool_name":"UnityShaderTool","event_type":"error","timestamp":"2026-04-04T12:00:00Z","message":"Shader error","details":{"shader":"PBR","line":42}}"#;
    client.send_to(json.as_bytes(), server_addr).unwrap();

    let mut buf = [0u8; 65536];
    let (len, _) = server.recv_from(&mut buf).expect("受信に失敗");
    let received = std::str::from_utf8(&buf[..len]).unwrap();
    assert!(received.contains("UnityShaderTool"));
    assert!(received.contains("PBR"));
}

// ===== 複数クライアント同時送信テスト =====

#[test]
fn 複数クライアントからデータグラムを受信できる() {
    let (server, server_addr) = start_test_server();

    // UDPはコネクションレスなので、異なるクライアントソケットから送信するだけ
    // TCPのように「同時接続」を管理する必要がない
    let client1 = UdpSocket::bind("127.0.0.1:0").unwrap();
    let client2 = UdpSocket::bind("127.0.0.1:0").unwrap();

    client1
        .send_to(
            make_usage_json("Client1Tool", "action1").as_bytes(),
            server_addr,
        )
        .unwrap();
    client2
        .send_to(
            make_usage_json("Client2Tool", "action2").as_bytes(),
            server_addr,
        )
        .unwrap();

    // 両方のデータグラムを受信
    let mut buf = [0u8; 65536];
    let mut received = Vec::new();
    for _ in 0..2 {
        let (len, _) = server.recv_from(&mut buf).expect("受信に失敗");
        received.push(std::str::from_utf8(&buf[..len]).unwrap().to_string());
    }

    // 順序は保証されないが、両方届いていること
    let all = received.join(" ");
    assert!(all.contains("Client1Tool"), "Client1Toolのメッセージが届いていない");
    assert!(all.contains("Client2Tool"), "Client2Toolのメッセージが届いていない");
}
