package main

import (
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

type AdminView struct {
	Page          string
	UserTablePage int
	SortCol       string
	SortDir       string
	ModalOpen     bool
	SelectedUser  int
	ConfirmOpen   bool
	DeleteTarget  int
	ToastVisible  bool
	ToastMessage  string
	ToastVariant  string

	// Settings
	ItemsPerPage int
	NotifyFreq   int

	// Calendar
	CalYear     int
	CalMonth    time.Month
	CalSelected *time.Time

	// Messages
	ChatMessages []gu.ChatMessage
	ChatDraft    string
}

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

func (v *AdminView) Mount(params gl.Params) error {
	v.Page = "dashboard"
	v.SortCol = "name"
	v.SortDir = "asc"
	v.SelectedUser = -1
	v.DeleteTarget = -1
	v.ItemsPerPage = 10
	v.NotifyFreq = 30
	now := time.Now()
	v.CalYear = now.Year()
	v.CalMonth = now.Month()
	v.ChatMessages = []gu.ChatMessage{
		{Author: "System", Content: "Welcome to Admin Messages.", Timestamp: "09:00", Sent: false},
		{Author: "Alice Johnson", Content: "The new user report is ready for review.", Timestamp: "09:15", Sent: false, Avatar: "A"},
		{Content: "Thanks, I'll take a look now.", Timestamp: "09:16", Sent: true},
	}
	if p := params["page"]; p != "" {
		v.Page = p
	}
	return nil
}

func (v *AdminView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Admin Panel"),
			gu.Theme(),
		),
		gd.Body(
			gu.AdminShell(
				v.renderSidebar(),
				v.renderContent(),
			),
			expr.If(v.ToastVisible,
				gul.Toast(v.ToastMessage, v.ToastVariant, "dismissToast"),
			),
		),
	}
}

func (v *AdminView) renderSidebar() g.ComponentFunc {
	return gu.Sidebar(
		gu.SidebarHeader("Admin Panel"),
		gu.SidebarLink("#", "Dashboard", v.Page == "dashboard",
			gl.Click("nav"), gl.ClickValue("dashboard")),
		gu.SidebarLink("#", "Users", v.Page == "users",
			gl.Click("nav"), gl.ClickValue("users")),
		gu.SidebarLink("#", "Messages", v.Page == "messages",
			gl.Click("nav"), gl.ClickValue("messages")),
		gu.SidebarDivider(),
		gu.SidebarLink("#", "Settings", v.Page == "settings",
			gl.Click("nav"), gl.ClickValue("settings")),
	)
}

func (v *AdminView) renderContent() g.ComponentFunc {
	var body g.ComponentFunc
	switch v.Page {
	case "users":
		body = v.pageUsers()
	case "messages":
		body = v.pageMessages()
	case "settings":
		body = v.pageSettings()
	default:
		body = v.pageDashboard()
	}
	return body
}

func (v *AdminView) pageDashboard() g.ComponentFunc {
	return gd.Div(
		gu.PageHeader("Dashboard",
			gu.Breadcrumb(gu.BreadcrumbItem{Label: "Dashboard", Href: ""}),
		),
		gd.Div(gp.Class("g-page-body"),
			gu.Stack(
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
							expr.If(v.CalSelected != nil,
								gd.Div(gp.Attr("style", "margin-top:8px"),
									gu.Badge(v.CalSelected.Format("2006-01-02"), "dark"),
								),
							),
						),
					),
				),
			),
		),
	)
}

func (v *AdminView) pageUsers() g.ComponentFunc {
	rows := userRows()
	pageSize := 8
	start := v.UserTablePage * pageSize
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	pageRows := rows[start:end]

	return gd.Div(
		gu.PageHeader("User Management",
			gu.Breadcrumb(
				gu.BreadcrumbItem{Label: "Dashboard", Href: "#"},
				gu.BreadcrumbItem{Label: "Users", Href: ""},
			),
		),
		gd.Div(gp.Class("g-page-body"),
			gul.DataTable(gul.DataTableOpts{
				Columns: []gul.Column{
					{Key: "name", Label: "Name", Sortable: true},
					{Key: "email", Label: "Email", Sortable: true},
					{Key: "role", Label: "Role", Sortable: true},
					{Key: "status", Label: "Status", Sortable: false},
					{Key: "actions", Label: "Actions", Sortable: false},
				},
				Rows:      pageRows,
				SortCol:   v.SortCol,
				SortDir:   v.SortDir,
				SortEvent: "sort",
				Page:      v.UserTablePage,
				PageSize:  pageSize,
				Total:     len(users),
				PageEvent: "userPage",
			}),
		),
		v.renderUserModal(),
		gul.Confirm(v.ConfirmOpen, "Delete User",
			"Are you sure you want to delete this user? This action cannot be undone.",
			"doDelete", "cancelDelete"),
	)
}

func (v *AdminView) renderUserModal() g.ComponentFunc {
	if v.SelectedUser < 0 || v.SelectedUser >= len(users) {
		return gul.Modal(false, "closeUserModal")
	}
	u := users[v.SelectedUser]
	return gul.Modal(v.ModalOpen, "closeUserModal",
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

func (v *AdminView) pageMessages() g.ComponentFunc {
	var msgViews []g.ComponentFunc
	for _, m := range v.ChatMessages {
		msgViews = append(msgViews, gu.ChatMessageView(m))
	}

	return gd.Div(
		gu.PageHeader("Messages",
			gu.Breadcrumb(
				gu.BreadcrumbItem{Label: "Dashboard", Href: "#"},
				gu.BreadcrumbItem{Label: "Messages", Href: ""},
			),
		),
		gd.Div(gp.Class("g-page-body"),
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
		),
	)
}

func (v *AdminView) pageSettings() g.ComponentFunc {
	min0, max100 := 0, 100
	min5, max50 := 5, 50

	return gd.Div(
		gu.PageHeader("Settings",
			gu.Breadcrumb(
				gu.BreadcrumbItem{Label: "Dashboard", Href: "#"},
				gu.BreadcrumbItem{Label: "Settings", Href: ""},
			),
		),
		gd.Div(gp.Class("g-page-body"),
			gu.Stack(
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
			),
		),
	)
}

func (v *AdminView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "nav":
		v.Page = payload["value"]
		v.ModalOpen = false
		v.ConfirmOpen = false
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
	case "userPage":
		fmt.Sscanf(payload["value"], "%d", &v.UserTablePage)
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

	// Messages
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

func activityRow(name, action, timeAgo string) g.ComponentFunc {
	return gd.Tr(
		gd.Td(gp.Value(name)),
		gd.Td(gp.Value(action)),
		gd.Td(gp.Attr("style", "color:var(--g-text-tertiary)"), gp.Value(timeAgo)),
	)
}

func statusBadge(status string) g.ComponentFunc {
	if status == "Active" {
		return gu.Badge("Active", "dark")
	}
	return gu.Badge("Inactive", "outline")
}

func userRows() [][]string {
	var rows [][]string
	for i, u := range users {
		rows = append(rows, []string{u.Name, u.Email, u.Role, u.Status, fmt.Sprintf("%d", i)})
	}
	return rows
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

func main() {
	addr := flag.String("addr", ":8910", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func() gl.View { return &AdminView{} }, opts...))
	log.Printf("admin running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
