using System;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Xunit;

namespace ToolLoggerClient.Tests
{
    public class ToolLoggerTests
    {
        /// <summary>空きポートを取得するヘルパー。</summary>
        private static int GetFreePort()
        {
            using var listener = new UdpClient(new IPEndPoint(IPAddress.Loopback, 0));
            return ((IPEndPoint)listener.Client.LocalEndPoint!).Port;
        }

        [Fact]
        public void Send_UDPで正しいJSONが届く()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new ToolLogger("test_tool", "1.0.0", port: port);
            logger.LogUsage("ボタン押下");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("test_tool", root.GetProperty("tool_name").GetString());
            Assert.Equal("usage", root.GetProperty("event_type").GetString());
            Assert.Equal("ボタン押下", root.GetProperty("message").GetString());
            Assert.Equal("1.0.0", root.GetProperty("tool_version").GetString());
            Assert.True(root.TryGetProperty("session_id", out _));
        }

        [Fact]
        public void Send_detailsが正しく埋め込まれる()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new ToolLogger("test_tool", port: port);
            logger.LogUsage("操作", "{\"action\":\"click\",\"target\":\"button_a\"}");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);
            var details = doc.RootElement.GetProperty("details");

            Assert.Equal("click", details.GetProperty("action").GetString());
            Assert.Equal("button_a", details.GetProperty("target").GetString());
        }

        [Fact]
        public void Send_middleware未起動でも例外を投げない()
        {
            // 誰もリッスンしていないポートに送信
            using var logger = new ToolLogger("test_tool", port: 59199);
            var ex = Record.Exception(() => logger.LogUsage("テスト"));
            Assert.Null(ex);
        }

        [Fact]
        public void Send_Dispose後でも例外を投げない()
        {
            var logger = new ToolLogger("test_tool", port: 59199);
            logger.Dispose();
            var ex = Record.Exception(() => logger.LogUsage("テスト"));
            Assert.Null(ex);
        }
    }

    public class SessionTests
    {
        private static int GetFreePort()
        {
            using var listener = new UdpClient(new IPEndPoint(IPAddress.Loopback, 0));
            return ((IPEndPoint)listener.Client.LocalEndPoint!).Port;
        }

        [Fact]
        public void コンストラクタでsession_idが生成される()
        {
            using var logger = new ToolLogger("test_tool", port: 59199);
            Assert.NotNull(logger.SessionId);
            Assert.Equal(8, logger.SessionId.Length);
        }

        [Fact]
        public void StartSession_session_startイベントが送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new ToolLogger("test_tool", port: port);
            logger.StartSession();

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal("session_start", doc.RootElement.GetProperty("event_type").GetString());
        }

        [Fact]
        public void StartSession_新しいsession_idが生成される()
        {
            using var logger = new ToolLogger("test_tool", port: 59199);
            var first = logger.SessionId;
            logger.StartSession();
            var second = logger.SessionId;

            Assert.NotEqual(first, second);
        }

        [Fact]
        public void EndSession_session_endイベントが送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new ToolLogger("test_tool", port: port);
            logger.EndSession();

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal("session_end", doc.RootElement.GetProperty("event_type").GetString());
        }

        [Fact]
        public void Send_session_idが自動付与される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new ToolLogger("test_tool", port: port);
            var expectedId = logger.SessionId;
            logger.LogUsage("テスト");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal(expectedId, doc.RootElement.GetProperty("session_id").GetString());
        }
    }
}
