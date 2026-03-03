# Clock チュートリアル

## 概要

ユーザー操作なしに自動更新されるリアルタイム時計とストップウォッチを作成します。このチュートリアルでは以下の機能を学びます。

- `gl.TickerView` インターフェース — サーバー駆動の定期 UI 更新
- `gl.View` インターフェース — ステートフルな LiveView コンポーネントの定義
- `ge.If` — 条件付きレンダリング（Start/Stop ボタンの切り替え）
- `gs.CSS` — インライン CSS スタイルの埋め込み

このサンプルは **サーバープッシュ更新** を示します。Go サーバーが一定間隔で `HandleTick` を呼び、再レンダーし、DOM パッチのみをブラウザに送信します。時計の更新にユーザーのクリックは必要ありません。

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: パッケージのインポート

```go
import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)
```

標準の gerbera エイリアスに加えて、以下を使用します:
- `ge`（`expr`）— `If` による条件付きレンダリング
- `gs`（`styles`）— `CSS()` によるインライン CSS
- `time` — 時計・ストップウォッチの時間管理

## ステップ 2: View 構造体の定義

```go
type ClockView struct {
	Now       time.Time
	Running   bool
	Elapsed   time.Duration
	StartedAt time.Time
}
```

構造体は 4 つの状態を保持します:
- `Now` — 現在時刻。毎ティックで更新
- `Running` — ストップウォッチが動作中かどうか
- `Elapsed` — ストップウォッチの累積経過時間
- `StartedAt` — ストップウォッチの開始時刻（一時停止を考慮して調整）

## ステップ 3: Mount の実装

```go
func (v *ClockView) Mount(_ gl.Params) error {
	v.Now = time.Now()
	return nil
}
```

`Mount` で現在時刻を初期化し、最初のレンダリングで正しい時刻を表示します。ストップウォッチ関連のフィールドはゼロ値（停止中、経過 0）のままです。

## ステップ 4: TickerView の実装

```go
func (v *ClockView) TickInterval() time.Duration {
	return 100 * time.Millisecond
}

func (v *ClockView) HandleTick() error {
	v.Now = time.Now()
	if v.Running {
		v.Elapsed = v.Now.Sub(v.StartedAt)
	}
	return nil
}
```

これがこのサンプルの中核機能です。`TickerView` インターフェースを実装すると、フレームワークが自動的に:
1. 指定した間隔で `time.Ticker` を作成
2. ティックごとに `HandleTick` を呼び出し
3. 再レンダーして DOM パッチをブラウザに送信

`TickInterval` は `100ms` を返し、ストップウォッチの 10 分の 1 秒表示を滑らかにします。`HandleTick` は現在時刻を更新し、ストップウォッチが動作中であれば経過時間を再計算します。

## ステップ 5: HandleEvent の実装

```go
func (v *ClockView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "start":
		if !v.Running {
			v.Running = true
			v.StartedAt = time.Now().Add(-v.Elapsed)
		}
	case "stop":
		if v.Running {
			v.Elapsed = time.Now().Sub(v.StartedAt)
			v.Running = false
		}
	case "reset":
		v.Running = false
		v.Elapsed = 0
	}
	return nil
}
```

3 つのユーザー操作でストップウォッチを制御します:
- **start** — 開始時刻を記録。過去に蓄積された経過時間分だけ遡ることで、再開時に正しく継続します
- **stop** — 経過時間をスナップショットして一時停止
- **reset** — 停止し、経過時間をクリア

## ステップ 6: Render の実装

```go
func (v *ClockView) Render() []g.ComponentFunc {
	clock := v.Now.Format("15:04:05")
	mins := int(v.Elapsed.Minutes())
	secs := int(v.Elapsed.Seconds()) % 60
	tenths := int(v.Elapsed.Milliseconds()/100) % 10
	stopwatch := fmt.Sprintf("%02d:%02d.%d", mins, secs, tenths)

	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Clock — Gerbera TickerView Demo"),
			gs.CSS(`...`),
		),
		gd.Body(
			gd.Div(
				gp.Class("container"),
				gd.H1(gp.Value("Gerbera Clock")),

				gd.Div(gp.Class("label"), gp.Value("Current Time")),
				gd.Div(gp.Class("clock"), gp.Value(clock)),

				gd.Div(gp.Class("label"), gp.Value("Stopwatch")),
				gd.Div(gp.Class("stopwatch"), gp.Value(stopwatch)),

				gd.Div(
					gp.Class("buttons"),
					ge.If(!v.Running,
						gd.Button(gp.Class("btn-start"), gl.Click("start"), gp.Value("Start")),
						gd.Button(gp.Class("btn-stop"), gl.Click("stop"), gp.Value("Stop")),
					),
					gd.Button(gp.Class("btn-reset"), gl.Click("reset"), gp.Value("Reset")),
				),
			),
		),
	}
}
```

ポイント:
- 時計は `HH:MM:SS`、ストップウォッチは `MM:SS.T`（10 分の 1 秒）形式でフォーマット
- `ge.If(!v.Running, startButton, stopButton)` で状態に応じて Start と Stop を切り替え
- `gs.CSS(...)` は `<head>` 内に `<style>` 要素を埋め込みます — 外部 CSS ファイルは不要
- Reset ボタンは常に表示

## ステップ 7: サーバーの起動

```go
func main() {
	addr := flag.String("addr", ":8870", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &ClockView{} }, opts...))
	log.Printf("clock running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

Counter サンプルと同じ標準的な LiveView サーバーセットアップです。

## 動作の仕組み

```
ブラウザ                                  Go サーバー
  |                                          |
  |  1. GET /                                |
  |  <-- 初期 HTML + gerbera.js              |
  |                                          |
  |  2. WebSocket 接続                       |
  |  <-> 双方向接続確立                       |
  |                                          |
  |  3. 100ms ごと: HandleTick               |
  |      Now = time.Now()                    |
  |      Elapsed = Now - StartedAt           |
  |      -> Render -> Diff -> パッチ生成      |
  |                                          |
  |  <-- [{"op":"text","path":[...],"val":   |
  |        "15:04:05"}]                      |
  |                                          |
  |  4. JS がパッチを DOM に適用              |
  |                                          |
  |  5. ユーザーが「Start」をクリック          |
  |  --> {"e":"start","p":{}}                |
  |      HandleEvent -> Running=true         |
```

Counter サンプルではユーザー操作が更新のトリガーですが、このサンプルでは **サーバー駆動の更新** を示しています。時計とストップウォッチはティッカーにより継続的に更新され、Start/Stop/Reset ボタンはユーザー駆動のイベントです。

## 実行方法

```bash
go run example/clock/clock.go
```

ブラウザで http://localhost:8870 を開きます。時計がリアルタイムで更新されます。**Start** でストップウォッチ開始、**Stop** で一時停止、**Reset** でクリアです。

### デバッグモード

```bash
go run example/clock/clock.go -debug
```

デバッグパネルの **Patches** タブを観察すると、ティッカーによって 100ms ごとにパッチが到着する様子を確認できます。

## 発展課題

1. 「Lap」ボタンをクリックしてスプリットタイムを記録するラップタイマーを追加してみましょう
2. ティック間隔を `1s` に変更し、ストップウォッチの表示が粗くなることを確認してみましょう
3. 時計の下に現在の日付を表示してみましょう
4. `ge.If` を使って、ストップウォッチが 60 秒を超えたら色を変えてみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gl.TickerView` | インターフェース: `View` を拡張し `TickInterval()` と `HandleTick()` を追加 |
| `TickInterval()` | ティック間隔を返す。`0` を返すと無効化 |
| `HandleTick()` | 各ティックで呼ばれる。ここで状態を更新 |
| `ge.If(cond, true, false)` | 条件付きレンダリング — 条件に応じて 1 つ目または 2 つ目の `ComponentFunc` を描画 |
| `gs.CSS(text)` | 指定した CSS テキストの `<style>` 要素を埋め込む |
