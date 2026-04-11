"""カテゴリを固定したログ送信オブジェクト。"""

import functools
import time
from typing import TYPE_CHECKING, Any, Callable, Optional, TypeVar

from .category import Category

if TYPE_CHECKING:
    from .client import WorkflowLens

T = TypeVar("T")


class CategoryLogger:
    """カテゴリ固定のログ送信オブジェクト。

    action指定時はコンテキストマネージャ/デコレータとして使用でき、
    ブロック全体の所要時間を自動計測してexit時に送信する。
    action省略時はグルーピング用（exit時にログは送信しない）。
    """

    __slots__ = ("_logger", "_category", "_action", "_start_time")

    def __init__(
        self,
        logger: "WorkflowLens",
        category: Category,
        action: Optional[str] = None,
    ) -> None:
        self._logger = logger
        self._category = category
        self._action = action
        self._start_time: Optional[float] = None

    def log(self, action: str, duration_ms: Optional[int] = None) -> None:
        """指定actionでログを即時送信する。"""
        self._logger.log(self._category, action, duration_ms=duration_ms)

    def __enter__(self) -> "CategoryLogger":
        if self._action is not None:
            self._start_time = time.perf_counter()
        return self

    def __exit__(self, *args: object) -> None:
        if self._action is not None and self._start_time is not None:
            elapsed_ms = int((time.perf_counter() - self._start_time) * 1000)
            self._logger.log(self._category, self._action, duration_ms=elapsed_ms)

    def __call__(self, func: Callable[..., T]) -> Callable[..., T]:
        """デコレータとして使用: @logger.build("compile")"""
        if self._action is None:
            raise ValueError("デコレータとして使用する場合はactionを指定してください")

        @functools.wraps(func)
        def wrapper(*args: Any, **kwargs: Any) -> T:
            start = time.perf_counter()
            try:
                return func(*args, **kwargs)
            finally:
                elapsed_ms = int((time.perf_counter() - start) * 1000)
                self._logger.log(self._category, self._action, duration_ms=elapsed_ms)

        return wrapper
