# ADR-0001: ログスキーマ v2 — category + action 2層構造への移行

## ステータス

Accepted (2026-04-11)

## コンテキスト

WorkflowLensはゲーム開発ツールの操作ログを収集・分析する基盤である。ツール制作者がクライアントライブラリを組み込んでログを送信する。

v1スキーマでは以下のフィールドで「何が起きたか」を表現していた：

- `event_type`（enum: usage / error / session_start / session_end / cancellation）
- `message`（自由記述テキスト）
- `details`（自由形式JSON）

### 問題

1. **認知負荷が高い** — ツール制作者はログ送信のたびに「event_typeは何を選ぶか」「messageに何を書くか」「detailsに何を入れるか」「messageとdetailsの境界はどこか」を判断する必要があった
2. **分析に使いにくい** — `message` は自然言語で記述がバラバラになり集計・フィルタに使えない。`details` は自由形式のためキー名が統一されずクエリが複雑になる
3. **event_typeの粒度が粗い** — `usage` に「ブラシ適用」も「ビルド実行」も含まれ、意味のある分析には `details` の中身を掘る必要があった

## 決定

### ログスキーマ v2

`event_type` + `message` + `details` を削除し、以下に置き換える：

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `tool_name` | string | 必須 | ツール名（v1から変更なし） |
| `category` | enum | 必須 | 操作の大分類 |
| `action` | string | 必須 | カテゴリ内の具体的操作 |
| `timestamp` | ISO 8601 | 自動 | 発生時刻（v1から変更なし） |
| `session_id` | string | 自動 | セッション識別子（v1から変更なし） |
| `user_id` | string | 任意 | ユーザー識別（新規追加） |
| `tool_version` | string | 任意 | バージョン（v1から変更なし） |
| `duration_ms` | number | 任意 | 操作所要時間（新規追加） |
| `traceparent` | string | 任意 | W3Cトレースコンテキスト（v1から変更なし） |

### categoryのenum値

| 値 | 用途 |
|---|---|
| `asset` | アセット関連操作（インポート、エクスポート、変換） |
| `build` | ビルド・コンパイル関連操作 |
| `edit` | 編集操作（ブラシ、パラメータ変更、配置） |
| `error` | エラー発生 |
| `session` | セッション管理（start / end） |

### 2層構造を選んだ理由

3つの案を検討した：

1. **操作ドメイン寄りのenum一本** — asset, build, edit, error, session_start, session_end をenumで定義。集計はしやすいがenum値だけでは粒度が足りない
2. **ユーザー自由定義の文字列一本** — 制約が少ないが統一性が下がり、v1の `message` と同じ問題が再発する
3. **2層構造（enum + string）** — enumで大分類を強制し集計の軸を確保、stringで詳細を自由に記述

2層構造は「集計の確実性」と「表現の自由度」を両立する。ツール制作者はenumを1つ選び、actionに短い文字列を書くだけでよい。

### message / details を削除した理由

- `message` — 人間向けの説明だが、ログ閲覧時は `category` + `action` で十分に操作内容がわかる。自然言語の揺れが集計の障害になっていた
- `details` — 自由形式JSONはキー統一が困難。よく使われるデータ（所要時間、ユーザーID）はトップレベルフィールドに昇格させた。ツール固有の任意データを送る手段は意図的に廃止し、スキーマの一貫性を優先した

### user_id を追加した理由

「誰がどのツールをどう使っているか」はワークフロー分析の基本軸。未指定時はOSユーザー名を自動取得し、ツール制作者の追加負担はゼロ。

### duration_ms を追加した理由

操作の所要時間はパフォーマンス分析の基礎データ。v1では `details` 内に非構造で埋もれていたものをトップレベルに昇格。スコープAPI（C#: `using`, Python: `with`）による自動計測も提供する。

### セッション管理の統一

v1の `session_start` / `session_end` というevent_typeと `StartSession()` / `EndSession()` メソッドを廃止。`Log(Category.Session, "start")` / `Log(Category.Session, "end")` に統一し、APIの表面積を削減する。

## 結果

### メリット

- ツール制作者が毎回考えるのは **categoryとactionの2つだけ**
- `category` によるenum集計が常に可能（ダッシュボード構築が容易）
- `user_id` / `duration_ms` がトップレベルにあるためSQL集計が直接的

### トレードオフ

- **破壊的変更**: v1との互換性なし。既存の `logs` テーブルは drop & recreate が必要（v0.xのため許容）
- **ツール固有データの送信手段がない**: `details` を廃止したため、スキーマ外のデータは送れない。必要になった場合はv3で再検討する
- **category追加にはコード変更が必要**: enumのためサーバー・ミドルウェア・クライアント全てで対応が必要。ただし追加頻度は低いと想定

### category追加ポリシー

新しいcategoryの追加は以下の手順で行う：

1. 仕様書（`log-message.md`）にcategoryを追加
2. サーバー・ミドルウェアの `validCategories` に追加
3. C# / Python クライアントのenum に追加
4. 全テスト通過を確認

### DB移行方針

v0.x（プレリリース）のため、既存テーブルの互換性は保証しない。`logs` テーブルを drop して再起動すれば新スキーマで自動作成される。v1.0以降でスキーマ変更が発生する場合はマイグレーション機構を検討する。
