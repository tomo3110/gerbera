# SNS（Twitter クローン）

Gerbera LiveView、MySQL、リアルタイム通知を使用したフル機能の Twitter クローンです。

機能:
- **タイムライン** — 投稿、いいね、リツイートとリアルタイムのカウント更新
- **ユーザープロフィール** — フォロー/フォロー解除、投稿履歴、自己紹介
- **ダイレクトメッセージ** — Hub pub/sub によるリアルタイム 1:1 チャット
- **コメント** — 投稿へのスレッド形式の返信
- **検索** — デバウンス付きのユーザー・投稿検索
- **設定** — プロフィール編集、パスワード変更、アバターアップロード
- **リアルタイム通知** — いいね、リツイート、フォロー、メッセージが `InfoReceiver` 経由で即座にプッシュ
- **モバイルファースト** — モバイルではドロワーナビゲーション、デスクトップではサイドバーのレスポンシブレイアウト
- **セッション認証** — CSRF 保護付きのログイン/登録、SHA-256 パスワードハッシュ

## 使い方

```bash
# Docker Compose ですべて起動（MySQL + アプリ）
cd example/sns && docker compose up

# バックグラウンドで起動
cd example/sns && docker compose up -d

# コード変更後に再ビルド
cd example/sns && docker compose up --build
```

### ローカル開発（アプリは Docker 外で実行）

```bash
# MySQL のみ起動
cd example/sns && docker compose up -d mysql

# アプリをローカルで実行
cd example/sns && go run .

# デバッグパネル付き
cd example/sns && go run . -debug

# カスタムポートと DSN
cd example/sns && go run . -addr :9000 -dsn "user:pass@tcp(host:3306)/dbname?parseTime=true"
```

## フラグ

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `-addr` | `:8930` | リッスンアドレス |
| `-dsn` | `sns:snspass@tcp(127.0.0.1:3306)/sns?parseTime=true` | MySQL DSN |
| `-debug` | `false` | Gerbera デバッグパネルを有効化 |

## 依存パッケージ

- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) — `database/sql` 用 MySQL ドライバ
- MySQL 8.0（`docker-compose.yml` 経由）

## アーキテクチャ

このサンプルは `replace` ディレクティブ付きの独立した `go.mod` を使用します（`example/mdviewer` と同じパターン）。

アプリは**マルチビュー SSR シェル**アーキテクチャに従います: SSR レイアウト（`layout.go`）がサイドバーナビゲーションを描画し、個別の LiveView エンドポイントを `LiveMount` で埋め込みます。各ページは独自の View 構造体（`TimelineView`、`ProfileView`、`PostDetailView`、`MessagesView`、`SearchView`、`SettingsView`）を持ち、分離された状態を管理します。すべてのビューは埋め込まれた `baseView` を通じて共通ロジックを共有します。

SSR シェルページがルーティングとバッジカウントを処理し、LiveView エンドポイントがリアルタイムのインタラクティビティを担当します:

```
GET /           → SSR シェル → LiveMount("/live/timeline")  → TimelineView
GET /profile    → SSR シェル → LiveMount("/live/profile")   → ProfileView
GET /post/{id}  → SSR シェル → LiveMount("/live/post?id=…") → PostDetailView
GET /messages   → SSR シェル → LiveMount("/live/messages")  → MessagesView
GET /search     → SSR シェル → LiveMount("/live/search")    → SearchView
GET /settings   → SSR シェル → LiveMount("/live/settings")  → SettingsView
```

認証は `session/` パッケージを使用した静的ログイン/登録ページ（`example/auth/` と同じパターン）で、すべての LiveView エンドポイントは `session.RequireKey` で保護されています。

リアルタイム通知は `example/chat/` の Hub pub/sub パターンを使用: 共有 `Hub` がユーザー ID を複数の `*live.Session` 参照（View/WebSocket ごとに 1 つ）にマッピングし、`SendInfo()` が型付き通知メッセージを正しいユーザーの `HandleInfo()` メソッドにすべてのアクティブ View を通じてプッシュします。

## ファイル構成

| ファイル | 役割 |
|----------|------|
| `main.go` | エントリポイント、ルーティング、MySQL 初期化、セッションストア |
| `db.go` | DB 接続、自動マイグレーション、全 SQL クエリ |
| `models.go` | データ構造体（User, Post, Like, Follow 等） |
| `hub.go` | マルチセッションユーザーマッピング付きリアルタイム通知 Hub |
| `auth.go` | ログイン/登録ページ + POST ハンドラ + パスワードハッシュ |
| `layout.go` | SSR シェルレイアウト（サイドバー、ドロワー、バッジカウント、LiveMount） |
| `view_base.go` | 共有ビューロジック: mountBase、unmountBase、投稿アクション、通知 |
| `view_timeline.go` | TimelineView: ホームフィード、投稿作成、投稿/削除 |
| `view_profile.go` | ProfileView: ユーザープロフィール、フォロー/フォロー解除 |
| `view_post_detail.go` | PostDetailView: 単一投稿、コメント、削除 |
| `view_messages.go` | MessagesView: 会話一覧、1:1 チャット |
| `view_search.go` | SearchView: デバウンス付きユーザー/投稿検索 |
| `view_settings.go` | SettingsView: プロフィール編集、パスワード変更、アバターアップロード |
| `components.go` | 共有 UI コンポーネント（投稿カード、投稿ボックス、アバター、SVG アイコン） |
| `css.go` | 全 CSS（モバイルファースト、レスポンシブブレークポイント） |
| `schema.sql` | リファレンス SQL スキーマ |
| `docker-compose.yml` | MySQL 8.0 コンテナ |
