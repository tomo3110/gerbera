# Auth サンプルチュートリアル

このサンプルは `session/` パッケージを使用したセッションベースの認証を実演します。CSRF 保護、セッションミドルウェア、プッシュベースのセッション無効化を含みます。

## 概要

auth サンプルが実装する機能:

- CSRF トークン保護付き**ログインフォーム**（サーバーサイドレンダリング）
- `session.RequireKey` ミドルウェアで保護された**ダッシュボード** — SSR ページ内に `gl.LiveMount` で LiveView を埋め込み
- **セッション無効化** — あるタブでログアウトすると `SessionExpiredHandler` 経由で他のすべてのタブが自動的にリダイレクト

## アーキテクチャ

ダッシュボードは SSR + LiveView のハイブリッドアプローチを採用しています:

- **SSR シェル** (`dashboardPage()`) — `g.Handler()` でレンダリングされ、`<head>`（タイトル、テーマ CSS）と外側のレイアウト（`<body>`、センタリング、コンテナ）を提供します。`gl.LiveMount("/live/dashboard")` で LiveView を埋め込みます。
- **LiveView** (`DashboardView`) — SSR シェル内にマウントされます。`Render()` は `<body>` コンテンツのみ（ウェルカムカード、セッション期限切れ警告、ログアウトボタン）を返します。`<head>` やページ全体のレイアウトはレンダリング**しません**。

この分離により、静的なページ構造は初回リクエストで通常の HTML として配信され、LiveView は動的な WebSocket 駆動の部分を担当します。

## 主要コンセプト

### MemoryStore

`session.MemoryStore` はセッションデータをメモリに保持し、HMAC-SHA256 署名付きクッキーを使用します。クッキーにはセッション ID のみが含まれ、すべてのデータはサーバー側に保持されます。

```go
key := []byte("example-secret-key-change-in-prod")
store := session.NewMemoryStore(key)
defer store.Close()  // バックグラウンド GC ゴルーチンを停止
```

オプションには `WithMaxAge(duration)`、`WithCookie(config)`、`WithGCInterval(duration)` があります。

### セッションミドルウェア

`session.Middleware` は HTTP ハンドラをラップし、各リクエストでクッキーからセッションを自動的に読み込み、ハンドラの処理後に保存します:

```go
sessionMW := session.Middleware(store)
handler := sessionMW(mux)
```

セッションは `session.FromContext(r.Context())` で取得できます。

### CSRF 保護

`session/` パッケージは定数時間比較を使用した CSRF トークンヘルパーを提供します:

```go
sess := session.FromContext(r.Context())
token := session.GenerateCSRFToken(sess)  // トークンを生成しセッションに保存
// hidden フォームフィールドにトークンを含める:
// <input type="hidden" name="csrf_token" value="...">

// フォーム送信時にバリデーション:
if !session.ValidCSRFToken(sess, r.FormValue("csrf_token")) {
    http.Error(w, "invalid CSRF token", http.StatusForbidden)
    return
}
```

### RequireKey 認証ガード

`session.RequireKey` は未認証ユーザーをリダイレクトするミドルウェアを作成します:

```go
authGuard := session.RequireKey("username", "/login")
mux.Handle("GET /{$}", authGuard(protectedHandler))
```

セッションに `"username"` キーがない場合、ユーザーは `/login` にリダイレクトされます。

### SSR シェルと LiveMount

ダッシュボードページは `gl.LiveMount` を使って LiveView を埋め込む SSR ハンドラです:

```go
func dashboardPage() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(
            gd.Title("Auth Demo"),
            gu.Theme(),
        ),
        gd.Body(
            gu.Center(
                gp.Attr("style", "min-height: 100vh"),
                gu.ContainerNarrow(
                    gu.Stack(
                        gd.H1(gp.Value("Auth Demo")),
                        gl.LiveMount("/live/dashboard"),
                    ),
                ),
            ),
        ),
    }
}
```

`g.Handler(dashboardPage()...)` でこれを標準の `http.Handler` として配信します。`gl.LiveMount("/live/dashboard")` はプレースホルダーを挿入し、クライアント側の JavaScript が WebSocket 接続の LiveView にアップグレードします。

対応する LiveView エンドポイントは別途登録します:

```go
mux.Handle("GET /{$}", authGuard(g.Handler(dashboardPage()...)))
mux.Handle("/live/dashboard", authGuard(gl.Handler(func(_ context.Context) gl.View {
    return &DashboardView{}
}, liveOpts...)))
```

### プッシュベースのセッション無効化

`gl.WithSessionStore(store)` を LiveView ハンドラに渡すと、WebSocket 接続はストアの `Broker` 経由でセッション無効化イベントを購読します。セッション破棄（例: ログアウト時）により、購読中のすべての接続に通知されます。

`SessionExpiredHandler` を実装した View はコールバックを受け取ります:

```go
func (v *DashboardView) OnSessionExpired() error {
    v.SessionExpired = true
    v.Navigate("/login")
    return nil
}
```

これによりリアルタイムのクロスタブログアウトが実現します: あるブラウザタブでログアウトすると、ダッシュボードを表示している他のすべてのタブがリダイレクトされます。

## ウォークスルー

### main()

1. 署名キーを指定して `MemoryStore` を作成
2. セッションの自動読み込み/保存のために `session.Middleware` を設定
3. 認証ガードとして `session.RequireKey("username", "/login")` を設定
4. ルートを登録:
   - `GET /login` — インラインの `HandleFunc` でログインフォームを表示（ログイン済みの場合は `/` にリダイレクト）
   - `POST /login` — 資格情報と CSRF トークンを検証。エラー時は `g.ExecuteTemplate()` でステータスコード制御付きレスポンスを返す
   - `/logout` — セッションを破棄し `/login` にリダイレクト
   - `GET /{$}` — SSR ダッシュボードシェル（`g.Handler`）に `gl.LiveMount("/live/dashboard")` を埋め込み
   - `/live/dashboard` — LiveView WebSocket エンドポイント（`gl.Handler`）、`WithSessionStore` 付き

### ログインフロー

1. ユーザーが `/login` にアクセス — `GET /login` ハンドラが `g.ExecuteTemplate()` で CSRF トークン付きの `loginPage()` を表示
2. ユーザーが資格情報を送信 — `POST /login` ハンドラが CSRF トークンとパスワードを検証
3. 失敗時、ハンドラは `w.WriteHeader(http.StatusUnauthorized)` を呼んだ後 `g.ExecuteTemplate()` でエラーメッセージ付きのログインフォームを再表示。ここで `g.ExecuteTemplate()` を使う理由は（`g.Handler()` ではなく）、レスポンスボディ書き込み前に HTTP ステータスコードを設定できるため
4. 成功時、`sess.Set("username", username)` でセッションにユーザーを保存し `/` にリダイレクト
5. ユーザーは `GET /{$}` にリダイレクトされ SSR シェルが配信される。埋め込まれた `LiveMount` が `DashboardView` LiveView に接続

### DashboardView

- `Render()` は `<body>` コンテンツのみを返す — ウェルカムカード、セッション期限切れ警告、ログアウトボタン。`<head>` やページ全体のレイアウトは含まず、それらは SSR シェルが提供する
- `Mount()` が `params.Conn.Session.GetString("username")` を読み取ってユーザー名を表示
- `OnSessionExpired()` が警告フラグを設定し `v.Navigate("/login")` でリダイレクト

## 実行方法

```bash
go run example/auth/auth.go          # http://localhost:8895
go run example/auth/auth.go -debug   # デバッグパネル付き
```

任意のユーザー名とパスワード `"password"` でログインできます。複数タブを開いてクロスタブセッション無効化をテストしてください。
