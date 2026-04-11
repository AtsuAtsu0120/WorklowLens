using System.Diagnostics;

namespace WorkflowLensClient
{
    /// <summary>
    /// WorkflowLensのトレーシング用ActivitySource。
    /// ホストアプリケーションがOpenTelemetry SDKを設定している場合、
    /// このActivitySourceを監視対象に追加することでトレースが収集される。
    /// OTel SDK未設定時はStartActivityがnullを返し、オーバーヘッドはゼロ。
    /// </summary>
    /// <example>
    /// // ホストアプリでのOTel SDK設定例（.NET Standard 2.1以上）:
    /// services.AddOpenTelemetry()
    ///     .WithTracing(builder => builder
    ///         .AddSource(WorkflowLensTelemetry.ActivitySourceName)
    ///         .AddOtlpExporter());
    /// </example>
    public static class WorkflowLensTelemetry
    {
        /// <summary>ActivitySource名。OTel SDK設定時にこの名前を登録する。</summary>
        public const string ActivitySourceName = "WorkflowLensClient";

        internal static readonly ActivitySource Source = new ActivitySource(ActivitySourceName, "0.1.0");
    }
}
