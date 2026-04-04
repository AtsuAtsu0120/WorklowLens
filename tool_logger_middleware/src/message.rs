use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// ログイベントの種類
///
/// C#の `enum` に近いが、Rustの `enum` は各バリアントにデータを持てる（今回は持たない単純な形）。
/// `#[serde(rename_all = "snake_case")]` により、JSONでは "usage", "error" というスネークケースで表現される。
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum EventType {
    /// ツールの通常利用
    Usage,
    /// エラー発生
    Error,
}

/// クライアントから送られるログメッセージ
///
/// C#でいう `record` や `class` に相当する構造体。
/// `#[derive(Deserialize)]` を付けることで、`serde_json` がJSONから自動的にこの構造体を組み立てる。
/// C#では `JsonSerializer.Deserialize<LogMessage>(json)` に相当。
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LogMessage {
    /// ツール名（例: "UnityTerrainEditor", "MayaRigTool"）
    pub tool_name: String,

    /// イベント種別
    pub event_type: EventType,

    /// ISO 8601形式のタイムスタンプ（例: "2026-04-04T10:30:00Z"）
    pub timestamp: DateTime<Utc>,

    /// 人間が読めるメッセージ
    pub message: String,

    /// 任意の追加情報（スタックトレース、パラメータなど）
    ///
    /// C#の `JsonElement?` や `JToken?` に近い。
    /// `Option<T>` はC#の `T?`（Nullable）に相当し、値がないときは `None`（C#の `null`）になる。
    /// `serde_json::Value` は任意のJSON値を表す型で、ツールごとに自由な構造を送れる。
    /// `#[serde(default)]` により、JSONでこのフィールドが省略された場合は自動的に `None` になる。
    #[serde(default)]
    pub details: Option<serde_json::Value>,
}

impl LogMessage {
    /// JSON文字列からLogMessageをパースする
    ///
    /// C#での `JsonSerializer.Deserialize<LogMessage>(json)` に相当。
    /// 成功時は `Ok(LogMessage)`、失敗時は `Err(serde_json::Error)` を返す。
    /// C#の `try/catch` の代わりに、Rustでは `Result` 型でエラーを表現する。
    pub fn parse(json_str: &str) -> Result<Self, serde_json::Error> {
        serde_json::from_str(json_str)
    }
}
