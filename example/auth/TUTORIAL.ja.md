# Auth サンプルチュートリアル

このサンプルは `session/` パッケージを使用したセッションベースの認証を実演します。CSRF 保護、セッションミドルウェア、プッシュベースのセッション無効化を含みます。

## 概要

auth サンプルが実装する機能:

- CSRF トークン保護付き**ログインフォーム**（サーバーサイドレンダリング）
- `session.RequireKey` ミドルウェアで保護された **LiveView ダッシュボード**
- **セッション無効化** — あるタブでログアウトすると `SessionExpiredHandler` 経由で他のすべてのタブが自動的にリダイレクト

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
mux.Handle("/", authGuard(protectedHandler))
```

セッションに `"username"` キーがない場合、ユーザーは `/login` にリダイレクトされます。

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
   - `GET /login` — ログインフォームを表示（ログイン済みの場合は `/` にリダイレクト）
   - `POST /login` — 資格情報と CSRF トークンを検証
   - `/logout` — セッションを破棄し `/login` にリダイレクト
   - `/` — `WithSessionStore` 付きの保護された LiveView ダッシュボード

### ログインフロー

1. ユーザーが `/login` にアクセス → `loginFormPage()` が CSRF トークン付きフォームを表示
2. ユーザーが資格情報を送信 → `loginPostHandler()` が CSRF トークンとパスワードを検証
3. 成功時、`sess.Set("username", username)` でセッションにユーザーを保存
4. ユーザーは `/` にリダイレクトされ `DashboardView` が表示される

### DashboardView

- `Mount()` が `params.Conn.Session.GetString("username")` を読み取ってユーザー名を表示
- `OnSessionExpired()` が警告フラグを設定し `v.Navigate("/login")` でリダイレクト

## 実行方法

```bash
go run example/auth/auth.go          # http://localhost:8895
go run example/auth/auth.go -debug   # デバッグパネル付き
```

任意のユーザー名とパスワード `"password"` でログインできます。複数タブを開いてクロスタブセッション無効化をテストしてください。
