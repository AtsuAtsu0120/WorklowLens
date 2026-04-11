using System.Collections.Generic;
using System.IO;
using System.Linq;
using Microsoft.CodeAnalysis;

namespace WorkflowLensClient.Generators.Tests
{
    /// <summary>
    /// テスト用の共通メタデータ参照取得ヘルパー。
    /// </summary>
    internal static class TestHelper
    {
        public static MetadataReference[] GetMetadataReferences()
        {
            var refs = new List<MetadataReference>
            {
                MetadataReference.CreateFromFile(typeof(object).Assembly.Location),
                MetadataReference.CreateFromFile(typeof(WorkflowLensClient.Category).Assembly.Location),
            };

            // .NET runtimeの参照追加
            var runtimePath = Path.GetDirectoryName(typeof(object).Assembly.Location)!;
            foreach (var name in new[] { "System.Runtime.dll", "netstandard.dll" })
            {
                var path = Path.Combine(runtimePath, name);
                if (File.Exists(path))
                {
                    refs.Add(MetadataReference.CreateFromFile(path));
                }
            }

            return refs.ToArray();
        }
    }
}
