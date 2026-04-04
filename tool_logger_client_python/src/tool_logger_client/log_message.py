"""JSONペイロードの組み立て。"""

import json
from datetime import datetime, timezone
from typing import Any, Dict, Optional


def build_json(
    tool_name: str,
    event_type: str,
    message: str,
    session_id: str,
    tool_version: Optional[str] = None,
    details: Optional[Dict[str, Any]] = None,
) -> str:
    """ログメッセージのJSON文字列を生成する。"""
    payload: Dict[str, Any] = {
        "tool_name": tool_name,
        "event_type": event_type,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "message": message,
        "session_id": session_id,
    }

    if tool_version is not None:
        payload["tool_version"] = tool_version

    if details is not None:
        payload["details"] = details

    return json.dumps(payload, ensure_ascii=False)
