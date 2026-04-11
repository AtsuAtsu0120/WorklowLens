---
title: "OpenTelemetry対応"
status: implemented
priority: medium
created: 2026-04-11
updated: 2026-04-11
related_files:
  - src/workflow_lens_client/client.py
  - src/workflow_lens_client/log_message.py
  - pyproject.toml
---

# OpenTelemetry対応

## 概要

PythonクライアントライブラリにOpenTelemetryのトレーシング対応を追加する。`opentelemetry-api`をオプショナル依存とし、未インストール時はno-opで動作する。

## 背景・目的

C#クライアントと同様に、Python側でもログ送信にスパンを生成し、traceparentをJSONペイロードに埋め込むことで、クライアント→middleware→サーバーの一貫したトレースを実現する。Maya環境ではOTel SDKをインストールできない場合もあるため、オプショナル依存として設計する。

## 要件

- [ ] `send()`呼び出し時にスパンを開始する
- [ ] スパンのTraceId/SpanIdからtraceparentを生成し、JSONペイロードに注入する
- [ ] `opentelemetry-api`が未インストールの場合でも正常に動作する（no-op）
- [ ] `pip install workflow-lens-client[telemetry]`でOTel依存を追加インストールできる

## 設計

### opentelemetry-apiのオプショナルインポート

`client.py`のモジュールレベルでtry/importを行い、未インストール時はno-opフラグを立てる。

```python
try:
    from opentelemetry import trace
    _tracer = trace.get_tracer("workflow_lens_client")
    _HAS_OTEL = True
except ImportError:
    _tracer = None
    _HAS_OTEL = False
```

### send()でのスパン開始

```python
def send(self, event_type: str, message: str, details: dict = None) -> None:
    traceparent = None

    if _HAS_OTEL:
        span = _tracer.start_span(
            "workflowlens.send",
            kind=trace.SpanKind.PRODUCER,
        )
        ctx = trace.set_span_in_context(span)
    else:
        span = None
        ctx = None

    try:
        if span is not None:
            span.set_attribute("tool.name", self._tool_name)
            span.set_attribute("event.type", event_type)
            span.set_attribute("session.id", self._session_id)

            # traceparent生成
            span_context = span.get_span_context()
            trace_id = format(span_context.trace_id, '032x')
            span_id = format(span_context.span_id, '016x')
            trace_flags = format(span_context.trace_flags, '02x')
            traceparent = f"00-{trace_id}-{span_id}-{trace_flags}"

        # JSONペイロード組み立て（traceparent含む）
        payload = LogMessage.build(
            self._tool_name, event_type, message,
            self._session_id, self._tool_version,
            details, traceparent)

        payload_bytes = payload.encode("utf-8")

        if span is not None:
            span.set_attribute(
                "messaging.message.payload_size_bytes",
                len(payload_bytes))

        # UDP送信
        self._sock.sendto(payload_bytes, self._address)
    except Exception:
        pass  # fire-and-forget
    finally:
        if span is not None:
            span.end()
```

### 属性

| 属性 | 型 | 説明 |
|------|-----|------|
| `tool.name` | string | ツール名 |
| `event.type` | string | イベント種別 |
| `session.id` | string | セッションID |
| `messaging.message.payload_size_bytes` | int | JSONペイロードのバイト数 |

### traceparent注入

スパンが有効な場合（`opentelemetry-api`がインストール済みかつSDKが構成されている場合）、W3C Trace Context形式のtraceparentをJSONペイロードに埋め込む。

**形式**: `00-{trace_id}-{span_id}-{trace_flags}`

**JSONへの埋め込み**:

```json
{
  "tool_name": "MayaBrushTool",
  "event_type": "usage",
  "timestamp": "2026-04-11T10:00:00Z",
  "message": "Brush applied",
  "session_id": "mbt-a1b2c3d4",
  "traceparent": "00-4bf92f3577b86cd53f044612e066a272-00f067aa0ba902b7-01"
}
```

`opentelemetry-api`が未インストールの場合、`traceparent`フィールドは省略する。

### LogMessageへの変更

`LogMessage.build()`に`traceparent`パラメータを追加する。

```python
@staticmethod
def build(
    tool_name: str, event_type: str, message: str,
    session_id: str = None, tool_version: str = None,
    details: dict = None, traceparent: str = None,
) -> str:
```

`traceparent`がNoneでない場合、JSONに`"traceparent": "..."`フィールドを追加する。

### pyproject.tomlの変更

オプショナル依存として`telemetry`エクストラを追加する。

```toml
[project.optional-dependencies]
telemetry = [
    "opentelemetry-api>=1.20.0",
]
dev = [
    "pytest>=7.0",
]
```

インストール方法:

```bash
# テレメトリなし（従来どおり）
pip install workflow-lens-client

# テレメトリあり
pip install workflow-lens-client[telemetry]
```

### 外部依存

| パッケージ | 必須/オプション | 用途 |
|-----------|---------------|------|
| `opentelemetry-api` | オプション（`[telemetry]`） | トレーシングAPI |

`opentelemetry-sdk`やエクスポーターはホストアプリ側で構成する。ライブラリはAPIのみに依存する。

## テスト方針

- [ ] 単体テスト: opentelemetry-api未インストール時にsend()が正常に動作すること
- [ ] 単体テスト: opentelemetry-apiインストール時にスパンが生成され、属性が正しく設定されること
- [ ] 単体テスト: traceparentがW3C形式でJSONに埋め込まれること
- [ ] 単体テスト: opentelemetry-api未インストール時にtraceparentフィールドがJSONに含まれないこと
- [ ] 単体テスト: LogMessage.build()にtraceparent=Noneを渡した場合の後方互換性

## 実装メモ

<!-- 実装時に追加 -->

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
