using System.Collections.Immutable;
using System.Text.RegularExpressions;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;
using Microsoft.CodeAnalysis.Diagnostics;

namespace WorkflowLensClient.Generators.Analyzers
{
    /// <summary>
    /// Log()/MeasureScope()のaction文字列がsnake_case命名規約に違反している場合に警告する。
    /// </summary>
    [DiagnosticAnalyzer(LanguageNames.CSharp)]
    public sealed class WL0002_ActionNamingAnalyzer : DiagnosticAnalyzer
    {
        private static readonly Regex SnakeCasePattern = new Regex(@"^[a-z][a-z0-9_]*$");

        public override ImmutableArray<DiagnosticDescriptor> SupportedDiagnostics
            => ImmutableArray.Create(Diagnostics.WL0002_ActionNaming);

        public override void Initialize(AnalysisContext context)
        {
            context.ConfigureGeneratedCodeAnalysis(GeneratedCodeAnalysisFlags.None);
            context.EnableConcurrentExecution();
            context.RegisterSyntaxNodeAction(AnalyzeInvocation, SyntaxKind.InvocationExpression);
        }

        private static void AnalyzeInvocation(SyntaxNodeAnalysisContext context)
        {
            var invocation = (InvocationExpressionSyntax)context.Node;

            // メソッド名を取得
            string? methodName = null;
            if (invocation.Expression is MemberAccessExpressionSyntax memberAccess)
                methodName = memberAccess.Name.Identifier.Text;
            else if (invocation.Expression is IdentifierNameSyntax identifier)
                methodName = identifier.Identifier.Text;

            if (methodName != "Log" && methodName != "MeasureScope")
                return;

            // シンボル解析で対象クラスを確認
            var symbolInfo = context.SemanticModel.GetSymbolInfo(invocation);
            if (symbolInfo.Symbol is not IMethodSymbol methodSymbol)
                return;

            var containingType = methodSymbol.ContainingType?.Name;
            if (containingType != "WorkflowLens" && containingType != "CategoryLogger")
                return;

            // action引数の位置を特定
            int actionArgIndex;
            if (containingType == "WorkflowLens" && methodName == "Log")
                actionArgIndex = 1; // Log(Category, string action, ...)
            else if (containingType == "WorkflowLens" && methodName == "MeasureScope")
                actionArgIndex = 1; // MeasureScope(Category, string action)
            else if (containingType == "CategoryLogger" && methodName == "Log")
                actionArgIndex = 0; // Log(string action, ...)
            else
                return;

            if (invocation.ArgumentList.Arguments.Count <= actionArgIndex)
                return;

            var actionArg = invocation.ArgumentList.Arguments[actionArgIndex];
            if (actionArg.Expression is LiteralExpressionSyntax literal &&
                literal.IsKind(SyntaxKind.StringLiteralExpression))
            {
                var value = literal.Token.ValueText;
                if (!SnakeCasePattern.IsMatch(value))
                {
                    context.ReportDiagnostic(Diagnostic.Create(
                        Diagnostics.WL0002_ActionNaming,
                        literal.GetLocation(),
                        value));
                }
            }
        }
    }
}
