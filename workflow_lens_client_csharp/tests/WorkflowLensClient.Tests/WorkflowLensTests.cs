using System;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using Xunit;

namespace WorkflowLensClient.Tests
{
    public class WorkflowLensTests
    {
        /// <summary>空きポートを取得するヘルパー。</summary>
        private static int GetFreePort()
        {
            using var listener = new UdpClient(new IPEndPoint(IPAddress.Loopback, 0));
            return ((IPEndPoint)listener.Client.LocalEndPoint!).Port;
        }

        [Fact]
        public void Log_UDPで正しいJSONが届く()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", "1.0.0", port: port, autoStartMiddleware: false);
            logger.Log(Category.Edit, "brush_apply");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("test_tool", root.GetProperty("tool_name").GetString());
            Assert.Equal("edit", root.GetProperty("category").GetString());
            Assert.Equal("brush_apply", root.GetProperty("action").GetString());
            Assert.Equal("1.0.0", root.GetProperty("tool_version").GetString());
            Assert.True(root.TryGetProperty("session_id", out _));
            Assert.True(root.TryGetProperty("user_id", out _));
        }

        [Fact]
        public void Log_durationMsが正しく設定される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", port: port, autoStartMiddleware: false);
            logger.Log(Category.Build, "compile", durationMs: 3200);

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal(3200, doc.RootElement.GetProperty("duration_ms").GetInt64());
        }

        [Fact]
        public void Log_middleware未起動でも例外を投げない()
        {
            using var logger = new WorkflowLens("test_tool", port: 59199, autoStartMiddleware: false);
            var ex = Record.Exception(() => logger.Log(Category.Edit, "test"));
            Assert.Null(ex);
        }

        [Fact]
        public void Log_Dispose後でも例外を投げない()
        {
            var logger = new WorkflowLens("test_tool", port: 59199, autoStartMiddleware: false);
            logger.Dispose();
            var ex = Record.Exception(() => logger.Log(Category.Edit, "test"));
            Assert.Null(ex);
        }

        [Fact]
        public void MeasureScope_durationMsが自動設定される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", port: port, autoStartMiddleware: false);
            using (logger.MeasureScope(Category.Build, "compile"))
            {
                Thread.Sleep(50);
            }

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("build", root.GetProperty("category").GetString());
            Assert.Equal("compile", root.GetProperty("action").GetString());
            Assert.True(root.GetProperty("duration_ms").GetInt64() >= 40);
        }

        [Fact]
        public void userId未指定時にOSユーザー名が自動設定される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", port: port, autoStartMiddleware: false);
            logger.Log(Category.Edit, "test");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            var userId = doc.RootElement.GetProperty("user_id").GetString();
            Assert.Equal(Environment.UserName, userId);
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
            using var logger = new WorkflowLens("test_tool", port: 59199, autoStartMiddleware: false);
            Assert.NotNull(logger.SessionId);
            Assert.Equal(8, logger.SessionId.Length);
        }

        [Fact]
        public void Session_startログが送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", port: port, autoStartMiddleware: false);
            logger.Log(Category.Session, "start");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal("session", doc.RootElement.GetProperty("category").GetString());
            Assert.Equal("start", doc.RootElement.GetProperty("action").GetString());
        }

        [Fact]
        public void Session_endログが送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", port: port, autoStartMiddleware: false);
            logger.Log(Category.Session, "end");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal("session", doc.RootElement.GetProperty("category").GetString());
            Assert.Equal("end", doc.RootElement.GetProperty("action").GetString());
        }

        [Fact]
        public void Log_session_idが自動付与される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", port: port, autoStartMiddleware: false);
            var expectedId = logger.SessionId;
            logger.Log(Category.Edit, "test");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal(expectedId, doc.RootElement.GetProperty("session_id").GetString());
        }
    }

    public class MiddlewareProcessTests
    {
        [Fact]
        public void middlewarePath未指定でプロセスが起動されない()
        {
            using var logger = new WorkflowLens("test_tool", port: 59199, autoStartMiddleware: false);
            logger.Dispose();
        }

        [Fact]
        public void 存在しないバイナリで例外がスローされる()
        {
            Assert.ThrowsAny<Exception>(() =>
                new WorkflowLens("test_tool", port: 59199,
                    middlewarePath: "/nonexistent/middleware"));
        }

        [Fact]
        public void middlewarePathが空文字でArgumentExceptionがスローされる()
        {
            Assert.Throws<ArgumentException>(() =>
                new WorkflowLens("test_tool", port: 59199, middlewarePath: ""));
        }

        [Fact]
        public void Dispose多重呼び出しで例外が出ない()
        {
            var logger = new WorkflowLens("test_tool", port: 59199, autoStartMiddleware: false);
            logger.Dispose();
            var ex = Record.Exception(() => logger.Dispose());
            Assert.Null(ex);
        }
    }

    public class MiddlewareAutoDiscoveryTests
    {
        [Fact]
        public void autoStartMiddleware_falseでプロセスが起動されない()
        {
            var result = WorkflowLens.ResolveMiddlewarePath(null, false);
            Assert.Null(result);
        }

        [Fact]
        public void middlewarePathが指定されていればそのまま返す()
        {
            var result = WorkflowLens.ResolveMiddlewarePath("/explicit/path", false);
            Assert.Equal("/explicit/path", result);
        }

        [Fact]
        public void middlewarePathが指定されていればautoStartMiddlewareより優先()
        {
            var result = WorkflowLens.ResolveMiddlewarePath("/explicit/path", true);
            Assert.Equal("/explicit/path", result);
        }

        [Fact]
        public void autoStartMiddleware_環境変数が設定されていればそれを使う()
        {
            var original = Environment.GetEnvironmentVariable(WorkflowLens.MiddlewarePathEnvVar);
            try
            {
                Environment.SetEnvironmentVariable(WorkflowLens.MiddlewarePathEnvVar, "/env/var/path");
                var result = WorkflowLens.ResolveMiddlewarePath(null, true);
                Assert.Equal("/env/var/path", result);
            }
            finally
            {
                Environment.SetEnvironmentVariable(WorkflowLens.MiddlewarePathEnvVar, original);
            }
        }

        [Fact]
        public void autoStartMiddleware_バイナリが見つからない場合わかりやすいエラー()
        {
            var original = Environment.GetEnvironmentVariable(WorkflowLens.MiddlewarePathEnvVar);
            try
            {
                Environment.SetEnvironmentVariable(WorkflowLens.MiddlewarePathEnvVar, null);
                var ex = Assert.Throws<InvalidOperationException>(
                    () => WorkflowLens.ResolveMiddlewarePath(null, true));
                Assert.Contains("ミドルウェアバイナリが見つかりません", ex.Message);
                Assert.Contains(WorkflowLens.MiddlewarePathEnvVar, ex.Message);
                Assert.Contains(WorkflowLens.MiddlewareBinaryName, ex.Message);
            }
            finally
            {
                Environment.SetEnvironmentVariable(WorkflowLens.MiddlewarePathEnvVar, original);
            }
        }
    }
}
