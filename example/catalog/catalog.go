package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

type CatalogView struct {
	Page         string
	ModalOpen    bool
	ConfirmOpen  bool
	ToastVisible bool
	ToastMessage string
	ToastVariant string
	DropdownOpen  bool
	DropdownValue string
	TablePage    int
	SortCol      string
	SortDir      string
	TabIndex     int

	// Drawer
	DrawerOpen bool
	DrawerSide string

	// SearchSelect
	SSQuery     string
	SSOpen      bool
	SSValue     string
	SSHighlight int

	// Form validation demo
	FormName      string
	FormEmail     string
	FormSubmitted bool

	// Tree
	TreeOpen map[string]bool

	// Checkbox/Radio demo
	CheckA  bool
	CheckB  bool
	RadioV  string

	// Mobile drawer for sidebar
	MobileNavOpen bool

	// Theme demo
	ThemeMode string // "light", "dark", "custom", "auto"

	// NumberInput demo
	NumVal int

	// Slider demo
	SliderVal int

	// Calendar demo
	CalYear     int
	CalMonth    time.Month
	CalSelected *time.Time

	// Chat demo
	ChatMessages []gu.ChatMessage
	ChatDraft    string

	// Pagination demo
	PaginationPage int

	// InfiniteScroll demo
	InfScrollItems   int
	InfScrollView    gu.InfiniteScrollView
	InfScrollLoading bool

	// ButtonGroup demo
	BtnGroupValue string

	// Accordion demo
	AccordionOpen [3]bool

	// Stepper demo
	StepperCurrent int

	// TimePicker demo
	TimeHour   int
	TimeMinute int
	TimeSecond int

	// Chart demo
	ChartType      string
	ChartHoverInfo string
	ChartClickInfo string

	// Avatar demo
	AvatarClickInfo string
}

func (v *CatalogView) Mount(params gl.Params) error {
	v.Page = "overview"
	v.SortCol = "name"
	v.SortDir = "asc"
	v.DrawerSide = "left"
	v.RadioV = "opt1"
	v.ThemeMode = "light"
	v.SSHighlight = -1
	v.TreeOpen = map[string]bool{"src": true}
	v.NumVal = 5
	v.SliderVal = 50
	now := time.Now()
	v.CalYear = now.Year()
	v.CalMonth = now.Month()
	v.InfScrollItems = 10
	v.InfScrollView = gu.InfiniteScrollList
	v.BtnGroupValue = "day"
	v.AccordionOpen = [3]bool{true, false, false}
	v.StepperCurrent = 1
	v.TimeHour = 14
	v.TimeMinute = 30
	v.TimeSecond = 0
	v.ChartType = "line"
	v.ChatMessages = []gu.ChatMessage{
		{Author: "Alice", Content: "Hello! Welcome to the chat demo.", Timestamp: "10:00", Sent: false, Avatar: "A"},
		{Content: "Thanks! This looks great.", Timestamp: "10:01", Sent: true},
		{Author: "Alice", Content: "Try sending a message below.", Timestamp: "10:02", Sent: false, Avatar: "A"},
	}
	if p := params["page"]; p != "" {
		v.Page = p
	}
	return nil
}

func (v *CatalogView) currentTheme() g.ComponentFunc {
	switch v.ThemeMode {
	case "dark":
		return gu.ThemeWith(gu.DarkTheme())
	case "custom":
		return gu.ThemeWith(gu.ThemeConfig{
			Accent:      "#1a73e8",
			AccentLight: "#e8f0fe",
			AccentHover: "#1557b0",
		})
	case "auto":
		return gu.ThemeAuto(gu.ThemeConfig{}, gu.ThemeConfig{})
	default:
		return gu.Theme()
	}
}

func (v *CatalogView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Gerbera UI Catalog"),
			gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
			v.currentTheme(),
		),
		gd.Body(
			gu.MobileHeader("UI Catalog", gl.Click("toggleMobileNav")),
			gu.AdminShell(
				v.renderSidebar(),
				v.renderContent(),
			),
			// Mobile nav drawer
			gul.Drawer(v.MobileNavOpen, "closeMobileNav", "left",
				gul.DrawerHeader("UI Catalog", "closeMobileNav"),
				gul.DrawerBody(v.renderNavLinks()),
			),
		),
	}
}

func (v *CatalogView) renderNavLinks() g.ComponentFunc {
	return gd.Div(
		navLink(v, "overview", "Overview"),
		navDividerLabel("Static Widgets"),
		navLink(v, "card", "Card"),
		navLink(v, "button", "Button"),
		navLink(v, "icon", "Icon"),
		navLink(v, "badge", "Badge & Alert"),
		navLink(v, "table", "Table"),
		navLink(v, "form", "Form"),
		navLink(v, "formadvanced", "Form (Advanced)"),
		navLink(v, "stat", "StatCard"),
		navLink(v, "nav", "Navigation"),
		navLink(v, "tree", "TreeView"),
		navLink(v, "spinner", "Spinner"),
		navLink(v, "misc", "Misc"),
		navLink(v, "layout", "Layout"),
		navDividerLabel("Live Widgets"),
		navLink(v, "pagination", "Pagination"),
		navLink(v, "buttongroup", "ButtonGroup"),
		navLink(v, "accordion", "Accordion"),
		navLink(v, "stepper", "Stepper"),
		navLink(v, "infinitescroll", "InfiniteScroll"),
		navLink(v, "numberinput", "NumberInput"),
		navLink(v, "slider", "Slider"),
		navLink(v, "timepicker", "TimePicker"),
		navLink(v, "calendar", "Calendar"),
		navLink(v, "chat", "Chat"),
		navLink(v, "modal", "Modal"),
		navLink(v, "toast", "Toast"),
		navLink(v, "datatable", "DataTable"),
		navLink(v, "dropdown", "Dropdown"),
		navLink(v, "tabs", "Tabs"),
		navLink(v, "drawer", "Drawer"),
		navLink(v, "searchselect", "SearchSelect"),
		navLink(v, "confirm", "Confirm"),
		navDividerLabel("Data Visualization"),
		navLink(v, "chart", "Charts"),
		navLink(v, "avatar", "Avatar"),
		navDividerLabel("Theming"),
		navLink(v, "theme", "Theme"),
	)
}

func navLink(v *CatalogView, page, label string) g.ComponentFunc {
	return gu.SidebarLink("#", label, v.Page == page,
		gl.Click("nav"), gl.ClickValue(page))
}

func navDividerLabel(label string) g.ComponentFunc {
	return gd.Div(
		gu.SidebarDivider(),
		gd.Div(gp.Attr("style", "padding:4px 24px;font-size:11px;font-weight:600;color:var(--g-text-tertiary);text-transform:uppercase;letter-spacing:0.04em"), gp.Value(label)),
	)
}

func (v *CatalogView) renderSidebar() g.ComponentFunc {
	return gu.Sidebar(
		gu.SidebarHeader("UI Catalog"),
		v.renderNavLinks(),
	)
}

func (v *CatalogView) renderContent() g.ComponentFunc {
	var body g.ComponentFunc
	switch v.Page {
	case "card":
		body = v.pageCard()
	case "button":
		body = v.pageButton()
	case "icon":
		body = v.pageIcon()
	case "badge":
		body = v.pageBadge()
	case "table":
		body = v.pageTable()
	case "form":
		body = v.pageForm()
	case "formadvanced":
		body = v.pageFormAdvanced()
	case "stat":
		body = v.pageStat()
	case "nav":
		body = v.pageNav()
	case "tree":
		body = v.pageTree()
	case "spinner":
		body = v.pageSpinner()
	case "misc":
		body = v.pageMisc()
	case "layout":
		body = v.pageLayout()
	case "pagination":
		body = v.pagePagination()
	case "buttongroup":
		body = v.pageButtonGroup()
	case "accordion":
		body = v.pageAccordion()
	case "stepper":
		body = v.pageStepper()
	case "infinitescroll":
		body = v.pageInfiniteScroll()
	case "numberinput":
		body = v.pageNumberInput()
	case "slider":
		body = v.pageSlider()
	case "timepicker":
		body = v.pageTimePicker()
	case "calendar":
		body = v.pageCalendar()
	case "chat":
		body = v.pageChat()
	case "modal":
		body = v.pageModal()
	case "toast":
		body = v.pageToast()
	case "datatable":
		body = v.pageDataTable()
	case "dropdown":
		body = v.pageDropdown()
	case "tabs":
		body = v.pageTabs()
	case "drawer":
		body = v.pageDrawer()
	case "searchselect":
		body = v.pageSearchSelect()
	case "confirm":
		body = v.pageConfirm()
	case "chart":
		body = v.pageChart()
	case "avatar":
		body = v.pageAvatar()
	case "theme":
		body = v.pageTheme()
	default:
		body = v.pageOverview()
	}
	return gd.Div(
		gu.PageHeader("Gerbera UI Catalog"),
		gd.Div(gp.Class("g-page-body"), body),
	)
}

func (v *CatalogView) pageOverview() g.ComponentFunc {
	return gu.Stack(
		gu.Alert("Welcome to the Gerbera UI Widget Catalog. Use the sidebar to browse components. Responsive: resize your browser to see mobile layout.", "info"),
		gu.Grid(gu.GridCols3,
			gu.StatCard("Static Widgets", "20"),
			gu.StatCard("Live Widgets", "14"),
			gu.StatCard("Icons", fmt.Sprintf("%d", len(gu.IconNames()))),
		),
	)
}

func (v *CatalogView) pageCard() g.ComponentFunc {
	return gd.Div(
		section("Card", "Container with optional header and footer."),
		gu.Card(
			gu.CardHeader("Card Title", gu.Button("Action", gu.ButtonSmall, gu.ButtonOutline)),
			gd.Div(gp.Class("g-page-body"),
				gd.P(gp.Value("Card content goes here.")),
			),
			gu.CardFooter(gd.Span(gp.Value("Footer text"))),
		),
	)
}

func (v *CatalogView) pageButton() g.ComponentFunc {
	return gu.Stack(
		section("Button", "Button variants."),
		gu.HStack(
			gu.Button("Default"),
			gu.Button("Primary", gu.ButtonPrimary),
			gu.Button("Outline", gu.ButtonOutline),
			gu.Button("Danger", gu.ButtonDanger),
			gu.Button("Small", gu.ButtonSmall),
			gu.Button("Small Primary", gu.ButtonSmall, gu.ButtonPrimary),
		),
		section("Button with Icon", "Combine Icon() with Button."),
		gu.HStack(
			gu.Button("Add", gu.ButtonPrimary, gu.Icon("plus", "sm")),
			gu.Button("Delete", gu.ButtonDanger, gu.Icon("trash", "sm")),
			gu.Button("Download", gu.ButtonOutline, gu.Icon("download", "sm")),
			gu.Button("Search", gu.ButtonSmall, gu.Icon("search", "sm")),
		),
	)
}

func (v *CatalogView) pageIcon() g.ComponentFunc {
	names := gu.IconNames()
	var iconCells []g.ComponentFunc
	for _, name := range names {
		iconCells = append(iconCells, gu.VStack(
			gp.Attr("style", "padding:12px;border:1px solid var(--g-border);border-radius:var(--g-radius);min-width:80px"),
			gu.Icon(name, "lg"),
			gd.Span(gp.Attr("style", "font-size:11px;color:var(--g-text-tertiary)"), gp.Value(name)),
		))
	}
	items := []g.ComponentFunc{section("Icon", fmt.Sprintf("SVG icons (%d available). Usage: gu.Icon(name, size).", len(names)))}
	hstackAttrs := append([]g.ComponentFunc{}, iconCells...)
	items = append(items, gu.HStack(hstackAttrs...))
	return gu.Stack(items...)
}

func (v *CatalogView) pageBadge() g.ComponentFunc {
	return gu.Stack(
		section("Badge", "Status indicator labels."),
		gu.HStack(
			gu.Badge("Default"),
			gu.Badge("Dark", "dark"),
			gu.Badge("Outline", "outline"),
			gu.Badge("Light", "light"),
		),
		section("Alert", "Notification messages."),
		gu.Alert("Info message.", "info"),
		gu.Alert("Success message.", "success"),
		gu.Alert("Warning message.", "warning"),
		gu.Alert("Danger message.", "danger"),
	)
}

func (v *CatalogView) pageTable() g.ComponentFunc {
	return gd.Div(
		section("Table", "Styled table with header. Scrollable on mobile."),
		gu.Card(
			gu.StyledTable(
				gu.THead("Name", "Email", "Role"),
				gd.Tbody(
					gd.Tr(gd.Td(gp.Value("Alice")), gd.Td(gp.Value("alice@example.com")), gd.Td(gu.Badge("Admin", "dark"))),
					gd.Tr(gd.Td(gp.Value("Bob")), gd.Td(gp.Value("bob@example.com")), gd.Td(gu.Badge("User"))),
					gd.Tr(gd.Td(gp.Value("Charlie")), gd.Td(gp.Value("charlie@example.com")), gd.Td(gu.Badge("User"))),
				),
			),
		),
	)
}

func (v *CatalogView) pageForm() g.ComponentFunc {
	return gd.Div(
		section("Form", "Basic styled form controls."),
		gu.Card(
			gd.Div(gp.Class("g-page-body"),
				gu.FormGroup(
					gu.FormLabel("Name", "name"),
					gu.FormInput("name", gp.ID("name"), gp.Placeholder("Enter your name")),
				),
				gu.FormGroup(
					gu.FormLabel("Email", "email"),
					gu.FormInput("email", gp.ID("email"), gp.Type("email"), gp.Placeholder("you@example.com")),
				),
				gu.FormGroup(
					gu.FormLabel("Bio", "bio"),
					gu.FormTextarea("bio", gp.ID("bio"), gp.Placeholder("Tell us about yourself...")),
				),
				gu.FormGroup(
					gu.FormLabel("Role", "role"),
					gu.FormSelect("role", []gu.FormOption{
						{Value: "admin", Label: "Admin"},
						{Value: "editor", Label: "Editor"},
						{Value: "viewer", Label: "Viewer"},
					}, gp.ID("role")),
				),
				gu.Button("Submit", gu.ButtonPrimary),
			),
		),
	)
}

func (v *CatalogView) pageFormAdvanced() g.ComponentFunc {
	nameErr := ""
	emailErr := ""
	if v.FormSubmitted {
		if v.FormName == "" {
			nameErr = "Name is required."
		}
		if v.FormEmail == "" {
			emailErr = "Email is required."
		} else if !strings.Contains(v.FormEmail, "@") {
			emailErr = "Please enter a valid email address."
		}
	}

	return gu.Stack(
		section("Form Validation", "Dynamic error messages with LiveView."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.FormGroup(
				gu.FormLabel("Name", "vname"),
				gu.FormInput("vname", gp.ID("vname"),
					gp.Attr("value", v.FormName),
					gp.Placeholder("Required"),
					gl.Input("formInput"),
					expr.If(nameErr != "", gu.FormInputError),
				),
				gu.FormError(nameErr),
			),
			gu.FormGroup(
				gu.FormLabel("Email", "vemail"),
				gu.FormInput("vemail", gp.ID("vemail"),
					gp.Attr("value", v.FormEmail),
					gp.Placeholder("you@example.com"),
					gl.Input("formInput"),
					expr.If(emailErr != "", gu.FormInputError),
				),
				gu.FormError(emailErr),
			),
			gu.Button("Validate", gu.ButtonPrimary, gl.Click("formValidate")),
			expr.If(v.FormSubmitted && nameErr == "" && emailErr == "",
				gd.Div(gp.Attr("style", "margin-top:12px"),
					gu.Alert("All fields are valid!", "success"),
				),
			),
		)),
		section("Checkbox", "Styled checkboxes with labels."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Checkbox("check-a", "Enable notifications", v.CheckA, gl.Click("toggleCheckA")),
			gu.Checkbox("check-b", "Accept terms and conditions", v.CheckB, gl.Click("toggleCheckB")),
		)),
		section("Radio", "Styled radio buttons with mutual exclusion."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Radio("plan", "opt1", "Free Plan", v.RadioV == "opt1", gl.Click("setRadio"), gl.ClickValue("opt1")),
			gu.Radio("plan", "opt2", "Pro Plan ($9/mo)", v.RadioV == "opt2", gl.Click("setRadio"), gl.ClickValue("opt2")),
			gu.Radio("plan", "opt3", "Enterprise (Contact us)", v.RadioV == "opt3", gl.Click("setRadio"), gl.ClickValue("opt3")),
			gd.Div(gp.Attr("style", "margin-top:8px"),
				gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"), gp.Value("Selected: "+v.RadioV)),
			),
		)),
	)
}

func (v *CatalogView) pageStat() g.ComponentFunc {
	return gu.Stack(
		section("StatCard", "KPI/metric display. Grid adapts on mobile."),
		gu.Grid(gu.GridCols4,
			gu.StatCard("Total Users", "12,345"),
			gu.StatCard("Active Now", "843"),
			gu.StatCard("Revenue", "$54.2K"),
			gu.StatCard("Growth", "+12.5%"),
		),
	)
}

func (v *CatalogView) pageNav() g.ComponentFunc {
	return gu.Stack(
		section("Breadcrumb", "Navigation trail."),
		gu.Breadcrumb(
			gu.BreadcrumbItem{Label: "Home", Href: "/"},
			gu.BreadcrumbItem{Label: "Users", Href: "/users"},
			gu.BreadcrumbItem{Label: "Alice", Href: ""},
		),
		section("MobileHeader", "Shown on screens <= 768px. Resize to see."),
		gu.Alert("The mobile header with hamburger menu is visible above on narrow screens.", "info"),
	)
}

func (v *CatalogView) pageTree() g.ComponentFunc {
	return gd.Div(
		section("TreeView", "Hierarchical tree with expand/collapse (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Tree([]gu.TreeNode{
				{
					Label: "src", Icon: "folder", Open: v.TreeOpen["src"],
					Attrs: []g.ComponentFunc{gl.Click("treeToggle"), gl.ClickValue("src")},
					Children: []gu.TreeNode{
						{
							Label: "components", Icon: "folder", Open: v.TreeOpen["components"],
							Attrs: []g.ComponentFunc{gl.Click("treeToggle"), gl.ClickValue("components")},
							Children: []gu.TreeNode{
								{Label: "Button.tsx", Icon: "file"},
								{Label: "Card.tsx", Icon: "file"},
								{Label: "Modal.tsx", Icon: "file"},
							},
						},
						{
							Label: "pages", Icon: "folder", Open: v.TreeOpen["pages"],
							Attrs: []g.ComponentFunc{gl.Click("treeToggle"), gl.ClickValue("pages")},
							Children: []gu.TreeNode{
								{Label: "index.tsx", Icon: "file"},
								{Label: "about.tsx", Icon: "file"},
							},
						},
						{Label: "main.ts", Icon: "file"},
					},
				},
				{Label: "package.json", Icon: "file"},
				{Label: "README.md", Icon: "file"},
			}),
		)),
	)
}

func (v *CatalogView) pageMisc() g.ComponentFunc {
	return gu.Stack(
		section("Divider", "Horizontal separator."),
		gd.P(gp.Value("Content above")),
		gu.Divider(),
		gd.P(gp.Value("Content below")),
		section("EmptyState", "Placeholder for empty content."),
		gu.Card(
			gu.EmptyState("No items found.", gu.Button("Add Item", gu.ButtonPrimary)),
		),
		section("Progress", "Progress bar."),
		gu.Column(gu.GapMd,
			gu.Progress(0),
			gu.Progress(35),
			gu.Progress(75),
			gu.Progress(100),
		),
	)
}

func (v *CatalogView) pageLayout() g.ComponentFunc {
	return gu.Stack(
		section("Row & Column", "Flex row (horizontal) and column (vertical) containers."),
		gu.Row(
			gu.Button("A"),
			gu.Button("B"),
			gu.Button("C"),
		),
		gu.Column(
			gu.Alert("First", "info"),
			gu.Alert("Second", "success"),
		),

		section("Stack & HStack & VStack", "Pre-configured flex stacks."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gd.P(gp.Value("Stack (vertical, gap:16px) wraps this whole card body.")),
			gu.HStack(
				gu.Button("Save", gu.ButtonPrimary),
				gu.Button("Cancel"),
				gu.Button("Delete", gu.ButtonDanger),
			),
			gu.VStack(
				gu.Icon("star", "lg"),
				gd.Span(gp.Value("Centered label")),
			),
		)),

		section("Center", "Center content horizontally and vertically."),
		gd.Div(gp.Attr("style", "height:120px;border:1px dashed var(--g-border);border-radius:var(--g-radius)"),
			gu.Center(gp.Attr("style", "height:100%"),
				gd.Span(gp.Value("Centered content")),
			),
		),

		section("Grid", "CSS Grid with column modifiers."),
		gu.Grid(gu.GridCols3,
			gu.StatCard("Users", "100"),
			gu.StatCard("Revenue", "$50K"),
			gu.StatCard("Growth", "+12%"),
		),
		gu.Grid(gu.GridCols4,
			gu.Badge("1"), gu.Badge("2"), gu.Badge("3"), gu.Badge("4"),
		),

		section("GridAutoFit", "Responsive auto-fit grid (min 200px)."),
		gu.Grid(gu.GridAutoFit("200px"),
			gu.StatCard("A", "1"),
			gu.StatCard("B", "2"),
			gu.StatCard("C", "3"),
			gu.StatCard("D", "4"),
			gu.StatCard("E", "5"),
		),

		section("GridSpan", "Grid child spanning multiple columns."),
		gu.Grid(gu.GridCols3,
			gd.Div(gu.GridSpan(2), gu.Alert("Spans 2 columns", "info")),
			gu.Alert("1 column", "success"),
		),

		section("Container", "Max-width containers (960 / 640 / 1280px)."),
		gu.Container(gu.Alert("Container (960px)", "info")),
		gu.ContainerNarrow(gu.Alert("ContainerNarrow (640px)", "warning")),
		gu.ContainerWide(gu.Alert("ContainerWide (1280px)", "success")),

		section("Spacer", "Push elements apart within a Row."),
		gu.Row(gu.AlignCenter,
			gd.Span(gp.Value("Logo")),
			gu.Spacer(),
			gu.Button("Logout", gu.ButtonOutline),
		),

		section("SpaceY", "Vertical spacing helpers."),
		gu.Column(
			gu.Alert("Above", "info"),
			gu.SpaceY("xl"),
			gu.Alert("Below (xl gap)", "success"),
		),

		section("Gap Modifiers", "Override default gap on any flex/grid container."),
		gu.Row(gu.GapLg,
			gu.Badge("Gap"), gu.Badge("Large"),
		),
		gu.Row(gu.GapNone,
			gu.Badge("Gap"), gu.Badge("None"),
		),

		section("Justify & Align", "Override justify-content and align-items."),
		gu.Row(gu.JustifyBetween,
			gu.Button("Left"),
			gu.Button("Right"),
		),
		gu.Row(gu.JustifyCenter,
			gu.Button("Centered"),
		),
	)
}

func (v *CatalogView) pagePagination() g.ComponentFunc {
	total := 120

	return gu.Stack(
		section("Pagination", "Page navigation with ellipsis for large page counts (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Pagination(gu.PaginationOpts{
				Page:      v.PaginationPage,
				PageSize:  10,
				Total:     total,
				PageEvent: "paginateTo",
			}),
			gd.Div(gp.Attr("style", "margin-top:12px"),
				gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"),
					gp.Value(fmt.Sprintf("Current page: %d (0-based)", v.PaginationPage))),
			),
		)),
		section("Static Pagination", "Non-interactive version without LiveView events."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Pagination(gu.PaginationOpts{Page: 3, PageSize: 10, Total: 120}),
		)),
	)
}

func (v *CatalogView) pageButtonGroup() g.ComponentFunc {
	return gu.Stack(
		section("ButtonGroup", "Segmented control / button group (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.ButtonGroup([]gu.ButtonGroupItem{
				{Label: "Day", Value: "day", Active: v.BtnGroupValue == "day"},
				{Label: "Week", Value: "week", Active: v.BtnGroupValue == "week"},
				{Label: "Month", Value: "month", Active: v.BtnGroupValue == "month"},
				{Label: "Year", Value: "year", Active: v.BtnGroupValue == "year"},
			}, gu.ButtonGroupOpts{ClickEvent: "btnGroupChange"}),
			gd.Div(gp.Attr("style", "margin-top:12px"),
				gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"),
					gp.Value("Selected: "+v.BtnGroupValue)),
			),
		)),
		section("Static ButtonGroup", "Non-interactive version without LiveView events."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.ButtonGroup([]gu.ButtonGroupItem{
				{Label: "Left", Value: "left", Active: true},
				{Label: "Center", Value: "center"},
				{Label: "Right", Value: "right"},
			}, gu.ButtonGroupOpts{}),
		)),
		section("Small ButtonGroup", "Compact variant with ButtonGroupSmall modifier."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.ButtonGroup([]gu.ButtonGroupItem{
				{Label: "S", Value: "s", Active: true},
				{Label: "M", Value: "m"},
				{Label: "L", Value: "l"},
			}, gu.ButtonGroupOpts{Small: true}),
		)),
	)
}

func (v *CatalogView) pageAccordion() g.ComponentFunc {
	return gu.Stack(
		section("Accordion", "Collapsible sections with server-controlled open/close (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Accordion([]gu.AccordionItem{
				{Title: "What is Gerbera?", Content: gd.P(gp.Value("Gerbera is a Go HTML template engine that uses functional composition instead of traditional template files.")), Open: v.AccordionOpen[0]},
				{Title: "How does it work?", Content: gd.P(gp.Value("HTML is built programmatically by composing ComponentFunc functions.")), Open: v.AccordionOpen[1]},
				{Title: "Is it production ready?", Content: gd.P(gp.Value("Gerbera uses sync.Pool for zero-allocation rendering and is suitable for production use.")), Open: v.AccordionOpen[2]},
			}, gu.AccordionOpts{Exclusive: true, ToggleEvent: "accordionToggle"}),
		)),
		section("Static Accordion", "Native details/summary based, no LiveView required."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Accordion([]gu.AccordionItem{
				{Title: "First Section", Content: gd.P(gp.Value("This uses native HTML details/summary elements.")), Open: true},
				{Title: "Second Section", Content: gd.P(gp.Value("Browser handles open/close without JavaScript.")), Open: false},
			}, gu.AccordionOpts{}),
		)),
	)
}

func (v *CatalogView) pageStepper() g.ComponentFunc {
	steps := []gu.Step{
		{Label: "Cart", Description: "Review items"},
		{Label: "Shipping", Description: "Enter address"},
		{Label: "Payment", Description: "Add payment method"},
		{Label: "Confirm", Description: "Place order"},
	}
	for i := range steps {
		switch {
		case i < v.StepperCurrent:
			steps[i].Status = gu.StepCompleted
		case i == v.StepperCurrent:
			steps[i].Status = gu.StepActive
		default:
			steps[i].Status = gu.StepUpcoming
		}
	}

	return gu.Stack(
		section("Stepper", "Step-by-step progress indicator (LiveView). Click completed steps to go back."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Stepper(steps, gu.StepperOpts{ClickEvent: "stepperClick"}),
			gd.Div(gp.Attr("style", "margin-top:16px"),
				gu.HStack(
					expr.If(v.StepperCurrent > 0,
						gu.Button("Back", gu.ButtonOutline, gl.Click("stepperPrev")),
					),
					expr.If(v.StepperCurrent < len(steps)-1,
						gu.Button("Next", gu.ButtonPrimary, gl.Click("stepperNext")),
					),
				),
			),
		)),
		section("Vertical Stepper", "Vertical layout, also responsive (auto-vertical on mobile)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Stepper([]gu.Step{
				{Label: "Sign Up", Status: gu.StepCompleted},
				{Label: "Verify Email", Status: gu.StepActive, Description: "Check your inbox"},
				{Label: "Set Profile", Status: gu.StepUpcoming},
			}, gu.StepperOpts{Vertical: true}),
		)),
	)
}

func (v *CatalogView) pageInfiniteScroll() g.ComponentFunc {
	var items []g.ComponentFunc
	for i := 0; i < v.InfScrollItems; i++ {
		items = append(items, gu.Card(
			gd.Div(gp.Class("g-page-body"),
				gd.Span(gp.Value(fmt.Sprintf("Item %d", i+1))),
			),
		))
	}

	return gu.Stack(
		section("InfiniteScroll", "Scrollable list with load-more, view toggle (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.InfiniteScroll(gu.InfiniteScrollOpts{
				View:          v.InfScrollView,
				Loading:       v.InfScrollLoading,
				ShowToggle:    true,
				LoadMoreEvent: "loadMore",
				ToggleEvent:   "toggleInfView",
			}, items...),
		)),
		section("Static InfiniteScroll", "Non-interactive version."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.InfiniteScroll(gu.InfiniteScrollOpts{View: gu.InfiniteScrollList, ShowToggle: true},
				gu.Card(gd.Div(gp.Class("g-page-body"), gd.Span(gp.Value("Static Item 1")))),
				gu.Card(gd.Div(gp.Class("g-page-body"), gd.Span(gp.Value("Static Item 2")))),
				gu.Card(gd.Div(gp.Class("g-page-body"), gd.Span(gp.Value("Static Item 3")))),
			),
		)),
	)
}

func (v *CatalogView) pageModal() g.ComponentFunc {
	return gd.Div(
		section("Modal", "Dialog overlay (LiveView)."),
		gu.Button("Open Modal", gu.ButtonPrimary, gl.Click("openModal")),
		gul.Modal(v.ModalOpen, "closeModal",
			gul.ModalHeader("Sample Modal", "closeModal"),
			gul.ModalBody(
				gd.P(gp.Value("This is a modal dialog built with gerbera/ui/live.")),
				gu.Alert("Modal content can include any widget.", "info"),
			),
			gul.ModalFooter(
				gu.Button("Close", gu.ButtonOutline, gl.Click("closeModal")),
				gu.Button("Save", gu.ButtonPrimary, gl.Click("closeModal")),
			),
		),
	)
}

func (v *CatalogView) pageToast() g.ComponentFunc {
	return gd.Div(
		section("Toast", "Notification popup (LiveView)."),
		gu.HStack(
			gu.Button("Info", gu.ButtonOutline, gl.Click("showToast"), gl.ClickValue("info")),
			gu.Button("Success", gu.ButtonOutline, gl.Click("showToast"), gl.ClickValue("success")),
			gu.Button("Warning", gu.ButtonOutline, gl.Click("showToast"), gl.ClickValue("warning")),
			gu.Button("Danger", gu.ButtonOutline, gl.Click("showToast"), gl.ClickValue("danger")),
		),
		expr.If(v.ToastVisible,
			gul.Toast(v.ToastMessage, v.ToastVariant, "dismissToast"),
		),
	)
}

func (v *CatalogView) pageDataTable() g.ComponentFunc {
	rows := sampleRows()
	pageSize := 5
	start := v.TablePage * pageSize
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	pageRows := rows[start:end]

	return gd.Div(
		section("DataTable", "Sortable, paginated table (LiveView)."),
		gul.DataTable(gul.DataTableOpts{
			Columns: []gul.Column{
				{Key: "name", Label: "Name", Sortable: true},
				{Key: "email", Label: "Email", Sortable: true},
				{Key: "role", Label: "Role", Sortable: true},
				{Key: "status", Label: "Status", Sortable: false},
			},
			Rows:      pageRows,
			SortCol:   v.SortCol,
			SortDir:   v.SortDir,
			SortEvent: "sort",
			Page:      v.TablePage,
			PageSize:  pageSize,
			Total:     len(rows),
			PageEvent: "tablePage",
		}),
	)
}

func (v *CatalogView) pageDropdown() g.ComponentFunc {
	label := "Options"
	if v.DropdownValue != "" {
		labels := map[string]string{"edit": "Edit", "duplicate": "Duplicate", "delete": "Delete"}
		if l, ok := labels[v.DropdownValue]; ok {
			label = l
		}
	}

	return gd.Div(
		section("Dropdown", "Toggle menu (LiveView)."),
		gul.Dropdown(v.DropdownOpen, "toggleDropdown",
			gu.Button(label, gu.ButtonOutline, gu.Icon("chevron-down", "sm")),
			gd.Div(
				gd.Button(gp.Class("g-dropdown-item"), gl.Click("dropdownAction"), gl.ClickValue("edit"), gp.Value("Edit")),
				gd.Button(gp.Class("g-dropdown-item"), gl.Click("dropdownAction"), gl.ClickValue("duplicate"), gp.Value("Duplicate")),
				gd.Button(gp.Class("g-dropdown-item"), gl.Click("dropdownAction"), gl.ClickValue("delete"), gp.Value("Delete")),
			),
		),
		expr.If(v.DropdownValue != "",
			gd.Div(gp.Attr("style", "margin-top:8px"),
				gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"),
					gp.Value("Selected: "+v.DropdownValue)),
			),
		),
	)
}

func (v *CatalogView) pageTabs() g.ComponentFunc {
	return gd.Div(
		section("Tabs", "Accessible tab panel with LiveView switching."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gul.Tabs("demo-tabs", v.TabIndex, []gul.Tab{
				{
					Label: "Overview",
					Content: gd.Div(
						gd.P(gp.Value("This is the Overview tab content.")),
						gu.Alert("Tabs switch instantly via LiveView events.", "info"),
					),
				},
				{
					Label: "Details",
					Content: gd.Div(
						gu.StyledTable(
							gu.THead("Property", "Value"),
							gd.Tbody(
								gd.Tr(gd.Td(gp.Value("Name")), gd.Td(gp.Value("Gerbera UI"))),
								gd.Tr(gd.Td(gp.Value("Version")), gd.Td(gp.Value("1.0.0"))),
							),
						),
					),
				},
				{
					Label: "Settings",
					Content: gd.Div(
						gu.FormGroup(
							gu.FormLabel("Display name", "tab-name"),
							gu.FormInput("tab-name", gp.ID("tab-name"), gp.Placeholder("Enter name")),
						),
						gu.Button("Save", gu.ButtonPrimary),
					),
				},
			}, "switchTab"),
		)),
	)
}

func (v *CatalogView) pageDrawer() g.ComponentFunc {
	return gd.Div(
		section("Drawer", "Slide-out panel from left or right edge (LiveView)."),
		gu.HStack(
			gu.Button("Open Left Drawer", gu.ButtonOutline, gl.Click("openDrawer"), gl.ClickValue("left")),
			gu.Button("Open Right Drawer", gu.ButtonOutline, gl.Click("openDrawer"), gl.ClickValue("right")),
		),
		gul.Drawer(v.DrawerOpen, "closeDrawer", v.DrawerSide,
			gul.DrawerHeader("Drawer Panel", "closeDrawer"),
			gul.DrawerBody(
				gd.P(gp.Value("This is a slide-out drawer panel.")),
				gu.Alert("Drawers are useful for mobile navigation, filters, or detail panels.", "info"),
				gu.SpaceY("md"),
				gu.FormGroup(
					gu.FormLabel("Filter", "drawer-filter"),
					gu.FormInput("drawer-filter", gp.ID("drawer-filter"), gp.Placeholder("Type to filter...")),
				),
			),
		),
	)
}

func (v *CatalogView) ssFilteredOptions() []gu.FormOption {
	var filtered []gu.FormOption
	q := strings.ToLower(v.SSQuery)
	for _, o := range allCountryOptions() {
		if q == "" || strings.Contains(strings.ToLower(o.Label), q) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

func (v *CatalogView) pageSearchSelect() g.ComponentFunc {
	filtered := v.ssFilteredOptions()

	selectedLabel := ""
	for _, o := range allCountryOptions() {
		if o.Value == v.SSValue {
			selectedLabel = o.Label
			break
		}
	}

	return gd.Div(
		section("SearchSelect", "Filterable combobox (LiveView). Type to search, click to select."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.FormGroup(
				gu.FormLabel("Country", "ss-country"),
				gul.SearchSelect(gul.SearchSelectOpts{
					Name:           "country",
					Query:          v.SSQuery,
					Options:        filtered,
					Selected:       v.SSValue,
					Open:           v.SSOpen,
					Placeholder:    "Type to search countries...",
					InputEvent:     "ssInput",
					SelectEvent:    "ssSelect",
					FocusEvent:     "ssFocus",
					KeydownEvent:   "ssKeydown",
					HighlightIndex: v.SSHighlight,
				}),
			),
			expr.If(v.SSValue != "",
				gd.Div(gp.Attr("style", "margin-top:8px"),
					gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"),
						gp.Value("Selected: "+selectedLabel+" ("+v.SSValue+")")),
				),
			),
		)),
	)
}

func (v *CatalogView) pageSpinner() g.ComponentFunc {
	return gu.Stack(
		section("Spinner", "CSS-only loading animation in three sizes."),
		gu.HStack(
			gu.VStack(
				gu.Spinner("sm"),
				gd.Span(gp.Attr("style", "font-size:11px;color:var(--g-text-tertiary)"), gp.Value("sm (16px)")),
			),
			gu.VStack(
				gu.Spinner("md"),
				gd.Span(gp.Attr("style", "font-size:11px;color:var(--g-text-tertiary)"), gp.Value("md (24px)")),
			),
			gu.VStack(
				gu.Spinner("lg"),
				gd.Span(gp.Attr("style", "font-size:11px;color:var(--g-text-tertiary)"), gp.Value("lg (40px)")),
			),
		),
		section("Inline Spinner", "Use SpinnerInline to display inline with text."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Row(gu.AlignCenter, gu.GapSm,
				gu.Spinner("sm", gu.SpinnerInline),
				gd.Span(gp.Value("Loading data...")),
			),
		)),
		section("Spinner in Button", "Combine with a button for loading states."),
		gu.HStack(
			gu.Button("Saving...", gu.ButtonPrimary, gp.Disabled(true), gu.Spinner("sm", gu.SpinnerInline)),
			gu.Button("Loading", gu.ButtonOutline, gp.Disabled(true), gu.Spinner("sm", gu.SpinnerInline)),
		),
	)
}

func (v *CatalogView) pageNumberInput() g.ComponentFunc {
	min, max := 0, 20

	return gu.Stack(
		section("NumberInput", "Number input with increment/decrement buttons (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.FormGroup(
				gu.FormLabel("Quantity", "num-qty"),
				gu.NumberInput("qty", v.NumVal, gu.NumberInputOpts{
					Min:            &min,
					Max:            &max,
					Step:           1,
					IncrementEvent: "numInc",
					DecrementEvent: "numDec",
					ChangeEvent:    "numChange",
				}),
			),
			gd.Div(gp.Attr("style", "margin-top:8px"),
				gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"),
					gp.Value(fmt.Sprintf("Value: %d (min: %d, max: %d)", v.NumVal, min, max))),
			),
		)),
		section("Static NumberInput", "Non-interactive version without LiveView events."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.NumberInput("items", 3, gu.NumberInputOpts{Min: &min, Max: &max, Step: 1}),
		)),
	)
}

func (v *CatalogView) pageSlider() g.ComponentFunc {
	return gu.Stack(
		section("Slider", "Range slider with label and live value display (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Slider("volume", v.SliderVal, gu.SliderOpts{
				Min:        0,
				Max:        100,
				Step:       1,
				Label:      "Volume",
				InputEvent: "slideInput",
			}),
			gd.Div(gp.Attr("style", "margin-top:12px"),
				gu.Progress(v.SliderVal),
			),
		)),
		section("Static Slider", "Non-interactive version without LiveView events."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Slider("brightness", 75, gu.SliderOpts{Min: 0, Max: 100, Label: "Brightness"}),
		)),
	)
}

func (v *CatalogView) pageTimePicker() g.ComponentFunc {
	return gu.Stack(
		section("TimePicker", "Time picker with hour/minute increment buttons (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.FormGroup(
				gu.FormLabel("Alarm Time (24h)", "tp-alarm"),
				gu.TimePicker("alarm", v.TimeHour, v.TimeMinute, v.TimeSecond, gu.TimePickerOpts{
					Use24H:      true,
					ChangeEvent: "timeChange",
				}),
			),
			gd.Div(gp.Attr("style", "margin-top:8px"),
				gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"),
					gp.Value(fmt.Sprintf("Value: %s", gu.FormatTime(v.TimeHour, v.TimeMinute, v.TimeSecond, false)))),
			),
		)),
		section("12-Hour Format", "TimePicker with AM/PM toggle."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.TimePicker("alarm12", v.TimeHour, v.TimeMinute, 0, gu.TimePickerOpts{
				Use24H:      false,
				ChangeEvent: "timeChange",
			}),
		)),
		section("With Seconds", "TimePicker showing seconds field."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.TimePicker("precise", v.TimeHour, v.TimeMinute, v.TimeSecond, gu.TimePickerOpts{
				Use24H:      true,
				ShowSec:     true,
				ChangeEvent: "timeChange",
			}),
		)),
		section("Static TimePicker", "Non-interactive version without LiveView events."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.TimePicker("meeting", 9, 0, 0, gu.TimePickerOpts{Use24H: true}),
		)),
	)
}

func (v *CatalogView) pageCalendar() g.ComponentFunc {
	return gu.Stack(
		section("Calendar", "Month-view calendar with navigation and date selection (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Calendar(gu.CalendarOpts{
				Year:             v.CalYear,
				Month:            v.CalMonth,
				Selected:         v.CalSelected,
				Today:            time.Now(),
				SelectEvent:      "calSelect",
				PrevMonthEvent:   "calPrev",
				NextMonthEvent:   "calNext",
				MonthChangeEvent: "calMonthChange",
				YearChangeEvent:  "calYearChange",
			}),
			expr.If(v.CalSelected != nil,
				gd.Div(gp.Attr("style", "margin-top:12px"),
					gu.Alert(fmt.Sprintf("Selected: %s", v.calSelectedStr()), "info"),
				),
			),
		)),
		section("Static Calendar", "Non-interactive version without LiveView events."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Calendar(gu.CalendarOpts{
				Year:  time.Now().Year(),
				Month: time.Now().Month(),
				Today: time.Now(),
			}),
		)),
	)
}

func (v *CatalogView) calSelectedStr() string {
	if v.CalSelected == nil {
		return ""
	}
	return v.CalSelected.Format("2006-01-02")
}

func (v *CatalogView) pageChat() g.ComponentFunc {
	var msgViews []g.ComponentFunc
	for _, m := range v.ChatMessages {
		msgViews = append(msgViews, gu.ChatMessageView(m))
	}

	return gu.Stack(
		section("Chat", "Chat message list with input area (LiveView)."),
		gu.Card(
			gu.ChatContainer(msgViews...),
			gu.ChatInput("chatMsg", v.ChatDraft, gu.ChatInputOpts{
				Placeholder:  "Type a message...",
				SendEvent:    "chatSend",
				InputEvent:   "chatInput",
				KeydownEvent: "chatKeydown",
			}),
		),
		section("Static Chat", "Non-interactive chat rendering."),
		gu.Card(
			gu.ChatContainer(
				gu.ChatMessageView(gu.ChatMessage{Author: "System", Content: "Welcome to the chat.", Timestamp: "09:00", Sent: false}),
				gu.ChatMessageView(gu.ChatMessage{Content: "Hello!", Timestamp: "09:01", Sent: true}),
			),
			gu.ChatInput("msg", "", gu.ChatInputOpts{}),
		),
	)
}

func (v *CatalogView) pageConfirm() g.ComponentFunc {
	return gd.Div(
		section("Confirm", "Confirmation dialog (LiveView)."),
		gu.Button("Delete Item", gu.ButtonDanger, gl.Click("openConfirm")),
		gul.Confirm(v.ConfirmOpen, "Delete Item", "Are you sure you want to delete this item? This action cannot be undone.", "doConfirm", "cancelConfirm"),
	)
}

func (v *CatalogView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "nav":
		v.Page = payload["value"]
		v.resetOverlays()
	case "toggleMobileNav":
		v.MobileNavOpen = !v.MobileNavOpen
	case "closeMobileNav":
		v.MobileNavOpen = false
	case "openModal":
		v.ModalOpen = true
	case "closeModal":
		v.ModalOpen = false
	case "showToast":
		variant := payload["value"]
		v.ToastVisible = true
		v.ToastVariant = variant
		v.ToastMessage = fmt.Sprintf("This is a %s notification.", variant)
	case "dismissToast":
		v.ToastVisible = false
	case "sort":
		col := payload["value"]
		if v.SortCol == col {
			if v.SortDir == "asc" {
				v.SortDir = "desc"
			} else {
				v.SortDir = "asc"
			}
		} else {
			v.SortCol = col
			v.SortDir = "asc"
		}
	case "tablePage":
		fmt.Sscanf(payload["value"], "%d", &v.TablePage)
	case "switchTab":
		fmt.Sscanf(payload["value"], "%d", &v.TabIndex)
	case "toggleDropdown":
		v.DropdownOpen = !v.DropdownOpen
	case "dropdownAction":
		v.DropdownValue = payload["value"]
		v.DropdownOpen = false
	case "openConfirm":
		v.ConfirmOpen = true
	case "doConfirm":
		v.ConfirmOpen = false
	case "cancelConfirm":
		v.ConfirmOpen = false

	// Drawer
	case "openDrawer":
		v.DrawerOpen = true
		v.DrawerSide = payload["value"]
	case "closeDrawer":
		v.DrawerOpen = false

	// SearchSelect
	case "ssInput":
		v.SSQuery = payload["value"]
		v.SSOpen = true
		v.SSHighlight = -1
	case "ssFocus":
		v.SSOpen = true
	case "ssSelect":
		v.SSValue = payload["value"]
		v.SSOpen = false
		v.SSHighlight = -1
		// Show selected label in query
		for _, c := range allCountryOptions() {
			if c.Value == v.SSValue {
				v.SSQuery = c.Label
				break
			}
		}
	case "ssKeydown":
		key := payload["key"]
		filtered := v.ssFilteredOptions()
		n := len(filtered)
		switch key {
		case "Escape":
			v.SSOpen = false
			v.SSHighlight = -1
		case "ArrowDown":
			if !v.SSOpen {
				v.SSOpen = true
				v.SSHighlight = 0
			} else if n > 0 {
				v.SSHighlight = (v.SSHighlight + 1) % n
			}
		case "ArrowUp":
			if v.SSOpen && n > 0 {
				if v.SSHighlight <= 0 {
					v.SSHighlight = n - 1
				} else {
					v.SSHighlight--
				}
			}
		case "Enter":
			if v.SSOpen && v.SSHighlight >= 0 && v.SSHighlight < n {
				selected := filtered[v.SSHighlight]
				v.SSValue = selected.Value
				v.SSQuery = selected.Label
				v.SSOpen = false
				v.SSHighlight = -1
			}
		}

	// Pagination
	case "paginateTo":
		fmt.Sscanf(payload["value"], "%d", &v.PaginationPage)

	// ButtonGroup
	case "btnGroupChange":
		v.BtnGroupValue = payload["value"]

	// Accordion
	case "accordionToggle":
		var idx int
		fmt.Sscanf(payload["value"], "%d", &idx)
		if idx >= 0 && idx < len(v.AccordionOpen) {
			// Exclusive mode: close others when opening
			if !v.AccordionOpen[idx] {
				for i := range v.AccordionOpen {
					v.AccordionOpen[i] = false
				}
			}
			v.AccordionOpen[idx] = !v.AccordionOpen[idx]
		}

	// Stepper
	case "stepperClick":
		var idx int
		fmt.Sscanf(payload["value"], "%d", &idx)
		if idx >= 0 && idx < v.StepperCurrent {
			v.StepperCurrent = idx
		}
	case "stepperPrev":
		if v.StepperCurrent > 0 {
			v.StepperCurrent--
		}
	case "stepperNext":
		if v.StepperCurrent < 3 {
			v.StepperCurrent++
		}

	// InfiniteScroll
	case "loadMore":
		if !v.InfScrollLoading {
			v.InfScrollLoading = true
			v.InfScrollItems += 5
			v.InfScrollLoading = false
		}
	case "toggleInfView":
		val := payload["value"]
		if val == "grid" {
			v.InfScrollView = gu.InfiniteScrollGrid
		} else {
			v.InfScrollView = gu.InfiniteScrollList
		}

	// NumberInput
	case "numInc":
		v.NumVal++
		if v.NumVal > 20 {
			v.NumVal = 20
		}
	case "numDec":
		v.NumVal--
		if v.NumVal < 0 {
			v.NumVal = 0
		}
	case "numChange":
		fmt.Sscanf(payload["value"], "%d", &v.NumVal)
		if v.NumVal < 0 {
			v.NumVal = 0
		}
		if v.NumVal > 20 {
			v.NumVal = 20
		}

	// TimePicker
	case "timeChange":
		val := payload["value"]
		switch val {
		case "hour-up":
			v.TimeHour = (v.TimeHour + 1) % 24
		case "hour-down":
			v.TimeHour = (v.TimeHour + 23) % 24
		case "minute-up":
			v.TimeMinute = (v.TimeMinute + 1) % 60
		case "minute-down":
			v.TimeMinute = (v.TimeMinute + 59) % 60
		case "second-up":
			v.TimeSecond = (v.TimeSecond + 1) % 60
		case "second-down":
			v.TimeSecond = (v.TimeSecond + 59) % 60
		case "am":
			if v.TimeHour >= 12 {
				v.TimeHour -= 12
			}
		case "pm":
			if v.TimeHour < 12 {
				v.TimeHour += 12
			}
		}

	// Slider
	case "slideInput":
		fmt.Sscanf(payload["value"], "%d", &v.SliderVal)

	// Calendar
	case "calSelect":
		if dateStr := payload["value"]; dateStr != "" {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				v.CalSelected = &t
			}
		}
	case "calPrev":
		v.CalMonth--
		if v.CalMonth < time.January {
			v.CalMonth = time.December
			v.CalYear--
		}
	case "calNext":
		v.CalMonth++
		if v.CalMonth > time.December {
			v.CalMonth = time.January
			v.CalYear++
		}
	case "calMonthChange":
		var m int
		fmt.Sscanf(payload["value"], "%d", &m)
		if m >= 1 && m <= 12 {
			v.CalMonth = time.Month(m)
		}
	case "calYearChange":
		var y int
		fmt.Sscanf(payload["value"], "%d", &y)
		if y > 0 {
			v.CalYear = y
		}

	// Chat
	case "chatInput":
		v.ChatDraft = payload["value"]
	case "chatSend":
		if strings.TrimSpace(v.ChatDraft) != "" {
			v.ChatMessages = append(v.ChatMessages, gu.ChatMessage{
				Content:   v.ChatDraft,
				Timestamp: time.Now().Format("15:04"),
				Sent:      true,
			})
			v.ChatDraft = ""
		}
	case "chatKeydown":
		if payload["key"] == "Enter" && strings.TrimSpace(v.ChatDraft) != "" {
			v.ChatMessages = append(v.ChatMessages, gu.ChatMessage{
				Content:   v.ChatDraft,
				Timestamp: time.Now().Format("15:04"),
				Sent:      true,
			})
			v.ChatDraft = ""
		}

	// Form validation
	case "formInput":
		name := payload["name"]
		val := payload["value"]
		switch name {
		case "vname":
			v.FormName = val
		case "vemail":
			v.FormEmail = val
		}
		if v.FormSubmitted {
			// Re-validate on input
		}
	case "formValidate":
		v.FormSubmitted = true

	// Tree
	case "treeToggle":
		key := payload["value"]
		v.TreeOpen[key] = !v.TreeOpen[key]

	// Checkbox / Radio
	case "toggleCheckA":
		v.CheckA = !v.CheckA
	case "toggleCheckB":
		v.CheckB = !v.CheckB
	case "setRadio":
		v.RadioV = payload["value"]

	// Theme
	case "setTheme":
		v.ThemeMode = payload["value"]

	// Chart
	case "chartTypeChange":
		v.ChartType = payload["value"]
	case "chartClick":
		v.ChartClickInfo = payload["value"]
	case "chartHover":
		v.ChartHoverInfo = payload["value"]
	case "chartLeave":
		v.ChartHoverInfo = ""

	// Avatar
	case "avatarClick":
		v.AvatarClickInfo = payload["value"]
	}
	return nil
}

func (v *CatalogView) resetOverlays() {
	v.ModalOpen = false
	v.ConfirmOpen = false
	v.ToastVisible = false
	v.DropdownOpen = false
	v.DrawerOpen = false
	v.SSOpen = false
	v.MobileNavOpen = false
}

func allCountryOptions() []gu.FormOption {
	return []gu.FormOption{
		{Value: "jp", Label: "Japan"},
		{Value: "us", Label: "United States"},
		{Value: "gb", Label: "United Kingdom"},
		{Value: "de", Label: "Germany"},
		{Value: "fr", Label: "France"},
		{Value: "ca", Label: "Canada"},
		{Value: "au", Label: "Australia"},
		{Value: "br", Label: "Brazil"},
		{Value: "in", Label: "India"},
		{Value: "kr", Label: "South Korea"},
	}
}

func (v *CatalogView) pageTheme() g.ComponentFunc {
	themeBtn := func(mode, label string) g.ComponentFunc {
		if v.ThemeMode == mode {
			return gu.Button(label, gu.ButtonPrimary, gl.Click("setTheme"), gl.ClickValue(mode))
		}
		return gu.Button(label, gu.ButtonOutline, gl.Click("setTheme"), gl.ClickValue(mode))
	}

	return gu.Stack(
		section("Theme", "Switch between light, dark, custom, and auto (OS-linked) themes."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gd.P(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary);margin:0 0 12px 0"),
				gp.Value("Click a button to change the theme. The entire page re-renders with the new CSS custom properties.")),
			gu.HStack(
				themeBtn("light", "Light"),
				themeBtn("dark", "Dark"),
				themeBtn("custom", "Custom Accent"),
				themeBtn("auto", "Auto (OS)"),
			),
			gd.Div(gp.Attr("style", "margin-top:12px"),
				gu.Alert(fmt.Sprintf("Current theme: %s", v.ThemeMode), "info"),
			),
		)),

		section("Preview", "Sample widgets rendered with the current theme."),
		gu.Grid(gu.GridCols3,
			gu.StatCard("Users", "1,234"),
			gu.StatCard("Revenue", "$56.7K"),
			gu.StatCard("Growth", "+8.3%"),
		),
		gu.Card(
			gu.CardHeader("Sample Card", gu.Button("Action", gu.ButtonSmall, gu.ButtonPrimary)),
			gd.Div(gp.Class("g-page-body"),
				gu.FormGroup(
					gu.FormLabel("Name", "theme-name"),
					gu.FormInput("theme-name", gp.ID("theme-name"), gp.Placeholder("Enter name")),
				),
				gu.HStack(
					gu.Button("Primary", gu.ButtonPrimary),
					gu.Button("Outline", gu.ButtonOutline),
					gu.Button("Danger", gu.ButtonDanger),
					gu.Button("Default"),
				),
			),
		),
		gu.HStack(
			gu.Badge("Default"),
			gu.Badge("Dark", "dark"),
			gu.Badge("Light", "light"),
			gu.Badge("Outline", "outline"),
		),
		gu.Alert("Info alert", "info"),
		gu.Alert("Success alert", "success"),
		gu.Alert("Warning alert", "warning"),
		gu.Alert("Danger alert", "danger"),
		gu.Progress(65),

		section("Usage", "Code examples for each theme mode."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gd.Pre(gp.Attr("style", "font-family:var(--g-font-mono);font-size:12px;background:var(--g-bg-inset);padding:12px;border-radius:var(--g-radius);overflow-x:auto;margin:0"),
				gp.Value(`// Default (light)
gu.Theme()

// Dark theme
gu.ThemeWith(gu.DarkTheme())

// Custom accent color
gu.ThemeWith(gu.ThemeConfig{
    Accent:      "#1a73e8",
    AccentLight: "#e8f0fe",
    AccentHover: "#1557b0",
})

// Auto (OS preference)
gu.ThemeAuto(gu.ThemeConfig{}, gu.ThemeConfig{})`),
			),
		)),
	)
}

func (v *CatalogView) chartSampleData() ([]gu.Series, []gu.DataPoint) {
	revenue := gu.Series{
		Name: "Revenue",
		Points: []gu.DataPoint{
			{Label: "Jan", Value: 420},
			{Label: "Feb", Value: 380},
			{Label: "Mar", Value: 510},
			{Label: "Apr", Value: 470},
			{Label: "May", Value: 600},
			{Label: "Jun", Value: 550},
		},
	}
	expenses := gu.Series{
		Name: "Expenses",
		Points: []gu.DataPoint{
			{Label: "Jan", Value: 300},
			{Label: "Feb", Value: 320},
			{Label: "Mar", Value: 350},
			{Label: "Apr", Value: 340},
			{Label: "May", Value: 380},
			{Label: "Jun", Value: 400},
		},
	}
	series := []gu.Series{revenue, expenses}
	pieData := []gu.DataPoint{
		{Label: "Product", Value: 45},
		{Label: "Service", Value: 30},
		{Label: "Support", Value: 15},
		{Label: "Other", Value: 10},
	}
	return series, pieData
}

func (v *CatalogView) pageChart() g.ComponentFunc {
	series, pieData := v.chartSampleData()

	chartTypes := []gu.ButtonGroupItem{
		{Label: "Line", Value: "line", Active: v.ChartType == "line"},
		{Label: "Column", Value: "column", Active: v.ChartType == "column"},
		{Label: "Bar", Value: "bar", Active: v.ChartType == "bar"},
		{Label: "Pie", Value: "pie", Active: v.ChartType == "pie"},
		{Label: "Scatter", Value: "scatter", Active: v.ChartType == "scatter"},
		{Label: "Histogram", Value: "histogram", Active: v.ChartType == "histogram"},
		{Label: "Stacked", Value: "stacked", Active: v.ChartType == "stacked"},
	}

	liveOpts := gu.ChartOpts{
		Width:           600,
		Height:          400,
		Title:           "Monthly Data",
		ShowGrid:        true,
		ShowLegend:      true,
		ShowTooltip:     true,
		ClickEvent:      "chartClick",
		MouseEnterEvent: "chartHover",
		MouseLeaveEvent: "chartLeave",
	}

	var liveChart g.ComponentFunc
	switch v.ChartType {
	case "column":
		liveChart = gu.ColumnChart(series, liveOpts)
	case "bar":
		liveChart = gu.BarChart(series, liveOpts)
	case "pie":
		liveChart = gu.PieChart(pieData, liveOpts)
	case "scatter":
		liveChart = gu.ScatterPlot(series, liveOpts)
	case "histogram":
		histValues := []float64{10, 15, 20, 25, 30, 30, 35, 40, 45, 50, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95}
		liveChart = gu.Histogram(histValues, gu.HistogramOpts{
			ChartOpts: gu.ChartOpts{
				Width: 600, Height: 400, Title: "Monthly Data",
				ShowGrid: true, ShowLegend: true, ShowTooltip: true,
				ClickEvent: "chartClick", MouseEnterEvent: "chartHover", MouseLeaveEvent: "chartLeave",
			},
			BinCount: 8,
		})
	case "stacked":
		liveChart = gu.StackedBarChart(series, liveOpts)
	default:
		liveChart = gu.LineChart(series, liveOpts)
	}

	items := []g.ComponentFunc{
		section("Charts (Live)", "Interactive charts with server-driven events. Click on data points."),
		gu.ButtonGroup(chartTypes, gu.ButtonGroupOpts{ClickEvent: "chartTypeChange"}),
		gd.Div(gp.Attr("style", "margin-top:16px"), liveChart),
	}

	// Hover info
	if v.ChartHoverInfo != "" {
		items = append(items, gd.Div(
			gp.Class("g-chart-tooltip"),
			gp.Attr("style", "position:relative;display:inline-block;margin-top:8px"),
			gp.Value("Hover: "+v.ChartHoverInfo),
		))
	}

	// Click info
	if v.ChartClickInfo != "" {
		items = append(items, gd.Div(gp.Attr("style", "margin-top:8px"),
			gu.Alert("Clicked: "+v.ChartClickInfo, "info"),
		))
	}

	// Static chart gallery
	staticOpts := gu.ChartOpts{
		Width:       400,
		Height:      280,
		ShowGrid:    true,
		ShowLegend:  true,
		ShowTooltip: true,
	}

	items = append(items,
		section("Static Chart Gallery", "All 7 chart types rendered as static SVG."),
		gu.Grid(gu.GridCols2,
			gu.Card(
				gu.CardHeader("Line Chart"),
				gd.Div(gp.Class("g-page-body"),
					gu.LineChart(series, gu.ChartOpts{Width: 400, Height: 280, ShowGrid: true, ShowLegend: true, ShowTooltip: true, Title: "Line"}),
				),
			),
			gu.Card(
				gu.CardHeader("Column Chart"),
				gd.Div(gp.Class("g-page-body"),
					gu.ColumnChart(series, gu.ChartOpts{Width: 400, Height: 280, ShowGrid: true, ShowLegend: true, Title: "Column"}),
				),
			),
			gu.Card(
				gu.CardHeader("Bar Chart"),
				gd.Div(gp.Class("g-page-body"),
					gu.BarChart(series, gu.ChartOpts{Width: 400, Height: 280, ShowGrid: true, ShowLegend: true, Title: "Bar"}),
				),
			),
			gu.Card(
				gu.CardHeader("Pie Chart"),
				gd.Div(gp.Class("g-page-body"),
					gu.PieChart(pieData, gu.ChartOpts{Width: 400, Height: 280, ShowLegend: true, Title: "Pie"}),
				),
			),
			gu.Card(
				gu.CardHeader("Scatter Plot"),
				gd.Div(gp.Class("g-page-body"),
					gu.ScatterPlot(series, staticOpts),
				),
			),
			gu.Card(
				gu.CardHeader("Histogram"),
				gd.Div(gp.Class("g-page-body"),
					gu.Histogram([]float64{10, 15, 20, 25, 30, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100},
						gu.HistogramOpts{ChartOpts: staticOpts, BinCount: 8}),
				),
			),
			gu.Card(
				gu.CardHeader("Stacked Bar Chart"),
				gd.Div(gp.Class("g-page-body"),
					gu.StackedBarChart(series, gu.ChartOpts{Width: 400, Height: 280, ShowGrid: true, ShowLegend: true, Title: "Stacked"}),
				),
			),
		),
	)

	return gu.Stack(items...)
}

func (v *CatalogView) pageAvatar() g.ComponentFunc {
	items := []g.ComponentFunc{
		section("Image Avatar", "Avatar with an image. Sizes: xs, sm, md, lg, xl."),
		gu.HStack(
			gu.ImageAvatar("https://i.pravatar.cc/64?u=1", gu.AvatarOpts{Size: "xs", Alt: "User 1"}),
			gu.ImageAvatar("https://i.pravatar.cc/64?u=2", gu.AvatarOpts{Size: "sm", Alt: "User 2"}),
			gu.ImageAvatar("https://i.pravatar.cc/64?u=3", gu.AvatarOpts{Size: "md", Alt: "User 3"}),
			gu.ImageAvatar("https://i.pravatar.cc/64?u=4", gu.AvatarOpts{Size: "lg", Alt: "User 4"}),
			gu.ImageAvatar("https://i.pravatar.cc/64?u=5", gu.AvatarOpts{Size: "xl", Alt: "User 5"}),
		),

		section("Rounded Shape", "Avatar with rounded corners instead of circle."),
		gu.HStack(
			gu.ImageAvatar("https://i.pravatar.cc/64?u=6", gu.AvatarOpts{Size: "lg", Shape: "rounded", Alt: "Rounded"}),
			gu.LetterAvatar("Rounded", gu.AvatarOpts{Size: "lg", Shape: "rounded"}),
		),

		section("Letter Avatar", "Avatar with initials. Background color is deterministic from the name."),
		gu.HStack(
			gu.LetterAvatar("Alice", gu.AvatarOpts{Size: "lg"}),
			gu.LetterAvatar("Bob", gu.AvatarOpts{Size: "lg"}),
			gu.LetterAvatar("Charlie", gu.AvatarOpts{Size: "lg"}),
			gu.LetterAvatar("Diana", gu.AvatarOpts{Size: "lg"}),
			gu.LetterAvatar("Eve", gu.AvatarOpts{Size: "lg"}),
			gu.LetterAvatar("Frank", gu.AvatarOpts{Size: "lg"}),
		),

		section("Avatar Group", "Overlapping avatar group with optional max display count."),
		gu.AvatarGroup([]g.ComponentFunc{
			gu.LetterAvatar("Alice", gu.AvatarOpts{}),
			gu.LetterAvatar("Bob", gu.AvatarOpts{}),
			gu.LetterAvatar("Charlie", gu.AvatarOpts{}),
			gu.LetterAvatar("Diana", gu.AvatarOpts{}),
			gu.LetterAvatar("Eve", gu.AvatarOpts{}),
		}, gu.AvatarGroupOpts{Max: 3}),

		section("Live Avatar (Click)", "Click on avatars to see the event payload."),
		gu.HStack(
			gu.ImageAvatar("https://i.pravatar.cc/64?u=10", gu.AvatarOpts{
				Size: "lg", Alt: "Clickable", ClickEvent: "avatarClick",
			}),
			gu.LetterAvatar("Alice", gu.AvatarOpts{
				Size: "lg", ClickEvent: "avatarClick",
			}),
			gu.LetterAvatar("Bob", gu.AvatarOpts{
				Size: "lg", ClickEvent: "avatarClick",
			}),
		),
	}

	if v.AvatarClickInfo != "" {
		items = append(items, gd.Div(gp.Attr("style", "margin-top:8px"),
			gu.Alert("Avatar clicked: "+v.AvatarClickInfo, "info"),
		))
	}

	return gu.Stack(items...)
}

func section(title, desc string) g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Attr("style", "font-size:18px;font-weight:600;margin:0 0 4px 0"), gp.Value(title)),
		gd.P(gp.Attr("style", "color:var(--g-text-secondary);margin:0 0 16px 0;font-size:13px"), gp.Value(desc)),
	)
}

func sampleRows() [][]string {
	return [][]string{
		{"Alice Johnson", "alice@example.com", "Admin", "Active"},
		{"Bob Smith", "bob@example.com", "Editor", "Active"},
		{"Charlie Brown", "charlie@example.com", "Viewer", "Inactive"},
		{"Diana Prince", "diana@example.com", "Editor", "Active"},
		{"Eve Adams", "eve@example.com", "Viewer", "Active"},
		{"Frank Castle", "frank@example.com", "Admin", "Active"},
		{"Grace Hopper", "grace@example.com", "Editor", "Inactive"},
		{"Henry Ford", "henry@example.com", "Viewer", "Active"},
		{"Iris West", "iris@example.com", "Editor", "Active"},
		{"Jack Ryan", "jack@example.com", "Viewer", "Active"},
		{"Karen Page", "karen@example.com", "Admin", "Inactive"},
		{"Leo Messi", "leo@example.com", "Viewer", "Active"},
	}
}

func main() {
	addr := flag.String("addr", ":8900", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &CatalogView{} }, opts...))
	log.Printf("catalog running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
