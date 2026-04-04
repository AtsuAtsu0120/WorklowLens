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

    public ToolLoggerClient(string host = "127.0.0.1", int port = 59100)
    {
        _client = new TcpClient(host, port);
        _stream = _client.GetStream();
    }

    public void Send(string toolName, string eventType, string message, string details = null)
    {
        var timestamp = DateTime.UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ");
        var detailsPart = details != null ? $", \"details\": {details}" : "";
        var json = $"{{\"tool_name\":\"{toolName}\",\"event_type\":\"{eventType}\","
                 + $"\"timestamp\":\"{timestamp}\",\"message\":\"{message}\"{detailsPart}}}\n";
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
// using var logger = new ToolLoggerClient();
// logger.Send("MyTool", "usage", "ブラシを適用しました", "{\"size\": 5}");
// logger.Send("MyTool", "error", "テクスチャの読み込みに失敗");
```

### Python (Maya)

```python
import socket
import json
from datetime import datetime, timezone

class ToolLoggerClient:
    def __init__(self, host="127.0.0.1", port=59100):
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._sock.connect((host, port))

    def send(self, tool_name, event_type, message, details=None):
        log = {
            "tool_name": tool_name,
            "event_type": event_type,
            "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "message": message,
        }
        if details is not None:
            log["details"] = details
        line = json.dumps(log) + "\n"
        self._sock.sendall(line.encode("utf-8"))

    def close(self):
        self._sock.close()

# 使用例
# client = ToolLoggerClient()
# client.send("MayaRigTool", "usage", "リグを適用しました", {"joint_count": 120})
# client.send("MayaRigTool", "error", "ウェイト計算に失敗")
# client.close()
```
