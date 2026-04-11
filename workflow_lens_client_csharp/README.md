# WorkflowLensClient (C#)

workflow_lens_middlewareへUDPでログを送信するC#クライアントライブラリ。

## インストール

NuGetパッケージとして追加:

```bash
dotnet add package WorkflowLensClient
```

またはソースコードを直接プロジェクトに含める。

## 使い方

### 基本

```csharp
using WorkflowLensClient;

// クライアント作成
using var logger = new WorkflowLens("my_tool", toolVersion: "1.0.0");

// セッション開始
logger.StartSession();

// 使用ログ
logger.LogUsage("ボタンAを押下");

// details付き（生JSON文字列）
logger.LogUsage("エクスポート完了", "{\"format\":\"fbx\",\"count\":42}");

// エラーログ
logger.LogError("ファイルが見つかりません", "{\"path\":\"/tmp/missing.txt\"}");

// キャンセルログ
logger.LogCancellation("エクスポートをキャンセル");

// セッション終了
logger.EndSession();
```

### 汎用送信

```csharp
logger.Send(EventType.Usage, "カスタムイベント", "{\"key\":\"value\"}");
```

### イベント種別

| 定数 | 値 | 用途 |
|------|---|------|
| `EventType.Usage` | `"usage"` | ツール機能の実行 |
| `EventType.Error` | `"error"` | エラー発生 |
| `EventType.SessionStart` | `"session_start"` | ツール起動 |
| `EventType.SessionEnd` | `"session_end"` | ツール終了 |
| `EventType.Cancellation` | `"cancellation"` | 操作キャンセル |

## 設計

- **通信**: UDP fire-and-forget（middleware未起動でも例外を投げない）
- **スレッドセーフ**: `UdpClient.Send()` がスレッドセーフ
- **details**: 生JSON文字列を渡す（シリアライザ非依存）
- **session_id**: `StartSession()` で自動生成（GUID先頭8文字）
- **依存**: なし（.NET Standard 2.1標準ライブラリのみ）
