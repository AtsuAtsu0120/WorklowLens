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
            string eventType,
            string message,
            string? sessionId,
            string? toolVersion,
            string? details,
            string? traceparent = null)
        {
            var sb = new StringBuilder(256);
            sb.Append('{');
            sb.Append("\"tool_name\":").Append(JsonEscape(toolName));
            sb.Append(",\"event_type\":").Append(JsonEscape(eventType));
            sb.Append(",\"timestamp\":").Append(JsonEscape(DateTime.UtcNow.ToString("o")));
            sb.Append(",\"message\":").Append(JsonEscape(message));

            if (sessionId != null)
            {
                sb.Append(",\"session_id\":").Append(JsonEscape(sessionId));
            }

            if (toolVersion != null)
            {
                sb.Append(",\"tool_version\":").Append(JsonEscape(toolVersion));
            }

            if (details != null)
            {
                // detailsは生JSON文字列なのでエスケープせずそのまま埋め込む
                sb.Append(",\"details\":").Append(details);
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
