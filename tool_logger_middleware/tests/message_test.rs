// LogMessage と EventType の単体テスト
//
// `tests/` ディレクトリに置かれたファイルはRustの「結合テスト」として扱われる。
// C#でいうと、別のテストプロジェクト（例: `ToolLogger.Tests`）に相当する。
// `use tool_logger::message::...` のように、公開APIのみアクセスできる。

use tool_logger::message::{EventType, LogMessage};

// ===== 正常系テスト =====

#[test]
fn 正常なjsonをパースできる() {
    let json = r#"{
        "tool_name": "UnityTerrainEditor",
        "event_type": "usage",
        "timestamp": "2026-04-04T10:30:00Z",
        "message": "Terrain brush applied",
        "details": {"brush_size": 5}
    }"#;

    let msg = LogMessage::parse(json).unwrap();
    assert_eq!(msg.tool_name, "UnityTerrainEditor");
    assert_eq!(msg.event_type, EventType::Usage);
    assert_eq!(msg.message, "Terrain brush applied");
    assert!(msg.details.is_some());
}

#[test]
fn errorイベントをパースできる() {
    let json = r#"{
        "tool_name": "MayaRigTool",
        "event_type": "error",
        "timestamp": "2026-04-04T11:00:00Z",
        "message": "Rig export failed"
    }"#;

    let msg = LogMessage::parse(json).unwrap();
    assert_eq!(msg.event_type, EventType::Error);
    assert_eq!(msg.tool_name, "MayaRigTool");
}

#[test]
fn details省略時にnoneになる() {
    let json = r#"{
        "tool_name": "MayaRigTool",
        "event_type": "error",
        "timestamp": "2026-04-04T11:00:00Z",
        "message": "Rig export failed"
    }"#;

    let msg = LogMessage::parse(json).unwrap();
    assert!(msg.details.is_none());
}

#[test]
fn detailsがnullの場合もnoneになる() {
    // JSONで明示的に `"details": null` を渡した場合
    let json = r#"{
        "tool_name": "TestTool",
        "event_type": "usage",
        "timestamp": "2026-04-04T10:00:00Z",
        "message": "test",
        "details": null
    }"#;

    let msg = LogMessage::parse(json).unwrap();
    // `Option<serde_json::Value>` は JSON の `null` を `None` として扱う。
    // C# の `Nullable<T>` が `null` のとき `HasValue == false` になるのと同じ。
    assert!(msg.details.is_none());
}

#[test]
fn detailsの中身を検証できる() {
    let json = r#"{
        "tool_name": "UnityTerrainEditor",
        "event_type": "usage",
        "timestamp": "2026-04-04T10:30:00Z",
        "message": "Terrain brush applied",
        "details": {"brush_size": 5, "scene": "Level01"}
    }"#;

    let msg = LogMessage::parse(json).unwrap();
    let details = msg.details.unwrap();

    // `details["brush_size"]` でインデックスアクセスできる。
    // C# の `jsonElement.GetProperty("brush_size")` に近い。
    assert_eq!(details["brush_size"], 5);
    assert_eq!(details["scene"], "Level01");
}

#[test]
fn detailsにネストしたオブジェクトを含められる() {
    let json = r#"{
        "tool_name": "UnityShaderTool",
        "event_type": "error",
        "timestamp": "2026-04-04T12:00:00Z",
        "message": "Shader compile error",
        "details": {
            "shader_name": "PBR_Standard",
            "errors": [
                {"line": 42, "message": "unexpected token"},
                {"line": 58, "message": "undefined variable"}
            ]
        }
    }"#;

    let msg = LogMessage::parse(json).unwrap();
    let details = msg.details.unwrap();
    let errors = details["errors"].as_array().unwrap();
    assert_eq!(errors.len(), 2);
    assert_eq!(errors[0]["line"], 42);
}

#[test]
fn シリアライズとデシリアライズが往復できる() {
    // 構造体 → JSON文字列 → 構造体 の往復（ラウンドトリップ）テスト。
    // C#で `Serialize` → `Deserialize` して元と一致するか確認するのと同じ。
    let original = LogMessage {
        tool_name: "MayaRigTool".to_string(),
        event_type: EventType::Error,
        timestamp: chrono::Utc::now(),
        message: "Rig export failed".to_string(),
        details: Some(serde_json::json!({"node_count": 42})),
    };

    // Serialize: 構造体 → JSON文字列
    let json_str = serde_json::to_string(&original).unwrap();

    // Deserialize: JSON文字列 → 構造体
    let restored = LogMessage::parse(&json_str).unwrap();

    assert_eq!(restored.tool_name, original.tool_name);
    assert_eq!(restored.event_type, original.event_type);
    assert_eq!(restored.message, original.message);
    assert_eq!(restored.details, original.details);
}

// ===== 異常系テスト =====

#[test]
fn 不正なevent_typeを拒否する() {
    let json = r#"{
        "tool_name": "TestTool",
        "event_type": "unknown_type",
        "timestamp": "2026-04-04T10:00:00Z",
        "message": "test"
    }"#;

    let result = LogMessage::parse(json);
    assert!(result.is_err());
}

#[test]
fn 必須フィールド欠落を拒否する_tool_name() {
    let json = r#"{
        "event_type": "usage",
        "timestamp": "2026-04-04T10:00:00Z",
        "message": "test"
    }"#;

    let result = LogMessage::parse(json);
    assert!(result.is_err());
}

#[test]
fn 必須フィールド欠落を拒否する_event_type() {
    let json = r#"{
        "tool_name": "TestTool",
        "timestamp": "2026-04-04T10:00:00Z",
        "message": "test"
    }"#;

    let result = LogMessage::parse(json);
    assert!(result.is_err());
}

#[test]
fn 必須フィールド欠落を拒否する_timestamp() {
    let json = r#"{
        "tool_name": "TestTool",
        "event_type": "usage",
        "message": "test"
    }"#;

    let result = LogMessage::parse(json);
    assert!(result.is_err());
}

#[test]
fn 必須フィールド欠落を拒否する_message() {
    let json = r#"{
        "tool_name": "TestTool",
        "event_type": "usage",
        "timestamp": "2026-04-04T10:00:00Z"
    }"#;

    let result = LogMessage::parse(json);
    assert!(result.is_err());
}

#[test]
fn タイムスタンプが不正な形式だと拒否する() {
    let json = r#"{
        "tool_name": "TestTool",
        "event_type": "usage",
        "timestamp": "not-a-date",
        "message": "test"
    }"#;

    let result = LogMessage::parse(json);
    assert!(result.is_err());
}

#[test]
fn 不正なjsonを拒否する() {
    let result = LogMessage::parse("not json at all");
    assert!(result.is_err());
}

#[test]
fn 空のjsonオブジェクトを拒否する() {
    let result = LogMessage::parse("{}");
    assert!(result.is_err());
}

#[test]
fn 空文字列を拒否する() {
    let result = LogMessage::parse("");
    assert!(result.is_err());
}
