using System.Collections.Immutable;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;
using Microsoft.CodeAnalysis.Diagnostics;

namespace WorkflowLensClient.Generators.Analyzers
{
    /// <summary>
    /// Log(Category.Session, ...)の直接呼び出しを検出し、
    /// AutoSession有効時は不要である旨をInfoレベルで通知する。
    /// </summary>
    [DiagnosticAnalyzer(LanguageNames.CSharp)]
    public sealed class WL0003_SessionDirectUseAnalyzer : DiagnosticAnalyzer
    {
        public override ImmutableArray<DiagnosticDescriptor> SupportedDiagnostics
            => ImmutableArray.Create(Diagnostics.WL0003_SessionDirectUse);

        public override void Initialize(AnalysisContext context)
        {
            context.ConfigureGeneratedCodeAnalysis(GeneratedCodeAnalysisFlags.None);
            context.EnableConcurrentExecution();
            context.RegisterSyntaxNodeAction(AnalyzeInvocation, SyntaxKind.InvocationExpression);
        }

        private static void AnalyzeInvocation(SyntaxNodeAnalysisContext context)
        {
            var invocation = (InvocationExpressionSyntax)context.Node;

            // メソッド名がLogか確認
            string? methodName = null;
            if (invocation.Expression is MemberAccessExpressionSyntax memberAccess)
                methodName = memberAccess.Name.Identifier.Text;
            else if (invocation.Expression is IdentifierNameSyntax identifier)
                methodName = identifier.Identifier.Text;

            if (methodName != "Log")
                return;

            // シンボル解析で対象クラスを確認
            var symbolInfo = context.SemanticModel.GetSymbolInfo(invocation);
            if (symbolInfo.Symbol is not IMethodSymbol methodSymbol)
                return;

            if (methodSymbol.ContainingType?.Name != "WorkflowLens")
                return;

            // 第1引数がCategory.Sessionか確認
            if (invocation.ArgumentList.Arguments.Count == 0)
                return;

            var firstArg = invocation.ArgumentList.Arguments[0];
            if (firstArg.Expression is MemberAccessExpressionSyntax categoryAccess &&
                categoryAccess.Expression is IdentifierNameSyntax categoryType &&
                categoryType.Identifier.Text == "Category" &&
                categoryAccess.Name.Identifier.Text == "Session")
            {
                context.ReportDiagnostic(Diagnostic.Create(
                    Diagnostics.WL0003_SessionDirectUse,
                    invocation.GetLocation()));
            }
        }
    }
}
