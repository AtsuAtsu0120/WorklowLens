"""client モジュールのテスト。"""

import json
import socket

import pytest

from workflow_lens_client import WorkflowLens


def _get_free_port() -> int:
    """空きポートを取得する。"""
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


def _receive_one(sock: socket.socket) -> dict:
    """UDPソケットから1メッセージを受信してJSONとしてパースする。"""
    data, _ = sock.recvfrom(65536)
    return json.loads(data.decode("utf-8"))


class TestSend:
    def test_UDPで正しいJSONが届く(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", "1.0.0", port=port, auto_start_middleware=False)
            logger.log_usage("ボタン押下")

            result = _receive_one(receiver)
            assert result["tool_name"] == "test_tool"
            assert result["event_type"] == "usage"
            assert result["message"] == "ボタン押下"
            assert result["tool_version"] == "1.0.0"
            assert "session_id" in result
            logger.close()
        finally:
            receiver.close()

    def test_detailsが正しく埋め込まれる(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            logger.log_usage("操作", {"action": "click", "target": "button_a"})

            result = _receive_one(receiver)
            assert result["details"]["action"] == "click"
            assert result["details"]["target"] == "button_a"
            logger.close()
        finally:
            receiver.close()

    def test_middleware未起動でも例外を投げない(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        logger.log_usage("テスト")  # 例外が発生しないこと
        logger.close()

    def test_close後でも例外を投げない(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        logger.close()
        logger.log_usage("テスト")  # 例外が発生しないこと


class TestSession:
    def test_コンストラクタでsession_idが生成される(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        assert logger.session_id is not None
        assert len(logger.session_id) == 8
        logger.close()

    def test_start_sessionでsession_startが送信される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            logger.start_session()

            result = _receive_one(receiver)
            assert result["event_type"] == "session_start"
            logger.close()
        finally:
            receiver.close()

    def test_start_sessionで新しいsession_idが生成される(self):
        logger = WorkflowLens("test_tool", port=59199, auto_start_middleware=False)
        first = logger.session_id
        logger.start_session()
        second = logger.session_id
        assert first != second
        logger.close()

    def test_end_sessionでsession_endが送信される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = WorkflowLens("test_tool", port=port, auto_start_middleware=False)
            logger.end_session()

            result = _receive_one(receiver)
            assert result["event_type"] == "session_end"
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
            logger.log_usage("テスト")

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
                logger.log_usage("操作")

            # session_start, usage, session_end の3メッセージが届く
            msg1 = _receive_one(receiver)
            msg2 = _receive_one(receiver)
            msg3 = _receive_one(receiver)

            assert msg1["event_type"] == "session_start"
            assert msg2["event_type"] == "usage"
            assert msg3["event_type"] == "session_end"
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
        logger.close()  # 2回目でも例外が出ないこと


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
