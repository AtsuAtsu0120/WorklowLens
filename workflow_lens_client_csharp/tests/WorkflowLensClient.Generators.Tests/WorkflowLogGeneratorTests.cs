using System.Linq;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using WorkflowLensClient.Generators;
using Xunit;

namespace WorkflowLensClient.Generators.Tests
{
    public class WorkflowLogGeneratorTests
    {
        private static GeneratorDriverRunResult RunGenerator(string source)
        {
            var syntaxTrees = new[]
            {
                CSharpSyntaxTree.ParseText(source),
                CSharpSyntaxTree.ParseText(AttributeSources.WorkflowLogSource),
                CSharpSyntaxTree.ParseText(AttributeSources.WorkflowLensLoggerSource),
            };
            // 属性はcompilationに含めているので、Generatorは重複生成しない

            var references = TestHelper.GetMetadataReferences();

            var compilation = CSharpCompilation.Create("TestAssembly",
                syntaxTrees,
                references,
                new CSharpCompilationOptions(OutputKind.DynamicallyLinkedLibrary));

            var generator = new WorkflowLogGenerator();
            var driver = CSharpGeneratorDriver.Create(generator);
            driver = (CSharpGeneratorDriver)driver.RunGeneratorsAndUpdateCompilation(
                compilation, out _, out _);

            return driver.GetRunResult();
        }

        [Fact]
        public void Coreメソッドからラッパーが正しく生成される()
        {
            var source = @"
using WorkflowLensClient;

public partial class MyTool
{
    [WorkflowLensLogger]
    private readonly WorkflowLens _logger = null!;

    [WorkflowLog(Category.Build, ""compile"")]
    private void CompileCore()
    {
    }
}
";
            var result = RunGenerator(source);
            var generatedSource = result.GeneratedTrees
                .FirstOrDefault(t => t.FilePath.Contains("WorkflowLog.g.cs"))
                ?.GetText().ToString();

            Assert.NotNull(generatedSource);
            Assert.Contains("public void Compile()", generatedSource);
            Assert.Contains("using (_logger.Build(\"compile\"))", generatedSource);
            Assert.Contains("CompileCore()", generatedSource);
        }

        [Fact]
        public void 戻り値ありメソッドが正しく生成される()
        {
            var source = @"
using WorkflowLensClient;

public partial class MyTool
{
    [WorkflowLensLogger]
    private readonly WorkflowLens _logger = null!;

    [WorkflowLog(Category.Edit, ""import"", MethodName = ""ImportAsset"")]
    private string ImportAssetCore(string path)
    {
        return path;
    }
}
";
            var result = RunGenerator(source);
            var generatedSource = result.GeneratedTrees
                .FirstOrDefault(t => t.FilePath.Contains("WorkflowLog.g.cs"))
                ?.GetText().ToString();

            Assert.NotNull(generatedSource);
            Assert.Contains("public string ImportAsset(string path)", generatedSource);
            Assert.Contains("return ImportAssetCore(path)", generatedSource);
        }

        [Fact]
        public void 非partialクラスで診断WLSG0002が出る()
        {
            var source = @"
using WorkflowLensClient;

public class MyTool
{
    [WorkflowLensLogger]
    private readonly WorkflowLens _logger = null!;

    [WorkflowLog(Category.Build, ""compile"")]
    private void CompileCore() { }
}
";
            var result = RunGenerator(source);
            Assert.Contains(result.Diagnostics, d => d.Id == "WLSG0002");
        }

        [Fact]
        public void loggerフィールド未指定で診断WLSG0003が出る()
        {
            var source = @"
using WorkflowLensClient;

public partial class MyTool
{
    private readonly WorkflowLens _logger = null!;

    [WorkflowLog(Category.Build, ""compile"")]
    private void CompileCore() { }
}
";
            var result = RunGenerator(source);
            Assert.Contains(result.Diagnostics, d => d.Id == "WLSG0003");
        }
    }
}
