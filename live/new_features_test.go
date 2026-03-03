package live

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)

// --- A4: Additional Event Bindings ---

func TestAdditionalEventBindings(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) g.ComponentFunc
		attr string
	}{
		{"DblClick", DblClick, "gerbera-dblclick"},
		{"MouseEnter", MouseEnter, "gerbera-mouseenter"},
		{"MouseLeave", MouseLeave, "gerbera-mouseleave"},
		{"TouchStart", TouchStart, "gerbera-touchstart"},
		{"TouchEnd", TouchEnd, "gerbera-touchend"},
		{"TouchMove", TouchMove, "gerbera-touchmove"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &g.Element{TagName: "div", Attr: make(g.AttrMap)}
			fn := tt.fn("test-event")
			if err := fn(el); err != nil {
				t.Fatal(err)
			}
			if el.Attr[tt.attr] != "test-event" {
				t.Errorf("expected %s=test-event, got %s", tt.attr, el.Attr[tt.attr])
			}
		})
	}
}

// --- A5: Debounce ---

func TestDebounce(t *testing.T) {
	el := &g.Element{TagName: "input", Attr: make(g.AttrMap)}
	fn := Debounce(300)
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["gerbera-debounce"] != "300" {
		t.Errorf("expected gerbera-debounce=300, got %s", el.Attr["gerbera-debounce"])
	}
}

// --- A3: Lifecycle Hooks ---

func TestHookBinding(t *testing.T) {
	el := &g.Element{TagName: "div", Attr: make(g.AttrMap)}
	fn := Hook("scroll-sync")
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["gerbera-hook"] != "scroll-sync" {
		t.Errorf("expected gerbera-hook=scroll-sync, got %s", el.Attr["gerbera-hook"])
	}
}

// --- B1: Live Navigation ---

func TestLiveLink(t *testing.T) {
	el := &g.Element{TagName: "a", Attr: make(g.AttrMap)}
	fn := LiveLink("/about")
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["href"] != "/about" {
		t.Errorf("expected href=/about, got %s", el.Attr["href"])
	}
	if el.Attr["gerbera-live-link"] != "/about" {
		t.Errorf("expected gerbera-live-link=/about, got %s", el.Attr["gerbera-live-link"])
	}
}

// --- B5: Conditional CSS Classes ---

func TestClassIf(t *testing.T) {
	el := &g.Element{TagName: "div"}
	fn := gp.ClassIf(true, "active")
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if _, ok := el.ClassNames["active"]; !ok {
		t.Error("expected 'active' class when condition is true")
	}

	el2 := &g.Element{TagName: "div"}
	fn2 := gp.ClassIf(false, "active")
	if err := fn2(el2); err != nil {
		t.Fatal(err)
	}
	if el2.ClassNames != nil {
		if _, ok := el2.ClassNames["active"]; ok {
			t.Error("expected no 'active' class when condition is false")
		}
	}
}

func TestClassMap(t *testing.T) {
	el := &g.Element{TagName: "div"}
	fn := gp.ClassMap(map[string]bool{
		"active":   true,
		"disabled": false,
		"primary":  true,
	})
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if _, ok := el.ClassNames["active"]; !ok {
		t.Error("expected 'active' class")
	}
	if _, ok := el.ClassNames["disabled"]; ok {
		t.Error("expected no 'disabled' class")
	}
	if _, ok := el.ClassNames["primary"]; !ok {
		t.Error("expected 'primary' class")
	}
}

// --- C4: ARIA Helpers ---

func TestAriaHelpers(t *testing.T) {
	tests := []struct {
		name  string
		fn    g.ComponentFunc
		key   string
		value string
	}{
		{"AriaLabel", gp.AriaLabel("Close"), "aria-label", "Close"},
		{"AriaDescribedBy", gp.AriaDescribedBy("desc"), "aria-describedby", "desc"},
		{"AriaLabelledBy", gp.AriaLabelledBy("title"), "aria-labelledby", "title"},
		{"Role", gp.Role("button"), "role", "button"},
		{"AriaHidden true", gp.AriaHidden(true), "aria-hidden", "true"},
		{"AriaHidden false", gp.AriaHidden(false), "aria-hidden", "false"},
		{"AriaExpanded true", gp.AriaExpanded(true), "aria-expanded", "true"},
		{"AriaLive polite", gp.AriaLive("polite"), "aria-live", "polite"},
		{"AriaControls", gp.AriaControls("menu"), "aria-controls", "menu"},
		{"AriaInvalid true", gp.AriaInvalid(true), "aria-invalid", "true"},
		{"AriaRequired true", gp.AriaRequired(true), "aria-required", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &g.Element{TagName: "div", Attr: make(g.AttrMap)}
			if err := tt.fn(el); err != nil {
				t.Fatal(err)
			}
			if el.Attr[tt.key] != tt.value {
				t.Errorf("expected %s=%s, got %s", tt.key, tt.value, el.Attr[tt.key])
			}
		})
	}
}

// --- Property Helpers ---

func TestPropertyHelpers(t *testing.T) {
	tests := []struct {
		name  string
		fn    g.ComponentFunc
		key   string
		value string
	}{
		{"Href", gp.Href("https://example.com"), "href", "https://example.com"},
		{"Src", gp.Src("/image.png"), "src", "/image.png"},
		{"Type", gp.Type("email"), "type", "email"},
		{"Placeholder", gp.Placeholder("Enter text"), "placeholder", "Enter text"},
		{"For", gp.For("email"), "for", "email"},
		{"DataAttr", gp.DataAttr("id", "123"), "data-id", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &g.Element{TagName: "input", Attr: make(g.AttrMap)}
			if err := tt.fn(el); err != nil {
				t.Fatal(err)
			}
			if el.Attr[tt.key] != tt.value {
				t.Errorf("expected %s=%s, got %s", tt.key, tt.value, el.Attr[tt.key])
			}
		})
	}
}

func TestDisabled(t *testing.T) {
	el := &g.Element{TagName: "button"}
	if err := gp.Disabled(true)(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["disabled"] != "disabled" {
		t.Error("expected disabled attribute when true")
	}

	el2 := &g.Element{TagName: "button"}
	if err := gp.Disabled(false)(el2); err != nil {
		t.Fatal(err)
	}
	if el2.Attr != nil && el2.Attr["disabled"] != "" {
		t.Error("expected no disabled attribute when false")
	}
}

// --- A2: JS Commands ---

func TestCommandQueue(t *testing.T) {
	q := &CommandQueue{}
	q.ScrollTo(".preview", "100", "0")
	q.Focus("#input")
	q.AddClass(".item", "active")
	q.Hide(".modal")

	cmds := q.DrainCommands()
	if len(cmds) != 4 {
		t.Errorf("expected 4 commands, got %d", len(cmds))
	}

	if cmds[0].Cmd != "scroll_to" || cmds[0].Target != ".preview" {
		t.Error("expected scroll_to command")
	}
	if cmds[1].Cmd != "focus" || cmds[1].Target != "#input" {
		t.Error("expected focus command")
	}
	if cmds[2].Cmd != "add_class" {
		t.Error("expected add_class command")
	}
	if cmds[3].Cmd != "hide" {
		t.Error("expected hide command")
	}

	// After drain, should be empty
	cmds2 := q.DrainCommands()
	if len(cmds2) != 0 {
		t.Error("expected empty queue after drain")
	}
}

// --- A1: TickerView ---

type tickerTestView struct {
	Count int
}

func (v *tickerTestView) Mount(params Params) error { return nil }
func (v *tickerTestView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Body(gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.Count)))),
	}
}
func (v *tickerTestView) HandleEvent(event string, payload Payload) error { return nil }
func (v *tickerTestView) TickInterval() time.Duration                     { return time.Second }
func (v *tickerTestView) HandleTick() error {
	v.Count++
	return nil
}

func TestTickerViewInterface(t *testing.T) {
	var _ TickerView = &tickerTestView{}
}

// --- C5: Test Utilities ---

func TestTestView(t *testing.T) {
	tv := NewTestView(&testView{})
	if err := tv.Mount(nil); err != nil {
		t.Fatal(err)
	}
	if !tv.Rendered {
		t.Error("expected Rendered to be true after Mount")
	}

	patches, err := tv.SimulateEvent("inc", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(patches) == 0 {
		t.Error("expected patches after increment")
	}

	html, err := tv.RenderHTML()
	if err != nil {
		t.Fatal(err)
	}
	if html == "" {
		t.Error("expected non-empty rendered HTML")
	}
}

func TestTestViewTick(t *testing.T) {
	tv := NewTestView(&tickerTestView{})
	if err := tv.Mount(nil); err != nil {
		t.Fatal(err)
	}

	patches, err := tv.SimulateTick()
	if err != nil {
		t.Fatal(err)
	}
	if len(patches) == 0 {
		t.Error("expected patches after tick")
	}
}

// --- B3 / B6: Form Builder ---

func TestField(t *testing.T) {
	el := &g.Element{TagName: "div", Children: make([]*g.Element, 0)}
	fn := Field("email", FieldOpts{
		Type:        "email",
		Label:       "Email",
		Placeholder: "Enter email",
		Value:       "test@example.com",
		Required:    true,
		Event:       "validate",
		Error:       "Invalid email",
	})

	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if len(el.Children) == 0 {
		t.Error("expected Field to produce children")
	}
}

func TestFieldWithoutError(t *testing.T) {
	el := &g.Element{TagName: "div", Children: make([]*g.Element, 0)}
	fn := Field("name", FieldOpts{
		Label: "Name",
		Value: "John",
	})

	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if len(el.Children) == 0 {
		t.Error("expected Field to produce children")
	}
}

func TestErrors(t *testing.T) {
	var errs Errors
	if errs.HasErrors() {
		t.Error("expected no errors initially")
	}

	errs.AddError("email", "required")
	if !errs.HasErrors() {
		t.Error("expected errors after AddError")
	}
	if !errs.HasError("email") {
		t.Error("expected email error")
	}
	if errs.ErrorFor("email") != "required" {
		t.Errorf("expected 'required', got '%s'", errs.ErrorFor("email"))
	}
	if errs.HasError("name") {
		t.Error("expected no error for name")
	}

	errs.Clear()
	if errs.HasErrors() {
		t.Error("expected no errors after Clear")
	}
}

// --- A6: Upload Binding ---

func TestUploadBinding(t *testing.T) {
	el := &g.Element{TagName: "input", Attr: make(g.AttrMap)}
	fn := Upload("upload-avatar")
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["gerbera-upload"] != "upload-avatar" {
		t.Errorf("expected gerbera-upload=upload-avatar, got %s", el.Attr["gerbera-upload"])
	}
}

// --- C6: Middleware ---

func TestWithMiddleware(t *testing.T) {
	called := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	h := Handler(func(_ *http.Request) View { return &testView{} }, WithMiddleware(mw))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !called {
		t.Error("expected middleware to be called")
	}
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Result().StatusCode)
	}
}

// --- C3: Component ---

func TestComponent(t *testing.T) {
	el := &g.Element{TagName: "div", Children: make([]*g.Element, 0)}
	fn := Component("chat", "/chat")
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if len(el.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(el.Children))
	}
	child := el.Children[0]
	if child.Attr["gerbera-component"] != "/chat" {
		t.Error("expected gerbera-component=/chat")
	}
	if child.Attr["gerbera-component-id"] != "chat" {
		t.Error("expected gerbera-component-id=chat")
	}
}
