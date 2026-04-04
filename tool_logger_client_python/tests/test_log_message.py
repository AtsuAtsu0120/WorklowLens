"""log_message モジュールのテスト。"""

import json
from datetime import datetime

from tool_logger_client.log_message import build_json


class TestBuildJson:
    def test_必須フィールドのみ(self):
        result = json.loads(build_json("my_tool", "usage", "テスト", "abc12345"))

        assert result["tool_name"] == "my_tool"
        assert result["event_type"] == "usage"
        assert result["message"] == "テスト"
        assert result["session_id"] == "abc12345"
        assert "timestamp" in result
        # ISO 8601形式であることを確認
        datetime.fromisoformat(result["timestamp"])
        # 任意フィールドは省略
        assert "tool_version" not in result
        assert "details" not in result

    def test_全フィールドあり(self):
        details = {"key": "value", "count": 42}
        result = json.loads(
            build_json("my_tool", "error", "エラー発生", "abc12345", "1.2.3", details)
        )

        assert result["tool_name"] == "my_tool"
        assert result["event_type"] == "error"
        assert result["message"] == "エラー発生"
        assert result["session_id"] == "abc12345"
        assert result["tool_version"] == "1.2.3"
        assert result["details"]["key"] == "value"
        assert result["details"]["count"] == 42

    def test_タイムスタンプがISO8601_UTC(self):
        result = json.loads(build_json("tool", "usage", "msg", "sess"))
        ts = result["timestamp"]
        parsed = datetime.fromisoformat(ts)
        # UTCオフセットが+00:00であること
        assert parsed.utcoffset().total_seconds() == 0

    def test_日本語メッセージ(self):
        result = json.loads(build_json("tool", "usage", "ボタンを押下しました", "sess"))
        assert result["message"] == "ボタンを押下しました"
