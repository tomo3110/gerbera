package live

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/ui"
)

func render(t *testing.T, c gerbera.ComponentFunc) string {
	t.Helper()
	var buf bytes.Buffer
	if err := gerbera.ExecuteTemplate(&buf, "en", c); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestModalClosed(t *testing.T) {
	out := render(t, Modal(false, "close"))
	if strings.Contains(out, "g-modal-overlay") {
		t.Error("Closed modal should not render overlay")
	}
}

func TestModalOpen(t *testing.T) {
	out := render(t, Modal(true, "close",
		ModalHeader("Title", "close"),
		ModalBody(gerbera.Literal("Content")),
		ModalFooter(gerbera.Literal("Footer")),
	))
	if !strings.Contains(out, "g-modal-overlay") {
		t.Error("Open modal should render overlay")
	}
	if !strings.Contains(out, "Title") {
		t.Error("Modal should contain header title")
	}
	if !strings.Contains(out, "Content") {
		t.Error("Modal should contain body content")
	}
	if !strings.Contains(out, `role="dialog"`) {
		t.Error("Modal should have dialog role")
	}
}

func TestToast(t *testing.T) {
	for _, v := range []string{"info", "success", "warning", "danger"} {
		t.Run(v, func(t *testing.T) {
			out := render(t, Toast("Hello", v, "dismiss"))
			if !strings.Contains(out, "g-toast-"+v) {
				t.Errorf("Toast(%q) should contain class g-toast-%s", v, v)
			}
			if !strings.Contains(out, `gerbera-click="dismiss"`) {
				t.Error("Toast dismiss button should have click event")
			}
		})
	}
}

func TestDataTable(t *testing.T) {
	out := render(t, DataTable(DataTableOpts{
		Columns: []Column{
			{Key: "name", Label: "Name", Sortable: true},
			{Key: "email", Label: "Email", Sortable: false},
		},
		Rows: [][]string{
			{"Alice", "alice@example.com"},
			{"Bob", "bob@example.com"},
		},
		SortCol:   "name",
		SortDir:   "asc",
		SortEvent: "sort",
		Page:      0,
		PageSize:  10,
		Total:     2,
		PageEvent: "page",
	}))
	if !strings.Contains(out, "g-table") {
		t.Error("DataTable should contain g-table class")
	}
	if !strings.Contains(out, "Alice") {
		t.Error("DataTable should contain row data")
	}
	if !strings.Contains(out, `gerbera-click="sort"`) {
		t.Error("Sortable column should have sort event")
	}
}

func TestDropdownClosed(t *testing.T) {
	out := render(t, Dropdown(false, "toggle",
		gerbera.Literal("Open"),
		gerbera.Literal("Menu"),
	))
	if strings.Contains(out, "g-dropdown-menu") {
		t.Error("Closed dropdown should not show menu")
	}
	if !strings.Contains(out, `aria-expanded="false"`) {
		t.Error("Closed dropdown should have aria-expanded=false")
	}
}

func TestDropdownOpen(t *testing.T) {
	out := render(t, Dropdown(true, "toggle",
		gerbera.Literal("Open"),
		gerbera.Literal("Menu"),
	))
	if !strings.Contains(out, "g-dropdown-menu") {
		t.Error("Open dropdown should show menu")
	}
	if !strings.Contains(out, `aria-expanded="true"`) {
		t.Error("Open dropdown should have aria-expanded=true")
	}
}

func TestConfirm(t *testing.T) {
	out := render(t, Confirm(true, "Delete?", "Are you sure?", "confirmDelete", "cancel"))
	if !strings.Contains(out, "g-modal-overlay") {
		t.Error("Confirm should render modal overlay")
	}
	if !strings.Contains(out, "Delete?") {
		t.Error("Confirm should show title")
	}
	if !strings.Contains(out, "Are you sure?") {
		t.Error("Confirm should show message")
	}
	if !strings.Contains(out, `gerbera-click="confirmDelete"`) {
		t.Error("Confirm should have confirm event")
	}
	if !strings.Contains(out, `gerbera-click="cancel"`) {
		t.Error("Confirm should have cancel event")
	}
}

func TestTabsRender(t *testing.T) {
	out := render(t, Tabs("my-tabs", 1, []Tab{
		{Label: "First", Content: gerbera.Literal("Content 1")},
		{Label: "Second", Content: gerbera.Literal("Content 2")},
		{Label: "Third", Content: gerbera.Literal("Content 3")},
	}, "switchTab"))

	if !strings.Contains(out, "g-tabs") {
		t.Error("Tabs should have g-tabs class")
	}
	if !strings.Contains(out, `role="tablist"`) {
		t.Error("Tabs should have tablist role")
	}
	if !strings.Contains(out, `role="tab"`) {
		t.Error("Tab buttons should have tab role")
	}
	if !strings.Contains(out, `role="tabpanel"`) {
		t.Error("Tab panels should have tabpanel role")
	}
	if !strings.Contains(out, `gerbera-click="switchTab"`) {
		t.Error("Tab buttons should have click event")
	}
	if !strings.Contains(out, `gerbera-value="0"`) {
		t.Error("First tab should have value 0")
	}
	if !strings.Contains(out, `gerbera-value="2"`) {
		t.Error("Third tab should have value 2")
	}
}

func TestTabsActivePanel(t *testing.T) {
	out := render(t, Tabs("t", 0, []Tab{
		{Label: "A", Content: gerbera.Literal("Panel A")},
		{Label: "B", Content: gerbera.Literal("Panel B")},
	}, "change"))

	// Active tab (index 0) should have aria-selected=true
	if !strings.Contains(out, `aria-selected="true"`) {
		t.Error("Active tab should have aria-selected=true")
	}
	// Inactive panel should be hidden
	if !strings.Contains(out, `hidden="hidden"`) {
		t.Error("Inactive panel should have hidden attribute")
	}
	// Active tab should have active class
	if !strings.Contains(out, "g-tab-active") {
		t.Error("Active tab should have g-tab-active class")
	}
}

func TestTabsAriaLinkage(t *testing.T) {
	out := render(t, Tabs("demo", 0, []Tab{
		{Label: "One", Content: gerbera.Literal("C1")},
	}, "ev"))

	if !strings.Contains(out, `id="demo-tab-0"`) {
		t.Error("Tab button should have correct ID")
	}
	if !strings.Contains(out, `id="demo-panel-0"`) {
		t.Error("Tab panel should have correct ID")
	}
	if !strings.Contains(out, `aria-controls="demo-panel-0"`) {
		t.Error("Tab should have aria-controls pointing to panel")
	}
	if !strings.Contains(out, `aria-labelledby="demo-tab-0"`) {
		t.Error("Panel should have aria-labelledby pointing to tab")
	}
}

func TestDrawerClosed(t *testing.T) {
	out := render(t, Drawer(false, "close", "left"))
	if strings.Contains(out, "g-drawer") {
		t.Error("Closed drawer should not render")
	}
}

func TestDrawerOpenLeft(t *testing.T) {
	out := render(t, Drawer(true, "close", "left",
		DrawerHeader("Menu", "close"),
		DrawerBody(gerbera.Literal("Content")),
	))
	if !strings.Contains(out, "g-drawer-overlay") {
		t.Error("Open drawer should render overlay")
	}
	if !strings.Contains(out, "g-drawer-left") {
		t.Error("Left drawer should have g-drawer-left class")
	}
	if !strings.Contains(out, "Menu") {
		t.Error("Drawer should contain header title")
	}
	if !strings.Contains(out, "Content") {
		t.Error("Drawer should contain body content")
	}
	if !strings.Contains(out, `role="dialog"`) {
		t.Error("Drawer should have dialog role")
	}
}

func TestDrawerOpenRight(t *testing.T) {
	out := render(t, Drawer(true, "close", "right"))
	if !strings.Contains(out, "g-drawer-right") {
		t.Error("Right drawer should have g-drawer-right class")
	}
}

func TestSearchSelectClosed(t *testing.T) {
	out := render(t, SearchSelect(SearchSelectOpts{
		Name:        "country",
		Query:       "",
		Options:     nil,
		Open:        false,
		InputEvent:  "searchCountry",
		SelectEvent: "selectCountry",
	}))
	if !strings.Contains(out, "g-searchselect") {
		t.Error("SearchSelect should have g-searchselect class")
	}
	if strings.Contains(out, "g-searchselect-list") {
		t.Error("Closed SearchSelect should not show list")
	}
	if !strings.Contains(out, `role="combobox"`) {
		t.Error("SearchSelect should have combobox role")
	}
}

func TestSearchSelectOpen(t *testing.T) {
	out := render(t, SearchSelect(SearchSelectOpts{
		Name:  "country",
		Query: "Jap",
		Options: []ui.FormOption{
			{Value: "jp", Label: "Japan"},
		},
		Selected:    "jp",
		Open:        true,
		InputEvent:  "search",
		SelectEvent: "select",
	}))
	if !strings.Contains(out, "g-searchselect-list") {
		t.Error("Open SearchSelect should show list")
	}
	if !strings.Contains(out, "Japan") {
		t.Error("SearchSelect should show matching options")
	}
	if !strings.Contains(out, "g-searchselect-option-active") {
		t.Error("Selected option should have active class")
	}
	if !strings.Contains(out, `gerbera-click="select"`) {
		t.Error("Options should have select event")
	}
}

func TestSearchSelectEmpty(t *testing.T) {
	out := render(t, SearchSelect(SearchSelectOpts{
		Name:        "x",
		Query:       "zzz",
		Options:     nil,
		Open:        true,
		InputEvent:  "s",
		SelectEvent: "sel",
	}))
	if !strings.Contains(out, "No matches") {
		t.Error("Empty SearchSelect should show no-matches message")
	}
}

func TestModalHeader(t *testing.T) {
	out := render(t, ModalHeader("Test", "close"))
	if !strings.Contains(out, "g-modal-header") {
		t.Error("ModalHeader should have g-modal-header class")
	}
	if !strings.Contains(out, `aria-label="Close"`) {
		t.Error("Close button should have aria-label")
	}
	_ = property.Class // ensure import is used
}

// --- Live NumberInput tests ---

func TestLiveNumberInput(t *testing.T) {
	out := render(t, NumberInput(NumberInputOpts{
		Name:           "qty",
		Value:          5,
		IncrementEvent: "inc",
		DecrementEvent: "dec",
		ChangeEvent:    "change",
	}))
	if !strings.Contains(out, "g-numberinput") {
		t.Error("Live NumberInput should have g-numberinput class")
	}
	if !strings.Contains(out, `role="spinbutton"`) {
		t.Error("Live NumberInput should have spinbutton role")
	}
	if !strings.Contains(out, `gerbera-click="inc"`) {
		t.Error("Increment button should have click event")
	}
	if !strings.Contains(out, `gerbera-click="dec"`) {
		t.Error("Decrement button should have click event")
	}
	if !strings.Contains(out, `gerbera-change="change"`) {
		t.Error("Input should have change event")
	}
	if !strings.Contains(out, `value="5"`) {
		t.Error("Input should have value")
	}
}

func TestLiveNumberInputMinMax(t *testing.T) {
	min, max := 0, 10
	out := render(t, NumberInput(NumberInputOpts{
		Name:  "qty",
		Value: 5,
		Min:   &min,
		Max:   &max,
	}))
	if !strings.Contains(out, `min="0"`) {
		t.Error("Live NumberInput should have min attribute")
	}
	if !strings.Contains(out, `max="10"`) {
		t.Error("Live NumberInput should have max attribute")
	}
	if !strings.Contains(out, `aria-valuemin="0"`) {
		t.Error("Live NumberInput should have aria-valuemin")
	}
	if !strings.Contains(out, `aria-valuemax="10"`) {
		t.Error("Live NumberInput should have aria-valuemax")
	}
}

// --- Live Slider tests ---

func TestLiveSlider(t *testing.T) {
	out := render(t, Slider(SliderOpts{
		Name:       "volume",
		Value:      75,
		Min:        0,
		Max:        100,
		Label:      "Volume",
		InputEvent: "slideVolume",
	}))
	if !strings.Contains(out, "g-slider") {
		t.Error("Live Slider should have g-slider class")
	}
	if !strings.Contains(out, `role="slider"`) {
		t.Error("Live Slider should have slider role")
	}
	if !strings.Contains(out, `gerbera-input="slideVolume"`) {
		t.Error("Slider should have input event")
	}
	if !strings.Contains(out, "Volume") {
		t.Error("Slider should display label")
	}
	if !strings.Contains(out, "75") {
		t.Error("Slider should display current value")
	}
}

// --- Live Calendar tests ---

func TestLiveCalendar(t *testing.T) {
	out := render(t, Calendar(CalendarOpts{
		Year:           2025,
		Month:          time.January,
		Today:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		SelectEvent:    "selectDay",
		PrevMonthEvent: "prevMonth",
		NextMonthEvent: "nextMonth",
	}))
	if !strings.Contains(out, "g-calendar") {
		t.Error("Live Calendar should have g-calendar class")
	}
	if !strings.Contains(out, `role="grid"`) {
		t.Error("Live Calendar should have grid role")
	}
	if !strings.Contains(out, "January 2025") {
		t.Error("Live Calendar should display month and year")
	}
	if !strings.Contains(out, `gerbera-click="selectDay"`) {
		t.Error("Day cells should have select event")
	}
	if !strings.Contains(out, `gerbera-click="prevMonth"`) {
		t.Error("Prev button should have click event")
	}
	if !strings.Contains(out, `gerbera-click="nextMonth"`) {
		t.Error("Next button should have click event")
	}
}

func TestLiveCalendarSelected(t *testing.T) {
	sel := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	out := render(t, Calendar(CalendarOpts{
		Year:        2025,
		Month:       time.January,
		Selected:    &sel,
		Today:       time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		SelectEvent: "select",
	}))
	if !strings.Contains(out, "g-calendar-day-selected") {
		t.Error("Live Calendar should highlight selected date")
	}
	if !strings.Contains(out, "g-calendar-day-today") {
		t.Error("Live Calendar should highlight today")
	}
}

func TestLiveCalendarNavButtons(t *testing.T) {
	out := render(t, Calendar(CalendarOpts{
		Year:           2025,
		Month:          time.March,
		Today:          time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		PrevMonthEvent: "prev",
		NextMonthEvent: "next",
	}))
	if !strings.Contains(out, `aria-label="Previous month"`) {
		t.Error("Prev button should have aria-label")
	}
	if !strings.Contains(out, `aria-label="Next month"`) {
		t.Error("Next button should have aria-label")
	}
}

// --- Live ChatInput tests ---

func TestLiveChatInput(t *testing.T) {
	out := render(t, ChatInput(ChatInputOpts{
		Name:         "msg",
		Value:        "hello",
		SendEvent:    "sendMsg",
		InputEvent:   "typeMsg",
		KeydownEvent: "keyMsg",
	}))
	if !strings.Contains(out, "g-chat-inputbar") {
		t.Error("Live ChatInput should have inputbar class")
	}
	if !strings.Contains(out, `gerbera-click="sendMsg"`) {
		t.Error("Send button should have click event")
	}
	if !strings.Contains(out, `gerbera-input="typeMsg"`) {
		t.Error("Input should have input event")
	}
	if !strings.Contains(out, `gerbera-keydown="keyMsg"`) {
		t.Error("Input should have keydown event")
	}
	if !strings.Contains(out, `value="hello"`) {
		t.Error("Input should have value")
	}
}

func TestLiveChatInputPlaceholder(t *testing.T) {
	out := render(t, ChatInput(ChatInputOpts{
		Name:        "msg",
		Placeholder: "Say something...",
	}))
	if !strings.Contains(out, "Say something...") {
		t.Error("ChatInput should use custom placeholder")
	}
}

func TestLiveChatInputDefaultPlaceholder(t *testing.T) {
	out := render(t, ChatInput(ChatInputOpts{Name: "msg"}))
	if !strings.Contains(out, "Type a message...") {
		t.Error("ChatInput should have default placeholder")
	}
}

// --- Live Pagination tests ---

func TestLivePagination(t *testing.T) {
	out := render(t, Pagination(PaginationOpts{
		Page:      2,
		PageSize:  10,
		Total:     100,
		PageEvent: "paginateTo",
	}))
	if !strings.Contains(out, "g-pagination") {
		t.Error("Live Pagination should have g-pagination class")
	}
	if !strings.Contains(out, `role="navigation"`) {
		t.Error("Live Pagination should have navigation role")
	}
	if !strings.Contains(out, `gerbera-click="paginateTo"`) {
		t.Error("Page buttons should have click event")
	}
	if !strings.Contains(out, `aria-current="page"`) {
		t.Error("Active page should have aria-current")
	}
}

func TestLivePaginationFirstPage(t *testing.T) {
	out := render(t, Pagination(PaginationOpts{
		Page:      0,
		PageSize:  10,
		Total:     50,
		PageEvent: "page",
	}))
	// Prev button should be disabled on first page
	if !strings.Contains(out, "g-pagination-prev") {
		t.Error("Should have prev button")
	}
}

// --- Live ButtonGroup tests ---

func TestLiveButtonGroup(t *testing.T) {
	out := render(t, ButtonGroup(ButtonGroupOpts{
		Items: []ui.ButtonGroupItem{
			{Label: "Day", Value: "day", Active: true},
			{Label: "Week", Value: "week"},
		},
		ChangeEvent: "btnGroupChange",
	}))
	if !strings.Contains(out, "g-btngroup") {
		t.Error("Live ButtonGroup should have g-btngroup class")
	}
	if !strings.Contains(out, `role="group"`) {
		t.Error("Live ButtonGroup should have group role")
	}
	if !strings.Contains(out, `gerbera-click="btnGroupChange"`) {
		t.Error("Buttons should have click event")
	}
	if !strings.Contains(out, `aria-pressed="true"`) {
		t.Error("Active button should have aria-pressed=true")
	}
	if !strings.Contains(out, `gerbera-value="week"`) {
		t.Error("Buttons should have click value")
	}
}

// --- Live Accordion tests ---

func TestLiveAccordion(t *testing.T) {
	out := render(t, Accordion(AccordionOpts{
		Items: []ui.AccordionItem{
			{Title: "Section 1", Content: gerbera.Literal("Content 1"), Open: true},
			{Title: "Section 2", Content: gerbera.Literal("Content 2"), Open: false},
		},
		ToggleEvent: "accordionToggle",
	}))
	if !strings.Contains(out, "g-accordion") {
		t.Error("Live Accordion should have g-accordion class")
	}
	if !strings.Contains(out, `gerbera-click="accordionToggle"`) {
		t.Error("Headers should have click event")
	}
	if !strings.Contains(out, `aria-expanded="true"`) {
		t.Error("Open item should have aria-expanded=true")
	}
	if !strings.Contains(out, `aria-expanded="false"`) {
		t.Error("Closed item should have aria-expanded=false")
	}
	if !strings.Contains(out, "Content 1") {
		t.Error("Open item should show content")
	}
	if strings.Contains(out, "Content 2") {
		t.Error("Closed item should not show content")
	}
	if !strings.Contains(out, `role="region"`) {
		t.Error("Body should have region role")
	}
}

func TestLiveAccordionAriaLinkage(t *testing.T) {
	out := render(t, Accordion(AccordionOpts{
		Items: []ui.AccordionItem{
			{Title: "First", Content: gerbera.Literal("Body"), Open: true},
		},
		ToggleEvent: "toggle",
	}))
	if !strings.Contains(out, `id="g-accordion-header-0"`) {
		t.Error("Header should have ID")
	}
	if !strings.Contains(out, `id="g-accordion-body-0"`) {
		t.Error("Body should have ID")
	}
	if !strings.Contains(out, `aria-controls="g-accordion-body-0"`) {
		t.Error("Header should have aria-controls")
	}
	if !strings.Contains(out, `aria-labelledby="g-accordion-header-0"`) {
		t.Error("Body should have aria-labelledby")
	}
}

// --- Live Stepper tests ---

func TestLiveStepper(t *testing.T) {
	out := render(t, Stepper(StepperOpts{
		Steps: []ui.Step{
			{Label: "Cart", Status: ui.StepCompleted},
			{Label: "Shipping", Status: ui.StepActive, Description: "Enter address"},
			{Label: "Payment", Status: ui.StepUpcoming},
		},
		ClickEvent: "stepperClick",
	}))
	if !strings.Contains(out, "g-stepper") {
		t.Error("Live Stepper should have g-stepper class")
	}
	if !strings.Contains(out, `role="list"`) {
		t.Error("Live Stepper should have list role")
	}
	if !strings.Contains(out, "g-stepper-completed") {
		t.Error("Completed step should have class")
	}
	if !strings.Contains(out, `gerbera-click="stepperClick"`) {
		t.Error("Completed step should have click event")
	}
	if !strings.Contains(out, `aria-current="step"`) {
		t.Error("Active step should have aria-current")
	}
}

func TestLiveStepperVertical(t *testing.T) {
	out := render(t, Stepper(StepperOpts{
		Steps: []ui.Step{
			{Label: "A", Status: ui.StepCompleted},
			{Label: "B", Status: ui.StepActive},
		},
		Vertical: true,
	}))
	if !strings.Contains(out, "g-stepper-vertical") {
		t.Error("Vertical stepper should have vertical class")
	}
}

// --- Live InfiniteScroll tests ---

func TestLiveInfiniteScroll(t *testing.T) {
	out := render(t, InfiniteScroll(InfiniteScrollOpts{
		View:          ui.InfiniteScrollList,
		Loading:       true,
		ShowToggle:    true,
		LoadMoreEvent: "loadMore",
		ToggleEvent:   "toggleView",
	}, gerbera.Literal("Item")))
	if !strings.Contains(out, "g-infinitescroll") {
		t.Error("Live InfiniteScroll should have g-infinitescroll class")
	}
	if !strings.Contains(out, "g-infinitescroll-toolbar") {
		t.Error("Should show toolbar when ShowToggle is true")
	}
	if !strings.Contains(out, `gerbera-scroll="loadMore"`) {
		t.Error("Content should have scroll event")
	}
	if !strings.Contains(out, `gerbera-click="toggleView"`) {
		t.Error("Toggle buttons should have click event")
	}
	if !strings.Contains(out, "g-infinitescroll-loader") {
		t.Error("Loading should show loader")
	}
	if !strings.Contains(out, "g-spinner") {
		t.Error("Loader should contain spinner")
	}
}

func TestLiveInfiniteScrollGrid(t *testing.T) {
	out := render(t, InfiniteScroll(InfiniteScrollOpts{
		View: ui.InfiniteScrollGrid,
	}))
	if !strings.Contains(out, "g-infinitescroll-grid") {
		t.Error("Grid view should have grid class")
	}
}

// --- Live TimePicker tests ---

func TestLiveTimePicker(t *testing.T) {
	out := render(t, TimePicker(TimePickerOpts{
		Name:        "alarm",
		Hour:        14,
		Minute:      30,
		Use24H:      true,
		ChangeEvent: "timeChange",
	}))
	if !strings.Contains(out, "g-timepicker") {
		t.Error("Live TimePicker should have g-timepicker class")
	}
	if !strings.Contains(out, `role="group"`) {
		t.Error("Live TimePicker should have group role")
	}
	if !strings.Contains(out, `aria-label="alarm"`) {
		t.Error("Live TimePicker should have aria-label")
	}
	if !strings.Contains(out, `value="14"`) {
		t.Error("Live TimePicker should display hour")
	}
	if !strings.Contains(out, `value="30"`) {
		t.Error("Live TimePicker should display minute")
	}
	if !strings.Contains(out, `gerbera-click="timeChange"`) {
		t.Error("Buttons should have click event")
	}
	if !strings.Contains(out, `gerbera-change="timeChange"`) {
		t.Error("Inputs should have change event")
	}
}

func TestLiveTimePickerShowSec(t *testing.T) {
	out := render(t, TimePicker(TimePickerOpts{
		Name:        "time",
		Hour:        10,
		Minute:      20,
		Second:      45,
		Use24H:      true,
		ShowSec:     true,
		ChangeEvent: "timeChange",
	}))
	if !strings.Contains(out, `value="45"`) {
		t.Error("Live TimePicker with ShowSec should display seconds")
	}
	if !strings.Contains(out, `aria-label="Second"`) {
		t.Error("Second unit should have aria-label")
	}
}

func TestLiveTimePicker12Hour(t *testing.T) {
	out := render(t, TimePicker(TimePickerOpts{
		Name:        "time",
		Hour:        15,
		Minute:      30,
		Use24H:      false,
		ChangeEvent: "timeChange",
	}))
	if !strings.Contains(out, `value="03"`) {
		t.Error("12-hour Live TimePicker should convert 15:00 to 03")
	}
	if !strings.Contains(out, "g-timepicker-ampm") {
		t.Error("12-hour Live TimePicker should have AM/PM toggle")
	}
	if !strings.Contains(out, `aria-pressed="true"`) {
		t.Error("Active AM/PM button should have aria-pressed=true")
	}
}

func TestLiveTimePickerDataValue(t *testing.T) {
	out := render(t, TimePicker(TimePickerOpts{
		Name:   "time",
		Hour:   14,
		Minute: 30,
		Use24H: true,
	}))
	if !strings.Contains(out, `data-value="14:30"`) {
		t.Error("Live TimePicker should have data-value with formatted time")
	}
}

func TestLiveTimePickerDataValueWithSec(t *testing.T) {
	out := render(t, TimePicker(TimePickerOpts{
		Name:    "time",
		Hour:    14,
		Minute:  30,
		Second:  15,
		Use24H:  true,
		ShowSec: true,
	}))
	if !strings.Contains(out, `data-value="14:30:15"`) {
		t.Error("Live TimePicker with ShowSec should have data-value with seconds")
	}
}

// Keep unused imports referenced
var _ = ui.FormOption{}
var _ = time.UTC
