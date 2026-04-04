# tool_logger_middleware — 使い方ガイド

tool_loggerの組み込み方法とクライアント実装例です。
プロジェクト概要やログ仕様については [ルートのREADME](../README.md) を参照してください。

## ツールへの同伴方法

tool_loggerのバイナリをツールのプロジェクトに同伴させ、ツール起動時にプロセスとして起動します。

### ディレクトリ配置例

```
MyUnityProject/
├── Assets/
├── Plugins/
│   └── ToolLogger/
│       ├── tool_logger.exe    # Windows
│       └── tool_logger        # macOS/Linux
└── ...

MyMayaScripts/
├── scripts/
│   └── my_tool.py
└── bin/
    ├── tool_logger.exe        # Windows
    └── tool_logger            # macOS/Linux
```

### Unity (C#) からの起動

```csharp
using System.Diagnostics;

public static class ToolLoggerProcess
{
    private static Process _loggerProcess;

    /// <summary>
    /// tool_loggerプロセスを起動する。ツールの初期化時に呼ぶ。
    /// </summary>
    public static void Start(int port = 59100)
    {
        var exePath = Path.Combine(Application.dataPath, "Plugins", "ToolLogger",
            Application.platform == RuntimePlatform.WindowsEditor
                ? "tool_logger.exe"
                : "tool_logger");

        _loggerProcess = new Process
        {
            StartInfo = new ProcessStartInfo
            {
                FileName = exePath,
                Arguments = port.ToString(),
                UseShellExecute = false,
                CreateNoWindow = true,
                RedirectStandardOutput = true,
            }
        };
        _loggerProcess.Start();
    }

    /// <summary>
    /// tool_loggerプロセスを停止する。ツールの終了時に呼ぶ。
    /// </summary>
    public static void Stop()
    {
        if (_loggerProcess is { HasExited: false })
        {
            _loggerProcess.Kill();
            _loggerProcess.Dispose();
            _loggerProcess = null;
        }
    }
}
```

### Maya (Python) からの起動

```python
import subprocess
import os

_logger_process = None

def start_logger(port=59100):
    """tool_loggerプロセスを起動する。ツールの初期化時に呼ぶ。"""
    global _logger_process

    # スクリプトの隣にある bin/ ディレクトリからバイナリを探す
    script_dir = os.path.dirname(os.path.abspath(__file__))
    bin_dir = os.path.join(script_dir, "..", "bin")

    if os.name == "nt":
        exe_path = os.path.join(bin_dir, "tool_logger.exe")
    else:
        exe_path = os.path.join(bin_dir, "tool_logger")

    _logger_process = subprocess.Popen(
        [exe_path, str(port)],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

def stop_logger():
    """tool_loggerプロセスを停止する。ツールの終了時に呼ぶ。"""
    global _logger_process
    if _logger_process is not None:
        _logger_process.terminate()
        _logger_process.wait()
        _logger_process = None
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

## クライアント実装例

### C# (Unity)

```csharp
using System;
using System.Net.Sockets;
using System.Text;

public class ToolLoggerClient : IDisposable
{
    private TcpClient _client;
    private NetworkStream _stream;
    private readonly string _sessionId;
    private readonly string _toolVersion;

    public ToolLoggerClient(string toolVersion = null, string host = "127.0.0.1", int port = 59100)
    {
        _client = new TcpClient(host, port);
        _stream = _client.GetStream();
        _sessionId = Guid.NewGuid().ToString("N")[..8];
        _toolVersion = toolVersion;
    }

    public void Send(string toolName, string eventType, string message, string details = null)
    {
        var timestamp = DateTime.UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ");
        var detailsPart = details != null ? $", \"details\": {details}" : "";
        var sessionPart = $", \"session_id\": \"{_sessionId}\"";
        var versionPart = _toolVersion != null ? $", \"tool_version\": \"{_toolVersion}\"" : "";
        var json = $"{{\"tool_name\":\"{toolName}\",\"event_type\":\"{eventType}\","
                 + $"\"timestamp\":\"{timestamp}\",\"message\":\"{message}\""
                 + $"{sessionPart}{versionPart}{detailsPart}}}\n";
        var bytes = Encoding.UTF8.GetBytes(json);
        _stream.Write(bytes, 0, bytes.Length);
    }

    public void Dispose()
    {
        _stream?.Dispose();
        _client?.Dispose();
    }
}

// 使用例
// using var logger = new ToolLoggerClient(toolVersion: "1.2.0");
// logger.Send("MyTool", "session_start", "ツールを起動しました");
// logger.Send("MyTool", "usage", "ブラシを適用しました", "{\"size\": 5}");
// logger.Send("MyTool", "cancellation", "エクスポートをキャンセルしました");
// logger.Send("MyTool", "session_end", "ツールを終了しました");
```

### Python (Maya)

```python
import socket
import json
import uuid
from datetime import datetime, timezone

class ToolLoggerClient:
    def __init__(self, tool_version=None, host="127.0.0.1", port=59100):
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._sock.connect((host, port))
        self._session_id = uuid.uuid4().hex[:8]
        self._tool_version = tool_version

    def send(self, tool_name, event_type, message, details=None):
        log = {
            "tool_name": tool_name,
            "event_type": event_type,
            "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "message": message,
            "session_id": self._session_id,
        }
        if self._tool_version is not None:
            log["tool_version"] = self._tool_version
        if details is not None:
            log["details"] = details
        line = json.dumps(log) + "\n"
        self._sock.sendall(line.encode("utf-8"))

    def close(self):
        self._sock.close()

# 使用例
# client = ToolLoggerClient(tool_version="1.0.3")
# client.send("MayaRigTool", "session_start", "ツールを起動しました")
# client.send("MayaRigTool", "usage", "リグを適用しました", {"joint_count": 120})
# client.send("MayaRigTool", "cancellation", "エクスポートをキャンセルしました")
# client.send("MayaRigTool", "session_end", "ツールを終了しました")
# client.close()
```
