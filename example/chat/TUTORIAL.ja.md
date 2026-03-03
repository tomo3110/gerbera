# Chat サンプルチュートリアル

このサンプルは `InfoReceiver`、`Unmounter`、`Session.SendInfo()` を使用して、接続された LiveView セッション間でメッセージをブロードキャストするマルチユーザーリアルタイムチャットルームを実演します。

## 概要

chat サンプルが実装する機能:

- **参加フォーム** — ユーザーがユーザー名を入力してチャットルームに参加
- **リアルタイムメッセージング** — メッセージが接続中のすべてのユーザーに即座にブロードキャスト
- **オンラインカウント** — 接続中のユーザー数を表示
- **システム通知** — 参加/退出イベントをシステムメッセージとして表示
- **自動スクロール** — チャットが最新メッセージに自動スクロール

## 主要コンセプト

### Hub（Pub/Sub）

`Hub` 構造体（`hub.go`）は接続中の LiveView セッションを追跡し、メッセージをブロードキャストするシンプルな pub/sub です:

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]*live.Session
}
```

- `Join(sess)` — LiveView セッションを登録
- `Leave(sessionID)` — セッションを削除
- `Broadcast(msg, senderID)` — 送信者以外のすべてのセッションに `sess.SendInfo(msg)` 経由でメッセージを送信
- `OnlineCount()` — 接続中のクライアント数を返す

すべての接続で共有される単一の `Hub` インスタンスを使用します:

```go
var hub = NewHub()
```

### InfoReceiver

`InfoReceiver` はサーバーサイドメッセージを受信できる LiveView インターフェースです。他のユーザーがチャットメッセージを送信すると、`HandleInfo()` 経由で配信されます:

```go
type InfoReceiver interface {
    View
    HandleInfo(msg any) error
}
```

```go
func (v *ChatView) HandleInfo(msg any) error {
    if cm, ok := msg.(ChatMessage); ok {
        v.Messages = append(v.Messages, cm)
        v.ScrollIntoPct("#chat-messages", "1.0")
    }
    return nil
}
```

### Session.SendInfo()

`SendInfo()` は LiveView セッションの info チャネルにメッセージをプッシュします。メッセージは View の `HandleInfo()` メソッドに配信され、再レンダリングがトリガーされます:

```go
func (h *Hub) Broadcast(msg ChatMessage, senderID string) {
    for id, sess := range h.clients {
        if id == senderID {
            continue
        }
        sess.SendInfo(msg)
    }
}
```

### Unmounter

`Unmounter` は WebSocket 接続が閉じられた時（例: ユーザーがタブを閉じた時）に呼ばれます。チャットでは Hub からの退出と退出通知のブロードキャストに使用します:

```go
type Unmounter interface {
    Unmount()
}
```

```go
func (v *ChatView) Unmount() {
    if v.Username != "" && v.session != nil {
        v.hub.Leave(v.session.ID)
        v.hub.Broadcast(ChatMessage{
            Author:  v.Username,
            Content: fmt.Sprintf("%s left the room", v.Username),
            System:  true,
        }, v.session.ID)
    }
}
```

### LiveSession アクセス

`Mount()` メソッドは `params.Conn.LiveSession` から `LiveSession` 参照を保存します。これは `SendInfo()` を提供する `*live.Session` です:

```go
func (v *ChatView) Mount(params gl.Params) error {
    v.hub = hub
    v.session = params.Conn.LiveSession
    return nil
}
```

## ウォークスルー

### chat.go — ChatView

1. **Mount** — 共有 Hub と LiveSession への参照を保存
2. **HandleEvent** — 3 つのイベントを処理:
   - `"join"` — ユーザー名を設定し、Hub に参加、参加通知をブロードキャスト
   - `"input"` — 下書きメッセージテキストを更新
   - `"send"` / `"keydown"`（Enter） — メッセージをローカルに追加、下書きをクリア、他のユーザーにブロードキャスト
3. **HandleInfo** — 他のユーザーからの `ChatMessage` を受信し、メッセージリストに追加
4. **Unmount** — Hub から退出し、退出通知をブロードキャスト
5. **Render** — `Username` が設定されているかどうかで参加フォームまたはチャット画面を表示

### hub.go — Hub

セッション ID から `*live.Session` へのスレッドセーフなマップです。`Broadcast()` は送信者以外のすべてのクライアントを反復し、`sess.SendInfo(msg)` を呼び出します。

### メッセージフロー

```
ユーザー A がメッセージを入力
  → ユーザー A の View で HandleEvent("send")
  → hub.Broadcast(msg, A.sessionID)
  → 他の各セッションに対して: sess.SendInfo(msg)
  → ユーザー B の HandleInfo(msg) が呼ばれる
  → ユーザー B の View が新しいメッセージで再レンダリング
```

## 実行方法

```bash
go run example/chat/chat.go          # http://localhost:8920
go run example/chat/chat.go -debug   # デバッグパネル付き
```

複数のブラウザタブを開いて複数ユーザーをシミュレートしてください。各タブで異なるユーザー名で参加できます。
