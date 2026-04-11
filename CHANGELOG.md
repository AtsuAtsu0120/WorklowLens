# Changelog

## v1.0.0 (2026-04-12)

### 概要
OpenTelemetry分散トレーシング対応、ログスキーマv2への移行、クライアントAPIの大幅強化を含むメジャーリリース。

### 新機能
- Go middleware/serverにOpenTelemetry計装を追加（OTLP gRPCエクスポーター、スパン・メトリクス）
- C#/PythonクライアントにOpenTelemetryトレーシングを追加（クライアント→middleware→serverの分散トレース）
- middlewareにHTTPバッチ転送を実装（100件 or 5秒ごとにバッチPOST、W3C Trace Context対応）
- モニタリング基盤をOTel Collector + Prometheusに移行、GrafanaダッシュボードをPromQLベースに刷新
- ログスキーマv2: category+actionの2層構造に統合、user_id自動取得・duration_ms対応（**Breaking Change: DB再作成が必要**）
- C#クライアントAPI最適化: Options Pattern、AutoSession、CategoryLogger、Source Generator、Roslyn Analyzer
- PythonクライアントAPI最適化: Options Pattern、AutoSession、CategoryLogger、デコレータ対応

### バグ修正
- .gitignoreのパターン修正（cmd/配下のmain.goが意図せず除外されていた問題）

### その他
- tool_logger_* → workflow_lens_* にリネーム
- README・仕様書をOTel対応 + HTTP転送に合わせて更新

---

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
