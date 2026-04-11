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
        private readonly bool _autoSession;
        private UdpClient? _client;
        private Process? _process;
        private readonly string _sessionId;
        private bool _disposed;
        private bool _sessionStartSent;
        private bool _sessionEndSent;

        /// <summary>現在のセッションID。</summary>
        public string SessionId => _sessionId;

        /// <summary>環境変数名。</summary>
        internal const string MiddlewarePathEnvVar = "WORKFLOW_LENS_MIDDLEWARE_PATH";

        /// <summary>PATH探索時のバイナリ名。</summary>
        internal const string MiddlewareBinaryName = "workflow_lens_middleware";

        /// <summary>
        /// Optionsパターンでインスタンスを生成する。
        /// </summary>
        /// <param name="toolName">ツール名（必須）。</param>
        /// <param name="configure">オプション設定コールバック（任意）。</param>
        public WorkflowLens(string toolName, Action<WorkflowLensOptions>? configure = null)
        {
            _toolName = toolName ?? throw new ArgumentNullException(nameof(toolName));

            var options = new WorkflowLensOptions();
            configure?.Invoke(options);

            _toolVersion = options.ToolVersion;
            _userId = options.UserId ?? Environment.UserName;
            _host = options.Host;
            _port = options.Port;
            _autoSession = options.AutoSession;
            _sessionId = GenerateSessionId();
            _client = new UdpClient();

            var resolvedPath = ResolveMiddlewarePath(options.MiddlewarePath, options.AutoStartMiddleware);
            if (resolvedPath != null)
            {
                try
                {
                    _process = new Process
                    {
                        StartInfo = new ProcessStartInfo
                        {
                            FileName = resolvedPath,
                            Arguments = _port.ToString(),
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

            // AutoSession: コンストラクタでSession/startを自動送信
            if (_autoSession)
            {
                Log(Category.Session, "start");
            }
        }

        /// <summary>
        /// 既存の互換コンストラクタ。内部でOptionsパターンに委譲する。
        /// AutoSessionはfalse（既存コードの動作を変えない）。
        /// </summary>
        /// <param name="toolName">ツール名（必須）。</param>
        /// <param name="toolVersion">ツールバージョン（任意）。</param>
        /// <param name="userId">ユーザーID（任意、未指定時はOSユーザー名）。</param>
        /// <param name="host">送信先ホスト。</param>
        /// <param name="port">送信先ポート。</param>
        /// <param name="middlewarePath">middlewareバイナリのパス（任意）。</param>
        /// <param name="autoStartMiddleware">trueの場合、middlewareバイナリを自動検出して起動する。</param>
        public WorkflowLens(string toolName, string? toolVersion,
                          string? userId = null,
                          string host = "127.0.0.1", int port = 59100,
                          string? middlewarePath = null,
                          bool autoStartMiddleware = true)
            : this(toolName, o =>
            {
                o.ToolVersion = toolVersion;
                o.UserId = userId;
                o.Host = host;
                o.Port = port;
                o.MiddlewarePath = middlewarePath;
                o.AutoStartMiddleware = autoStartMiddleware;
                o.AutoSession = false;
            })
        { }

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
            // AutoSession: 重複防止
            if (category == Category.Session)
            {
                if (action == "start")
                {
                    if (_sessionStartSent) return;
                    _sessionStartSent = true;
                }
                else if (action == "end")
                {
                    if (_sessionEndSent) return;
                    _sessionEndSent = true;
                }
            }

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
        /// カテゴリを固定したCategoryLoggerを生成する。
        /// action指定時はusingブロック全体の所要時間を自動計測してDispose時に送信する。
        /// action省略時はグルーピング用（Dispose時にログは送信しない）。
        /// </summary>
        public CategoryLogger Asset(string? action = null) => new CategoryLogger(this, Category.Asset, action);

        /// <inheritdoc cref="Asset(string?)"/>
        public CategoryLogger Build(string? action = null) => new CategoryLogger(this, Category.Build, action);

        /// <inheritdoc cref="Asset(string?)"/>
        public CategoryLogger Edit(string? action = null) => new CategoryLogger(this, Category.Edit, action);

        /// <inheritdoc cref="Asset(string?)"/>
        public CategoryLogger Error(string? action = null) => new CategoryLogger(this, Category.Error, action);

        /// <summary>
        /// 操作時間を自動計測するスコープを開始する。
        /// usingブロックで囲むとDispose時に自動的にLogが呼ばれる。
        /// </summary>
        public IDisposable MeasureScope(Category category, string action)
        {
            return new MeasureScopeHandle(this, category, action);
        }

        public void Dispose()
        {
            if (_disposed) return;
            _disposed = true;

            // AutoSession: DisposeでSession/endを自動送信
            if (_autoSession && !_sessionEndSent)
            {
                Log(Category.Session, "end");
            }

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
