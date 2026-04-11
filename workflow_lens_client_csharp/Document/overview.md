# workflow_lens_client_csharp

## 概要

workflow_lens_middlewareへUDPでログを送信するC#クライアントライブラリ。
主にUnity上のゲーム開発ツールからの利用を想定する。

## アーキテクチャ

```
┌─────────────────────┐     ┌──────────────────┐
│ Unity ツール (C#)    │     │                  │
│                     │──UDP──▶ workflow_lens      │
│ WorkflowLens クラス    │     │  middleware       │
└─────────────────────┘     └──────────────────┘
  このライブラリ              ローカル中継
```

- **通信プロトコル**: UDP（1データグラム = 1 JSONメッセージ）
- **デフォルト送信先**: `127.0.0.1:59100`
- **fire-and-forget**: 送信失敗時は例外を投げない

## モジュール一覧

| ファイル | 役割 |
|---------|------|
| `src/WorkflowLensClient/WorkflowLens.cs` | メインクラス。UdpClientでJSON送信、MeasureScope |
| `src/WorkflowLensClient/LogMessage.cs` | JSONペイロードの組み立て |
| `src/WorkflowLensClient/Category.cs` | カテゴリenum定義 |
| `src/WorkflowLensClient/WorkflowLensOptions.cs` | コンストラクタ設定のOptionsクラス |
| `src/WorkflowLensClient/CategoryLogger.cs` | カテゴリ固定のログ送信プロパティ |
| `src/WorkflowLensClient.Generators/` | Source Generator + Analyzer（netstandard2.0） |

## 設計判断

| 判断 | 選択 | 理由 |
|------|------|------|
| ターゲット | .NET Standard 2.1 | Unity 2021.2+対応。Nullable参照型が使える |
| middleware探索 | PATH → 環境変数 → 明示パス | 開発者がバイナリパスを管理する手間を削減 |
| JSON組み立て | 文字列補間 | 固定構造なのでシリアライザ不要。外部依存ゼロ |
| スレッドセーフ | ロック不要 | UdpClient.Send()自体がスレッドセーフ |
| エラーハンドリング | サイレントキャッチ | ログ送信がツール本体を壊してはならない |
| session_id | Guid先頭8文字 | 短くて実用上十分 |
| user_id | デフォルトでEnvironment.UserName | ツール制作者の追加負担ゼロ |

## 機能仕様インデックス

| 機能名 | ファイル | status |
|--------|---------|--------|
| UDP送信 | [udp-sender.md](features/udp-sender.md) | implemented |
| セッション管理 | [session-management.md](features/session-management.md) | implemented |
| Middlewareプロセス管理 | [middleware-process.md](features/middleware-process.md) | implemented |
| Optionsパターン | [options-pattern.md](features/options-pattern.md) | implemented |
| セッション自動化 | [auto-session.md](features/auto-session.md) | implemented |
| カテゴリロガー | [category-scope.md](features/category-scope.md) | implemented |
| Source Generator | [source-generator.md](features/source-generator.md) | implemented |
| Analyzer | [analyzer.md](features/analyzer.md) | implemented |
