using System;
using System.Text.Json;
using Xunit;

namespace WorkflowLensClient.Tests
{
    public class LogMessageTests
    {
        [Fact]
        public void BuildJson_必須フィールドのみ()
        {
            var json = LogMessage.BuildJson("my_tool", "edit", "brush_apply", "abc12345", null, null, null);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("my_tool", root.GetProperty("tool_name").GetString());
            Assert.Equal("edit", root.GetProperty("category").GetString());
            Assert.Equal("brush_apply", root.GetProperty("action").GetString());
            Assert.Equal("abc12345", root.GetProperty("session_id").GetString());
            Assert.True(root.TryGetProperty("timestamp", out var ts));
            // ISO 8601形式であることを確認
            Assert.True(DateTimeOffset.TryParse(ts.GetString(), out _));
            // オプションフィールドは省略
            Assert.False(root.TryGetProperty("tool_version", out _));
            Assert.False(root.TryGetProperty("user_id", out _));
            Assert.False(root.TryGetProperty("duration_ms", out _));
        }

        [Fact]
        public void BuildJson_全フィールドあり()
        {
            var json = LogMessage.BuildJson("my_tool", "error", "shader_compile", "abc12345", "1.2.3", "tanaka", 3200);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("my_tool", root.GetProperty("tool_name").GetString());
            Assert.Equal("error", root.GetProperty("category").GetString());
            Assert.Equal("shader_compile", root.GetProperty("action").GetString());
            Assert.Equal("abc12345", root.GetProperty("session_id").GetString());
            Assert.Equal("1.2.3", root.GetProperty("tool_version").GetString());
            Assert.Equal("tanaka", root.GetProperty("user_id").GetString());
            Assert.Equal(3200, root.GetProperty("duration_ms").GetInt64());
        }

        [Fact]
        public void BuildJson_特殊文字のエスケープ()
        {
            var json = LogMessage.BuildJson("tool", "edit", "行1\n行2\t\"引用\"", "sess", null, null, null);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("行1\n行2\t\"引用\"", root.GetProperty("action").GetString());
        }

        [Fact]
        public void BuildJson_タイムスタンプがISO8601形式()
        {
            var json = LogMessage.BuildJson("tool", "edit", "test", "sess", null, null, null);
            using var doc = JsonDocument.Parse(json);
            var ts = doc.RootElement.GetProperty("timestamp").GetString()!;

            // ISO 8601形式: "2026-04-04T10:00:00.0000000Z" のようなパターン
            var parsed = DateTimeOffset.Parse(ts);
            Assert.Equal(TimeSpan.Zero, parsed.Offset);
        }
    }
}
