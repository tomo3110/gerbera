# Clock Tutorial

## Overview

Build a real-time clock and stopwatch that updates automatically without any user interaction. This tutorial covers the following features:

- `gl.TickerView` interface — Server-driven periodic UI updates
- `gl.View` interface — Defining a stateful LiveView component
- `ge.If` — Conditional rendering (toggling Start/Stop buttons)
- `gs.CSS` — Embedding inline CSS styles

This example demonstrates **server-push updates**: the Go server triggers `HandleTick` at a fixed interval, re-renders, and sends only the DOM patches to the browser. No user clicks are needed to keep the clock ticking.

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Importing Packages

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

In addition to the standard gerbera aliases, this example uses:
- `ge` (`expr`) — Conditional rendering with `If`
- `gs` (`styles`) — Inline CSS via `CSS()`
- `time` — For clock/stopwatch time tracking

## Step 2: Defining the View Struct

```go
type ClockView struct {
	Now       time.Time
	Running   bool
	Elapsed   time.Duration
	StartedAt time.Time
}
```

The struct holds four pieces of state:
- `Now` — Current time, updated every tick
- `Running` — Whether the stopwatch is active
- `Elapsed` — Accumulated stopwatch time
- `StartedAt` — When the stopwatch was last started (adjusted for pauses)

## Step 3: Implementing Mount

```go
func (v *ClockView) Mount(_ gl.Params) error {
	v.Now = time.Now()
	return nil
}
```

`Mount` initializes the current time so the very first render shows the correct clock. The stopwatch fields default to their zero values (stopped, 0 elapsed).

## Step 4: Implementing TickerView

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

This is the key feature of this example. By implementing the `TickerView` interface, the framework automatically:
1. Creates a `time.Ticker` with the returned interval
2. Calls `HandleTick` on each tick
3. Re-renders and sends DOM patches to the browser

`TickInterval` returns `100ms`, giving the stopwatch a smooth tenths-of-a-second display. `HandleTick` updates the current time and, if the stopwatch is running, recalculates the elapsed duration.

## Step 5: Implementing HandleEvent

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

Three user actions control the stopwatch:
- **start** — Records the start time, adjusting backwards by any previously accumulated elapsed time so the stopwatch resumes correctly
- **stop** — Snapshots the elapsed time and pauses
- **reset** — Stops and clears the elapsed time

## Step 6: Implementing Render

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

Key points:
- The clock is formatted as `HH:MM:SS` and the stopwatch as `MM:SS.T` (tenths of a second)
- `ge.If(!v.Running, startButton, stopButton)` toggles between Start and Stop based on state
- `gs.CSS(...)` embeds a `<style>` element in the `<head>` — no external CSS files needed
- The Reset button is always visible

## Step 7: Starting the Server

```go
func main() {
	addr := flag.String("addr", ":8870", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func() gl.View { return &ClockView{} }, opts...))
	log.Printf("clock running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

Standard LiveView server setup — same pattern as the counter example.

## How It Works

```
Browser                                   Go Server
  |                                          |
  |  1. GET /                                |
  |  <-- Initial HTML + gerbera.js           |
  |                                          |
  |  2. WebSocket connect                    |
  |  <-> Bidirectional connection            |
  |                                          |
  |  3. Every 100ms: HandleTick              |
  |      Now = time.Now()                    |
  |      Elapsed = Now - StartedAt           |
  |      -> Render -> Diff -> Patches        |
  |                                          |
  |  <-- [{"op":"text","path":[...],"val":   |
  |        "15:04:05"}]                      |
  |                                          |
  |  4. JS applies patches to DOM            |
  |                                          |
  |  5. User clicks "Start"                  |
  |  --> {"e":"start","p":{}}                |
  |      HandleEvent -> Running=true         |
```

Unlike the counter example where updates are user-driven, this example shows **server-driven updates**. The clock and stopwatch update continuously via the ticker, while the Start/Stop/Reset buttons are user-driven events.

## Running

```bash
go run example/clock/clock.go
```

Open http://localhost:8870 in your browser. The clock updates in real time. Click **Start** to begin the stopwatch, **Stop** to pause, and **Reset** to clear.

### Debug Mode

```bash
go run example/clock/clock.go -debug
```

In the debug panel, observe the **Patches** tab — you will see a stream of patches arriving every 100ms as the ticker drives UI updates.

## Exercises

1. Add a lap timer that records split times when a "Lap" button is clicked
2. Change the tick interval to `1s` and observe how the stopwatch display becomes less smooth
3. Add a date display below the clock showing the current date
4. Use `ge.If` to highlight the stopwatch in a different color when it exceeds 60 seconds

## API Reference

| Function | Description |
|----------|-------------|
| `gl.TickerView` | Interface: extends `View` with `TickInterval()` and `HandleTick()` |
| `TickInterval()` | Returns the tick interval; return `0` to disable |
| `HandleTick()` | Called on each tick; update state here |
| `ge.If(cond, true, false)` | Conditional rendering — renders first or second `ComponentFunc` |
| `gs.CSS(text)` | Embeds a `<style>` element with the given CSS text |
