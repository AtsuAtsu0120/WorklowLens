"""オプショナルなOpenTelemetryトレーシングサポート。

opentelemetry-apiがインストールされている場合はTracerを使用し、
未インストールの場合はno-opで動作する。
"""

try:
    from opentelemetry import trace

    tracer = trace.get_tracer("workflow_lens_client", "0.1.0")
    _HAS_OTEL = True
except ImportError:
    tracer = None
    _HAS_OTEL = False


def start_span(name, attributes=None):
    """スパンを開始する。OTel未インストール時はno-opコンテキストマネージャを返す。"""
    if tracer is None:
        return _NoOpContextManager()
    return tracer.start_as_current_span(name, attributes=attributes)


def get_traceparent():
    """現在のスパンコンテキストからtraceparent文字列を生成する。

    OTel未インストール時またはスパンが無効な場合はNoneを返す。
    """
    if not _HAS_OTEL:
        return None
    span = trace.get_current_span()
    ctx = span.get_span_context()
    if not ctx.is_valid:
        return None
    flags = "01" if ctx.trace_flags.sampled else "00"
    return f"00-{format(ctx.trace_id, '032x')}-{format(ctx.span_id, '016x')}-{flags}"


class _NoOpContextManager:
    """OTel未インストール時に使用するno-opコンテキストマネージャ。"""

    def __enter__(self):
        return self

    def __exit__(self, *args):
        pass
