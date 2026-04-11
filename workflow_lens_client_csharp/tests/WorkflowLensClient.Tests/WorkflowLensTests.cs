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

        /// <summary>テスト用ロガーを生成するヘルパー。AutoSession無効、middleware無効。</summary>
        private static WorkflowLens CreateTestLogger(int port, string? toolVersion = null)
        {
            return new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
                o.ToolVersion = toolVersion;
            });
        }

        [Fact]
        public void Log_UDPで正しいJSONが届く()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = CreateTestLogger(port, toolVersion: "1.0.0");
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

            using var logger = CreateTestLogger(port);
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
            using var logger = CreateTestLogger(59199);
            var ex = Record.Exception(() => logger.Log(Category.Edit, "test"));
            Assert.Null(ex);
        }

        [Fact]
        public void Log_Dispose後でも例外を投げない()
        {
            var logger = CreateTestLogger(59199);
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

            using var logger = CreateTestLogger(port);
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

            using var logger = CreateTestLogger(port);
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
            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = 59199;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });
            Assert.NotNull(logger.SessionId);
            Assert.Equal(8, logger.SessionId.Length);
        }

        [Fact]
        public void Session_startログが送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });
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

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });
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

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });
            var expectedId = logger.SessionId;
            logger.Log(Category.Edit, "test");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal(expectedId, doc.RootElement.GetProperty("session_id").GetString());
        }
    }

    public class AutoSessionTests
    {
        private static int GetFreePort()
        {
            using var listener = new UdpClient(new IPEndPoint(IPAddress.Loopback, 0));
            return ((IPEndPoint)listener.Client.LocalEndPoint!).Port;
        }

        [Fact]
        public void AutoSession有効でSession_startが自動送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = true;
            });

            // コンストラクタでSession/startが自動送信されているはず
            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal("session", doc.RootElement.GetProperty("category").GetString());
            Assert.Equal("start", doc.RootElement.GetProperty("action").GetString());
        }

        [Fact]
        public void AutoSession有効でDisposeでSession_endが自動送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = true;
            });

            // startを受信して捨てる
            var ep = new IPEndPoint(IPAddress.Any, 0);
            receiver.Receive(ref ep);

            // Disposeでendが送信される
            logger.Dispose();

            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal("session", doc.RootElement.GetProperty("category").GetString());
            Assert.Equal("end", doc.RootElement.GetProperty("action").GetString());
        }

        [Fact]
        public void AutoSession無効でSession_startが送信されない()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 500;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });

            // タイムアウト = 何も受信しない
            var ep = new IPEndPoint(IPAddress.Any, 0);
            Assert.Throws<SocketException>(() => receiver.Receive(ref ep));
        }

        [Fact]
        public void AutoSession有効で手動start後に重複送信されない()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = true;
            });

            // 自動startを受信
            var ep = new IPEndPoint(IPAddress.Any, 0);
            receiver.Receive(ref ep);

            // 手動startは重複防止で送信されないはず
            logger.Log(Category.Session, "start");

            // 別のログを送信して受信確認
            logger.Log(Category.Edit, "test");
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            // 次に来るのはeditであること（重複startではない）
            Assert.Equal("edit", doc.RootElement.GetProperty("category").GetString());
        }

        [Fact]
        public void 既存コンストラクタではAutoSessionが無効()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 500;

            using var logger = new WorkflowLens("test_tool", null,
                port: port, autoStartMiddleware: false);

            // AutoSession=falseなのでSession/startは送信されない
            var ep = new IPEndPoint(IPAddress.Any, 0);
            Assert.Throws<SocketException>(() => receiver.Receive(ref ep));
        }
    }

    public class CategoryLoggerTests
    {
        private static int GetFreePort()
        {
            using var listener = new UdpClient(new IPEndPoint(IPAddress.Loopback, 0));
            return ((IPEndPoint)listener.Client.LocalEndPoint!).Port;
        }

        [Fact]
        public void CategoryLogger_Logが正しいカテゴリで送信される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });

            using (var edit = logger.Edit())
            {
                edit.Log("brush_apply");
            }

            var ep = new IPEndPoint(IPAddress.Any, 0);
            var data = receiver.Receive(ref ep);
            var json = Encoding.UTF8.GetString(data);
            using var doc = JsonDocument.Parse(json);

            Assert.Equal("edit", doc.RootElement.GetProperty("category").GetString());
            Assert.Equal("brush_apply", doc.RootElement.GetProperty("action").GetString());
        }

        [Fact]
        public void CategoryLogger_action指定で自動計測される()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });

            using (logger.Build("compile"))
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
        public void CategoryLogger_action省略でDispose時にログ送信しない()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 500;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });

            using (var edit = logger.Edit())
            {
                // 中でLogを呼ばない
            }

            // Dispose時に何も送信されないことを確認
            var ep = new IPEndPoint(IPAddress.Any, 0);
            Assert.Throws<SocketException>(() => receiver.Receive(ref ep));
        }

        [Fact]
        public void CategoryLogger_各ファクトリメソッドが正しいカテゴリを返す()
        {
            var port = GetFreePort();
            using var receiver = new UdpClient(new IPEndPoint(IPAddress.Loopback, port));
            receiver.Client.ReceiveTimeout = 3000;

            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = port;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });

            var categories = new[] { "asset", "build", "edit", "error" };
            using (var a = logger.Asset()) a.Log("test");
            using (var b = logger.Build()) b.Log("test");
            using (var e = logger.Edit()) e.Log("test");
            using (var r = logger.Error()) r.Log("test");

            var ep = new IPEndPoint(IPAddress.Any, 0);
            foreach (var expected in categories)
            {
                var data = receiver.Receive(ref ep);
                var json = Encoding.UTF8.GetString(data);
                using var doc = JsonDocument.Parse(json);
                Assert.Equal(expected, doc.RootElement.GetProperty("category").GetString());
            }
        }

        [Fact]
        public void CategoryLogger_Dispose多重呼び出しで例外が出ない()
        {
            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = 59199;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });

            var scope = logger.Build("compile");
            scope.Dispose();
            var ex = Record.Exception(() => scope.Dispose());
            Assert.Null(ex);
        }
    }

    public class MiddlewareProcessTests
    {
        [Fact]
        public void middlewarePath未指定でプロセスが起動されない()
        {
            using var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = 59199;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });
            logger.Dispose();
        }

        [Fact]
        public void 存在しないバイナリで例外がスローされる()
        {
            Assert.ThrowsAny<Exception>(() =>
                new WorkflowLens("test_tool", o =>
                {
                    o.Port = 59199;
                    o.MiddlewarePath = "/nonexistent/middleware";
                    o.AutoSession = false;
                }));
        }

        [Fact]
        public void middlewarePathが空文字でArgumentExceptionがスローされる()
        {
            Assert.Throws<ArgumentException>(() =>
                new WorkflowLens("test_tool", o =>
                {
                    o.Port = 59199;
                    o.MiddlewarePath = "";
                    o.AutoSession = false;
                }));
        }

        [Fact]
        public void Dispose多重呼び出しで例外が出ない()
        {
            var logger = new WorkflowLens("test_tool", o =>
            {
                o.Port = 59199;
                o.AutoStartMiddleware = false;
                o.AutoSession = false;
            });
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
