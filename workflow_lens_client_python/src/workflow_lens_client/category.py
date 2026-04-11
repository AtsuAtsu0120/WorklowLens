"""ログカテゴリの定数。"""

from enum import Enum


class Category(Enum):
    """ログカテゴリ。"""

    ASSET = "asset"
    BUILD = "build"
    EDIT = "edit"
    ERROR = "error"
    SESSION = "session"
