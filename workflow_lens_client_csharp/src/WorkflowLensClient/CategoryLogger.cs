using System;
using System.Diagnostics;

namespace WorkflowLensClient
{
    /// <summary>
    /// カテゴリを固定したログ送信オブジェクト。
    /// action指定時はusingブロック全体の所要時間を自動計測してDispose時に送信する。
    /// action省略時はグルーピング用（Dispose時にログは送信しない）。
    /// </summary>
    public sealed class CategoryLogger : IDisposable
    {
        private readonly WorkflowLens _logger;
        private readonly Category _category;
        private readonly string? _action;
        private readonly Stopwatch? _stopwatch;
        private bool _disposed;

        internal CategoryLogger(WorkflowLens logger, Category category, string? action = null)
        {
            _logger = logger;
            _category = category;
            _action = action;

            if (action != null)
            {
                _stopwatch = Stopwatch.StartNew();
            }
        }

        /// <summary>指定actionでログを即時送信する。</summary>
        /// <param name="action">アクション。</param>
        /// <param name="durationMs">操作時間（ミリ秒、任意）。</param>
        public void Log(string action, long? durationMs = null)
            => _logger.Log(_category, action, durationMs);

        /// <summary>
        /// action指定時: 自動計測ログを送信する。
        /// action省略時: 何もしない。
        /// </summary>
        public void Dispose()
        {
            if (_disposed) return;
            _disposed = true;

            if (_action != null && _stopwatch != null)
            {
                _stopwatch.Stop();
                _logger.Log(_category, _action, _stopwatch.ElapsedMilliseconds);
            }
        }
    }
}
