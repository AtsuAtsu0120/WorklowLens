"""log_message モジュールのテスト。"""

import json
from datetime import datetime

from workflow_lens_client.log_message import build_json


class TestBuildJson:
    def test_必須フィールドのみ(self):
        result = json.loads(build_json("my_tool", "edit", "brush_apply", "abc12345"))

        assert result["tool_name"] == "my_tool"
        assert result["category"] == "edit"
        assert result["action"] == "brush_apply"
        assert result["session_id"] == "abc12345"
        assert "timestamp" in result
        # ISO 8601形式であることを確認
        datetime.fromisoformat(result["timestamp"])
        # 任意フィールドは省略
        assert "tool_version" not in result
        assert "user_id" not in result
        assert "duration_ms" not in result

    def test_全フィールドあり(self):
        result = json.loads(
            build_json("my_tool", "error", "shader_compile", "abc12345", "1.2.3", "tanaka", 3200)
        )

        assert result["tool_name"] == "my_tool"
        assert result["category"] == "error"
        assert result["action"] == "shader_compile"
        assert result["session_id"] == "abc12345"
        assert result["tool_version"] == "1.2.3"
        assert result["user_id"] == "tanaka"
        assert result["duration_ms"] == 3200

    def test_タイムスタンプがISO8601_UTC(self):
        result = json.loads(build_json("tool", "edit", "test", "sess"))
        ts = result["timestamp"]
        parsed = datetime.fromisoformat(ts)
        # UTCオフセットが+00:00であること
        assert parsed.utcoffset().total_seconds() == 0

    def test_日本語アクション(self):
        result = json.loads(build_json("tool", "edit", "ブラシ適用", "sess"))
        assert result["action"] == "ブラシ適用"
