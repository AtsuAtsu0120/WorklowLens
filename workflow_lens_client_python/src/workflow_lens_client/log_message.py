"""JSONペイロードの組み立て。"""

import json
from datetime import datetime, timezone
from typing import Any, Dict, Optional


def build_json(
    tool_name: str,
    category: str,
    action: str,
    session_id: str,
    tool_version: Optional[str] = None,
    user_id: Optional[str] = None,
    duration_ms: Optional[int] = None,
    traceparent: Optional[str] = None,
) -> str:
    """ログメッセージのJSON文字列を生成する。"""
    payload: Dict[str, Any] = {
        "tool_name": tool_name,
        "category": category,
        "action": action,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "session_id": session_id,
    }

    if tool_version is not None:
        payload["tool_version"] = tool_version

    if user_id is not None:
        payload["user_id"] = user_id

    if duration_ms is not None:
        payload["duration_ms"] = duration_ms

    if traceparent is not None:
        payload["traceparent"] = traceparent

    return json.dumps(payload, ensure_ascii=False)
