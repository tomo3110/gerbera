package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

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
		navLink(v, "misc", "Misc"),
		navLink(v, "layout", "Layout"),
		navDividerLabel("Live Widgets"),
		navLink(v, "modal", "Modal"),
		navLink(v, "toast", "Toast"),
		navLink(v, "datatable", "DataTable"),
		navLink(v, "dropdown", "Dropdown"),
		navLink(v, "tabs", "Tabs"),
		navLink(v, "drawer", "Drawer"),
		navLink(v, "searchselect", "SearchSelect"),
		navLink(v, "confirm", "Confirm"),
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
	case "misc":
		body = v.pageMisc()
	case "layout":
		body = v.pageLayout()
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
			gu.StatCard("Static Widgets", "17"),
			gu.StatCard("Live Widgets", "8"),
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

	http.Handle("/", gl.Handler(func() gl.View { return &CatalogView{} }, opts...))
	log.Printf("catalog running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
