"""client モジュールのテスト。"""

import json
import socket

from tool_logger_client import ToolLogger


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
            logger = ToolLogger("test_tool", "1.0.0", port=port)
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
            logger = ToolLogger("test_tool", port=port)
            logger.log_usage("操作", {"action": "click", "target": "button_a"})

            result = _receive_one(receiver)
            assert result["details"]["action"] == "click"
            assert result["details"]["target"] == "button_a"
            logger.close()
        finally:
            receiver.close()

    def test_middleware未起動でも例外を投げない(self):
        logger = ToolLogger("test_tool", port=59199)
        logger.log_usage("テスト")  # 例外が発生しないこと
        logger.close()

    def test_close後でも例外を投げない(self):
        logger = ToolLogger("test_tool", port=59199)
        logger.close()
        logger.log_usage("テスト")  # 例外が発生しないこと


class TestSession:
    def test_コンストラクタでsession_idが生成される(self):
        logger = ToolLogger("test_tool", port=59199)
        assert logger.session_id is not None
        assert len(logger.session_id) == 8
        logger.close()

    def test_start_sessionでsession_startが送信される(self):
        port = _get_free_port()
        receiver = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        receiver.bind(("127.0.0.1", port))
        receiver.settimeout(3)

        try:
            logger = ToolLogger("test_tool", port=port)
            logger.start_session()

            result = _receive_one(receiver)
            assert result["event_type"] == "session_start"
            logger.close()
        finally:
            receiver.close()

    def test_start_sessionで新しいsession_idが生成される(self):
        logger = ToolLogger("test_tool", port=59199)
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
            logger = ToolLogger("test_tool", port=port)
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
            logger = ToolLogger("test_tool", port=port)
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
            with ToolLogger("test_tool", port=port) as logger:
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
