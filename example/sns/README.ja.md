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

アプリは `example/admin/` の**単一 LiveView + 内部ルーティング**パターンに従います: 単一の `SNSView` 構造体が `v.page` 状態と `renderContent()` の `switch` により全ページを管理します。ナビゲーションイベントがページ状態を更新し、diff エンジンが変更された DOM のみをパッチします。

認証は `session/` パッケージを使用した静的ログイン/登録ページ（`example/auth/` と同じパターン）で、メインアプリは `session.RequireKey` で保護された LiveView です。

リアルタイム通知は `example/chat/` の Hub pub/sub パターンを使用: 共有 `Hub` がユーザー ID を `*live.Session` 参照にマッピングし、`SendInfo()` が型付き通知メッセージを正しいユーザーの `HandleInfo()` メソッドにプッシュします。

## ファイル構成

| ファイル | 役割 |
|----------|------|
| `main.go` | エントリポイント、ルーティング、MySQL 初期化、セッションストア |
| `db.go` | DB 接続、自動マイグレーション、全 SQL クエリ |
| `models.go` | データ構造体（User, Post, Like, Follow 等） |
| `hub.go` | リアルタイム通知 Hub |
| `auth.go` | ログイン/登録ページ + POST ハンドラ + パスワードハッシュ |
| `view.go` | SNSView 構造体、Mount, HandleEvent, HandleUpload, HandleInfo, Unmount |
| `pages.go` | Render メソッド + ページレンダラー（タイムライン、プロフィール、投稿、メッセージ、設定、検索） |
| `components.go` | 共有 UI コンポーネント（投稿カード、ナビリンク、投稿ボックス、SVG アイコン） |
| `css.go` | 全 CSS（モバイルファースト、レスポンシブブレークポイント） |
| `schema.sql` | リファレンス SQL スキーマ |
| `docker-compose.yml` | MySQL 8.0 コンテナ |
