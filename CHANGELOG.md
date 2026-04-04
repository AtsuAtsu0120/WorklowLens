# Changelog

## v0.1.0 (2026-04-04)

### 概要
tool_loggerプロジェクトの初回リリース。ゲームツールからのログをUDP経由で収集・保存・閲覧するシステム一式を提供する。

### 新機能
- tool_logger_middleware: UDP受信→HTTP転送の中継サーバー（Go実装）
- tool_logger_server: ログ保存・閲覧のWebサーバー（Go + PostgreSQL）
- tool_logger_client_csharp: C# (.NET Standard 2.1) 用UDPログ送信クライアントライブラリ
- tool_logger_client_python: Python 3.7+ 用UDPログ送信クライアントライブラリ

### その他
- READMEにセットアップ手順・アーキテクチャ説明を追加
- 各サブプロジェクトに仕様書（Document/）を整備
