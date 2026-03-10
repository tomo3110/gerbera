# SNS（Twitter クローン）チュートリアル

このサンプルは MySQL 永続化、セッション認証、マルチビュー LiveView アーキテクチャ、リアルタイム Hub 通知、ファイルアップロード、モバイルファーストレスポンシブデザインを組み合わせたフル機能の Twitter クローンを実演します。Gerbera の `ui/` コンポーネントライブラリを活用しています。

## 概要

SNS サンプルが実装する機能:

- **セッション認証** — CSRF 保護と SHA-256 パスワードハッシュ付きのログイン/登録（静的ページ）
- **マルチビュー SSR シェル** — 各ページが独自の View 構造体を持ち、`LiveMount` 経由で SSR シェルレイアウト内にマウント
- **MySQL 永続化** — ユーザー、投稿、いいね、リツイート、フォロー、コメント、メッセージ、通知
- **リアルタイム Hub** — `InfoReceiver` + `SendInfo()` 経由の型付き通知メッセージプッシュ
- **ファイルアップロード** — `UploadHandler` によるアバター・投稿画像アップロード
- **モバイルファースト UI** — モバイルではドロワーナビゲーション、デスクトップではサイドバー、レスポンシブな投稿カード

## 前提条件

- Go 1.22 以降がインストール済み
- Docker および Docker Compose がインストール済み（MySQL 用）
- このリポジトリがローカルにクローン済み

## 主要コンセプト

### 独立モジュール

`example/mdviewer` と同様に、このサンプルはローカルの gerbera モジュールを参照する `replace` ディレクティブ付きの独自 `go.mod` を持ちます。MySQL ドライバの依存（`go-sql-driver/mysql`）をルートモジュールから分離します。

```go
module github.com/tomo3110/gerbera/example/sns

require (
    github.com/go-sql-driver/mysql v1.8.1
    github.com/tomo3110/gerbera v0.0.0
)

replace github.com/tomo3110/gerbera => ../..
```

### SSR シェル + LiveMount アーキテクチャ

`example/admin/` の単一 SPA パターンとは異なり、このサンプルは SSR シェルページが個別の LiveView エンドポイントを `LiveMount` で埋め込む方式を採用しています。各ページは分離された状態を持つ独自の View 構造体を持ちます:

```
GET /           → SSR シェル (layout.go) → LiveMount("/live/timeline")  → TimelineView
GET /profile    → SSR シェル             → LiveMount("/live/profile")   → ProfileView
GET /post/{id}  → SSR シェル             → LiveMount("/live/post?id=…") → PostDetailView
GET /messages   → SSR シェル             → LiveMount("/live/messages")  → MessagesView
GET /search     → SSR シェル             → LiveMount("/live/search")    → SearchView
GET /settings   → SSR シェル             → LiveMount("/live/settings")  → SettingsView
```

SSR シェル（`layout.go`）がサイドバーナビゲーション、モバイルドロワー、バッジカウントを描画します。`LiveMount` コンポーネントが対応する LiveView エンドポイントに WebSocket で接続します。

### baseView — 共通ロジック

すべての View 構造体が `baseView` を埋め込み、共通フィールドとヘルパーを提供します:

```go
type baseView struct {
    gl.CommandQueue
    db   *sql.DB
    hub  *Hub
    sess *gl.Session

    userID int64
    user   *User

    toastMessage string
    toastVariant string
    toastVisible bool
}
```

- `mountBase(params)` — HTTP セッションからユーザー ID を読み取り、DB からユーザーを読み込み、Hub に参加
- `unmountBase()` — WebSocket が閉じた時に Hub から離脱
- `handlePostAction(event, payload)` — いいね/リツイート/共有/トースト非表示イベントの共通処理
- `handleBaseInfo(msg)` — 共通通知処理（いいね、リツイート、フォロー、コメントのトースト表示）

### パスワードハッシュ

SHA-256 とランダムソルトを使用（標準ライブラリのみ、bcrypt 依存なし）:

```go
func hashPassword(password string) string {
    salt := make([]byte, 16)
    rand.Read(salt)
    hash := sha256.Sum256(append(salt, []byte(password)...))
    return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash[:])
}

func verifyPassword(stored, password string) bool {
    parts := strings.SplitN(stored, ":", 2)
    if len(parts) != 2 {
        return false
    }
    salt, err := hex.DecodeString(parts[0])
    if err != nil {
        return false
    }
    hash := sha256.Sum256(append(salt, []byte(password)...))
    return parts[1] == hex.EncodeToString(hash[:])
}
```

### マルチセッション対応ユーザーマッピング Hub

SNS Hub はデータベースのユーザー ID を複数の LiveView セッション（View/WebSocket ごとに 1 つ）にマッピングし、特定ユーザーへの通知送信を可能にします:

```go
type Hub struct {
    mu       sync.RWMutex
    clients  map[string]*live.Session         // セッション ID → Session
    userSess map[int64]map[string]struct{} // ユーザー DB ID → セッション ID セット
}

func (h *Hub) Notify(userID int64, msg any) {
    if sids, ok := h.userSess[userID]; ok {
        for sid := range sids {
            if sess, ok := h.clients[sid]; ok {
                sess.SendInfo(msg)
            }
        }
    }
}

func (h *Hub) Broadcast(msg any, senderSessionID string) {
    for id, sess := range h.clients {
        if id == senderSessionID {
            continue
        }
        sess.SendInfo(msg)
    }
}
```

`HandleInfo` で区別できるよう 7 つの型付き通知メッセージを使用します:

```go
type NewLikeNotif struct{ PostID int64; ActorName string }
type NewRetweetNotif struct{ PostID int64; ActorName string }
type NewFollowNotif struct{ ActorName string }
type NewMessageNotif struct{ SenderID int64; Content string }
type NewCommentNotif struct{ PostID int64; ActorName string }
type NewPostNotif struct{ AuthorName string }
type NewCommentOnViewedPostNotif struct{ PostID int64; ActorName string }
```

### 自動マイグレーション

手動 SQL 実行の代わりに、`db.go` が起動時に `CREATE TABLE IF NOT EXISTS` 文を実行します:

```go
func autoMigrate(db *sql.DB) error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS users (...)`,
        `CREATE TABLE IF NOT EXISTS posts (...)`,
        // ...
    }
    for _, q := range queries {
        if _, err := db.Exec(q); err != nil {
            return fmt.Errorf("migration: %w", err)
        }
    }
    return nil
}
```

## ウォークスルー

### main.go — エントリポイント

1. フラグをパース（`-addr`、`-dsn`、`-debug`）、環境変数オーバーライド対応（`DATABASE_DSN`、`LISTEN_ADDR`）
2. MySQL 接続を開き、自動マイグレーションを実行
3. `session.MemoryStore` とセッションミドルウェアを作成
4. アップロードディレクトリを作成（`uploads/avatars/`）
5. ルートを登録:
   - `GET /avatars/` — アップロード済みアバターの静的ファイル配信
   - `GET /login` — ログインフォームを表示（認証済みの場合は `/` にリダイレクト）
   - `POST /login` — 資格情報を検証し、セッションに `user_id` を設定
   - `GET /register` — 登録フォームを表示
   - `POST /register` — ユーザーを作成し、セッションに `user_id` を設定
   - `/logout` — セッションを破棄し `/login` にリダイレクト
   - SSR シェルページ（`/`、`/profile`、`/profile/{id}`、`/post/{id}`、`/messages`、`/messages/{id}`、`/search`、`/settings`）— すべて `session.RequireKey("user_id", "/login")` で保護
   - LiveView エンドポイント（`/live/timeline`、`/live/profile`、`/live/post`、`/live/messages`、`/live/search`、`/live/settings`）— それぞれ独自の View 構造体を生成

```go
mux.Handle("/live/timeline", authGuard(gl.Handler(func(_ context.Context) gl.View {
    return NewTimelineView(db, hub)
}, liveOpts...)))
```

### auth.go — 静的認証ページ

ログインと登録ページは `g.Handler()` を使用したサーバーサイドレンダリング（LiveView ではない）です。`isRegister` フラグに基づいて HTML を生成する `renderAuthPage()` 関数を共有し、`loginPage()` と `registerPage()` が便利なラッパーとして機能します。POST エラーレスポンスではステータスコード制御のために `g.ExecuteTemplate()` を直接使用しています。

CSRF トークンは `session.GenerateCSRFToken(sess)` で生成し、フォーム送信時にバリデーションします。ユーザー名は正規表現 `^[a-zA-Z0-9_]{1,30}$` でバリデーションされます。

### layout.go — SSR シェルレイアウト

`snsPage()` 関数がサイドバーナビゲーション付きの完全な HTML シェルを描画します:

```go
func snsPage(title, activePage, liveEndpoint string, badges badgeCounts) g.Components {
    return g.Components{
        gd.Head(/* タイトル、ビューポート、テーマ、CSS */),
        gd.Body(
            // モバイルヘッダー（ドロワートグルボタン付き）
            // メインレイアウト: デスクトップサイドバー + LiveMount コンテンツエリア
            // モバイルドロワーオーバーレイ（ナビゲーションリンク付き）
        ),
    }
}
```

`sidebarLinks()` は 6 つのナビゲーションアイテム（Home、Search、Notifications、Messages、Profile、Settings）とログアウトボタンを描画します。各リンクはオプションのバッジカウントをサポートする `sidebarLink()` を使用します。

`fetchBadgeCounts()` は現在のユーザーセッションの通知数とメッセージ数をデータベースからクエリします。

### view_base.go — 共有ビューロジック

`baseView` 構造体が共通フィールドを保持し、すべてのビューで使用されるヘルパーメソッドを提供します:

- `mountBase(params)` — HTTP セッションからユーザー ID を読み取り、ユーザーを読み込み、Hub に参加
- `unmountBase()` — WebSocket 切断時に Hub から離脱
- `showToast(msg, variant)` — 表示用のトースト状態を設定
- `handlePostAction(event, payload)` — ビュー間で共有される `toggleLike`、`toggleRetweet`、`sharePost`、`dismissToast` イベントの処理
- `handleBaseInfo(msg)` — 共通通知処理（いいね、リツイート、フォロー、コメント）のトースト表示

### view_timeline.go — TimelineView

投稿ボックスと投稿カード付きのホームタイムラインを表示します。

**状態**: `posts`、`composeDraft`、`composeChars`、`confirmOpen`、`confirmAction`

**イベント**: `composeInput`、`submitPost`、`confirmDeletePost`、`doDelete`、`cancelDelete`、および共有投稿アクション

**HandleUpload**: `postPhoto` アップロードを受信（現在はログ記録のみ）

**HandleInfo**: `NewPostNotif`、`NewLikeNotif` 等でタイムラインを更新

### view_profile.go — ProfileView

アバター、統計、フォローボタン、投稿リスト付きのユーザープロフィールを表示します。

**Mount**: `id` クエリパラメータを読み取り、省略時は自分のプロフィールを表示

**イベント**: `toggleFollow`、および共有投稿アクション

### view_post_detail.go — PostDetailView

コメントとコメントフォーム付きの単一投稿を表示します。投稿者は削除可能です。

**イベント**: `commentInput`、`submitComment`、`confirmDeletePost`、`doDelete`、`cancelDelete`、および共有投稿アクション

**HandleInfo**: `NewCommentOnViewedPostNotif` でコメントを更新

### view_messages.go — MessagesView

会話リストまたはアクティブチャットビューを表示します。

**HandleParams**: URL パスに基づいて会話リストとチャットを切り替え

**イベント**: `openChat`、`backToConversations`、`chatInput`、`chatSend`、`chatKeydown`

**HandleInfo**: 該当の会話を表示中なら `NewMessageNotif` でリアルタイムにチャットを更新

### view_search.go — SearchView

デバウンス付きのユーザー・投稿検索を提供します。

**イベント**: `gl.Debounce(300)` 付き `searchInput`、および共有投稿アクション

**HandleParams**: `keyword` クエリパラメータから結果を読み込み

### view_settings.go — SettingsView

プロフィール編集、パスワード変更、アバターアップロードを提供します。

**イベント**: `settingsDisplayNameInput`、`settingsEmailInput`、`settingsBioInput`、`saveProfile`、`savePassword`

**HandleUpload**: `avatarUpload` を `uploads/avatars/` に保存し DB を更新

### components.go — 共有コンポーネント

複数のビューで使用される再利用可能な `ComponentFunc` ビルダー:

- **`postCard(p PostWithMeta)`** — アバター、著者リンク、コンテンツリンク、画像、アクションボタン（いいね/リツイート/コメント/共有）付きの投稿をレンダリング。いいねとリツイートボタンは状態に基づいて色が変化。ナビゲーションは `<a href>` リンクを使用。
- **`composeBox(draft, charCount)`** — 文字カウンター（250 文字上限）と写真アップロードボタン付きの投稿作成エリア。
- **`renderSearchUserItem(u User)`** — アバターと名前付きのクリック可能なユーザー検索結果。
- **`avatarComponent(u User)`** — ユーザーにアバターパスがあるかに基づいて `gu.ImageAvatar` または `gu.LetterAvatar` を返す。
- **`iconHeart(filled)`**、**`iconRetweet()`**、**`iconComment()`**、**`iconShare()`** — `g.Literal()` によるインライン SVG アイコン。
- **`formatTimeAgo(t)`** — 相対時間文字列（"now"、"5m"、"2h"、"3d"、"Jan 2"）。
- **`truncate(s, maxLen)`** — "..." サフィックス付きの文字列切り詰め。

### css.go — モバイルファースト CSS

すべての CSS は Go 文字列定数として定義。主要なレスポンシブパターン:

- ベーススタイルはモバイルを対象（シングルカラム、全幅カード）
- `@media (min-width: 769px)` でデスクトップサイドバーを表示しモバイルヘッダーを非表示
- 投稿アクションボタンはアイコン + カウントの `inline-flex`
- アクティブないいね: `color: #ef4444`（赤）、アクティブなリツイート: `color: #22c55e`（緑）
- Gerbera テーマ変数（`var(--g-*)`）を全体で使用し一貫したテーマを実現
- ドロワー CSS は `layout.go` にドロワーマークアップと共に定義

### db.go — データベースレイヤー

すべてのデータベース操作は `*sql.DB` を受け取るプレーン関数:

- **タイムラインクエリ** は posts を users と結合し、現在のユーザーのインタラクション状態に対する `EXISTS` サブクエリでいいね/リツイート/コメントカウントを集約
- **トグル操作**（いいね、リツイート、フォロー）は既存行をチェックし、挿入/削除を実行
- **会話リスト** はサブクエリで会話パートナーごとの最新メッセージを検索

## 通知フロー

```
ユーザー A がユーザー B の投稿にいいね
  → ユーザー A の TimelineView（または ProfileView 等）で HandleEvent("toggleLike")
  → baseView.doToggleLike(): dbToggleLike(db, A.userID, postID)
  → dbCreateNotification(db, B.userID, A.userID, "like", &postID)
  → hub.Notify(B.userID, NewLikeNotif{PostID, A.DisplayName})
  → ユーザー B のすべてのアクティブ View で HandleInfo(NewLikeNotif{...}) が呼ばれる
  → baseView.handleBaseInfo() がトースト表示: "Alice liked your post"
```

## 実行方法

```bash
# Docker Compose ですべて起動（MySQL + アプリ）
cd example/sns && docker compose up

# またはローカル開発: MySQL は Docker、アプリはローカル
cd example/sns && docker compose up -d mysql
cd example/sns && go run .             # http://localhost:8930
cd example/sns && go run . -debug      # デバッグパネル付き
```

1. http://localhost:8930 を開く — ログインページにリダイレクトされます
2. 「Register」をクリックして新しいアカウントを作成
3. 別のブラウザ/シークレットウィンドウで 2 つ目のアカウントを作成
4. 投稿、いいね、フォロー、メッセージ送信 — アクションが両方のタブにリアルタイムで表示されます

## 演習

1. **無限スクロール** — タイムラインの固定 50 件制限を `gu.Pagination` に置き換えるか、オフセットを追跡して `"loadMore"` イベントで追加投稿を読み込む無限スクロールを実装。
2. **投稿画像アップロード** — 現在 `TimelineView` の `HandleUpload` はファイルをログに記録するだけで保存しません。`SettingsView` のアバター処理と同様にアップロード画像をディスクに保存し、パスを `posts.image_path` に格納してください。
3. **通知ページ** — `NotificationsView` と `/live/notifications` エンドポイントを追加し、`notifications` テーブルから全通知を既読/未読状態付きでリスト表示。
4. **ブックマーク機能** — ブックマークテーブルと投稿カードにブックマークボタンを追加。保存した投稿を表示する `BookmarksView` を作成。
5. **ダークモード** — 設定にテーマ切り替えを追加。ユーザーのセッションに設定を保存し、反転した CSS 変数を持つ `dark` クラスを `<body>` に条件付きで適用。
