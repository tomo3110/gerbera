package live

import (
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/session"
)

// --- test views ---

type loopTestView struct {
	Count int
}

func (v *loopTestView) Mount(_ Params) error   { v.Count = 0; return nil }
func (v *loopTestView) HandleEvent(event string, _ Payload) error {
	if event == "inc" {
		v.Count++
	}
	return nil
}
func (v *loopTestView) Render() g.Components {
	return g.Components{
		gd.Head(gd.Title("Test")),
		gd.Body(gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.Count)))),
	}
}

type loopTickerView struct {
	loopTestView
	Ticks int
}

func (v *loopTickerView) TickInterval() time.Duration { return 10 * time.Millisecond }
func (v *loopTickerView) HandleTick() error           { v.Ticks++; return nil }

type loopInfoView struct {
	loopTestView
	LastInfo any
}

func (v *loopInfoView) HandleInfo(msg any) error { v.LastInfo = msg; return nil }

type loopSessionExpiredView struct {
	loopTestView
	Expired bool
}

func (v *loopSessionExpiredView) OnSessionExpired() error {
	v.Expired = true
	return nil
}

type loopUnmountView struct {
	loopTestView
	Unmounted bool
}

func (v *loopUnmountView) Unmount() { v.Unmounted = true }

type loopPatcherView struct {
	loopTestView
	Page string
}

func (v *loopPatcherView) HandleParams(path string, params url.Values) error {
	switch path {
	case "/profile":
		v.Page = "profile"
	case "/search":
		v.Page = "search"
	default:
		v.Page = "home"
	}
	return nil
}

func (v *loopPatcherView) Render() g.Components {
	return g.Components{
		gd.Head(gd.Title("Test")),
		gd.Body(gd.H1(gp.Value(fmt.Sprintf("Page: %s", v.Page)))),
	}
}

// --- tests ---

func TestViewLoopBasicEvent(t *testing.T) {
	view := &loopTestView{}
	view.Mount(Params{})

	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	var loopErr error
	go func() {
		defer wg.Done()
		loopErr = ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
		})
	}()

	// Send an event
	tr.PushEvent("inc", nil)

	// Give the loop time to process
	time.Sleep(50 * time.Millisecond)

	if view.Count != 1 {
		t.Errorf("expected Count=1, got %d", view.Count)
	}

	// Check that a Message was sent
	tr.mu.Lock()
	msgCount := len(tr.Messages)
	tr.mu.Unlock()
	if msgCount == 0 {
		t.Error("expected at least one message to be sent")
	}

	// Close transport to end the loop
	tr.Close()
	wg.Wait()

	if loopErr != nil {
		t.Errorf("expected nil error, got %v", loopErr)
	}
}

func TestViewLoopMultipleEvents(t *testing.T) {
	view := &loopTestView{}
	view.Mount(Params{})

	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
		})
	}()

	for i := 0; i < 5; i++ {
		tr.PushEvent("inc", nil)
		time.Sleep(10 * time.Millisecond)
	}

	if view.Count != 5 {
		t.Errorf("expected Count=5, got %d", view.Count)
	}

	tr.Close()
	wg.Wait()
}

func TestViewLoopTicker(t *testing.T) {
	view := &loopTickerView{}
	view.Mount(Params{})

	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
		})
	}()

	// Wait for a few ticks
	time.Sleep(80 * time.Millisecond)

	tr.Close()
	wg.Wait()

	if view.Ticks < 3 {
		t.Errorf("expected at least 3 ticks, got %d", view.Ticks)
	}
}

func TestViewLoopInfoReceiver(t *testing.T) {
	view := &loopInfoView{}
	view.Mount(Params{})

	infoCh := make(chan any, 8)
	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
			InfoCh:    infoCh,
		})
	}()

	infoCh <- "hello-info"
	time.Sleep(50 * time.Millisecond)

	if view.LastInfo != "hello-info" {
		t.Errorf("expected LastInfo='hello-info', got %v", view.LastInfo)
	}

	tr.Close()
	wg.Wait()
}

func TestViewLoopSessionExpired(t *testing.T) {
	view := &loopSessionExpiredView{}
	view.Mount(Params{})

	tr := NewTestTransport()
	broker := session.NewBroker()
	httpSessionID := "http-sess-123"

	var wg sync.WaitGroup
	wg.Add(1)
	var loopErr error
	go func() {
		defer wg.Done()
		loopErr = ViewLoop(view, tr, ViewLoopConfig{
			SessionID:     "test-sess",
			Lang:          "en",
			Broker:        broker,
			HTTPSessionID: httpSessionID,
		})
	}()

	// Invalidate the session
	time.Sleep(20 * time.Millisecond)
	broker.Invalidate(httpSessionID)
	wg.Wait()

	if loopErr != ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got %v", loopErr)
	}
	if !view.Expired {
		t.Error("expected OnSessionExpired to be called")
	}
}

func TestViewLoopCloseReturnsNil(t *testing.T) {
	view := &loopTestView{}
	view.Mount(Params{})

	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	var loopErr error
	go func() {
		defer wg.Done()
		loopErr = ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
		})
	}()

	tr.Close()
	wg.Wait()

	if loopErr != nil {
		t.Errorf("expected nil on clean close, got %v", loopErr)
	}
}

func TestViewLoopUnmount(t *testing.T) {
	view := &loopUnmountView{}
	view.Mount(Params{})

	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
		})
	}()

	tr.Close()
	wg.Wait()

	if !view.Unmounted {
		t.Error("expected Unmount() to be called on ViewLoop exit")
	}
}

func TestPushPatchCommand(t *testing.T) {
	var q CommandQueue
	q.PushPatch("/profile")
	cmds := q.DrainCommands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Cmd != "push_patch" {
		t.Errorf("expected cmd=push_patch, got %s", cmds[0].Cmd)
	}
	if cmds[0].Args["path"] != "/profile" {
		t.Errorf("expected path=/profile, got %s", cmds[0].Args["path"])
	}
}

func TestReplacePatchCommand(t *testing.T) {
	var q CommandQueue
	q.ReplacePatch("/?sort=date")
	cmds := q.DrainCommands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Cmd != "replace_patch" {
		t.Errorf("expected cmd=replace_patch, got %s", cmds[0].Cmd)
	}
	if cmds[0].Args["path"] != "/?sort=date" {
		t.Errorf("expected path=/?sort=date, got %s", cmds[0].Args["path"])
	}
}

func TestPushNavigateCommand(t *testing.T) {
	var q CommandQueue
	q.PushNavigate("/other-page")
	cmds := q.DrainCommands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Cmd != "push_navigate" {
		t.Errorf("expected cmd=push_navigate, got %s", cmds[0].Cmd)
	}
	if cmds[0].Args["path"] != "/other-page" {
		t.Errorf("expected path=/other-page, got %s", cmds[0].Args["path"])
	}
}

func TestViewLoopParamsChange(t *testing.T) {
	view := &loopPatcherView{}
	view.Mount(Params{})
	view.Page = "home"

	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
		})
	}()

	// Simulate browser back/forward changing URL to /profile
	tr.PushEvent("gerbera:params_change", Payload{
		"_path": "/profile",
	})

	time.Sleep(50 * time.Millisecond)

	if view.Page != "profile" {
		t.Errorf("expected Page=profile, got %s", view.Page)
	}

	// Check that a message was sent
	tr.mu.Lock()
	msgCount := len(tr.Messages)
	tr.mu.Unlock()
	if msgCount == 0 {
		t.Error("expected at least one message to be sent")
	}

	tr.Close()
	wg.Wait()
}

func TestViewLoopParamsChangeNonPatcher(t *testing.T) {
	view := &loopTestView{}
	view.Mount(Params{})

	tr := NewTestTransport()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ViewLoop(view, tr, ViewLoopConfig{
			SessionID: "test-sess",
			Lang:      "en",
		})
	}()

	// Send params_change to a non-Patcher view — should be silently ignored
	tr.PushEvent("gerbera:params_change", Payload{
		"_path": "/profile",
	})

	time.Sleep(50 * time.Millisecond)

	// No message should be sent (the event is dropped for non-Patcher views)
	tr.mu.Lock()
	msgCount := len(tr.Messages)
	tr.mu.Unlock()
	if msgCount != 0 {
		t.Errorf("expected 0 messages for non-Patcher view, got %d", msgCount)
	}

	tr.Close()
	wg.Wait()
}
