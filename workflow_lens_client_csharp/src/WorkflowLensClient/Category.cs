using System;

namespace WorkflowLensClient
{
    /// <summary>ログカテゴリ。</summary>
    public enum Category
    {
        Asset,
        Build,
        Edit,
        Error,
        Session,
    }

    /// <summary>Categoryの拡張メソッド。</summary>
    internal static class CategoryExtensions
    {
        /// <summary>JSONに使う小文字文字列を返す。</summary>
        internal static string ToJsonString(this Category category) => category switch
        {
            Category.Asset => "asset",
            Category.Build => "build",
            Category.Edit => "edit",
            Category.Error => "error",
            Category.Session => "session",
            _ => throw new ArgumentOutOfRangeException(nameof(category)),
        };
    }
}
