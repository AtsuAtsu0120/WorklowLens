"""WorkflowLensの初期化オプション。"""

from dataclasses import dataclass
from typing import Optional


@dataclass
class WorkflowLensOptions:
    """WorkflowLensの設定をまとめるデータクラス。"""

    tool_version: Optional[str] = None
    user_id: Optional[str] = None
    host: str = "127.0.0.1"
    port: int = 59100
    middleware_path: Optional[str] = None
    auto_start_middleware: bool = True
    auto_session: bool = True
