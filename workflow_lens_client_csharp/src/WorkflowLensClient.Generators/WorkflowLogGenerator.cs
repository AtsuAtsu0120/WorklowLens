using System.Collections.Generic;
using System.Linq;
using System.Text;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;
using Microsoft.CodeAnalysis.Text;

namespace WorkflowLensClient.Generators
{
    /// <summary>
    /// [WorkflowLog] 属性が付いたprivateメソッドに対し、
    /// MeasureScopeで囲むpublicラッパーメソッドを自動生成する。
    /// </summary>
    [Generator]
    public class WorkflowLogGenerator : ISourceGenerator
    {
        public void Initialize(GeneratorInitializationContext context)
        {
            context.RegisterForSyntaxNotifications(() => new MethodSyntaxReceiver());
        }

        public void Execute(GeneratorExecutionContext context)
        {
            // 属性が既にcompilationに存在しなければソースを注入
            if (context.Compilation.GetTypeByMetadataName(AttributeSources.WorkflowLogAttributeName) == null)
            {
                context.AddSource("WorkflowLogAttribute.g.cs",
                    SourceText.From(AttributeSources.WorkflowLogSource, Encoding.UTF8));
            }
            if (context.Compilation.GetTypeByMetadataName(AttributeSources.WorkflowLensLoggerAttributeName) == null)
            {
                context.AddSource("WorkflowLensLoggerAttribute.g.cs",
                    SourceText.From(AttributeSources.WorkflowLensLoggerSource, Encoding.UTF8));
            }

            if (context.SyntaxReceiver is not MethodSyntaxReceiver receiver)
                return;

            var compilation = context.Compilation;

            // クラスごとにグルーピング
            var methodsByClass = new Dictionary<INamedTypeSymbol, List<(IMethodSymbol method, AttributeData attr, MethodDeclarationSyntax syntax)>>(
                SymbolEqualityComparer.Default);

            foreach (var methodDecl in receiver.CandidateMethods)
            {
                var model = compilation.GetSemanticModel(methodDecl.SyntaxTree);
                var methodSymbol = model.GetDeclaredSymbol(methodDecl) as IMethodSymbol;
                if (methodSymbol == null) continue;

                var attr = methodSymbol.GetAttributes().FirstOrDefault(a =>
                    a.AttributeClass?.Name == "WorkflowLogAttribute" ||
                    a.AttributeClass?.ToDisplayString() == AttributeSources.WorkflowLogAttributeName);
                if (attr == null) continue;

                var containingType = methodSymbol.ContainingType;
                if (!methodsByClass.ContainsKey(containingType))
                {
                    methodsByClass[containingType] = new List<(IMethodSymbol, AttributeData, MethodDeclarationSyntax)>();
                }
                methodsByClass[containingType].Add((methodSymbol, attr, methodDecl));
            }

            foreach (var kvp in methodsByClass)
            {
                var classSymbol = kvp.Key;
                var methods = kvp.Value;

                // partialチェック
                var classDecl = classSymbol.DeclaringSyntaxReferences
                    .Select(r => r.GetSyntax())
                    .OfType<ClassDeclarationSyntax>()
                    .FirstOrDefault();

                if (classDecl == null) continue;

                if (!classDecl.Modifiers.Any(m => m.IsKind(SyntaxKind.PartialKeyword)))
                {
                    context.ReportDiagnostic(Diagnostic.Create(
                        Diagnostics.WLSG0002_ClassMustBePartial,
                        classDecl.Identifier.GetLocation(),
                        classSymbol.Name));
                    continue;
                }

                // [WorkflowLensLogger]フィールドを探す
                var loggerField = classSymbol.GetMembers()
                    .OfType<IFieldSymbol>()
                    .FirstOrDefault(f => f.GetAttributes().Any(a =>
                        a.AttributeClass?.Name == "WorkflowLensLoggerAttribute" ||
                        a.AttributeClass?.ToDisplayString() == AttributeSources.WorkflowLensLoggerAttributeName));

                if (loggerField == null)
                {
                    context.ReportDiagnostic(Diagnostic.Create(
                        Diagnostics.WLSG0003_LoggerFieldNotFound,
                        classDecl.Identifier.GetLocation(),
                        classSymbol.Name));
                    continue;
                }

                var source = GeneratePartialClass(classSymbol, loggerField, methods, context);
                if (source != null)
                {
                    var fileName = classSymbol.ToDisplayString().Replace('.', '_');
                    context.AddSource($"{fileName}.WorkflowLog.g.cs",
                        SourceText.From(source, Encoding.UTF8));
                }
            }
        }

        private static string? GeneratePartialClass(
            INamedTypeSymbol classSymbol,
            IFieldSymbol loggerField,
            List<(IMethodSymbol method, AttributeData attr, MethodDeclarationSyntax syntax)> methods,
            GeneratorExecutionContext context)
        {
            var sb = new StringBuilder();
            sb.AppendLine("// <auto-generated />");
            sb.AppendLine("#nullable enable");
            sb.AppendLine("using System;");
            sb.AppendLine("using WorkflowLensClient;");
            sb.AppendLine();

            // namespace
            var ns = classSymbol.ContainingNamespace;
            var hasNamespace = !ns.IsGlobalNamespace;
            if (hasNamespace)
            {
                sb.AppendLine($"namespace {ns.ToDisplayString()}");
                sb.AppendLine("{");
            }

            var indent = hasNamespace ? "    " : "";
            sb.AppendLine($"{indent}public partial class {classSymbol.Name}");
            sb.AppendLine($"{indent}{{");

            foreach (var (method, attr, syntax) in methods)
            {
                if (attr.ConstructorArguments.Length < 2) continue;

                var categoryValue = attr.ConstructorArguments[0].Value;
                var action = attr.ConstructorArguments[1].Value as string;
                if (categoryValue == null || action == null) continue;

                var categoryName = GetCategoryName((int)categoryValue);
                if (categoryName == null) continue;

                // 公開メソッド名の決定
                string publicMethodName;
                var methodNameProp = attr.NamedArguments
                    .FirstOrDefault(na => na.Key == "MethodName").Value;
                if (!methodNameProp.IsNull && methodNameProp.Value is string customName)
                {
                    publicMethodName = customName;
                }
                else if (method.Name.EndsWith("Core"))
                {
                    publicMethodName = method.Name.Substring(0, method.Name.Length - 4);
                }
                else
                {
                    // Core サフィックスなし — 警告を出してメソッド名をそのまま使う
                    context.ReportDiagnostic(Diagnostic.Create(
                        Diagnostics.WLSG0004_MethodNameConvention,
                        syntax.Identifier.GetLocation(),
                        method.Name));
                    publicMethodName = method.Name;
                }

                // パラメータ
                var parameters = method.Parameters;
                var paramDecl = string.Join(", ", parameters.Select(p => $"{p.Type.ToDisplayString()} {p.Name}"));
                var paramCall = string.Join(", ", parameters.Select(p => p.Name));

                // 戻り値
                var returnType = method.ReturnType.ToDisplayString();
                var hasReturn = returnType != "void";

                sb.AppendLine($"{indent}    public {returnType} {publicMethodName}({paramDecl})");
                sb.AppendLine($"{indent}    {{");
                sb.AppendLine($"{indent}        using ({loggerField.Name}.{categoryName}(\"{action}\"))");
                sb.AppendLine($"{indent}        {{");
                if (hasReturn)
                {
                    sb.AppendLine($"{indent}            return {method.Name}({paramCall});");
                }
                else
                {
                    sb.AppendLine($"{indent}            {method.Name}({paramCall});");
                }
                sb.AppendLine($"{indent}        }}");
                sb.AppendLine($"{indent}    }}");
                sb.AppendLine();
            }

            sb.AppendLine($"{indent}}}");
            if (hasNamespace) sb.AppendLine("}");

            return sb.ToString();
        }

        private static string? GetCategoryName(int value)
        {
            return value switch
            {
                0 => "Asset",
                1 => "Build",
                2 => "Edit",
                3 => "Error",
                4 => "Session",
                _ => null,
            };
        }

        private class MethodSyntaxReceiver : ISyntaxReceiver
        {
            public List<MethodDeclarationSyntax> CandidateMethods { get; } = new List<MethodDeclarationSyntax>();

            public void OnVisitSyntaxNode(SyntaxNode syntaxNode)
            {
                if (syntaxNode is MethodDeclarationSyntax methodDecl &&
                    methodDecl.AttributeLists.Count > 0)
                {
                    CandidateMethods.Add(methodDecl);
                }
            }
        }
    }
}
