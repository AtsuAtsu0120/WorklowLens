"""workflow_lens_middlewareへUDPでログを送信するクライアント。"""

import getpass
import os
import shutil
import socket
import subprocess
import threading
import time
import uuid
from contextlib import contextmanager
from typing import Iterator, Optional

from .category import Category
from .log_message import build_json


class WorkflowLens:
    """workflow_lens_middlewareへUDPでログを送信するクライアント。

    スレッドセーフ。コンテキストマネージャ対応。
    """

    MIDDLEWARE_PATH_ENV_VAR = "WORKFLOW_LENS_MIDDLEWARE_PATH"
    """環境変数名。"""

    MIDDLEWARE_BINARY_NAME = "workflow_lens_middleware"
    """PATH探索時のバイナリ名。"""

    def __init__(
        self,
        tool_name: str,
        tool_version: Optional[str] = None,
        user_id: Optional[str] = None,
        host: str = "127.0.0.1",
        port: int = 59100,
        middleware_path: Optional[str] = None,
        auto_start_middleware: bool = True,
    ) -> None:
        self._tool_name = tool_name
        self._tool_version = tool_version
        self._user_id = user_id or self._resolve_user_id()
        self._host = host
        self._port = port
        self._session_id = self._generate_session_id()
        self._sock: Optional[socket.socket] = socket.socket(
            socket.AF_INET, socket.SOCK_DGRAM
        )
        self._lock = threading.Lock()
        self._process: Optional[subprocess.Popen] = None  # type: ignore[type-arg]

        resolved_path = self._resolve_middleware_path(
            middleware_path, auto_start_middleware
        )
        if resolved_path is not None:
            try:
                self._process = subprocess.Popen(
                    [resolved_path, str(port)],
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                )
            except Exception:
                self._sock.close()
                self._sock = None
                raise

    @classmethod
    def _resolve_middleware_path(
        cls, middleware_path: Optional[str], auto_start_middleware: bool
    ) -> Optional[str]:
        """middlewareバイナリのパスを解決する。

        優先順位: middleware_path(明示指定) → 環境変数 → PATH探索。
        """
        if middleware_path is not None:
            if not middleware_path:
                raise ValueError("middleware_pathが空文字です")
            return middleware_path

        if not auto_start_middleware:
            return None

        # 環境変数から取得
        env_path = os.environ.get(cls.MIDDLEWARE_PATH_ENV_VAR)
        if env_path:
            return env_path

        # PATH探索
        found = shutil.which(cls.MIDDLEWARE_BINARY_NAME)
        if found is not None:
            return found

        raise FileNotFoundError(
            f"ミドルウェアバイナリが見つかりません。"
            f"PATH に {cls.MIDDLEWARE_BINARY_NAME} を配置するか、"
            f"環境変数 {cls.MIDDLEWARE_PATH_ENV_VAR} を設定してください。"
        )

    @staticmethod
    def _resolve_user_id() -> str:
        """OSユーザー名を取得する。"""
        try:
            return getpass.getuser()
        except Exception:
            return os.environ.get("USER", "unknown")

    @property
    def session_id(self) -> str:
        """現在のセッションID。"""
        return self._session_id

    def log(
        self,
        category: Category,
        action: str,
        duration_ms: Optional[int] = None,
    ) -> None:
        """ログを送信する。"""
        from ._telemetry import get_traceparent, start_span

        try:
            with start_span(
                "workflowlens.send",
                attributes={
                    "tool.name": self._tool_name,
                    "category": category.value,
                    "action": action,
                    "session.id": self._session_id,
                },
            ):
                with self._lock:
                    sock = self._sock
                    if sock is None:
                        return
                    traceparent = get_traceparent()
                    json_str = build_json(
                        self._tool_name,
                        category.value,
                        action,
                        self._session_id,
                        self._tool_version,
                        self._user_id,
                        duration_ms,
                        traceparent=traceparent,
                    )
                    sock.sendto(json_str.encode("utf-8"), (self._host, self._port))
        except OSError:
            pass

    @contextmanager
    def measure(self, category: Category, action: str) -> Iterator[None]:
        """操作時間を自動計測するコンテキストマネージャ。"""
        start = time.perf_counter()
        try:
            yield
        finally:
            elapsed_ms = int((time.perf_counter() - start) * 1000)
            self.log(category, action, duration_ms=elapsed_ms)

    def close(self) -> None:
        """ソケットを閉じ、middlewareプロセスがあれば停止する。"""
        with self._lock:
            if self._sock is not None:
                self._sock.close()
                self._sock = None

        if self._process is not None:
            try:
                self._process.terminate()
                self._process.wait(timeout=5)
            except Exception:
                pass
            self._process = None

    def __enter__(self) -> "WorkflowLens":
        """コンテキストマネージャ: セッション開始を自動送信。"""
        self.log(Category.SESSION, "start")
        return self

    def __exit__(self, *args: object) -> None:
        """コンテキストマネージャ: セッション終了 + close。"""
        self.log(Category.SESSION, "end")
        self.close()

    @staticmethod
    def _generate_session_id() -> str:
        return uuid.uuid4().hex[:8]
