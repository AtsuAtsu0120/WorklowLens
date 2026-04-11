namespace WorkflowLensClient
{
    /// <summary>イベント種別の定数。</summary>
    public static class EventType
    {
        public const string Usage = "usage";
        public const string Error = "error";
        public const string SessionStart = "session_start";
        public const string SessionEnd = "session_end";
        public const string Cancellation = "cancellation";
    }
}
