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
            var json = LogMessage.BuildJson("my_tool", "usage", "テスト", "abc12345", null, null);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("my_tool", root.GetProperty("tool_name").GetString());
            Assert.Equal("usage", root.GetProperty("event_type").GetString());
            Assert.Equal("テスト", root.GetProperty("message").GetString());
            Assert.Equal("abc12345", root.GetProperty("session_id").GetString());
            Assert.True(root.TryGetProperty("timestamp", out var ts));
            // ISO 8601形式であることを確認
            Assert.True(DateTimeOffset.TryParse(ts.GetString(), out _));
            // tool_version, detailsは省略
            Assert.False(root.TryGetProperty("tool_version", out _));
            Assert.False(root.TryGetProperty("details", out _));
        }

        [Fact]
        public void BuildJson_全フィールドあり()
        {
            var details = "{\"key\":\"value\",\"count\":42}";
            var json = LogMessage.BuildJson("my_tool", "error", "エラー発生", "abc12345", "1.2.3", details);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("my_tool", root.GetProperty("tool_name").GetString());
            Assert.Equal("error", root.GetProperty("event_type").GetString());
            Assert.Equal("エラー発生", root.GetProperty("message").GetString());
            Assert.Equal("abc12345", root.GetProperty("session_id").GetString());
            Assert.Equal("1.2.3", root.GetProperty("tool_version").GetString());
            Assert.Equal("value", root.GetProperty("details").GetProperty("key").GetString());
            Assert.Equal(42, root.GetProperty("details").GetProperty("count").GetInt32());
        }

        [Fact]
        public void BuildJson_特殊文字のエスケープ()
        {
            var json = LogMessage.BuildJson("tool", "usage", "行1\n行2\t\"引用\"", "sess", null, null);
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            Assert.Equal("行1\n行2\t\"引用\"", root.GetProperty("message").GetString());
        }

        [Fact]
        public void BuildJson_タイムスタンプがISO8601形式()
        {
            var json = LogMessage.BuildJson("tool", "usage", "msg", "sess", null, null);
            using var doc = JsonDocument.Parse(json);
            var ts = doc.RootElement.GetProperty("timestamp").GetString()!;

            // ISO 8601形式: "2026-04-04T10:00:00.0000000Z" のようなパターン
            var parsed = DateTimeOffset.Parse(ts);
            Assert.Equal(TimeSpan.Zero, parsed.Offset);
        }
    }
}
