"""workflow_lens_middlewareへUDPでログを送信するクライアント。"""

import functools
import getpass
import os
import shutil
import socket
import subprocess
import threading
import time
import uuid
import warnings
from typing import Any, Callable, Optional, TypeVar

from .category import Category
from .log_message import build_json
from .options import WorkflowLensOptions

T = TypeVar("T")


class _MeasureDecorator:
    """操作時間を自動計測するコンテキストマネージャ兼デコレータ。"""

    __slots__ = ("_logger", "_category", "_action", "_start_time")

    def __init__(self, logger: "WorkflowLens", category: Category, action: str) -> None:
        self._logger = logger
        self._category = category
        self._action = action
        self._start_time: Optional[float] = None

    def __enter__(self) -> "_MeasureDecorator":
        self._start_time = time.perf_counter()
        return self

    def __exit__(self, *args: object) -> None:
        if self._start_time is not None:
            elapsed_ms = int((time.perf_counter() - self._start_time) * 1000)
            self._logger.log(self._category, self._action, duration_ms=elapsed_ms)

    def __call__(self, func: Callable[..., T]) -> Callable[..., T]:
        """デコレータとして使用: @logger.measure(Category.BUILD, "compile")"""

        @functools.wraps(func)
        def wrapper(*args: Any, **kwargs: Any) -> T:
            start = time.perf_counter()
            try:
                return func(*args, **kwargs)
            finally:
                elapsed_ms = int((time.perf_counter() - start) * 1000)
                self._logger.log(self._category, self._action, duration_ms=elapsed_ms)

        return wrapper


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
        *,
        configure: Optional[Callable[[WorkflowLensOptions], None]] = None,
        options: Optional[WorkflowLensOptions] = None,
    ) -> None:
        # オプション解決
        if options is not None:
            opts = options
        elif configure is not None:
            opts = WorkflowLensOptions()
            configure(opts)
        else:
            # 従来の位置引数パス — auto_session=False で後方互換
            opts = WorkflowLensOptions(
                tool_version=tool_version,
                user_id=user_id,
                host=host,
                port=port,
                middleware_path=middleware_path,
                auto_start_middleware=auto_start_middleware,
                auto_session=False,
            )

        self._tool_name = tool_name
        self._tool_version = opts.tool_version
        self._user_id = opts.user_id or self._resolve_user_id()
        self._host = opts.host
        self._port = opts.port
        self._session_id = self._generate_session_id()
        self._sock: Optional[socket.socket] = socket.socket(
            socket.AF_INET, socket.SOCK_DGRAM
        )
        self._lock = threading.Lock()
        self._process: Optional[subprocess.Popen] = None  # type: ignore[type-arg]

        # Auto-Session フラグ
        self._auto_session: bool = opts.auto_session
        self._session_start_sent: bool = False
        self._session_end_sent: bool = False

        resolved_path = self._resolve_middleware_path(
            opts.middleware_path, opts.auto_start_middleware
        )
        if resolved_path is not None:
            try:
                self._process = subprocess.Popen(
                    [resolved_path, str(opts.port)],
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                )
            except Exception:
                self._sock.close()
                self._sock = None
                raise

        # Auto-Session: コンストラクタ完了時に session/start を送信
        if self._auto_session:
            self.log(Category.SESSION, "start")

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

        # バイナリ未発見 — 警告のみ（Mayaプラグイン等の起動時クラッシュを防ぐ）
        warnings.warn(
            f"ミドルウェアバイナリが見つかりません。"
            f"PATH に {cls.MIDDLEWARE_BINARY_NAME} を配置するか、"
            f"環境変数 {cls.MIDDLEWARE_PATH_ENV_VAR} を設定してください。",
            stacklevel=3,
        )
        return None

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
        # セッション重複送信防止
        if category == Category.SESSION:
            if action == "start":
                if self._session_start_sent:
                    return
                self._session_start_sent = True
            elif action == "end":
                if self._session_end_sent:
                    return
                self._session_end_sent = True

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

    def measure(self, category: Category, action: str) -> _MeasureDecorator:
        """操作時間を自動計測するコンテキストマネージャ兼デコレータ。"""
        return _MeasureDecorator(self, category, action)

    # --- Category-Scoped ファクトリメソッド ---

    def asset(self, action: Optional[str] = None) -> "CategoryLogger":
        """Assetカテゴリのスコープロガーを返す。"""
        from .category_logger import CategoryLogger

        return CategoryLogger(self, Category.ASSET, action)

    def build(self, action: Optional[str] = None) -> "CategoryLogger":
        """Buildカテゴリのスコープロガーを返す。"""
        from .category_logger import CategoryLogger

        return CategoryLogger(self, Category.BUILD, action)

    def edit(self, action: Optional[str] = None) -> "CategoryLogger":
        """Editカテゴリのスコープロガーを返す。"""
        from .category_logger import CategoryLogger

        return CategoryLogger(self, Category.EDIT, action)

    def error(self, action: Optional[str] = None) -> "CategoryLogger":
        """Errorカテゴリのスコープロガーを返す。"""
        from .category_logger import CategoryLogger

        return CategoryLogger(self, Category.ERROR, action)

    def close(self) -> None:
        """ソケットを閉じ、middlewareプロセスがあれば停止する。"""
        # Auto-Session: close時に session/end を自動送信
        if self._auto_session and not self._session_end_sent:
            self.log(Category.SESSION, "end")

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
