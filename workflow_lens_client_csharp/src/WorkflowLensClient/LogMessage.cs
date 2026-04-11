using System;
using System.Text;

namespace WorkflowLensClient
{
    /// <summary>JSONペイロードを組み立てる内部ヘルパー。</summary>
    internal static class LogMessage
    {
        /// <summary>ログメッセージのJSON文字列を生成する。</summary>
        internal static string BuildJson(
            string toolName,
            string category,
            string action,
            string? sessionId,
            string? toolVersion,
            string? userId,
            long? durationMs,
            string? traceparent = null)
        {
            var sb = new StringBuilder(256);
            sb.Append('{');
            sb.Append("\"tool_name\":").Append(JsonEscape(toolName));
            sb.Append(",\"category\":").Append(JsonEscape(category));
            sb.Append(",\"action\":").Append(JsonEscape(action));
            sb.Append(",\"timestamp\":").Append(JsonEscape(DateTime.UtcNow.ToString("o")));

            if (sessionId != null)
            {
                sb.Append(",\"session_id\":").Append(JsonEscape(sessionId));
            }

            if (toolVersion != null)
            {
                sb.Append(",\"tool_version\":").Append(JsonEscape(toolVersion));
            }

            if (userId != null)
            {
                sb.Append(",\"user_id\":").Append(JsonEscape(userId));
            }

            if (durationMs != null)
            {
                sb.Append(",\"duration_ms\":").Append(durationMs.Value);
            }

            if (traceparent != null)
            {
                sb.Append(",\"traceparent\":").Append(JsonEscape(traceparent));
            }

            sb.Append('}');
            return sb.ToString();
        }

        /// <summary>文字列をJSONの二重引用符付き文字列にエスケープする。</summary>
        private static string JsonEscape(string value)
        {
            var sb = new StringBuilder(value.Length + 2);
            sb.Append('"');
            foreach (var c in value)
            {
                switch (c)
                {
                    case '"': sb.Append("\\\""); break;
                    case '\\': sb.Append("\\\\"); break;
                    case '\n': sb.Append("\\n"); break;
                    case '\r': sb.Append("\\r"); break;
                    case '\t': sb.Append("\\t"); break;
                    default:
                        if (c < 0x20)
                            sb.AppendFormat("\\u{0:x4}", (int)c);
                        else
                            sb.Append(c);
                        break;
                }
            }
            sb.Append('"');
            return sb.ToString();
        }
    }
}
