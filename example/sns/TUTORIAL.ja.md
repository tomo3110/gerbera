# SNS（Twitter クローン）チュートリアル

このサンプルは MySQL 永続化、セッション認証、マルチページ LiveView ルーティング、リアルタイム Hub 通知、ファイルアップロード、モバイルファーストレスポンシブデザインを組み合わせたフル機能の Twitter クローンを実演します。Gerbera の `ui/` コンポーネントライブラリを活用しています。

## 概要

SNS サンプルが実装する機能:

- **セッション認証** — CSRF 保護と SHA-256 パスワードハッシュ付きのログイン/登録（静的ページ）
- **単一 LiveView SPA** — `v.page` 状態による内部ページルーティング（`example/admin/` と同じパターン）
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
    salt, _ := hex.DecodeString(parts[0])
    hash := sha256.Sum256(append(salt, []byte(password)...))
    return parts[1] == hex.EncodeToString(hash[:])
}
```

### ユーザーマッピング付き Hub

chat サンプルのシンプルなセッションベース Hub とは異なり、SNS Hub はデータベースのユーザー ID を LiveView セッションにマッピングし、特定ユーザーへの通知送信を可能にします:

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]*live.Session // ライブセッション ID → セッション
    userMap map[int64]string         // ユーザー DB ID → ライブセッション ID
}

func (h *Hub) Notify(userID int64, msg any) {
    if sessID, ok := h.userMap[userID]; ok {
        if sess, ok := h.clients[sessID]; ok {
            sess.SendInfo(msg)
        }
    }
}
```

`HandleInfo` で区別できるよう型付き通知メッセージを使用します:

```go
type NewLikeNotif struct {
    PostID    int64
    ActorName string
}
type NewMessageNotif struct {
    SenderID int64
    Content  string
}
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

1. フラグをパース（`-addr`、`-dsn`、`-debug`）
2. MySQL 接続を開き、自動マイグレーションを実行
3. `session.MemoryStore` とセッションミドルウェアを作成
4. ルートを登録:
   - `GET /login` — ログインフォームを表示（認証済みの場合は `/` にリダイレクト）
   - `POST /login` — 資格情報を検証し、セッションに `user_id` を設定
   - `GET /register` — 登録フォームを表示
   - `POST /register` — ユーザーを作成し、セッションに `user_id` を設定
   - `/logout` — セッションを破棄し `/login` にリダイレクト
   - `/` — `session.RequireKey("user_id", "/login")` で保護された LiveView（`SNSView`）

LiveView ハンドラは `gl.WithSessionStore(store)` を渡して `Mount()` でのセッションアクセスを有効化:

```go
mux.Handle("/", authGuard(gl.Handler(func(_ context.Context) gl.View {
    return &SNSView{
        db:  db,
        hub: hub,
    }
}, liveOpts...)))
```

### auth.go — 静的認証ページ

ログインと登録ページは `g.Handler()` を使用したサーバーサイドレンダリング（LiveView ではない）です。`isRegister` フラグに基づいて HTML を生成する `renderAuthPage()` 関数を共有します。POST エラーレスポンスではステータスコード制御のために `g.ExecuteTemplate()` を直接使用しています。

CSRF トークンは `session.GenerateCSRFToken(sess)` で生成し、フォーム送信時にバリデーションします。

### view.go — SNSView

`SNSView` 構造体がすべてのアプリケーション状態を保持:

```go
type SNSView struct {
    gl.CommandQueue
    db      *sql.DB
    hub     *Hub
    session *gl.Session

    userID int64
    user   *User
    page   string  // "home", "profile", "post", "messages", "settings", "search"

    // ページ固有の状態...
    posts        []PostWithMeta
    composeDraft string
    // ...
}
```

**Mount** は HTTP セッションからユーザー ID を読み取り、データベースからユーザーを読み込み、Hub に参加し、初期ページデータを読み込みます:

```go
func (v *SNSView) Mount(params gl.Params) error {
    v.session = params.Conn.LiveSession
    if params.Conn.Session != nil {
        if uid, ok := params.Conn.Session.Get("user_id").(int64); ok {
            v.userID = uid
        }
    }
    // ユーザー読み込み、Hub 参加、ページデータ読み込み...
}
```

**HandleEvent** は 25 以上のイベントを機能別にルーティング:

- **ナビゲーション**: `"nav"`、`"toggleDrawer"`、`"closeDrawer"`
- **投稿作成**: `"composeInput"`、`"submitPost"`
- **投稿アクション**: `"toggleLike"`、`"toggleRetweet"`、`"viewPost"`、`"viewProfile"`、`"sharePost"`
- **フォロー**: `"toggleFollow"`
- **コメント**: `"commentInput"`、`"submitComment"`
- **メッセージ**: `"openChat"`、`"backToConversations"`、`"chatInput"`、`"chatSend"`、`"chatKeydown"`
- **設定**: `"settingsDisplayNameInput"`、`"settingsEmailInput"`、`"settingsBioInput"`、`"saveProfile"`、`"savePassword"`
- **検索**: `"searchInput"`（`gl.Debounce(300)` 付き）
- **削除**: `"confirmDeletePost"`、`"doDelete"`、`"cancelDelete"`

**HandleInfo** は Hub から型付き通知を受信し、UI を更新:

```go
func (v *SNSView) HandleInfo(msg any) error {
    switch m := msg.(type) {
    case NewLikeNotif:
        v.showToast(fmt.Sprintf("%s liked your post", m.ActorName), "info")
        v.unreadNotifications++
        v.refreshCurrentPosts()
    case NewMessageNotif:
        if v.page == "messages" && v.chatPartnerID == m.SenderID {
            v.loadChat(m.SenderID)  // リアルタイムでチャットを更新
        } else {
            v.showToast("New message received", "info")
        }
    // ...
    }
}
```

**Unmount** は WebSocket が閉じた時に Hub から離脱:

```go
func (v *SNSView) Unmount() {
    if v.session != nil {
        v.hub.Leave(v.session.ID, v.userID)
    }
}
```

### pages.go — ページレンダリング

`Render()` メソッドがフルページレイアウトを構築:

```go
func (v *SNSView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(
            gd.Title("SNS"),
            gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
            gu.Theme(),
            gs.CSS(snsCSS),
        ),
        gd.Body(
            // モバイルヘッダー（CSS でデスクトップでは非表示）
            // モバイルドロワー（gul.Drawer）
            // メインレイアウト: サイドバー + コンテンツ
            // トーストオーバーレイ
            // 確認ダイアログ
        ),
    }
}
```

`renderContent()` は `switch v.page` によりページ固有メソッドにディスパッチします。

各ページメソッド（例: `renderTimeline()`、`renderProfile()`、`renderMessages()`）は `ui/` および `ui/live/` コンポーネントを組み合わせます:

- **タイムライン**: `composeBox()` + `postCard()` コンポーネントのリスト
- **プロフィール**: `gu.LetterAvatar` / `gu.ImageAvatar`、フォローボタン、プロフィール統計、投稿リスト
- **投稿詳細**: `postCard()` + コメントリスト + コメントフォーム
- **メッセージ**: 会話リストまたはアクティブチャット用の `gu.ChatContainer` / `gu.ChatInput`
- **設定**: プロフィール編集用の `gu.FormGroup` / `gu.FormInput` / `gu.FormTextarea`
- **検索**: `gl.Debounce(300)` 付き `gu.FormInput`、ユーザーと投稿に分割された結果表示

### components.go — 共有コンポーネント

複数ページで使用される再利用可能な `ComponentFunc` ビルダー:

- **`postCard(p PostWithMeta)`** — アバター、著者名、コンテンツ、画像、アクションボタン（いいね/リツイート/コメント/共有）付きの投稿をレンダリング。いいねとリツイートボタンは状態に基づいて色が変化。
- **`navLink(icon, label, page, active, badgeCount)`** — オプションの未読バッジ付きサイドバー/ドロワーナビゲーションアイテム。
- **`composeBox(draft, charCount)`** — 文字カウンターと写真アップロードボタン付きの投稿作成エリア。
- **`iconHeart(filled)`**、**`iconRetweet()`**、**`iconComment()`**、**`iconShare()`** — `g.Literal()` によるインライン SVG アイコン。
- **`avatarComponent(u User)`** — ユーザーにアバターパスがあるかに基づいて `gu.ImageAvatar` または `gu.LetterAvatar` を返す。

### css.go — モバイルファースト CSS

すべての CSS は Go 文字列定数として定義。主要なレスポンシブパターン:

- ベーススタイルはモバイルを対象（シングルカラム、全幅カード）
- `@media (min-width: 769px)` でデスクトップサイドバーを表示しモバイルヘッダーを非表示
- 投稿アクションボタンはアイコン + カウントの `inline-flex`
- アクティブないいね: `color: #ef4444`（赤）、アクティブなリツイート: `color: #22c55e`（緑）
- Gerbera テーマ変数（`var(--g-*)`）を全体で使用し一貫したテーマを実現

### db.go — データベースレイヤー

すべてのデータベース操作は `*sql.DB` を受け取るプレーン関数:

- **タイムラインクエリ** は posts を users と結合し、現在のユーザーのインタラクション状態に対する `EXISTS` サブクエリでいいね/リツイート/コメントカウントを集約
- **トグル操作**（いいね、リツイート、フォロー）は既存行をチェックし、挿入/削除を実行
- **会話リスト** はサブクエリで会話パートナーごとの最新メッセージを検索

## 通知フロー

```
ユーザー A がユーザー B の投稿にいいね
  → ユーザー A の View で HandleEvent("toggleLike")
  → dbToggleLike(db, A.userID, postID)
  → dbCreateNotification(db, B.userID, A.userID, "like", &postID)
  → hub.Notify(B.userID, NewLikeNotif{PostID, A.DisplayName})
  → ユーザー B の HandleInfo(NewLikeNotif{...}) が呼ばれる
  → ユーザー B にトースト表示: "Alice liked your post"
  → ユーザー B の投稿カウントがリアルタイムで更新
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
2. **画像アップロード** — 現在 `HandleUpload` はファイルをログに記録するだけで保存しません。アップロード画像をディスクに保存し、パスを `posts.image_path` または `users.avatar_path` に格納してください。
3. **通知ページ** — `notifications` テーブルから全通知を既読/未読状態付きでリスト表示する専用通知ページを追加。
4. **ブックマーク機能** — ブックマークテーブルと投稿カードにブックマークボタンを追加。保存した投稿を表示する「ブックマーク」ページを作成。
5. **ダークモード** — 設定にテーマ切り替えを追加。ユーザーのセッションに設定を保存し、反転した CSS 変数を持つ `dark` クラスを `<body>` に条件付きで適用。
