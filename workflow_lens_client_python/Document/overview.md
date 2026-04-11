# workflow_lens_client_python

## 概要

workflow_lens_middlewareへUDPでログを送信するPythonクライアントライブラリ。
主にMaya上のゲーム開発ツールからの利用を想定する。

## アーキテクチャ

```
┌─────────────────────┐     ┌──────────────────┐
│ Maya ツール (Python) │     │                  │
│                     │──UDP──▶ workflow_lens    │
│ WorkflowLens クラス  │     │  middleware       │
└─────────────────────┘     └──────────────────┘
  このライブラリ              ローカル中継
```

- **通信プロトコル**: UDP（1データグラム = 1 JSONメッセージ）
- **デフォルト送信先**: `127.0.0.1:59100`
- **fire-and-forget**: 送信失敗時は例外を投げない

## モジュール一覧

| ファイル | 役割 |
|---------|------|
| `src/workflow_lens_client/client.py` | メインクラス。socketでJSON送信 |
| `src/workflow_lens_client/log_message.py` | JSONペイロードの組み立て |
| `src/workflow_lens_client/event_type.py` | イベント種別の定数定義 |
| `src/workflow_lens_client/__init__.py` | 公開APIのre-export |

## 設計判断

| 判断 | 選択 | 理由 |
|------|------|------|
| Python最低バージョン | 3.7 | Maya 2022+がPython 3.7+ |
| middleware探索 | PATH → 環境変数 → 明示パス | 開発者がバイナリパスを管理する手間を削減 |
| details型 | `dict` | 標準ライブラリの`json.dumps()`でシリアライズ |
| スレッドセーフ | `threading.Lock` | socketの`sendto()`はスレッドセーフではない |
| コンテキストマネージャ | 対応 | `with`文でstart_session/end_sessionを自動化 |
| エラーハンドリング | サイレントキャッチ | ログ送信がツール本体を壊してはならない |
| session_id | UUID先頭8文字 | 短くて実用上十分 |

## 機能仕様インデックス

| 機能名 | ファイル | status |
|--------|---------|--------|
| UDP送信 | [udp-sender.md](features/udp-sender.md) | implemented |
| セッション管理 | [session-management.md](features/session-management.md) | implemented |
| Middlewareプロセス管理 | [middleware-process.md](features/middleware-process.md) | implemented |
