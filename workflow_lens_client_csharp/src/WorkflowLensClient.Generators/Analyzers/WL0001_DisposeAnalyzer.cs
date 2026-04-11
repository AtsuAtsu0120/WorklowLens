using System.Collections.Immutable;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;
using Microsoft.CodeAnalysis.Diagnostics;

namespace WorkflowLensClient.Generators.Analyzers
{
    /// <summary>
    /// WorkflowLensインスタンスがusingで囲まれていない場合に警告する。
    /// </summary>
    [DiagnosticAnalyzer(LanguageNames.CSharp)]
    public sealed class WL0001_DisposeAnalyzer : DiagnosticAnalyzer
    {
        public override ImmutableArray<DiagnosticDescriptor> SupportedDiagnostics
            => ImmutableArray.Create(Diagnostics.WL0001_DisposeNotUsing);

        public override void Initialize(AnalysisContext context)
        {
            context.ConfigureGeneratedCodeAnalysis(GeneratedCodeAnalysisFlags.None);
            context.EnableConcurrentExecution();
            context.RegisterSyntaxNodeAction(AnalyzeObjectCreation, SyntaxKind.ObjectCreationExpression);
        }

        private static void AnalyzeObjectCreation(SyntaxNodeAnalysisContext context)
        {
            var creation = (ObjectCreationExpressionSyntax)context.Node;

            // 型がWorkflowLensかチェック
            var typeInfo = context.SemanticModel.GetTypeInfo(creation);
            if (typeInfo.Type?.Name != "WorkflowLens" ||
                typeInfo.Type.ContainingNamespace?.ToDisplayString() != "WorkflowLensClient")
                return;

            // 親を辿ってusingで囲まれているか確認
            var parent = creation.Parent;

            // EqualsValueClause → VariableDeclarator → VariableDeclaration → ...
            if (parent is EqualsValueClauseSyntax equalsClause &&
                equalsClause.Parent is VariableDeclaratorSyntax declarator &&
                declarator.Parent is VariableDeclarationSyntax declaration)
            {
                var declParent = declaration.Parent;

                // using var x = new WorkflowLens(...)
                if (declParent is LocalDeclarationStatementSyntax localDecl &&
                    localDecl.UsingKeyword.IsKind(SyntaxKind.UsingKeyword))
                    return;

                // using (var x = new WorkflowLens(...)) { }
                if (declParent is UsingStatementSyntax)
                    return;

                // フィールド代入の場合はスキップ
                if (declParent is FieldDeclarationSyntax)
                    return;
            }

            // 代入式 (_field = new WorkflowLens(...)) の場合もスキップ
            if (parent is AssignmentExpressionSyntax)
                return;

            context.ReportDiagnostic(Diagnostic.Create(
                Diagnostics.WL0001_DisposeNotUsing,
                creation.GetLocation()));
        }
    }
}
