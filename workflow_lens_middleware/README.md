# workflow_lens_middleware — 使い方ガイド

workflow_lensの組み込み方法とクライアント実装例です。
プロジェクト概要やログ仕様については [ルートのREADME](../README.md) を参照してください。

## セットアップ

### 方法1: PATHに配置（推奨）

バイナリを `workflow_lens_middleware` という名前でPATHの通ったディレクトリに配置します。クライアントライブラリが自動的に検出します。

```bash
# ビルドしてPATHの通ったディレクトリにコピー
go build -o workflow_lens_middleware ./cmd/middleware
cp workflow_lens_middleware /usr/local/bin/
```

### 方法2: 環境変数で指定

環境変数 `WORKFLOW_LENS_MIDDLEWARE_PATH` にバイナリのフルパスを設定します。

```bash
export WORKFLOW_LENS_MIDDLEWARE_PATH=/path/to/workflow_lens_middleware
```

### 方法3: プロジェクトに同伴

バイナリをツールのプロジェクトに同伴させ、クライアントライブラリの`middlewarePath`パラメータで明示指定します。

```
MyUnityProject/
├── Assets/
├── Plugins/
│   └── WorkflowLens/
│       ├── workflow_lens_middleware.exe    # Windows
│       └── workflow_lens_middleware        # macOS/Linux
└── ...

MyMayaScripts/
├── scripts/
│   └── my_tool.py
└── bin/
    ├── workflow_lens_middleware.exe        # Windows
    └── workflow_lens_middleware            # macOS/Linux
```

## イベントタイプと使い所

### `session_start` — ツール起動時に送信

ツールのウィンドウを開いた、プラグインがロードされた等のタイミングで1回送信する。

```json
{"tool_name":"UnityTerrainEditor","event_type":"session_start","timestamp":"2026-04-04T10:00:00Z","message":"Tool window opened","session_id":"ute-a1b2c3d4","tool_version":"2.1.0"}
```

- `session_id` はこのタイミングで生成し、以降のすべてのイベントに付与する
- **得られる指標**: セッション数/週、DAU（日次アクティブユーザー）、利用頻度の推移
- **判断例**: 直近30日間でセッションがゼロのツールは削除候補

### `usage` — ツールの機能を実行したときに送信

ボタン押下、ブラシ適用、エクスポート実行など、ユーザーが機能を使うたびに送信する。

```json
{"tool_name":"UnityTerrainEditor","event_type":"usage","timestamp":"2026-04-04T10:05:00Z","message":"Terrain brush applied","session_id":"ute-a1b2c3d4","tool_version":"2.1.0","details":{"feature":"paint_height","brush_size":5}}
```

- `details.feature` にどの機能を使ったかを入れると、機能別の利用分布が取れる
- **得られる指標**: 機能別利用回数、セッション内の操作数
- **判断例**: 10機能あるツールで1機能しか使われていなければ、そのツールは簡素化できる
- **判断例**: ツールAとツールBで同じ `feature` 名のusageが多ければ統合候補

### `error` — エラー発生時に送信

例外の発生、処理の失敗など、エラーが起きたタイミングで送信する。

```json
{"tool_name":"MayaRigTool","event_type":"error","timestamp":"2026-04-04T11:00:00Z","message":"Rig export failed: unsupported joint type","session_id":"mrt-x9y8z7w6","tool_version":"1.0.3","details":{"error_code":"EXPORT_001","feature":"export_rig","severity":"critical"}}
```

- `tool_version` を付けておくと、バージョンアップ後にエラーが減ったか確認できる
- `details.severity` で重要度を分けると、致命的なエラーだけ集計できる
- **得られる指標**: エラー率（エラー数 / セッション数）、バージョン別エラー推移
- **判断例**: エラー率30%超かつバージョンアップしても改善しないツールは削除候補

### `cancellation` — 操作をキャンセルしたときに送信

ダイアログを開いたがキャンセルした、処理を途中で中断した等のタイミングで送信する。

```json
{"tool_name":"UnityTerrainEditor","event_type":"cancellation","timestamp":"2026-04-04T10:12:00Z","message":"Export dialog cancelled","session_id":"ute-a1b2c3d4","tool_version":"2.1.0","details":{"feature":"export_heightmap","step":"format_selection"}}
```

- `details.step` で「どの段階でキャンセルしたか」を記録するとUXの問題箇所が特定できる
- **得られる指標**: キャンセル率（キャンセル数 / (usage数 + キャンセル数)）
- **判断例**: キャンセル率50%超の機能はUIが分かりにくい可能性がある
- **判断例**: ツールAでキャンセル後にツールBで同じ操作をしていれば、統合候補

### `session_end` — ツール終了時に送信

ツールのウィンドウを閉じた、プラグインがアンロードされた等のタイミングで送信する。

```json
{"tool_name":"UnityTerrainEditor","event_type":"session_end","timestamp":"2026-04-04T10:47:00Z","message":"Tool window closed","session_id":"ute-a1b2c3d4","tool_version":"2.1.0","details":{"usage_count":23,"errors_during_session":0}}
```

- `session_start` と組み合わせてセッション時間を算出する（上の例では47分間）
- `session_start` があるのに `session_end` がない場合、ツールがクラッシュした可能性がある
- **得られる指標**: セッション時間、クラッシュ率（session_endなし / session_start数）
- **判断例**: 平均セッション時間5秒のツールは「開いたが使わなかった」=不要の可能性
- **判断例**: 2つのツールのセッションが常に同時に開かれていれば統合候補

### フィールド補足

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `tool_name` | はい | ツール名。集計のキーになる |
| `event_type` | はい | 上記5種類のいずれか |
| `timestamp` | はい | ISO 8601形式（例: `2026-04-04T10:30:00Z`） |
| `message` | はい | 人間が読めるメッセージ |
| `session_id` | いいえ | ツール起動ごとにユニークなID。同一セッション内のイベントを紐付ける |
| `tool_version` | いいえ | ツールのバージョン。バージョン別のエラー率比較に使う |
| `details` | いいえ | 任意のJSON。`feature`, `error_code`, `severity` など自由に使える |

`session_id` と `tool_version` は省略可能なので、既存のクライアントはそのまま動作する。

## クライアントライブラリ

ログ送信には専用のクライアントライブラリを使用してください。通信はUDP fire-and-forgetで、middlewareが未起動でも例外を投げません。

- **C#**: [workflow_lens_client_csharp](../workflow_lens_client_csharp/) — `dotnet add package WorkflowLensClient` またはソース直接参照
- **Python**: [workflow_lens_client_python](../workflow_lens_client_python/) — `pip install workflow-lens-client` またはソース直接参照

### C# (Unity)

```csharp
using WorkflowLensClient;

// middlewareを自動探索して起動（PATH or 環境変数から検出）
using var logger = new WorkflowLens("UnityTerrainEditor", toolVersion: "2.1.0",
                                  autoStartMiddleware: true);

// セッション開始
logger.StartSession();

// 機能の使用ログ
logger.LogUsage("ブラシを適用しました", "{\"feature\":\"paint_height\",\"brush_size\":5}");

// キャンセル
logger.LogCancellation("エクスポートをキャンセルしました");

// エラー
logger.LogError("テクスチャの読み込みに失敗", "{\"error_code\":\"TEX_001\",\"severity\":\"warning\"}");

// セッション終了
logger.EndSession();

// パスを明示指定することも可能
using var logger2 = new WorkflowLens("MyTool",
    middlewarePath: "/path/to/workflow_lens_middleware");
```

### Python (Maya)

```python
from workflow_lens_client import WorkflowLens

# middlewareを自動探索して起動（PATH or 環境変数から検出）
with WorkflowLens("MayaRigTool", "1.0.3", auto_start_middleware=True) as logger:
    logger.log_usage("リグを適用しました", {"joint_count": 120})
    logger.log_cancellation("エクスポートをキャンセルしました")
    logger.log_error("不正なジョイントタイプ", {"error_code": "EXPORT_001", "severity": "critical"})

# パスを明示指定することも可能
logger = WorkflowLens("MyTool", middleware_path="/path/to/workflow_lens_middleware")
```
