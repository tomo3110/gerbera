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
	_ "github.com/tomo3110/gerbera/assets"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

// ---------------------------------------------------------------------------
// Shared data
// ---------------------------------------------------------------------------

type user struct {
	Name   string
	Email  string
	Role   string
	Status string
}

var users = []user{
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
	{"Maria Garcia", "maria@example.com", "Editor", "Active"},
	{"Nick Fury", "nick@example.com", "Admin", "Active"},
	{"Olivia Pope", "olivia@example.com", "Editor", "Active"},
}

func countActive() int {
	n := 0
	for _, u := range users {
		if u.Status == "Active" {
			n++
		}
	}
	return n
}

func countByRole(role string) int {
	n := 0
	for _, u := range users {
		if u.Role == role {
			n++
		}
	}
	return n
}

func userRows() [][]string {
	var rows [][]string
	for i, u := range users {
		rows = append(rows, []string{u.Name, u.Email, u.Role, u.Status, fmt.Sprintf("%d", i)})
	}
	return rows
}

// ---------------------------------------------------------------------------
// AdminView — single LiveView managing all pages
// ---------------------------------------------------------------------------

type AdminView struct {
	Page string

	// Dashboard state
	CalYear     int
	CalMonth    time.Month
	CalSelected *time.Time

	// Users state
	TablePage    int
	SortCol      string
	SortDir      string
	ModalOpen    bool
	SelectedUser int
	ConfirmOpen  bool
	DeleteTarget int
	ToastVisible bool
	ToastMessage string
	ToastVariant string

	// Messages state
	ChatMessages []gu.ChatMessage
	ChatDraft    string

	// Settings state
	ItemsPerPage int
	NotifyFreq   int
}

func (v *AdminView) Mount(_ gl.Params) error {
	v.Page = "dashboard"
	v.initDashboard()
	v.initUsers()
	v.initMessages()
	v.initSettings()
	return nil
}

func (v *AdminView) initDashboard() {
	now := time.Now()
	v.CalYear = now.Year()
	v.CalMonth = now.Month()
}

func (v *AdminView) initUsers() {
	v.SortCol = "name"
	v.SortDir = "asc"
	v.SelectedUser = -1
	v.DeleteTarget = -1
}

func (v *AdminView) initMessages() {
	v.ChatMessages = []gu.ChatMessage{
		{Author: "System", Content: "Welcome to Admin Messages.", Timestamp: "09:00", Sent: false},
		{Author: "Alice Johnson", Content: "The new user report is ready for review.", Timestamp: "09:15", Sent: false, Avatar: "A"},
		{Content: "Thanks, I'll take a look now.", Timestamp: "09:16", Sent: true},
	}
}

func (v *AdminView) initSettings() {
	v.ItemsPerPage = 10
	v.NotifyFreq = 30
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func (v *AdminView) Render() g.Components {
	return g.Components{
		gd.Head(
			gd.Title("Admin Panel"),
			gu.Theme(),
		),
		gd.Body(
			gu.AdminShell(
				v.sidebar(),
				gd.Div(gp.Class("g-page-body"),
					expr.If(v.Page == "dashboard", v.renderDashboard()),
					expr.If(v.Page == "users", v.renderUsers()),
					expr.If(v.Page == "messages", v.renderMessages()),
					expr.If(v.Page == "settings", v.renderSettings()),
				),
			),
		),
	}
}

func (v *AdminView) sidebar() g.ComponentFunc {
	return gu.Sidebar(
		gu.SidebarHeader("Admin Panel"),
		gu.SidebarLink("#", "Dashboard", v.Page == "dashboard", gl.Click("nav"), gl.ClickValue("dashboard")),
		gu.SidebarLink("#", "Users", v.Page == "users", gl.Click("nav"), gl.ClickValue("users")),
		gu.SidebarLink("#", "Messages", v.Page == "messages", gl.Click("nav"), gl.ClickValue("messages")),
		gu.SidebarDivider(),
		gu.SidebarLink("#", "Settings", v.Page == "settings", gl.Click("nav"), gl.ClickValue("settings")),
	)
}

// ---------------------------------------------------------------------------
// Dashboard page
// ---------------------------------------------------------------------------

func (v *AdminView) renderDashboard() g.ComponentFunc {
	return func(n g.Node) {
		gu.Stack(
			gu.PageHeader("Dashboard",
				gu.Breadcrumb(gu.BreadcrumbItem{Label: "Dashboard"}),
			),
			gu.Grid(gu.GridCols4,
				gu.StatCard("Total Users", fmt.Sprintf("%d", len(users))),
				gu.StatCard("Active Users", fmt.Sprintf("%d", countActive())),
				gu.StatCard("Admins", fmt.Sprintf("%d", countByRole("Admin"))),
				gu.StatCard("Editors", fmt.Sprintf("%d", countByRole("Editor"))),
			),
			gu.Grid(gu.GridCols2,
				gu.Card(
					gu.CardHeader("Recent Activity"),
					gu.StyledTable(
						gu.THead("User", "Action", "Time"),
						gd.Tbody(
							activityRow("Alice Johnson", "Updated profile", "2 min ago"),
							activityRow("Bob Smith", "Uploaded document", "15 min ago"),
							activityRow("Diana Prince", "Created post", "1 hour ago"),
							activityRow("Frank Castle", "Deleted comment", "3 hours ago"),
							activityRow("Grace Hopper", "Logged in", "5 hours ago"),
						),
					),
				),
				gu.Card(
					gu.CardHeader("Calendar"),
					gd.Div(gp.Class("g-page-body"),
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
						func(n g.Node) {
							if v.CalSelected != nil {
								gd.Div(gp.Attr("style", "margin-top:8px"),
									gu.Badge(v.CalSelected.Format("2006-01-02"), "dark"),
								)(n)
							}
						},
					),
				),
			),
		)(n)
	}
}

func activityRow(name, action, timeAgo string) g.ComponentFunc {
	return gd.Tr(
		gd.Td(gp.Value(name)),
		gd.Td(gp.Value(action)),
		gd.Td(gp.Attr("style", "color:var(--g-text-tertiary)"), gp.Value(timeAgo)),
	)
}

// ---------------------------------------------------------------------------
// Users page
// ---------------------------------------------------------------------------

func (v *AdminView) renderUsers() g.ComponentFunc {
	return func(n g.Node) {
		rows := userRows()
		pageSize := 8
		start := v.TablePage * pageSize
		end := start + pageSize
		if end > len(rows) {
			end = len(rows)
		}
		pageRows := rows[start:end]

		gd.Div(
			gu.PageHeader("User Management",
				gu.Breadcrumb(
					gu.BreadcrumbItem{Label: "Dashboard"},
					gu.BreadcrumbItem{Label: "Users"},
				),
			),
			gul.DataTable(gul.DataTableOpts{
				Columns: []gul.Column{
					{Key: "name", Label: "Name", Sortable: true},
					{Key: "email", Label: "Email", Sortable: true},
					{Key: "role", Label: "Role", Sortable: true},
					{Key: "status", Label: "Status"},
					{Key: "actions", Label: "Actions"},
				},
				Rows:      pageRows,
				SortCol:   v.SortCol,
				SortDir:   v.SortDir,
				SortEvent: "sort",
				Page:      v.TablePage,
				PageSize:  pageSize,
				Total:     len(users),
				PageEvent: "userPage",
			}),
			renderUserModal(v.ModalOpen, v.SelectedUser),
			gul.Confirm(v.ConfirmOpen, "Delete User",
				"Are you sure you want to delete this user? This action cannot be undone.",
				"doDelete", "cancelDelete"),
			gul.Toast(v.ToastVisible, v.ToastMessage, v.ToastVariant, "dismissToast"),
		)(n)
	}
}

func renderUserModal(open bool, idx int) g.ComponentFunc {
	if idx < 0 || idx >= len(users) {
		return gul.Modal(false, "closeUserModal")
	}
	u := users[idx]
	return gul.Modal(open, "closeUserModal",
		gul.ModalHeader("User Details", "closeUserModal"),
		gul.ModalBody(
			gu.FormGroup(
				gu.FormLabel("Name", "detail-name"),
				gu.FormInput("name", gp.ID("detail-name"), gp.Attr("value", u.Name), gp.Readonly(true)),
			),
			gu.FormGroup(
				gu.FormLabel("Email", "detail-email"),
				gu.FormInput("email", gp.ID("detail-email"), gp.Attr("value", u.Email), gp.Readonly(true)),
			),
			gu.FormGroup(
				gu.FormLabel("Role", "detail-role"),
				gu.FormInput("role", gp.ID("detail-role"), gp.Attr("value", u.Role), gp.Readonly(true)),
			),
			gu.Row(gu.AlignCenter,
				gd.Span(gp.Attr("style", "font-size:13px;color:var(--g-text-secondary)"), gp.Value("Status:")),
				statusBadge(u.Status),
			),
		),
		gul.ModalFooter(
			gu.Button("Close", gu.ButtonOutline, gl.Click("closeUserModal")),
		),
	)
}

func statusBadge(status string) g.ComponentFunc {
	if status == "Active" {
		return gu.Badge("Active", "dark")
	}
	return gu.Badge("Inactive", "outline")
}

// ---------------------------------------------------------------------------
// Messages page
// ---------------------------------------------------------------------------

func (v *AdminView) renderMessages() g.ComponentFunc {
	return func(n g.Node) {
		var msgViews g.Components
		for _, m := range v.ChatMessages {
			msgViews = append(msgViews, gu.ChatMessageView(m))
		}

		gd.Div(
			gu.PageHeader("Messages",
				gu.Breadcrumb(
					gu.BreadcrumbItem{Label: "Dashboard"},
					gu.BreadcrumbItem{Label: "Messages"},
				),
			),
			gu.Card(
				gu.CardHeader("Team Chat"),
				gd.Div(gp.Attr("style", "height:400px;display:flex;flex-direction:column"),
					gd.Div(gp.Attr("style", "flex:1;overflow-y:auto"),
						gu.ChatContainer(msgViews...),
					),
					gu.ChatInput("chatMsg", v.ChatDraft, gu.ChatInputOpts{
						Placeholder:  "Send a message to the team...",
						SendEvent:    "chatSend",
						InputEvent:   "chatInput",
						KeydownEvent: "chatKeydown",
					}),
				),
			),
		)(n)
	}
}

// ---------------------------------------------------------------------------
// Settings page
// ---------------------------------------------------------------------------

func (v *AdminView) renderSettings() g.ComponentFunc {
	return func(n g.Node) {
		min0, max100 := 0, 100
		min5, max50 := 5, 50

		gu.Stack(
			gu.PageHeader("Settings",
				gu.Breadcrumb(
					gu.BreadcrumbItem{Label: "Dashboard"},
					gu.BreadcrumbItem{Label: "Settings"},
				),
			),
			gu.Card(
				gu.CardHeader("General Settings"),
				gd.Div(gp.Class("g-page-body"),
					gu.FormGroup(
						gu.FormLabel("Site Name", "site-name"),
						gu.FormInput("site-name", gp.ID("site-name"), gp.Attr("value", "My Admin Panel")),
					),
					gu.FormGroup(
						gu.FormLabel("Language", "lang"),
						gu.FormSelect("lang", []gu.FormOption{
							{Value: "ja", Label: "Japanese"},
							{Value: "en", Label: "English"},
						}, gp.ID("lang")),
					),
					gu.Button("Save", gu.ButtonPrimary),
				),
			),
			gu.Card(
				gu.CardHeader("Display Settings"),
				gd.Div(gp.Class("g-page-body"),
					gu.FormGroup(
						gu.FormLabel("Items per page", "items-per-page"),
						gu.NumberInput("items-per-page", v.ItemsPerPage, gu.NumberInputOpts{
							Min:            &min5,
							Max:            &max50,
							Step:           5,
							IncrementEvent: "settingsItemsInc",
							DecrementEvent: "settingsItemsDec",
						}),
					),
					gu.FormGroup(
						gu.FormLabel("Notification frequency (minutes)", "notify-freq"),
						gu.Slider("notify-freq", v.NotifyFreq, gu.SliderOpts{
							Min:        min0,
							Max:        max100,
							Step:       5,
							Label:      "Frequency",
							InputEvent: "settingsNotifyFreq",
						}),
					),
				),
			),
		)(n)
	}
}

// ---------------------------------------------------------------------------
// HandleEvent — all events for all pages
// ---------------------------------------------------------------------------

func (v *AdminView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	// Navigation
	case "nav":
		v.Page = payload["value"]

	// Dashboard — calendar
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

	// Users — table, modal, confirm, toast
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
	case "userPage":
		fmt.Sscanf(payload["value"], "%d", &v.TablePage)
	case "viewUser":
		fmt.Sscanf(payload["value"], "%d", &v.SelectedUser)
		v.ModalOpen = true
	case "closeUserModal":
		v.ModalOpen = false
	case "confirmDeleteUser":
		fmt.Sscanf(payload["value"], "%d", &v.DeleteTarget)
		v.ConfirmOpen = true
	case "doDelete":
		v.ConfirmOpen = false
		v.ToastVisible = true
		v.ToastMessage = "User deleted successfully."
		v.ToastVariant = "success"
	case "cancelDelete":
		v.ConfirmOpen = false
	case "dismissToast":
		v.ToastVisible = false

	// Messages — chat
	case "chatInput":
		v.ChatDraft = payload["value"]
	case "chatSend", "chatKeydown":
		if event == "chatKeydown" && payload["key"] != "Enter" {
			return nil
		}
		if strings.TrimSpace(v.ChatDraft) != "" {
			v.ChatMessages = append(v.ChatMessages, gu.ChatMessage{
				Content:   v.ChatDraft,
				Timestamp: time.Now().Format("15:04"),
				Sent:      true,
			})
			v.ChatDraft = ""
		}

	// Settings
	case "settingsItemsInc":
		v.ItemsPerPage += 5
		if v.ItemsPerPage > 50 {
			v.ItemsPerPage = 50
		}
	case "settingsItemsDec":
		v.ItemsPerPage -= 5
		if v.ItemsPerPage < 5 {
			v.ItemsPerPage = 5
		}
	case "settingsNotifyFreq":
		fmt.Sscanf(payload["value"], "%d", &v.NotifyFreq)
	}
	return nil
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	addr := flag.String("addr", ":8910", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	mux := http.NewServeMux()
	mux.Handle("/", gl.Handler(func(_ context.Context) gl.View {
		return &AdminView{}
	}, opts...))

	log.Printf("admin running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, g.Serve(mux)))
}
