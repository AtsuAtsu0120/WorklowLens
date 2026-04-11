"""client モジュールのテスト。"""

import json
import socket
import time

import pytest

from workflow_lens_client import Category, WorkflowLens


def _get_free_port() -> int:
    """空きポートを取得する。"""
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


def _receive_one(sock: socket.socket) -> dict:
    """UDPソケットから1メッセージを受信してJSONとしてパースする。"""
    data, _ = sock.recvfrom(65536)
    return json.loads(data.decode("utf-8"))


class TestLog:
    def test_UDPで正しいJSONが届く(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", "1.0.0", port=port, auto_start_middleware=False)
            logger.log(Category.EDIT, "brush_apply")

            result = _receive_one(receiver)
            assert result["tool_name"] == "test_tool"
            assert result["category"] == "edit"
            assert result["action"] == "brush_apply"
            assert result["tool_version"] == "1.0.0"
            assert "session_id" in result
            assert "user_id" in result
            logger.close()
        finally:
            receiver.close()

    def test_durationMsが正しく設定される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            logger.log(Category.BUILD, "compile", duration_ms=3200)

            result = _receive_one(receiver)
            assert result["duration_ms"] == 3200
            logger.close()
        finally:
            receiver.close()

    def test_middleware未起動でも例外を投げない(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        logger.log(Category.EDIT, "test")
        logger.close()

    def test_close後でも例外を投げない(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        logger.close()
        logger.log(Category.EDIT, "test")

    def test_measureでduration_msが自動設定される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            with logger.measure(Category.BUILD, "compile"):
                time.sleep(0.05)

            result = _receive_one(receiver)
            assert result["category"] == "build"
            assert result["action"] == "compile"
            assert result["duration_ms"] >= 40
            logger.close()
        finally:
            receiver.close()

    def test_userId未指定時にOSユーザー名が自動設定される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            logger.log(Category.EDIT, "test")

            result = _receive_one(receiver)
            assert result["user_id"] is not None
            assert len(result["user_id"]) > 0
            logger.close()
        finally:
            receiver.close()


class TestSession:
    def test_コンストラクタでsession_idが生成される(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        assert logger.session_id is not None
        assert len(logger.session_id) == 8
        logger.close()

    def test_session_startログが送信される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            logger.log(Category.SESSION, "start")

            result = _receive_one(receiver)
            assert result["category"] == "session"
            assert result["action"] == "start"
            logger.close()
        finally:
            receiver.close()

    def test_session_endログが送信される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            logger.log(Category.SESSION, "end")

            result = _receive_one(receiver)
            assert result["category"] == "session"
            assert result["action"] == "end"
            logger.close()
        finally:
            receiver.close()

    def test_session_idが自動付与される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            expected_id = logger.session_id
            logger.log(Category.EDIT, "test")

            result = _receive_one(receiver)
            assert result["session_id"] == expected_id
            logger.close()
        finally:
            receiver.close()

    def test_コンテキストマネージャでstart_endが自動送信される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            with WorkflowLens("test_tool", port=port, auto_start_middleware=False) as logger:
                logger.log(Category.EDIT, "brush_apply")

            # session/start, edit/brush_apply, session/end の3メッセージが届く
            msg1 = _receive_one(receiver)
            msg2 = _receive_one(receiver)
            msg3 = _receive_one(receiver)

            assert msg1["category"] == "session"
            assert msg1["action"] == "start"
            assert msg2["category"] == "edit"
            assert msg2["action"] == "brush_apply"
            assert msg3["category"] == "session"
            assert msg3["action"] == "end"
        finally:
            receiver.close()


class TestMiddlewareProcess:
    def test_middleware_path未指定でプロセスが起動されない(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        assert logger._process is None
        logger.close()

    def test_存在しないバイナリでFileNotFoundError(self):
        with pytest.raises(FileNotFoundError):
            WorkflowLens("test_tool", port=59199,
                       middleware_path="/nonexistent/middleware")

    def test_middleware_pathが空文字でValueError(self):
        with pytest.raises(ValueError):
            WorkflowLens("test_tool", port=59199, middleware_path="")

    def test_close多重呼び出しで例外が出ない(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        logger.close()
        logger.close()


class TestMiddlewareAutoDiscovery:
    def test_auto_start_middleware_falseでプロセスが起動されない(self):
        result = WorkflowLens._resolve_middleware_path(None, False)
        assert result is None

    def test_middleware_pathが指定されていればそのまま返す(self):
        result = WorkflowLens._resolve_middleware_path("/explicit/path", False)
        assert result == "/explicit/path"

    def test_middleware_pathが指定されていればauto_start_middlewareより優先(self):
        result = WorkflowLens._resolve_middleware_path("/explicit/path", True)
        assert result == "/explicit/path"

    def test_auto_start_middleware_環境変数が設定されていればそれを使う(self, monkeypatch):
        monkeypatch.setenv(WorkflowLens.MIDDLEWARE_PATH_ENV_VAR, "/env/var/path")
        result = WorkflowLens._resolve_middleware_path(None, True)
        assert result == "/env/var/path"

    def test_auto_start_middleware_バイナリが見つからない場合わかりやすいエラー(
        self, monkeypatch
    ):
        monkeypatch.delenv(WorkflowLens.MIDDLEWARE_PATH_ENV_VAR, raising=False)
        monkeypatch.setattr("shutil.which", lambda name: None)
        with pytest.raises(FileNotFoundError, match="ミドルウェアバイナリが見つかりません"):
            WorkflowLens._resolve_middleware_path(None, True)
