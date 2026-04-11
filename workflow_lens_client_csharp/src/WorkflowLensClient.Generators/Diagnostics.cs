using Microsoft.CodeAnalysis;

namespace WorkflowLensClient.Generators
{
    /// <summary>
    /// Source Generator / Analyzer の診断ID定義。
    /// </summary>
    internal static class Diagnostics
    {
        private const string Category = "WorkflowLensClient";

        // Source Generator 診断
        public static readonly DiagnosticDescriptor WLSG0001_EnumMustBePublic = new DiagnosticDescriptor(
            "WLSG0001",
            "[WorkflowActions]を付けたenumはpublicである必要があります",
            "enum '{0}' はpublicである必要があります",
            Category,
            DiagnosticSeverity.Error,
            isEnabledByDefault: true);

        public static readonly DiagnosticDescriptor WLSG0002_ClassMustBePartial = new DiagnosticDescriptor(
            "WLSG0002",
            "[WorkflowLog]を使用するクラスはpartialである必要があります",
            "クラス '{0}' はpartialである必要があります",
            Category,
            DiagnosticSeverity.Error,
            isEnabledByDefault: true);

        public static readonly DiagnosticDescriptor WLSG0003_LoggerFieldNotFound = new DiagnosticDescriptor(
            "WLSG0003",
            "[WorkflowLensLogger]が付いたフィールドが見つかりません",
            "クラス '{0}' に[WorkflowLensLogger]が付いたWorkflowLensフィールドがありません",
            Category,
            DiagnosticSeverity.Error,
            isEnabledByDefault: true);

        public static readonly DiagnosticDescriptor WLSG0004_MethodNameConvention = new DiagnosticDescriptor(
            "WLSG0004",
            "[WorkflowLog]メソッド名が\"Core\"で終わっていません",
            "メソッド '{0}' は\"Core\"サフィックスで終わることを推奨します（MethodNameプロパティで上書き可能）",
            Category,
            DiagnosticSeverity.Warning,
            isEnabledByDefault: true);

        // Analyzer 診断
        public static readonly DiagnosticDescriptor WL0001_DisposeNotUsing = new DiagnosticDescriptor(
            "WL0001",
            "WorkflowLensインスタンスがusingで囲まれていません",
            "WorkflowLensはIDisposableです。usingで囲むことを推奨します",
            "Usage",
            DiagnosticSeverity.Warning,
            isEnabledByDefault: true);

        public static readonly DiagnosticDescriptor WL0002_ActionNaming = new DiagnosticDescriptor(
            "WL0002",
            "action文字列がsnake_case命名規約に違反しています",
            "action \"{0}\" はsnake_case（小文字英数字とアンダースコア）で記述してください",
            "Naming",
            DiagnosticSeverity.Warning,
            isEnabledByDefault: true);

        public static readonly DiagnosticDescriptor WL0003_SessionDirectUse = new DiagnosticDescriptor(
            "WL0003",
            "AutoSessionが有効な場合、Category.Sessionの直接使用は不要です",
            "AutoSessionが有効な場合、Category.Sessionのログは自動管理されます",
            "Usage",
            DiagnosticSeverity.Info,
            isEnabledByDefault: true);
    }
}
