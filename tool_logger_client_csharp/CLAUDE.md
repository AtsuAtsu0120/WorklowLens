# tool_logger_client_csharp

tool_logger_middlewareへUDPでログを送信するC#クライアントライブラリ。

## ビルド・テスト

```bash
dotnet build              # ビルド
dotnet test               # 全テスト
dotnet pack               # NuGetパッケージ作成
```

## プロジェクト構成

```
src/ToolLoggerClient/
  ToolLogger.cs             メインクラス（UdpClient, IDisposable）
  LogMessage.cs             JSONペイロード組み立て
  EventType.cs              イベント種別定数
tests/ToolLoggerClient.Tests/
  LogMessageTests.cs        JSON組み立てのユニットテスト
  ToolLoggerTests.cs        UDP送信の結合テスト・セッションテスト
Document/                   仕様書 (SDD: 仕様書駆動設計)
```

## 仕様書駆動設計 (SDD)

ルートの `CLAUDE.md` を参照。機能の実装・修正前に必ず `Document/features/` の該当仕様書を確認すること。

## コーディング規約

- .NET Standard 2.1 / C# 8.0
- コメント・ドキュメントは日本語
- 外部依存なし（標準ライブラリのみ）
- テストは xunit

## ユーザーコンテキスト

C#に精通。冗長な言語基礎の説明は不要。
