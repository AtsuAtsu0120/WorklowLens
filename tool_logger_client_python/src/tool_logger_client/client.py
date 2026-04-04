"""tool_logger_middlewareへUDPでログを送信するクライアント。"""

import socket
import threading
import uuid
from typing import Any, Dict, Optional

from . import event_type as EventType
from .log_message import build_json


class ToolLogger:
    """tool_logger_middlewareへUDPでログを送信するクライアント。

    スレッドセーフ。コンテキストマネージャ対応。
    """

    def __init__(
        self,
        tool_name: str,
        tool_version: Optional[str] = None,
        host: str = "127.0.0.1",
        port: int = 59100,
    ) -> None:
        self._tool_name = tool_name
        self._tool_version = tool_version
        self._host = host
        self._port = port
        self._session_id = self._generate_session_id()
        self._sock: Optional[socket.socket] = socket.socket(
            socket.AF_INET, socket.SOCK_DGRAM
        )
        self._lock = threading.Lock()

    @property
    def session_id(self) -> str:
        """現在のセッションID。"""
        return self._session_id

    def start_session(
        self,
        message: str = "Session started",
        details: Optional[Dict[str, Any]] = None,
    ) -> None:
        """セッションを開始する。新しいsession_idを生成し、session_startイベントを送信する。"""
        self._session_id = self._generate_session_id()
        self.send(EventType.SESSION_START, message, details)

    def end_session(
        self,
        message: str = "Session ended",
        details: Optional[Dict[str, Any]] = None,
    ) -> None:
        """セッションを終了する。session_endイベントを送信する。"""
        self.send(EventType.SESSION_END, message, details)

    def send(
        self,
        event_type: str,
        message: str,
        details: Optional[Dict[str, Any]] = None,
    ) -> None:
        """ログメッセージを送信する。"""
        try:
            with self._lock:
                sock = self._sock
                if sock is None:
                    return
                json_str = build_json(
                    self._tool_name,
                    event_type,
                    message,
                    self._session_id,
                    self._tool_version,
                    details,
                )
                sock.sendto(json_str.encode("utf-8"), (self._host, self._port))
        except OSError:
            pass

    def log_usage(
        self, message: str, details: Optional[Dict[str, Any]] = None
    ) -> None:
        """使用ログを送信する。"""
        self.send(EventType.USAGE, message, details)

    def log_error(
        self, message: str, details: Optional[Dict[str, Any]] = None
    ) -> None:
        """エラーログを送信する。"""
        self.send(EventType.ERROR, message, details)

    def log_cancellation(
        self, message: str, details: Optional[Dict[str, Any]] = None
    ) -> None:
        """キャンセルログを送信する。"""
        self.send(EventType.CANCELLATION, message, details)

    def close(self) -> None:
        """ソケットを閉じる。"""
        with self._lock:
            if self._sock is not None:
                self._sock.close()
                self._sock = None

    def __enter__(self) -> "ToolLogger":
        """コンテキストマネージャ: start_sessionを自動呼出し。"""
        self.start_session()
        return self

    def __exit__(self, *args: object) -> None:
        """コンテキストマネージャ: end_session + close。"""
        self.end_session()
        self.close()

    @staticmethod
    def _generate_session_id() -> str:
        return uuid.uuid4().hex[:8]
