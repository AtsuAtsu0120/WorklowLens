using System.Linq;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using WorkflowLensClient.Generators;
using Xunit;

namespace WorkflowLensClient.Generators.Tests
{
    public class CategoryApiGeneratorTests
    {
        private static GeneratorDriverRunResult RunGenerator(string source)
        {
            var syntaxTrees = new[]
            {
                CSharpSyntaxTree.ParseText(source),
                CSharpSyntaxTree.ParseText(AttributeSources.WorkflowActionsSource),
            };

            var references = TestHelper.GetMetadataReferences();

            var compilation = CSharpCompilation.Create("TestAssembly",
                syntaxTrees,
                references,
                new CSharpCompilationOptions(OutputKind.DynamicallyLinkedLibrary));

            var generator = new CategoryApiGenerator();
            var driver = CSharpGeneratorDriver.Create(generator);
            driver = (CSharpGeneratorDriver)driver.RunGeneratorsAndUpdateCompilation(
                compilation, out var outputCompilation, out var diagnostics);

            return driver.GetRunResult();
        }

        [Fact]
        public void enumメンバーから正しい拡張メソッドが生成される()
        {
            var source = @"
using WorkflowLensClient;

[WorkflowActions(Category.Build)]
public enum BuildAction
{
    Compile,
    Link,
}
";
            var result = RunGenerator(source);
            var generatedSource = result.GeneratedTrees
                .FirstOrDefault(t => t.FilePath.Contains("BuildExtensions"))
                ?.GetText().ToString();

            Assert.NotNull(generatedSource);
            Assert.Contains("public static void Compile(this CategoryLogger logger", generatedSource);
            Assert.Contains("=> logger.Log(\"compile\"", generatedSource);
            Assert.Contains("public static void Link(this CategoryLogger logger", generatedSource);
            Assert.Contains("=> logger.Log(\"link\"", generatedSource);
        }

        [Fact]
        public void PascalCase_snake_case変換が正しい()
        {
            var source = @"
using WorkflowLensClient;

[WorkflowActions(Category.Edit)]
public enum EditAction
{
    BrushApply,
    LayerChange,
    ShaderCompile,
}
";
            var result = RunGenerator(source);
            var generatedSource = result.GeneratedTrees
                .FirstOrDefault(t => t.FilePath.Contains("EditExtensions"))
                ?.GetText().ToString();

            Assert.NotNull(generatedSource);
            Assert.Contains("\"brush_apply\"", generatedSource);
            Assert.Contains("\"layer_change\"", generatedSource);
            Assert.Contains("\"shader_compile\"", generatedSource);
        }

        [Fact]
        public void PascalToSnakeCase_略語ケース()
        {
            Assert.Equal("lod_generate", PascalToSnakeCase.Convert("LODGenerate"));
            Assert.Equal("compile", PascalToSnakeCase.Convert("Compile"));
            Assert.Equal("shader_compile", PascalToSnakeCase.Convert("ShaderCompile"));
            Assert.Equal("brush_apply", PascalToSnakeCase.Convert("BrushApply"));
        }

        [Fact]
        public void 非publicなenumで診断WLSG0001が出る()
        {
            var source = @"
using WorkflowLensClient;

[WorkflowActions(Category.Build)]
internal enum BuildAction
{
    Compile,
}
";
            var result = RunGenerator(source);
            var diags = result.Diagnostics;
            Assert.Contains(diags, d => d.Id == "WLSG0001");
        }
    }
}
