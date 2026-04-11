using System;
using System.Diagnostics;
using System.IO;
using System.Net.Sockets;
using System.Runtime.InteropServices;
using System.Text;

namespace WorkflowLensClient
{
    /// <summary>
    /// workflow_lens_middlewareへUDPでログを送信するクライアント。
    /// スレッドセーフ。IDisposableでUdpClientを解放する。
    /// </summary>
    public class WorkflowLens : IDisposable
    {
        private readonly string _toolName;
        private readonly string? _toolVersion;
        private readonly string _userId;
        private readonly string _host;
        private readonly int _port;
        private UdpClient? _client;
        private Process? _process;
        private readonly string _sessionId;
        private bool _disposed;

        /// <summary>現在のセッションID。</summary>
        public string SessionId => _sessionId;

        /// <summary>環境変数名。</summary>
        internal const string MiddlewarePathEnvVar = "WORKFLOW_LENS_MIDDLEWARE_PATH";

        /// <summary>PATH探索時のバイナリ名。</summary>
        internal const string MiddlewareBinaryName = "workflow_lens_middleware";

        /// <param name="toolName">ツール名（必須）。</param>
        /// <param name="toolVersion">ツールバージョン（任意）。</param>
        /// <param name="userId">ユーザーID（任意、未指定時はOSユーザー名）。</param>
        /// <param name="host">送信先ホスト。</param>
        /// <param name="port">送信先ポート。</param>
        /// <param name="middlewarePath">middlewareバイナリのパス（任意）。指定時はプロセスを自動起動・停止する。</param>
        /// <param name="autoStartMiddleware">trueの場合、環境変数→PATH探索でmiddlewareバイナリを自動検出して起動する。</param>
        public WorkflowLens(string toolName, string? toolVersion = null,
                          string? userId = null,
                          string host = "127.0.0.1", int port = 59100,
                          string? middlewarePath = null,
                          bool autoStartMiddleware = true)
        {
            _toolName = toolName ?? throw new ArgumentNullException(nameof(toolName));
            _toolVersion = toolVersion;
            _userId = userId ?? Environment.UserName;
            _host = host;
            _port = port;
            _sessionId = GenerateSessionId();
            _client = new UdpClient();

            var resolvedPath = ResolveMiddlewarePath(middlewarePath, autoStartMiddleware);
            if (resolvedPath != null)
            {
                try
                {
                    _process = new Process
                    {
                        StartInfo = new ProcessStartInfo
                        {
                            FileName = resolvedPath,
                            Arguments = port.ToString(),
                            UseShellExecute = false,
                            CreateNoWindow = true,
                            RedirectStandardOutput = true,
                            RedirectStandardError = true,
                        }
                    };
                    _process.Start();
                }
                catch
                {
                    _client.Dispose();
                    _client = null;
                    throw;
                }
            }
        }

        /// <summary>
        /// middlewareバイナリのパスを解決する。
        /// 優先順位: middlewarePath(明示指定) → 環境変数 → PATH探索。
        /// </summary>
        internal static string? ResolveMiddlewarePath(string? middlewarePath, bool autoStartMiddleware)
        {
            if (middlewarePath != null)
            {
                if (string.IsNullOrWhiteSpace(middlewarePath))
                    throw new ArgumentException("middlewarePathが空文字です。", nameof(middlewarePath));
                return middlewarePath;
            }

            if (!autoStartMiddleware)
                return null;

            // 環境変数から取得
            var envPath = Environment.GetEnvironmentVariable(MiddlewarePathEnvVar);
            if (!string.IsNullOrEmpty(envPath))
                return envPath;

            // PATH探索
            var found = FindInPath(MiddlewareBinaryName);
            if (found != null)
                return found;

            throw new InvalidOperationException(
                $"ミドルウェアバイナリが見つかりません。" +
                $"PATH に {MiddlewareBinaryName} を配置するか、" +
                $"環境変数 {MiddlewarePathEnvVar} を設定してください。");
        }

        /// <summary>PATH環境変数からバイナリを探索する。</summary>
        internal static string? FindInPath(string binaryName)
        {
            var pathEnv = Environment.GetEnvironmentVariable("PATH");
            if (string.IsNullOrEmpty(pathEnv))
                return null;

            var separator = RuntimeInformation.IsOSPlatform(OSPlatform.Windows) ? ';' : ':';
            var dirs = pathEnv.Split(separator);
            var isWindows = RuntimeInformation.IsOSPlatform(OSPlatform.Windows);

            foreach (var dir in dirs)
            {
                if (string.IsNullOrWhiteSpace(dir))
                    continue;

                var candidate = Path.Combine(dir, binaryName);
                if (File.Exists(candidate))
                    return candidate;

                // Windowsでは拡張子付きも探索
                if (isWindows && File.Exists(candidate + ".exe"))
                    return candidate + ".exe";
            }

            return null;
        }

        /// <summary>ログを送信する。</summary>
        /// <param name="category">カテゴリ。</param>
        /// <param name="action">アクション。</param>
        /// <param name="durationMs">操作時間（ミリ秒、任意）。</param>
        public void Log(Category category, string action, long? durationMs = null)
        {
            try
            {
                var client = _client;
                if (client == null) return;

                using var activity = WorkflowLensTelemetry.Source.StartActivity(
                    "workflowlens.send",
                    System.Diagnostics.ActivityKind.Producer);

                activity?.SetTag("tool.name", _toolName);
                activity?.SetTag("category", category.ToJsonString());
                activity?.SetTag("action", action);
                activity?.SetTag("session.id", _sessionId);

                // ActivityからW3C traceparentを生成
                string? traceparent = null;
                if (activity != null && activity.Id != null)
                {
                    var flags = (activity.ActivityTraceFlags & System.Diagnostics.ActivityTraceFlags.Recorded) != 0
                        ? "01" : "00";
                    traceparent = $"00-{activity.TraceId}-{activity.SpanId}-{flags}";
                }

                var json = LogMessage.BuildJson(
                    _toolName, category.ToJsonString(), action,
                    _sessionId, _toolVersion, _userId, durationMs, traceparent);
                var bytes = Encoding.UTF8.GetBytes(json);
                client.Send(bytes, bytes.Length, _host, _port);

                activity?.SetTag("messaging.message.payload_size_bytes", bytes.Length);
            }
            catch (SocketException) { }
            catch (ObjectDisposedException) { }
        }

        /// <summary>
        /// 操作時間を自動計測するスコープを開始する。
        /// usingブロックで囲むとDisposeTime時に自動的にLogが呼ばれる。
        /// </summary>
        public IDisposable MeasureScope(Category category, string action)
        {
            return new MeasureScopeHandle(this, category, action);
        }

        public void Dispose()
        {
            if (_disposed) return;
            _disposed = true;

            _client?.Dispose();
            _client = null;

            if (_process is { HasExited: false })
            {
                try { _process.Kill(); } catch { }
            }
            _process?.Dispose();
            _process = null;
        }

        private static string GenerateSessionId()
            => Guid.NewGuid().ToString("N").Substring(0, 8);

        /// <summary>MeasureScopeの内部実装。</summary>
        private sealed class MeasureScopeHandle : IDisposable
        {
            private readonly WorkflowLens _logger;
            private readonly Category _category;
            private readonly string _action;
            private readonly Stopwatch _stopwatch;

            internal MeasureScopeHandle(WorkflowLens logger, Category category, string action)
            {
                _logger = logger;
                _category = category;
                _action = action;
                _stopwatch = Stopwatch.StartNew();
            }

            public void Dispose()
            {
                _stopwatch.Stop();
                _logger.Log(_category, _action, _stopwatch.ElapsedMilliseconds);
            }
        }
    }
}
