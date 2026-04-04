using System;
using System.Net.Sockets;
using System.Text;

namespace ToolLoggerClient
{
    /// <summary>
    /// tool_logger_middlewareへUDPでログを送信するクライアント。
    /// スレッドセーフ。IDisposableでUdpClientを解放する。
    /// </summary>
    public class ToolLogger : IDisposable
    {
        private readonly string _toolName;
        private readonly string? _toolVersion;
        private readonly string _host;
        private readonly int _port;
        private UdpClient? _client;
        private string _sessionId;
        private bool _disposed;

        /// <summary>現在のセッションID。</summary>
        public string SessionId => _sessionId;

        /// <param name="toolName">ツール名（必須）。</param>
        /// <param name="toolVersion">ツールバージョン（任意）。</param>
        /// <param name="host">送信先ホスト。</param>
        /// <param name="port">送信先ポート。</param>
        public ToolLogger(string toolName, string? toolVersion = null,
                          string host = "127.0.0.1", int port = 59100)
        {
            _toolName = toolName ?? throw new ArgumentNullException(nameof(toolName));
            _toolVersion = toolVersion;
            _host = host;
            _port = port;
            _sessionId = GenerateSessionId();
            _client = new UdpClient();
        }

        /// <summary>セッションを開始する。新しいsession_idを生成し、session_startイベントを送信する。</summary>
        public void StartSession(string message = "Session started", string? details = null)
        {
            _sessionId = GenerateSessionId();
            Send(EventType.SessionStart, message, details);
        }

        /// <summary>セッションを終了する。session_endイベントを送信する。</summary>
        public void EndSession(string message = "Session ended", string? details = null)
        {
            Send(EventType.SessionEnd, message, details);
        }

        /// <summary>ログメッセージを送信する。</summary>
        /// <param name="eventType">イベント種別（EventType定数を使用）。</param>
        /// <param name="message">メッセージ本文。</param>
        /// <param name="details">追加情報（生JSON文字列、任意）。</param>
        public void Send(string eventType, string message, string? details = null)
        {
            try
            {
                var client = _client;
                if (client == null) return;

                var json = LogMessage.BuildJson(
                    _toolName, eventType, message,
                    _sessionId, _toolVersion, details);
                var bytes = Encoding.UTF8.GetBytes(json);
                client.Send(bytes, bytes.Length, _host, _port);
            }
            catch (SocketException) { }
            catch (ObjectDisposedException) { }
        }

        /// <summary>使用ログを送信する。</summary>
        public void LogUsage(string message, string? details = null)
            => Send(EventType.Usage, message, details);

        /// <summary>エラーログを送信する。</summary>
        public void LogError(string message, string? details = null)
            => Send(EventType.Error, message, details);

        /// <summary>キャンセルログを送信する。</summary>
        public void LogCancellation(string message, string? details = null)
            => Send(EventType.Cancellation, message, details);

        public void Dispose()
        {
            if (_disposed) return;
            _disposed = true;
            _client?.Dispose();
            _client = null;
        }

        private static string GenerateSessionId()
            => Guid.NewGuid().ToString("N").Substring(0, 8);
    }
}
