using System.Text;

namespace WorkflowLensClient.Generators
{
    /// <summary>
    /// PascalCase → snake_case 変換ユーティリティ。
    /// 連続大文字（略語）は最後の1文字を次の単語の先頭として扱う。
    /// 例: ShaderCompile → shader_compile, LODGenerate → lod_generate
    /// </summary>
    internal static class PascalToSnakeCase
    {
        public static string Convert(string pascalCase)
        {
            if (string.IsNullOrEmpty(pascalCase))
                return pascalCase;

            var sb = new StringBuilder();
            for (int i = 0; i < pascalCase.Length; i++)
            {
                var c = pascalCase[i];

                if (char.IsUpper(c))
                {
                    if (i > 0)
                    {
                        // 前の文字が小文字 or 次の文字が小文字（略語の末尾）ならアンダースコア
                        var prevIsLower = char.IsLower(pascalCase[i - 1]);
                        var nextIsLower = i + 1 < pascalCase.Length && char.IsLower(pascalCase[i + 1]);

                        if (prevIsLower || nextIsLower)
                        {
                            sb.Append('_');
                        }
                    }
                    sb.Append(char.ToLowerInvariant(c));
                }
                else
                {
                    sb.Append(c);
                }
            }

            return sb.ToString();
        }
    }
}
