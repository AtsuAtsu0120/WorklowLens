using System.Collections.Immutable;
using System.Threading.Tasks;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.Diagnostics;
using WorkflowLensClient.Generators.Analyzers;
using Xunit;

namespace WorkflowLensClient.Generators.Tests
{
    public class AnalyzerTests
    {
        private static async Task<ImmutableArray<Diagnostic>> GetDiagnosticsAsync<TAnalyzer>(string source)
            where TAnalyzer : DiagnosticAnalyzer, new()
        {
            var syntaxTree = CSharpSyntaxTree.ParseText(source);

            var references = TestHelper.GetMetadataReferences();

            var compilation = CSharpCompilation.Create("TestAssembly",
                new[] { syntaxTree },
                references,
                new CSharpCompilationOptions(OutputKind.DynamicallyLinkedLibrary));

            var analyzer = new TAnalyzer();
            var compilationWithAnalyzers = compilation.WithAnalyzers(
                ImmutableArray.Create<DiagnosticAnalyzer>(analyzer));

            return await compilationWithAnalyzers.GetAnalyzerDiagnosticsAsync();
        }

        // WL0001 テスト

        [Fact]
        public async Task WL0001_using文で囲んだ場合に診断が出ない()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M()
    {
        using (var logger = new WorkflowLens(""tool"", o => { o.AutoStartMiddleware = false; }))
        {
        }
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0001_DisposeAnalyzer>(source);
            Assert.DoesNotContain(diags, d => d.Id == "WL0001");
        }

        [Fact]
        public async Task WL0001_通常のローカル変数で診断が出る()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M()
    {
        var logger = new WorkflowLens(""tool"", o => { o.AutoStartMiddleware = false; });
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0001_DisposeAnalyzer>(source);
            Assert.Contains(diags, d => d.Id == "WL0001");
        }

        [Fact]
        public async Task WL0001_フィールド代入で診断が出ない()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    private WorkflowLens _logger = new WorkflowLens(""tool"", o => { o.AutoStartMiddleware = false; });
}
";
            var diags = await GetDiagnosticsAsync<WL0001_DisposeAnalyzer>(source);
            Assert.DoesNotContain(diags, d => d.Id == "WL0001");
        }

        // WL0002 テスト

        [Fact]
        public async Task WL0002_正しいsnake_caseで診断が出ない()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M(WorkflowLens logger)
    {
        logger.Log(Category.Edit, ""brush_apply"");
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0002_ActionNamingAnalyzer>(source);
            Assert.DoesNotContain(diags, d => d.Id == "WL0002");
        }

        [Fact]
        public async Task WL0002_大文字始まりで診断が出る()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M(WorkflowLens logger)
    {
        logger.Log(Category.Edit, ""Compile"");
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0002_ActionNamingAnalyzer>(source);
            Assert.Contains(diags, d => d.Id == "WL0002");
        }

        [Fact]
        public async Task WL0002_スペース含みで診断が出る()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M(WorkflowLens logger)
    {
        logger.Log(Category.Edit, ""brush apply"");
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0002_ActionNamingAnalyzer>(source);
            Assert.Contains(diags, d => d.Id == "WL0002");
        }

        [Fact]
        public async Task WL0002_変数を渡した場合は診断が出ない()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M(WorkflowLens logger)
    {
        var action = ""test"";
        logger.Log(Category.Edit, action);
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0002_ActionNamingAnalyzer>(source);
            Assert.DoesNotContain(diags, d => d.Id == "WL0002");
        }

        // WL0003 テスト

        [Fact]
        public async Task WL0003_Category_Sessionで診断が出る()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M(WorkflowLens logger)
    {
        logger.Log(Category.Session, ""start"");
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0003_SessionDirectUseAnalyzer>(source);
            Assert.Contains(diags, d => d.Id == "WL0003");
        }

        [Fact]
        public async Task WL0003_Category_Buildで診断が出ない()
        {
            var source = @"
using WorkflowLensClient;
class Test
{
    void M(WorkflowLens logger)
    {
        logger.Log(Category.Build, ""compile"");
    }
}
";
            var diags = await GetDiagnosticsAsync<WL0003_SessionDirectUseAnalyzer>(source);
            Assert.DoesNotContain(diags, d => d.Id == "WL0003");
        }
    }
}
