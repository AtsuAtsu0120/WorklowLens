# workflow_lens_client_csharp

workflow_lens_middlewareへUDPでログを送信するC#クライアントライブラリ。

## ビルド・テスト

```bash
dotnet build              # ビルド
dotnet test               # 全テスト
dotnet pack               # NuGetパッケージ作成
```

## プロジェクト構成

```
src/WorkflowLensClient/
  WorkflowLens.cs             メインクラス（UdpClient, IDisposable）
  WorkflowLensOptions.cs      コンストラクタ設定のOptionsクラス
  CategoryLogger.cs           カテゴリ固定ログ送信（IDisposable自動計測）
  LogMessage.cs               JSONペイロード組み立て
  Category.cs                 カテゴリenum定義
src/WorkflowLensClient.Generators/
  CategoryApiGenerator.cs     [WorkflowActions] enum → 型安全API生成
  WorkflowLogGenerator.cs     [WorkflowLog] → Core/ラッパー自動計測生成
  Analyzers/                  WL0001〜WL0003 Analyzer
tests/WorkflowLensClient.Tests/
  WorkflowLensTests.cs        UDP送信・セッション・CategoryLogger結合テスト
  LogMessageTests.cs          JSON組み立てのユニットテスト
tests/WorkflowLensClient.Generators.Tests/
  CategoryApiGeneratorTests.cs  Generator生成コードの検証
  WorkflowLogGeneratorTests.cs  Generator生成コードの検証
  AnalyzerTests.cs              Analyzer診断の検証
Document/                       仕様書 (SDD: 仕様書駆動設計)
```

## 仕様書駆動設計 (SDD)

ルートの `CLAUDE.md` を参照。機能の実装・修正前に必ず `Document/features/` の該当仕様書を確認すること。

## コーディング規約

- .NET Standard 2.1 / C# 8.0
- コメント・ドキュメントは日本語
- 外部依存: System.Diagnostics.DiagnosticSource (OTel用), Microsoft.CodeAnalysis.CSharp 4.0.1 (Generator用)
- テストは xunit

## ユーザーコンテキスト

C#に精通。冗長な言語基礎の説明は不要。
