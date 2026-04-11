# workflow_lens_client_python

workflow_lens_middlewareへUDPでログを送信するPythonクライアントライブラリ。

## テスト

```bash
pip install -e ".[dev]"   # 開発用インストール
pytest -v                  # 全テスト
```

## プロジェクト構成

```
src/workflow_lens_client/
  client.py                 メインクラス（socket, threading.Lock）
  log_message.py            JSONペイロード組み立て
  event_type.py             イベント種別定数
  __init__.py               公開APIのre-export
tests/
  test_log_message.py       JSON組み立てのユニットテスト
  test_client.py            UDP送信の結合テスト・セッションテスト
Document/                   仕様書 (SDD: 仕様書駆動設計)
```

## 仕様書駆動設計 (SDD)

ルートの `CLAUDE.md` を参照。機能の実装・修正前に必ず `Document/features/` の該当仕様書を確認すること。

## コーディング規約

- Python 3.7+
- コメント・ドキュメントは日本語
- 外部依存なし（標準ライブラリのみ）
- テストは pytest

## ユーザーコンテキスト

C#に精通、Goもある程度わかる。冗長な言語基礎の説明は不要。
