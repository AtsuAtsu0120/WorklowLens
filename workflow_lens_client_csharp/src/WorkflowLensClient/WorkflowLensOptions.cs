namespace WorkflowLensClient
{
    /// <summary>
    /// WorkflowLensの初期化オプション。
    /// </summary>
    public class WorkflowLensOptions
    {
        /// <summary>ツールバージョン（任意）。</summary>
        public string? ToolVersion { get; set; }

        /// <summary>ユーザーID（任意、未指定時はOSユーザー名）。</summary>
        public string? UserId { get; set; }

        /// <summary>送信先ホスト。</summary>
        public string Host { get; set; } = "127.0.0.1";

        /// <summary>送信先ポート。</summary>
        public int Port { get; set; } = 59100;

        /// <summary>middlewareバイナリのパス（任意）。指定時はプロセスを自動起動・停止する。</summary>
        public string? MiddlewarePath { get; set; }

        /// <summary>trueの場合、環境変数→PATH探索でmiddlewareバイナリを自動検出して起動する。</summary>
        public bool AutoStartMiddleware { get; set; } = true;

        /// <summary>trueの場合、コンストラクタでSession/start、DisposeでSession/endを自動送信する。</summary>
        public bool AutoSession { get; set; } = true;
    }
}
