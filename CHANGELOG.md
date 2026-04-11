# Changelog

## v0.1.0 (2026-04-04)

### 概要
workflow_lensプロジェクトの初回リリース。ゲームツールからのログをUDP経由で収集・保存・閲覧するシステム一式を提供する。

### 新機能
- workflow_lens_middleware: UDP受信→HTTP転送の中継サーバー（Go実装）
- workflow_lens_server: ログ保存・閲覧のWebサーバー（Go + PostgreSQL）
- workflow_lens_client_csharp: C# (.NET Standard 2.1) 用UDPログ送信クライアントライブラリ
- workflow_lens_client_python: Python 3.7+ 用UDPログ送信クライアントライブラリ

### その他
- READMEにセットアップ手順・アーキテクチャ説明を追加
- 各サブプロジェクトに仕様書（Document/）を整備
