---
title: "Analyzer"
status: implemented
priority: low
created: 2026-04-11
updated: 2026-04-11
related_files:
  - src/WorkflowLensClient.Generators/Analyzers/WL0001_DisposeAnalyzer.cs
  - src/WorkflowLensClient.Generators/Analyzers/WL0002_ActionNamingAnalyzer.cs
  - src/WorkflowLensClient.Generators/Analyzers/WL0003_SessionDirectUseAnalyzer.cs
---

# Analyzer

## 概要

Roslyn Analyzerにより、WorkflowLensClientの典型的なミスをコンパイル時に検出する。Source Generatorと同じDLLに同居させる。

## 背景・目的

- Dispose忘れによるリソースリーク検出
- action文字列のtypo・命名規約違反の検出
- AutoSession有効時のSession手動呼び出しの警告

## 要件

### WL0001: Dispose忘れ検出

- [ ] `new WorkflowLens(...)`が`using`文/宣言で囲まれていない場合にWarningを出す
- [ ] フィールド代入の場合は検出対象外（クラスがIDisposableを実装していることを期待）
- [ ] Severity: Warning

### WL0002: action文字列の命名規約チェック

- [ ] `Log()`の第2引数（action）がリテラル文字列の場合、`^[a-z][a-z0-9_]*$`に一致しなければWarningを出す
- [ ] `MeasureScope()`の第2引数も同様
- [ ] 変数・式の場合は検出対象外（静的解析の限界）
- [ ] Severity: Warning

### WL0003: Session直接使用の警告

- [ ] `Log(Category.Session, ...)`の呼び出しを検出してInfoを出す
- [ ] メッセージ: 「AutoSessionが有効な場合、Category.Sessionの直接使用は不要です」
- [ ] Severity: Info

## 設計

### Analyzerプロジェクト配置

Source Generator（`WorkflowLensClient.Generators`）と同じプロジェクトに同居する。

```
src/WorkflowLensClient.Generators/
  Analyzers/
    WL0001_DisposeAnalyzer.cs
    WL0002_ActionNamingAnalyzer.cs
    WL0003_SessionDirectUseAnalyzer.cs
```

### WL0001: Dispose忘れ

```csharp
[DiagnosticAnalyzer(LanguageNames.CSharp)]
public sealed class WL0001_DisposeAnalyzer : DiagnosticAnalyzer
```

検出ロジック:
1. `ObjectCreationExpressionSyntax`で型が`WorkflowLens`のものを検出
2. 親ノードが以下のいずれかであることを確認:
   - `UsingStatementSyntax`（`using (var x = new ...)`）
   - `LocalDeclarationStatementSyntax`で`using`修飾子付き（`using var x = new ...`）
   - `EqualsValueClauseSyntax` → `VariableDeclaratorSyntax` → フィールド宣言
3. いずれにも該当しない場合にWarningを出す

### WL0002: action文字列命名規約

```csharp
[DiagnosticAnalyzer(LanguageNames.CSharp)]
public sealed class WL0002_ActionNamingAnalyzer : DiagnosticAnalyzer
```

検出ロジック:
1. `InvocationExpressionSyntax`で`Log`または`MeasureScope`メソッド呼び出しを検出
2. 対象クラスが`WorkflowLens`であることをシンボル解析で確認
3. action引数（第2引数）がリテラル文字列であれば正規表現`^[a-z][a-z0-9_]*$`で検証
4. 一致しなければWarning

### WL0003: Session直接使用

```csharp
[DiagnosticAnalyzer(LanguageNames.CSharp)]
public sealed class WL0003_SessionDirectUseAnalyzer : DiagnosticAnalyzer
```

検出ロジック:
1. `InvocationExpressionSyntax`で`Log`メソッド呼び出しを検出
2. 第1引数が`Category.Session`であることを確認
3. Infoレベルの診断を出す

### 診断一覧

| ID | タイトル | Severity | カテゴリ |
|----|---------|----------|---------|
| WL0001 | WorkflowLensインスタンスがusingで囲まれていません | Warning | Usage |
| WL0002 | action文字列がsnake_case命名規約に違反しています | Warning | Naming |
| WL0003 | AutoSessionが有効な場合、Category.Sessionの直接使用は不要です | Info | Usage |

## テスト方針

- [ ] WL0001: using文で囲んだ場合に診断が出ないこと
- [ ] WL0001: using varで宣言した場合に診断が出ないこと
- [ ] WL0001: フィールド代入の場合に診断が出ないこと
- [ ] WL0001: 通常のローカル変数代入で診断が出ること
- [ ] WL0002: `"compile"`（正しい）で診断が出ないこと
- [ ] WL0002: `"Compile"`（大文字始まり）で診断が出ること
- [ ] WL0002: `"brush apply"`（スペース含む）で診断が出ること
- [ ] WL0002: 変数を渡した場合に診断が出ないこと
- [ ] WL0003: `Log(Category.Session, "start")`で診断が出ること
- [ ] WL0003: `Log(Category.Build, "compile")`で診断が出ないこと
- [ ] テストは`DiagnosticVerifier`（Microsoft.CodeAnalysis.Testing）を使用する

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-11 | 初版作成 |
