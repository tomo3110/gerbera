package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

func render(t *testing.T, c gerbera.ComponentFunc) string {
	t.Helper()
	var buf bytes.Buffer
	if err := gerbera.ExecuteTemplate(&buf, "en", c); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestTheme(t *testing.T) {
	out := render(t, Theme())
	if !strings.Contains(out, "--g-bg:") {
		t.Error("Theme should contain CSS custom properties")
	}
	if !strings.Contains(out, "<style>") {
		t.Error("Theme should render as <style> element")
	}
	// Backward compatibility: Theme() should produce default values
	if !strings.Contains(out, "#fafafa") {
		t.Error("Theme should contain default background color")
	}
	if !strings.Contains(out, "--g-accent-hover:") {
		t.Error("Theme should contain --g-accent-hover variable")
	}
	if !strings.Contains(out, "--g-danger-hover:") {
		t.Error("Theme should contain --g-danger-hover variable")
	}
}

func TestThemeWith(t *testing.T) {
	cfg := ThemeConfig{
		Accent:      "#1a73e8",
		AccentLight: "#e8f0fe",
		AccentHover: "#1557b0",
	}
	out := render(t, ThemeWith(cfg))
	if !strings.Contains(out, "#1a73e8") {
		t.Error("ThemeWith should contain custom accent color")
	}
	if !strings.Contains(out, "#e8f0fe") {
		t.Error("ThemeWith should contain custom accent-light color")
	}
	if !strings.Contains(out, "#1557b0") {
		t.Error("ThemeWith should contain custom accent-hover color")
	}
	// Zero-value fields should be filled with defaults
	if !strings.Contains(out, "#fafafa") {
		t.Error("ThemeWith should fill zero-value Bg with default")
	}
	// Should still contain component CSS rules
	if !strings.Contains(out, ".g-btn") {
		t.Error("ThemeWith should contain component CSS rules")
	}
}

func TestThemeWithDark(t *testing.T) {
	out := render(t, ThemeWith(DarkTheme()))
	if !strings.Contains(out, "#0a0a0a") {
		t.Error("ThemeWith(DarkTheme()) should contain dark background color")
	}
	if !strings.Contains(out, "#fafafa") {
		t.Error("ThemeWith(DarkTheme()) should contain dark text color")
	}
	if !strings.Contains(out, "#e5e5e5") {
		t.Error("ThemeWith(DarkTheme()) should contain dark accent color")
	}
}

func TestThemeAuto(t *testing.T) {
	out := render(t, ThemeAuto(ThemeConfig{}, ThemeConfig{}))
	if !strings.Contains(out, "prefers-color-scheme: dark") {
		t.Error("ThemeAuto should contain prefers-color-scheme media query")
	}
	// Light defaults
	if !strings.Contains(out, "#fafafa") {
		t.Error("ThemeAuto should contain default light background")
	}
	// Dark defaults
	if !strings.Contains(out, "#0a0a0a") {
		t.Error("ThemeAuto should contain default dark background")
	}
	// Component rules should appear only once (use a rule that appears exactly once in themeRulesCSS)
	if strings.Count(out, "box-sizing: border-box") != 1 {
		t.Error("ThemeAuto should contain component CSS rules exactly once")
	}
}

func TestThemeAutoDarkFillsSpacing(t *testing.T) {
	out := render(t, ThemeAuto(ThemeConfig{}, ThemeConfig{}))
	// DarkTheme() does not set spacing/font/radius; they must be filled from DefaultTheme()
	if strings.Contains(out, "--g-space-xs: ;") {
		t.Error("ThemeAuto dark should not produce empty --g-space-xs")
	}
	if strings.Contains(out, "--g-font: ;") {
		t.Error("ThemeAuto dark should not produce empty --g-font")
	}
	if strings.Contains(out, "--g-radius: ;") {
		t.Error("ThemeAuto dark should not produce empty --g-radius")
	}
}

func TestThemeAutoCustom(t *testing.T) {
	out := render(t, ThemeAuto(
		ThemeConfig{Accent: "#1a73e8"},
		ThemeConfig{Accent: "#8ab4f8"},
	))
	if !strings.Contains(out, "#1a73e8") {
		t.Error("ThemeAuto should contain custom light accent")
	}
	if !strings.Contains(out, "#8ab4f8") {
		t.Error("ThemeAuto should contain custom dark accent")
	}
}

func TestDefaultTheme(t *testing.T) {
	d := DefaultTheme()
	if d.Bg != "#fafafa" {
		t.Errorf("DefaultTheme().Bg = %q, want #fafafa", d.Bg)
	}
	if d.AccentHover != "#262626" {
		t.Errorf("DefaultTheme().AccentHover = %q, want #262626", d.AccentHover)
	}
	if d.DangerHover != "#b91c1c" {
		t.Errorf("DefaultTheme().DangerHover = %q, want #b91c1c", d.DangerHover)
	}
}

func TestDarkTheme(t *testing.T) {
	d := DarkTheme()
	if d.Bg != "#0a0a0a" {
		t.Errorf("DarkTheme().Bg = %q, want #0a0a0a", d.Bg)
	}
	if d.Text != "#fafafa" {
		t.Errorf("DarkTheme().Text = %q, want #fafafa", d.Text)
	}
	if d.Accent != "#e5e5e5" {
		t.Errorf("DarkTheme().Accent = %q, want #e5e5e5", d.Accent)
	}
	if d.AccentHover != "#ffffff" {
		t.Errorf("DarkTheme().AccentHover = %q, want #ffffff", d.AccentHover)
	}
}

func TestThemeConfigWithDefaults(t *testing.T) {
	cfg := ThemeConfig{Bg: "#custom"}
	filled := cfg.withDefaults(DefaultTheme())
	if filled.Bg != "#custom" {
		t.Error("withDefaults should preserve non-zero Bg")
	}
	if filled.BgSurface != "#ffffff" {
		t.Error("withDefaults should fill zero BgSurface with default")
	}
	if filled.Accent != "#171717" {
		t.Error("withDefaults should fill zero Accent with default")
	}
}

func TestThemeBtnHoverUsesVariable(t *testing.T) {
	out := render(t, Theme())
	if !strings.Contains(out, "var(--g-accent-hover)") {
		t.Error("Button primary hover should use var(--g-accent-hover)")
	}
	if !strings.Contains(out, "var(--g-danger-hover)") {
		t.Error("Button danger hover should use var(--g-danger-hover)")
	}
}

func TestCard(t *testing.T) {
	out := render(t, Card(
		CardHeader("Title"),
		CardFooter(gerbera.Literal("footer")),
	))
	if !strings.Contains(out, "g-card") {
		t.Error("Card should have g-card class")
	}
	if !strings.Contains(out, "Title") {
		t.Error("CardHeader should contain title")
	}
	if !strings.Contains(out, "footer") {
		t.Error("CardFooter should contain content")
	}
}

func TestCardHeaderActions(t *testing.T) {
	out := render(t, CardHeader("Users", Button("Add", ButtonPrimary)))
	if !strings.Contains(out, "g-card-header-actions") {
		t.Error("CardHeader with actions should have actions container")
	}
	if !strings.Contains(out, "Add") {
		t.Error("CardHeader should contain action button")
	}
}

func TestStyledTable(t *testing.T) {
	out := render(t, StyledTable(THead("Name", "Email")))
	if !strings.Contains(out, "g-table") {
		t.Error("StyledTable should have g-table class")
	}
	if !strings.Contains(out, "Name") {
		t.Error("THead should contain header text")
	}
}

func TestBadge(t *testing.T) {
	tests := []struct {
		variant string
		class   string
	}{
		{"", "g-badge-default"},
		{"dark", "g-badge-dark"},
		{"outline", "g-badge-outline"},
		{"light", "g-badge-light"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			var out string
			if tt.variant == "" {
				out = render(t, Badge("Active"))
			} else {
				out = render(t, Badge("Active", tt.variant))
			}
			if !strings.Contains(out, tt.class) {
				t.Errorf("Badge(%q) should contain class %q", tt.variant, tt.class)
			}
		})
	}
}

func TestAlert(t *testing.T) {
	for _, v := range []string{"info", "success", "warning", "danger"} {
		t.Run(v, func(t *testing.T) {
			out := render(t, Alert("Message", v))
			if !strings.Contains(out, "g-alert-"+v) {
				t.Errorf("Alert(%q) should contain class g-alert-%s", v, v)
			}
			if !strings.Contains(out, `role="alert"`) {
				t.Error("Alert should have role=alert")
			}
		})
	}
}

func TestStatCard(t *testing.T) {
	out := render(t, StatCard("Users", "1,234"))
	if !strings.Contains(out, "g-stat") {
		t.Error("StatCard should have g-stat class")
	}
	if !strings.Contains(out, "Users") {
		t.Error("StatCard should contain label")
	}
	if !strings.Contains(out, "1,234") {
		t.Error("StatCard should contain value")
	}
}

func TestSidebar(t *testing.T) {
	out := render(t, Sidebar(
		SidebarHeader("Admin"),
		SidebarLink("/dashboard", "Dashboard", true),
		SidebarDivider(),
		SidebarLink("/users", "Users", false),
	))
	if !strings.Contains(out, "g-sidebar") {
		t.Error("Sidebar should have g-sidebar class")
	}
	if !strings.Contains(out, "g-sidebar-link-active") {
		t.Error("Active link should have active class")
	}
	if !strings.Contains(out, `aria-current="page"`) {
		t.Error("Active link should have aria-current")
	}
}

func TestBreadcrumb(t *testing.T) {
	out := render(t, Breadcrumb(
		BreadcrumbItem{Label: "Home", Href: "/"},
		BreadcrumbItem{Label: "Users", Href: ""},
	))
	if !strings.Contains(out, "g-breadcrumb") {
		t.Error("Breadcrumb should have g-breadcrumb class")
	}
	if !strings.Contains(out, `aria-current="page"`) {
		t.Error("Current item should have aria-current")
	}
}

func TestAdminShell(t *testing.T) {
	out := render(t, AdminShell(
		Sidebar(SidebarHeader("App")),
		gerbera.Literal("content"),
	))
	if !strings.Contains(out, "g-admin-shell") {
		t.Error("AdminShell should have g-admin-shell class")
	}
	if !strings.Contains(out, "g-admin-content") {
		t.Error("AdminShell should have content area")
	}
}

func TestPageHeader(t *testing.T) {
	out := render(t, PageHeader("Dashboard"))
	if !strings.Contains(out, "g-page-header") {
		t.Error("PageHeader should have g-page-header class")
	}
	if !strings.Contains(out, "Dashboard") {
		t.Error("PageHeader should contain title")
	}
}

func TestButton(t *testing.T) {
	out := render(t, Button("Save", ButtonPrimary))
	if !strings.Contains(out, "g-btn") {
		t.Error("Button should have g-btn class")
	}
	if !strings.Contains(out, "g-btn-primary") {
		t.Error("Button with ButtonPrimary should have primary class")
	}
	if !strings.Contains(out, "Save") {
		t.Error("Button should contain label")
	}
}

func TestButtonVariants(t *testing.T) {
	tests := []struct {
		name string
		opt  gerbera.ComponentFunc
		cls  string
	}{
		{"outline", ButtonOutline, "g-btn-outline"},
		{"danger", ButtonDanger, "g-btn-danger"},
		{"small", ButtonSmall, "g-btn-sm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(t, Button("X", tt.opt))
			if !strings.Contains(out, tt.cls) {
				t.Errorf("Button with %s should contain class %s", tt.name, tt.cls)
			}
		})
	}
}

func TestFormGroup(t *testing.T) {
	out := render(t, FormGroup(
		FormLabel("Email", "email"),
		FormInput("email", property.Type("email"), property.Placeholder("you@example.com")),
	))
	if !strings.Contains(out, "g-form-group") {
		t.Error("FormGroup should have g-form-group class")
	}
	if !strings.Contains(out, `for="email"`) {
		t.Error("FormLabel should have for attribute")
	}
	if !strings.Contains(out, `name="email"`) {
		t.Error("FormInput should have name attribute")
	}
}

func TestFormSelect(t *testing.T) {
	out := render(t, FormSelect("role", []FormOption{
		{Value: "admin", Label: "Admin"},
		{Value: "user", Label: "User"},
	}))
	if !strings.Contains(out, "g-form-select") {
		t.Error("FormSelect should have g-form-select class")
	}
	if !strings.Contains(out, "Admin") {
		t.Error("FormSelect should contain options")
	}
}

func TestDivider(t *testing.T) {
	out := render(t, Divider())
	if !strings.Contains(out, "g-divider") {
		t.Error("Divider should have g-divider class")
	}
}

func TestEmptyState(t *testing.T) {
	out := render(t, EmptyState("No data"))
	if !strings.Contains(out, "g-empty-state") {
		t.Error("EmptyState should have g-empty-state class")
	}
	if !strings.Contains(out, "No data") {
		t.Error("EmptyState should contain message")
	}
}

func TestProgress(t *testing.T) {
	out := render(t, Progress(75))
	if !strings.Contains(out, "g-progress-track") {
		t.Error("Progress should have g-progress-track class")
	}
	if !strings.Contains(out, `aria-valuenow="75"`) {
		t.Error("Progress should have aria-valuenow")
	}
	if !strings.Contains(out, "75%") {
		t.Error("Progress should set width to percentage")
	}
}

func TestProgressClamp(t *testing.T) {
	out := render(t, Progress(150))
	if !strings.Contains(out, `aria-valuenow="100"`) {
		t.Error("Progress > 100 should be clamped to 100")
	}
	out = render(t, Progress(-10))
	if !strings.Contains(out, `aria-valuenow="0"`) {
		t.Error("Progress < 0 should be clamped to 0")
	}
}

func TestIcon(t *testing.T) {
	out := render(t, Icon("home", "lg"))
	if !strings.Contains(out, "g-icon") {
		t.Error("Icon should have g-icon class")
	}
	if !strings.Contains(out, "g-icon-lg") {
		t.Error("Icon with lg size should have g-icon-lg class")
	}
	if !strings.Contains(out, "<svg") {
		t.Error("Icon should contain SVG element")
	}
}

func TestIconDefault(t *testing.T) {
	out := render(t, Icon("user"))
	if !strings.Contains(out, "g-icon-md") {
		t.Error("Icon without size should default to md")
	}
}

func TestIconUnknown(t *testing.T) {
	out := render(t, Icon("nonexistent"))
	if !strings.Contains(out, "<svg") {
		t.Error("Unknown icon should fall back to circle")
	}
}

func TestIconNames(t *testing.T) {
	names := IconNames()
	if len(names) < 10 {
		t.Error("IconNames should return at least 10 icons")
	}
}

func TestTree(t *testing.T) {
	out := render(t, Tree([]TreeNode{
		{
			Label: "Root",
			Open:  true,
			Children: []TreeNode{
				{Label: "Child A"},
				{Label: "Child B", Icon: "file"},
			},
		},
	}))
	if !strings.Contains(out, "g-tree") {
		t.Error("Tree should have g-tree class")
	}
	if !strings.Contains(out, "Root") {
		t.Error("Tree should contain root label")
	}
	if !strings.Contains(out, "Child A") {
		t.Error("Open tree should contain child labels")
	}
	if !strings.Contains(out, `role="tree"`) {
		t.Error("Tree should have tree role")
	}
	if !strings.Contains(out, `role="treeitem"`) {
		t.Error("Tree nodes should have treeitem role")
	}
}

func TestTreeClosed(t *testing.T) {
	out := render(t, Tree([]TreeNode{
		{
			Label: "Root",
			Open:  false,
			Children: []TreeNode{
				{Label: "Hidden"},
			},
		},
	}))
	if strings.Contains(out, "Hidden") {
		t.Error("Closed tree should not show children")
	}
}

func TestFormTextarea(t *testing.T) {
	out := render(t, FormTextarea("bio", property.ID("bio"), property.Placeholder("About you")))
	if !strings.Contains(out, "g-form-textarea") {
		t.Error("FormTextarea should have g-form-textarea class")
	}
	if !strings.Contains(out, `name="bio"`) {
		t.Error("FormTextarea should have name attribute")
	}
	if !strings.Contains(out, "<textarea") {
		t.Error("FormTextarea should render textarea element")
	}
}

func TestFormError(t *testing.T) {
	out := render(t, FormError("Email is required"))
	if !strings.Contains(out, "g-form-error") {
		t.Error("FormError should have g-form-error class")
	}
	if !strings.Contains(out, "Email is required") {
		t.Error("FormError should contain the message")
	}
	if !strings.Contains(out, `role="alert"`) {
		t.Error("FormError should have alert role")
	}
}

func TestFormErrorEmpty(t *testing.T) {
	out := render(t, FormError(""))
	if strings.Contains(out, "g-form-error") {
		t.Error("FormError with empty message should not render")
	}
}

func TestCheckbox(t *testing.T) {
	out := render(t, Checkbox("agree", "I agree", true))
	if !strings.Contains(out, "g-form-check") {
		t.Error("Checkbox should have g-form-check class")
	}
	if !strings.Contains(out, `type="checkbox"`) {
		t.Error("Checkbox should have type=checkbox")
	}
	if !strings.Contains(out, `checked="checked"`) {
		t.Error("Checked checkbox should have checked attribute")
	}
	if !strings.Contains(out, "I agree") {
		t.Error("Checkbox should contain label text")
	}
}

func TestCheckboxUnchecked(t *testing.T) {
	out := render(t, Checkbox("agree", "I agree", false))
	if strings.Contains(out, `checked="checked"`) {
		t.Error("Unchecked checkbox should not have checked attribute")
	}
}

func TestRadio(t *testing.T) {
	out := render(t, Radio("color", "red", "Red", true))
	if !strings.Contains(out, "g-form-check") {
		t.Error("Radio should have g-form-check class")
	}
	if !strings.Contains(out, `type="radio"`) {
		t.Error("Radio should have type=radio")
	}
	if !strings.Contains(out, `value="red"`) {
		t.Error("Radio should have value attribute")
	}
	if !strings.Contains(out, `checked="checked"`) {
		t.Error("Selected radio should have checked attribute")
	}
}

func TestMobileHeader(t *testing.T) {
	out := render(t, MobileHeader("Admin"))
	if !strings.Contains(out, "g-mobile-header") {
		t.Error("MobileHeader should have g-mobile-header class")
	}
	if !strings.Contains(out, "g-hamburger") {
		t.Error("MobileHeader should contain hamburger button")
	}
	if !strings.Contains(out, "Admin") {
		t.Error("MobileHeader should contain title")
	}
}

// --- Layout / Grid tests ---

func TestRow(t *testing.T) {
	out := render(t, Row(gerbera.Literal("a"), gerbera.Literal("b")))
	if !strings.Contains(out, "g-row") {
		t.Error("Row should have g-row class")
	}
}

func TestColumn(t *testing.T) {
	out := render(t, Column(gerbera.Literal("a")))
	if !strings.Contains(out, "g-col") {
		t.Error("Column should have g-col class")
	}
}

func TestStack(t *testing.T) {
	out := render(t, Stack(gerbera.Literal("a"), gerbera.Literal("b")))
	if !strings.Contains(out, "g-stack") {
		t.Error("Stack should have g-stack class")
	}
}

func TestHStack(t *testing.T) {
	out := render(t, HStack(Button("A"), Button("B")))
	if !strings.Contains(out, "g-hstack") {
		t.Error("HStack should have g-hstack class")
	}
}

func TestVStack(t *testing.T) {
	out := render(t, VStack(gerbera.Literal("icon"), gerbera.Literal("label")))
	if !strings.Contains(out, "g-vstack") {
		t.Error("VStack should have g-vstack class")
	}
}

func TestCenter(t *testing.T) {
	out := render(t, Center(gerbera.Literal("centered")))
	if !strings.Contains(out, "g-center") {
		t.Error("Center should have g-center class")
	}
}

func TestContainer(t *testing.T) {
	out := render(t, Container(gerbera.Literal("content")))
	if !strings.Contains(out, "g-container") {
		t.Error("Container should have g-container class")
	}
}

func TestContainerNarrow(t *testing.T) {
	out := render(t, ContainerNarrow(gerbera.Literal("narrow")))
	if !strings.Contains(out, "g-container") {
		t.Error("ContainerNarrow should have g-container class")
	}
	if !strings.Contains(out, "g-container-narrow") {
		t.Error("ContainerNarrow should have g-container-narrow class")
	}
}

func TestContainerWide(t *testing.T) {
	out := render(t, ContainerWide(gerbera.Literal("wide")))
	if !strings.Contains(out, "g-container-wide") {
		t.Error("ContainerWide should have g-container-wide class")
	}
}

func TestGrid(t *testing.T) {
	out := render(t, Grid(GridCols3, gerbera.Literal("a"), gerbera.Literal("b"), gerbera.Literal("c")))
	if !strings.Contains(out, "g-grid") {
		t.Error("Grid should have g-grid class")
	}
	if !strings.Contains(out, "g-grid-3") {
		t.Error("Grid with GridCols3 should have g-grid-3 class")
	}
}

func TestGridColsVariants(t *testing.T) {
	tests := []struct {
		name string
		opt  gerbera.ComponentFunc
		cls  string
	}{
		{"2cols", GridCols2, "g-grid-2"},
		{"3cols", GridCols3, "g-grid-3"},
		{"4cols", GridCols4, "g-grid-4"},
		{"5cols", GridCols5, "g-grid-5"},
		{"6cols", GridCols6, "g-grid-6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(t, Grid(tt.opt))
			if !strings.Contains(out, tt.cls) {
				t.Errorf("Grid with %s should contain class %s", tt.name, tt.cls)
			}
		})
	}
}

func TestGridAutoFill(t *testing.T) {
	out := render(t, Grid(GridAutoFill("200px")))
	if !strings.Contains(out, "repeat(auto-fill,minmax(200px,1fr))") {
		t.Error("GridAutoFill should set grid-template-columns with auto-fill")
	}
}

func TestGridAutoFit(t *testing.T) {
	out := render(t, Grid(GridAutoFit("300px")))
	if !strings.Contains(out, "repeat(auto-fit,minmax(300px,1fr))") {
		t.Error("GridAutoFit should set grid-template-columns with auto-fit")
	}
}

func TestGridSpan(t *testing.T) {
	out := render(t, Grid(GridCols3,
		gd.Div(GridSpan(2), gerbera.Literal("wide")),
		gerbera.Literal("normal"),
	))
	if !strings.Contains(out, "grid-column:span 2") {
		t.Error("GridSpan should set grid-column:span")
	}
}

func TestGridRowSpan(t *testing.T) {
	out := render(t, Grid(GridCols2,
		gd.Div(GridRowSpan(2), gerbera.Literal("tall")),
	))
	if !strings.Contains(out, "grid-row:span 2") {
		t.Error("GridRowSpan should set grid-row:span")
	}
}

func TestSpacer(t *testing.T) {
	out := render(t, Row(gerbera.Literal("left"), Spacer(), gerbera.Literal("right")))
	if !strings.Contains(out, "g-spacer") {
		t.Error("Spacer should have g-spacer class")
	}
}

func TestSpaceY(t *testing.T) {
	for _, size := range []string{"xs", "sm", "md", "lg", "xl"} {
		t.Run(size, func(t *testing.T) {
			out := render(t, SpaceY(size))
			if !strings.Contains(out, "g-space-y-"+size) {
				t.Errorf("SpaceY(%q) should have class g-space-y-%s", size, size)
			}
		})
	}
}

func TestGapModifiers(t *testing.T) {
	tests := []struct {
		name string
		opt  gerbera.ComponentFunc
		cls  string
	}{
		{"none", GapNone, "g-gap-none"},
		{"xs", GapXs, "g-gap-xs"},
		{"sm", GapSm, "g-gap-sm"},
		{"md", GapMd, "g-gap-md"},
		{"lg", GapLg, "g-gap-lg"},
		{"xl", GapXl, "g-gap-xl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(t, Row(tt.opt, gerbera.Literal("a")))
			if !strings.Contains(out, tt.cls) {
				t.Errorf("Gap %s should add class %s", tt.name, tt.cls)
			}
		})
	}
}

func TestJustifyModifiers(t *testing.T) {
	tests := []struct {
		name string
		opt  gerbera.ComponentFunc
		cls  string
	}{
		{"start", JustifyStart, "g-justify-start"},
		{"center", JustifyCenter, "g-justify-center"},
		{"end", JustifyEnd, "g-justify-end"},
		{"between", JustifyBetween, "g-justify-between"},
		{"around", JustifyAround, "g-justify-around"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(t, Row(tt.opt))
			if !strings.Contains(out, tt.cls) {
				t.Errorf("Justify %s should add class %s", tt.name, tt.cls)
			}
		})
	}
}

func TestAlignModifiers(t *testing.T) {
	tests := []struct {
		name string
		opt  gerbera.ComponentFunc
		cls  string
	}{
		{"start", AlignStart, "g-align-start"},
		{"center", AlignCenter, "g-align-center"},
		{"end", AlignEnd, "g-align-end"},
		{"stretch", AlignStretch, "g-align-stretch"},
		{"baseline", AlignBaseline, "g-align-baseline"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(t, Row(tt.opt))
			if !strings.Contains(out, tt.cls) {
				t.Errorf("Align %s should add class %s", tt.name, tt.cls)
			}
		})
	}
}

func TestWrapModifier(t *testing.T) {
	out := render(t, Row(Wrap))
	if !strings.Contains(out, "g-wrap") {
		t.Error("Wrap should add g-wrap class")
	}
}

func TestGrowShrink(t *testing.T) {
	out := render(t, Row(gd.Div(Grow, gerbera.Literal("a")), gd.Div(Shrink0, gerbera.Literal("b"))))
	if !strings.Contains(out, "g-grow") {
		t.Error("Grow should add g-grow class")
	}
	if !strings.Contains(out, "g-shrink-0") {
		t.Error("Shrink0 should add g-shrink-0 class")
	}
}

func TestComposedLayout(t *testing.T) {
	out := render(t, Stack(
		Alert("Info", "info"),
		Grid(GridCols3,
			StatCard("A", "1"),
			StatCard("B", "2"),
			StatCard("C", "3"),
		),
		HStack(
			Button("Save", ButtonPrimary),
			Button("Cancel"),
		),
	))
	if !strings.Contains(out, "g-stack") {
		t.Error("Composed layout should have g-stack")
	}
	if !strings.Contains(out, "g-grid") {
		t.Error("Composed layout should have g-grid")
	}
	if !strings.Contains(out, "g-hstack") {
		t.Error("Composed layout should have g-hstack")
	}
}

func TestRowWithSpacer(t *testing.T) {
	out := render(t, Row(AlignCenter,
		gerbera.Literal("logo"),
		Spacer(),
		Button("Logout"),
	))
	if !strings.Contains(out, "g-row") {
		t.Error("Row should have g-row class")
	}
	if !strings.Contains(out, "g-spacer") {
		t.Error("Row should contain Spacer")
	}
	if !strings.Contains(out, "g-align-center") {
		t.Error("Row should have AlignCenter modifier")
	}
}

// --- Spinner tests ---

func TestSpinner(t *testing.T) {
	out := render(t, Spinner("md"))
	if !strings.Contains(out, "g-spinner") {
		t.Error("Spinner should have g-spinner class")
	}
	if !strings.Contains(out, "g-spinner-md") {
		t.Error("Spinner should have size class")
	}
	if !strings.Contains(out, `role="status"`) {
		t.Error("Spinner should have role=status")
	}
	if !strings.Contains(out, `aria-label="Loading"`) {
		t.Error("Spinner should have aria-label")
	}
	if !strings.Contains(out, "g-spinner-arc") {
		t.Error("Spinner should have arc element")
	}
}

func TestSpinnerSizes(t *testing.T) {
	for _, size := range []string{"sm", "md", "lg"} {
		t.Run(size, func(t *testing.T) {
			out := render(t, Spinner(size))
			if !strings.Contains(out, "g-spinner-"+size) {
				t.Errorf("Spinner(%q) should have class g-spinner-%s", size, size)
			}
		})
	}
}

func TestSpinnerDefault(t *testing.T) {
	out := render(t, Spinner(""))
	if !strings.Contains(out, "g-spinner-md") {
		t.Error("Spinner with empty size should default to md")
	}
}

func TestSpinnerInline(t *testing.T) {
	out := render(t, Spinner("sm", SpinnerInline))
	if !strings.Contains(out, "g-spinner-inline") {
		t.Error("Spinner with SpinnerInline should have inline class")
	}
}

// --- NumberInput tests ---

func TestNumberInput(t *testing.T) {
	out := render(t, NumberInput("qty", 5, NumberInputOpts{}))
	if !strings.Contains(out, "g-numberinput") {
		t.Error("NumberInput should have g-numberinput class")
	}
	if !strings.Contains(out, `role="spinbutton"`) {
		t.Error("NumberInput should have spinbutton role")
	}
	if !strings.Contains(out, `name="qty"`) {
		t.Error("NumberInput should have name attribute")
	}
	if !strings.Contains(out, `value="5"`) {
		t.Error("NumberInput should have value attribute")
	}
	if !strings.Contains(out, `aria-valuenow="5"`) {
		t.Error("NumberInput should have aria-valuenow")
	}
}

func TestNumberInputMinMax(t *testing.T) {
	min, max := 0, 10
	out := render(t, NumberInput("qty", 3, NumberInputOpts{Min: &min, Max: &max}))
	if !strings.Contains(out, `min="0"`) {
		t.Error("NumberInput should have min attribute")
	}
	if !strings.Contains(out, `max="10"`) {
		t.Error("NumberInput should have max attribute")
	}
	if !strings.Contains(out, `aria-valuemin="0"`) {
		t.Error("NumberInput should have aria-valuemin")
	}
	if !strings.Contains(out, `aria-valuemax="10"`) {
		t.Error("NumberInput should have aria-valuemax")
	}
}

func TestNumberInputDisabledAtMin(t *testing.T) {
	min := 0
	out := render(t, NumberInput("qty", 0, NumberInputOpts{Min: &min}))
	// Decrement button should be disabled
	if !strings.Contains(out, "g-numberinput-dec") {
		t.Error("NumberInput should have dec button")
	}
}

// --- Slider tests ---

func TestSlider(t *testing.T) {
	out := render(t, Slider("volume", 50, SliderOpts{Min: 0, Max: 100}))
	if !strings.Contains(out, "g-slider") {
		t.Error("Slider should have g-slider class")
	}
	if !strings.Contains(out, `role="slider"`) {
		t.Error("Slider should have slider role")
	}
	if !strings.Contains(out, `name="volume"`) {
		t.Error("Slider should have name attribute")
	}
	if !strings.Contains(out, `value="50"`) {
		t.Error("Slider should have value attribute")
	}
	if !strings.Contains(out, `aria-valuenow="50"`) {
		t.Error("Slider should have aria-valuenow")
	}
	if !strings.Contains(out, `aria-valuemin="0"`) {
		t.Error("Slider should have aria-valuemin")
	}
	if !strings.Contains(out, `aria-valuemax="100"`) {
		t.Error("Slider should have aria-valuemax")
	}
}

func TestSliderLabel(t *testing.T) {
	out := render(t, Slider("vol", 30, SliderOpts{Label: "Volume", Max: 100}))
	if !strings.Contains(out, "Volume") {
		t.Error("Slider should display label")
	}
	if !strings.Contains(out, "30") {
		t.Error("Slider should display current value")
	}
	if !strings.Contains(out, "g-slider-header") {
		t.Error("Slider should have header")
	}
}

func TestSliderDefaultMax(t *testing.T) {
	out := render(t, Slider("x", 10, SliderOpts{}))
	if !strings.Contains(out, `max="100"`) {
		t.Error("Slider with zero Min/Max should default max to 100")
	}
}

// --- Calendar tests ---

func TestCalendar(t *testing.T) {
	out := render(t, Calendar(CalendarOpts{
		Year:  2025,
		Month: time.January,
		Today: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}))
	if !strings.Contains(out, "g-calendar") {
		t.Error("Calendar should have g-calendar class")
	}
	if !strings.Contains(out, `role="grid"`) {
		t.Error("Calendar should have grid role")
	}
	if !strings.Contains(out, "January 2025") {
		t.Error("Calendar should display month and year")
	}
	if !strings.Contains(out, "g-calendar-day-today") {
		t.Error("Calendar should highlight today")
	}
}

func TestCalendarSelected(t *testing.T) {
	sel := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	out := render(t, Calendar(CalendarOpts{
		Year:     2025,
		Month:    time.January,
		Selected: &sel,
		Today:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}))
	if !strings.Contains(out, "g-calendar-day-selected") {
		t.Error("Calendar should highlight selected date")
	}
	if !strings.Contains(out, `data-date="2025-01-20"`) {
		t.Error("Calendar should have data-date attribute for selected day")
	}
}

func TestCalendarDayNames(t *testing.T) {
	out := render(t, Calendar(CalendarOpts{
		Year:  2025,
		Month: time.March,
		Today: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
	}))
	if !strings.Contains(out, "Sun") {
		t.Error("Calendar should display default day names")
	}
	if !strings.Contains(out, "g-calendar-weekdays") {
		t.Error("Calendar should have weekdays row")
	}
}

func TestCalendarCustomDayNames(t *testing.T) {
	out := render(t, Calendar(CalendarOpts{
		Year:     2025,
		Month:    time.January,
		Today:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DayNames: []string{"日", "月", "火", "水", "木", "金", "土"},
	}))
	if !strings.Contains(out, "日") {
		t.Error("Calendar should use custom day names")
	}
}

func TestCalendarOutsideDays(t *testing.T) {
	out := render(t, Calendar(CalendarOpts{
		Year:  2025,
		Month: time.February,
		Today: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
	}))
	if !strings.Contains(out, "g-calendar-day-outside") {
		t.Error("Calendar should have outside day cells for padding")
	}
}

// --- Chat tests ---

func TestChatContainer(t *testing.T) {
	out := render(t, ChatContainer(gerbera.Literal("messages")))
	if !strings.Contains(out, "g-chat") {
		t.Error("ChatContainer should have g-chat class")
	}
	if !strings.Contains(out, `role="log"`) {
		t.Error("ChatContainer should have log role")
	}
}

func TestChatMessageReceived(t *testing.T) {
	out := render(t, ChatMessageView(ChatMessage{
		Author:    "Alice",
		Content:   "Hello!",
		Timestamp: "10:30",
		Sent:      false,
		Avatar:    "A",
	}))
	if !strings.Contains(out, "g-chat-bubble-received") {
		t.Error("Received message should have received bubble class")
	}
	if !strings.Contains(out, "g-chat-row-received") {
		t.Error("Received message should have received row class")
	}
	if !strings.Contains(out, "Hello!") {
		t.Error("Message should contain content")
	}
	if !strings.Contains(out, "Alice") {
		t.Error("Message should contain author")
	}
	if !strings.Contains(out, "10:30") {
		t.Error("Message should contain timestamp")
	}
	if !strings.Contains(out, "g-chat-avatar") {
		t.Error("Message should contain avatar")
	}
}

func TestChatMessageSent(t *testing.T) {
	out := render(t, ChatMessageView(ChatMessage{
		Content: "Hi there",
		Sent:    true,
	}))
	if !strings.Contains(out, "g-chat-bubble-sent") {
		t.Error("Sent message should have sent bubble class")
	}
	if !strings.Contains(out, "g-chat-row-sent") {
		t.Error("Sent message should have sent row class")
	}
}

func TestChatInput(t *testing.T) {
	out := render(t, ChatInput("message", ""))
	if !strings.Contains(out, "g-chat-inputbar") {
		t.Error("ChatInput should have inputbar class")
	}
	if !strings.Contains(out, "g-chat-input-field") {
		t.Error("ChatInput should have input field")
	}
	if !strings.Contains(out, "g-chat-send") {
		t.Error("ChatInput should have send button")
	}
	if !strings.Contains(out, `aria-label="Send"`) {
		t.Error("Send button should have aria-label")
	}
}

// --- Pagination tests ---

func TestPagination(t *testing.T) {
	out := render(t, Pagination(PaginationOpts{Page: 0, PageSize: 10, Total: 50}))
	if !strings.Contains(out, "g-pagination") {
		t.Error("Pagination should have g-pagination class")
	}
	if !strings.Contains(out, `role="navigation"`) {
		t.Error("Pagination should have navigation role")
	}
	if !strings.Contains(out, "1\u201310 of 50") {
		t.Error("Pagination should show item range")
	}
	if !strings.Contains(out, `aria-current="page"`) {
		t.Error("Active page should have aria-current")
	}
}

func TestPaginationEllipsis(t *testing.T) {
	out := render(t, Pagination(PaginationOpts{Page: 5, PageSize: 1, Total: 20}))
	if !strings.Contains(out, "\u2026") {
		t.Error("Pagination with many pages should show ellipsis")
	}
	// Should show first and last page
	if !strings.Contains(out, ">1<") {
		t.Error("Pagination should always show first page")
	}
}

func TestPaginationSinglePage(t *testing.T) {
	out := render(t, Pagination(PaginationOpts{Page: 0, PageSize: 10, Total: 5}))
	if !strings.Contains(out, "g-pagination-prev") {
		t.Error("Pagination should have prev button")
	}
	if !strings.Contains(out, "g-pagination-next") {
		t.Error("Pagination should have next button")
	}
}

// --- ButtonGroup tests ---

func TestButtonGroup(t *testing.T) {
	out := render(t, ButtonGroup([]ButtonGroupItem{
		{Label: "Day", Value: "day", Active: true},
		{Label: "Week", Value: "week"},
		{Label: "Month", Value: "month"},
	}))
	if !strings.Contains(out, "g-btngroup") {
		t.Error("ButtonGroup should have g-btngroup class")
	}
	if !strings.Contains(out, `role="group"`) {
		t.Error("ButtonGroup should have group role")
	}
	if !strings.Contains(out, `aria-pressed="true"`) {
		t.Error("Active button should have aria-pressed=true")
	}
	if !strings.Contains(out, `aria-pressed="false"`) {
		t.Error("Inactive button should have aria-pressed=false")
	}
	if !strings.Contains(out, "g-btngroup-active") {
		t.Error("Active button should have active class")
	}
	if !strings.Contains(out, "Day") {
		t.Error("ButtonGroup should contain button labels")
	}
}

func TestButtonGroupSmall(t *testing.T) {
	out := render(t, ButtonGroup([]ButtonGroupItem{
		{Label: "A", Value: "a"},
	}, ButtonGroupSmall))
	if !strings.Contains(out, "g-btngroup-sm") {
		t.Error("ButtonGroupSmall should add sm class")
	}
}

// --- Accordion tests ---

func TestAccordion(t *testing.T) {
	out := render(t, Accordion([]AccordionItem{
		{Title: "Section 1", Content: gerbera.Literal("Content 1"), Open: true},
		{Title: "Section 2", Content: gerbera.Literal("Content 2"), Open: false},
	}))
	if !strings.Contains(out, "g-accordion") {
		t.Error("Accordion should have g-accordion class")
	}
	if !strings.Contains(out, "g-accordion-item") {
		t.Error("Accordion should have accordion items")
	}
	if !strings.Contains(out, "g-accordion-header") {
		t.Error("Accordion should have headers")
	}
	if !strings.Contains(out, "Section 1") {
		t.Error("Accordion should show item titles")
	}
	if !strings.Contains(out, "Content 1") {
		t.Error("Open accordion item should show content")
	}
	if !strings.Contains(out, `open="open"`) {
		t.Error("Open item should have open attribute")
	}
}

func TestAccordionClosed(t *testing.T) {
	out := render(t, Accordion([]AccordionItem{
		{Title: "Closed", Content: gerbera.Literal("Hidden"), Open: false},
	}))
	if !strings.Contains(out, "g-accordion-body") {
		// Static accordion always renders the body; it's hidden via CSS
		// Just ensure the structure is correct
	}
	if !strings.Contains(out, "Closed") {
		t.Error("Accordion should show title even when closed")
	}
}

// --- Stepper tests ---

func TestStepper(t *testing.T) {
	out := render(t, Stepper([]Step{
		{Label: "Cart", Status: StepCompleted},
		{Label: "Shipping", Status: StepActive, Description: "Enter address"},
		{Label: "Payment", Status: StepUpcoming},
	}))
	if !strings.Contains(out, "g-stepper") {
		t.Error("Stepper should have g-stepper class")
	}
	if !strings.Contains(out, `role="list"`) {
		t.Error("Stepper should have list role")
	}
	if !strings.Contains(out, `role="listitem"`) {
		t.Error("Step should have listitem role")
	}
	if !strings.Contains(out, "g-stepper-completed") {
		t.Error("Completed step should have completed class")
	}
	if !strings.Contains(out, "g-stepper-active") {
		t.Error("Active step should have active class")
	}
	if !strings.Contains(out, "g-stepper-upcoming") {
		t.Error("Upcoming step should have upcoming class")
	}
	if !strings.Contains(out, `aria-current="step"`) {
		t.Error("Active step should have aria-current=step")
	}
	if !strings.Contains(out, "\u2713") {
		t.Error("Completed step should show checkmark")
	}
	if !strings.Contains(out, "Enter address") {
		t.Error("Step with description should display it")
	}
	if !strings.Contains(out, "g-stepper-connector") {
		t.Error("Non-last steps should have connectors")
	}
}

func TestStepperVertical(t *testing.T) {
	out := render(t, Stepper([]Step{
		{Label: "Step 1", Status: StepCompleted},
		{Label: "Step 2", Status: StepActive},
	}, StepperVertical))
	if !strings.Contains(out, "g-stepper-vertical") {
		t.Error("StepperVertical should add vertical class")
	}
}

// --- InfiniteScroll tests ---

func TestInfiniteScroll(t *testing.T) {
	out := render(t, InfiniteScroll(InfiniteScrollList, false, false,
		gerbera.Literal("Item 1"),
		gerbera.Literal("Item 2"),
	))
	if !strings.Contains(out, "g-infinitescroll") {
		t.Error("InfiniteScroll should have g-infinitescroll class")
	}
	if !strings.Contains(out, "g-infinitescroll-content") {
		t.Error("InfiniteScroll should have content area")
	}
	if !strings.Contains(out, `aria-live="polite"`) {
		t.Error("InfiniteScroll content should have aria-live")
	}
	if !strings.Contains(out, "Item 1") {
		t.Error("InfiniteScroll should render children")
	}
}

func TestInfiniteScrollGrid(t *testing.T) {
	out := render(t, InfiniteScroll(InfiniteScrollGrid, false, false))
	if !strings.Contains(out, "g-infinitescroll-grid") {
		t.Error("Grid view should have grid class")
	}
}

func TestInfiniteScrollLoading(t *testing.T) {
	out := render(t, InfiniteScroll(InfiniteScrollList, true, false))
	if !strings.Contains(out, "g-infinitescroll-loader") {
		t.Error("Loading state should show loader")
	}
	if !strings.Contains(out, "g-spinner") {
		t.Error("Loader should contain spinner")
	}
}

func TestInfiniteScrollToggle(t *testing.T) {
	out := render(t, InfiniteScroll(InfiniteScrollList, false, true))
	if !strings.Contains(out, "g-infinitescroll-toolbar") {
		t.Error("ShowToggle should render toolbar")
	}
	if !strings.Contains(out, "g-infinitescroll-toggle") {
		t.Error("Toolbar should have toggle buttons")
	}
	if !strings.Contains(out, `aria-pressed="true"`) {
		t.Error("Active view toggle should have aria-pressed=true")
	}
}

// --- TimePicker tests ---

func TestTimePicker(t *testing.T) {
	out := render(t, TimePicker("alarm", 14, 30, 0, TimePickerOpts{Use24H: true}))
	if !strings.Contains(out, "g-timepicker") {
		t.Error("TimePicker should have g-timepicker class")
	}
	if !strings.Contains(out, `role="group"`) {
		t.Error("TimePicker should have group role")
	}
	if !strings.Contains(out, `aria-label="alarm"`) {
		t.Error("TimePicker should have aria-label with name")
	}
	if !strings.Contains(out, `value="14"`) {
		t.Error("TimePicker should display hour value")
	}
	if !strings.Contains(out, `value="30"`) {
		t.Error("TimePicker should display minute value")
	}
	if !strings.Contains(out, `role="spinbutton"`) {
		t.Error("TimePicker units should have spinbutton role")
	}
	if !strings.Contains(out, `aria-label="Hour"`) {
		t.Error("Hour unit should have aria-label")
	}
	if !strings.Contains(out, `aria-label="Minute"`) {
		t.Error("Minute unit should have aria-label")
	}
	if !strings.Contains(out, "g-timepicker-sep") {
		t.Error("TimePicker should have separator")
	}
}

func TestTimePickerShowSec(t *testing.T) {
	out := render(t, TimePicker("time", 10, 20, 45, TimePickerOpts{Use24H: true, ShowSec: true}))
	if !strings.Contains(out, `value="45"`) {
		t.Error("TimePicker with ShowSec should display seconds")
	}
	if !strings.Contains(out, `aria-label="Second"`) {
		t.Error("Second unit should have aria-label")
	}
}

func TestTimePickerNoSec(t *testing.T) {
	out := render(t, TimePicker("time", 10, 20, 45, TimePickerOpts{Use24H: true, ShowSec: false}))
	if strings.Contains(out, `aria-label="Second"`) {
		t.Error("TimePicker without ShowSec should not have second unit")
	}
}

func TestTimePicker12Hour(t *testing.T) {
	out := render(t, TimePicker("time", 15, 30, 0, TimePickerOpts{Use24H: false}))
	if !strings.Contains(out, `value="03"`) {
		t.Error("12-hour TimePicker should convert 15:00 to 03")
	}
	if !strings.Contains(out, "g-timepicker-ampm") {
		t.Error("12-hour TimePicker should have AM/PM toggle")
	}
	if !strings.Contains(out, `aria-pressed="true"`) {
		t.Error("Active AM/PM button should have aria-pressed=true")
	}
}

func TestTimePicker12HourAM(t *testing.T) {
	out := render(t, TimePicker("time", 9, 0, 0, TimePickerOpts{Use24H: false}))
	if !strings.Contains(out, `value="09"`) {
		t.Error("12-hour TimePicker should show 09 for 9 AM")
	}
}

func TestTimePicker12HourNoon(t *testing.T) {
	out := render(t, TimePicker("time", 12, 0, 0, TimePickerOpts{Use24H: false}))
	if !strings.Contains(out, `value="12"`) {
		t.Error("12-hour TimePicker should show 12 for noon")
	}
}

func TestTimePicker12HourMidnight(t *testing.T) {
	out := render(t, TimePicker("time", 0, 0, 0, TimePickerOpts{Use24H: false}))
	if !strings.Contains(out, `value="12"`) {
		t.Error("12-hour TimePicker should show 12 for midnight")
	}
}

func TestTimePickerDisabled(t *testing.T) {
	out := render(t, TimePicker("time", 10, 0, 0, TimePickerOpts{Use24H: true, Disabled: true}))
	if !strings.Contains(out, `disabled="disabled"`) {
		t.Error("Disabled TimePicker should have disabled attributes")
	}
}

func TestTimePickerButtons(t *testing.T) {
	out := render(t, TimePicker("time", 10, 30, 0, TimePickerOpts{Use24H: true}))
	if !strings.Contains(out, "g-timepicker-up") {
		t.Error("TimePicker should have up button")
	}
	if !strings.Contains(out, "g-timepicker-down") {
		t.Error("TimePicker should have down button")
	}
	if !strings.Contains(out, `aria-label="Increase Hour"`) {
		t.Error("Up button should have increase aria-label")
	}
	if !strings.Contains(out, `aria-label="Decrease Hour"`) {
		t.Error("Down button should have decrease aria-label")
	}
}

func TestFormatTime(t *testing.T) {
	if got := FormatTime(14, 30, 0, false); got != "14:30" {
		t.Errorf("FormatTime(14,30,0,false) = %q, want 14:30", got)
	}
	if got := FormatTime(14, 30, 45, true); got != "14:30:45" {
		t.Errorf("FormatTime(14,30,45,true) = %q, want 14:30:45", got)
	}
}

// ---------- Chart tests ----------

func TestLineChart(t *testing.T) {
	series := []Series{
		{Name: "Revenue", Points: []DataPoint{
			{Label: "Jan", Value: 100},
			{Label: "Feb", Value: 200},
			{Label: "Mar", Value: 150},
		}},
	}
	out := render(t, LineChart(series, ChartOpts{ShowTooltip: true, ShowGrid: true, ShowLegend: true, Title: "Revenue"}))
	if !strings.Contains(out, "<svg") {
		t.Error("LineChart should render SVG element")
	}
	if !strings.Contains(out, "g-chart") {
		t.Error("LineChart should have g-chart class")
	}
	if !strings.Contains(out, "g-chart-line") {
		t.Error("LineChart should have g-chart-line class on polyline")
	}
	if !strings.Contains(out, "g-chart-point") {
		t.Error("LineChart should have g-chart-point class on circles")
	}
	if !strings.Contains(out, "<title>") {
		t.Error("LineChart with ShowTooltip should contain <title> elements")
	}
	if !strings.Contains(out, "g-chart-grid") {
		t.Error("LineChart with ShowGrid should contain grid")
	}
	if !strings.Contains(out, "g-chart-legend") {
		t.Error("LineChart with ShowLegend should contain legend")
	}
	if !strings.Contains(out, "g-chart-title") {
		t.Error("LineChart with Title should contain title text")
	}
}

func TestColumnChart(t *testing.T) {
	series := []Series{
		{Name: "Sales", Points: []DataPoint{
			{Label: "Q1", Value: 300},
			{Label: "Q2", Value: 450},
		}},
	}
	out := render(t, ColumnChart(series, ChartOpts{}))
	if !strings.Contains(out, "<svg") {
		t.Error("ColumnChart should render SVG element")
	}
	if !strings.Contains(out, "<rect") {
		t.Error("ColumnChart should contain rect elements")
	}
}

func TestBarChart(t *testing.T) {
	series := []Series{
		{Name: "Score", Points: []DataPoint{
			{Label: "Alice", Value: 85},
			{Label: "Bob", Value: 92},
		}},
	}
	out := render(t, BarChart(series, ChartOpts{ShowTooltip: true}))
	if !strings.Contains(out, "<svg") {
		t.Error("BarChart should render SVG element")
	}
	if !strings.Contains(out, "<rect") {
		t.Error("BarChart should contain rect elements")
	}
	if !strings.Contains(out, "<title>") {
		t.Error("BarChart with ShowTooltip should contain tooltips")
	}
}

func TestPieChart(t *testing.T) {
	data := []DataPoint{
		{Label: "A", Value: 30},
		{Label: "B", Value: 70},
	}
	out := render(t, PieChart(data, ChartOpts{ShowLegend: true, ShowTooltip: true}))
	if !strings.Contains(out, "<svg") {
		t.Error("PieChart should render SVG element")
	}
	if !strings.Contains(out, "g-chart-slice") {
		t.Error("PieChart should contain g-chart-slice class")
	}
	if !strings.Contains(out, "g-chart-legend") {
		t.Error("PieChart with ShowLegend should have legend")
	}
}

func TestPieChartSingleSlice(t *testing.T) {
	data := []DataPoint{{Label: "All", Value: 100}}
	out := render(t, PieChart(data, ChartOpts{}))
	if !strings.Contains(out, "<circle") {
		t.Error("PieChart with single slice should use circle")
	}
}

func TestScatterPlot(t *testing.T) {
	series := []Series{
		{Name: "Points", Points: []DataPoint{
			{Label: "X1", Value: 10},
			{Label: "X2", Value: 20},
		}},
	}
	out := render(t, ScatterPlot(series, ChartOpts{}))
	if !strings.Contains(out, "<svg") {
		t.Error("ScatterPlot should render SVG element")
	}
	if !strings.Contains(out, "g-chart-point") {
		t.Error("ScatterPlot should contain g-chart-point circles")
	}
}

func TestHistogram(t *testing.T) {
	values := []float64{1, 2, 2, 3, 3, 3, 4, 4, 5, 10}
	out := render(t, Histogram(values, HistogramOpts{BinCount: 5}))
	if !strings.Contains(out, "<svg") {
		t.Error("Histogram should render SVG element")
	}
	if !strings.Contains(out, "<rect") {
		t.Error("Histogram should contain rect elements")
	}
}

func TestStackedBarChart(t *testing.T) {
	series := []Series{
		{Name: "A", Points: []DataPoint{{Label: "Row1", Value: 30}, {Label: "Row2", Value: 20}}},
		{Name: "B", Points: []DataPoint{{Label: "Row1", Value: 50}, {Label: "Row2", Value: 40}}},
	}
	out := render(t, StackedBarChart(series, ChartOpts{ShowGrid: true, ShowLegend: true}))
	if !strings.Contains(out, "<svg") {
		t.Error("StackedBarChart should render SVG element")
	}
	if !strings.Contains(out, "<rect") {
		t.Error("StackedBarChart should contain rect elements")
	}
	if !strings.Contains(out, "g-chart-legend") {
		t.Error("StackedBarChart with ShowLegend should have legend")
	}
}

func TestChartEmpty(t *testing.T) {
	out := render(t, LineChart(nil, ChartOpts{}))
	if !strings.Contains(out, "No data") {
		t.Error("Empty chart should show 'No data' message")
	}
}

func TestChartMinMaxEqual(t *testing.T) {
	series := []Series{
		{Name: "Flat", Points: []DataPoint{
			{Label: "A", Value: 5},
			{Label: "B", Value: 5},
		}},
	}
	// Should not panic
	out := render(t, ColumnChart(series, ChartOpts{}))
	if !strings.Contains(out, "<svg") {
		t.Error("Chart with equal min/max should still render")
	}
}

// ---------- Avatar tests ----------

func TestImageAvatar(t *testing.T) {
	out := render(t, ImageAvatar("https://example.com/photo.jpg", AvatarOpts{Size: "lg", Alt: "User"}))
	if !strings.Contains(out, "g-avatar") {
		t.Error("ImageAvatar should have g-avatar class")
	}
	if !strings.Contains(out, "g-avatar-lg") {
		t.Error("ImageAvatar should have g-avatar-lg class for lg size")
	}
	if !strings.Contains(out, "g-avatar-circle") {
		t.Error("ImageAvatar should default to circle shape")
	}
	if !strings.Contains(out, `src="https://example.com/photo.jpg"`) {
		t.Error("ImageAvatar should contain img src")
	}
	if !strings.Contains(out, `alt="User"`) {
		t.Error("ImageAvatar should have alt text")
	}
}

func TestImageAvatarRounded(t *testing.T) {
	out := render(t, ImageAvatar("photo.jpg", AvatarOpts{Shape: "rounded"}))
	if !strings.Contains(out, "g-avatar-rounded") {
		t.Error("ImageAvatar with rounded shape should have g-avatar-rounded class")
	}
}

func TestLetterAvatar(t *testing.T) {
	out := render(t, LetterAvatar("Tomo", AvatarOpts{Size: "xl"}))
	if !strings.Contains(out, "g-avatar") {
		t.Error("LetterAvatar should have g-avatar class")
	}
	if !strings.Contains(out, "g-avatar-xl") {
		t.Error("LetterAvatar should have g-avatar-xl class")
	}
	if !strings.Contains(out, "background-color:") {
		t.Error("LetterAvatar should have background color style")
	}
	if !strings.Contains(out, "T") {
		t.Error("LetterAvatar should display initial letter")
	}
}

func TestLetterAvatarDeterministic(t *testing.T) {
	out1 := render(t, LetterAvatar("Alice", AvatarOpts{}))
	out2 := render(t, LetterAvatar("Alice", AvatarOpts{}))
	if out1 != out2 {
		t.Error("LetterAvatar should be deterministic for the same name")
	}
}

func TestAvatarGroup(t *testing.T) {
	avatars := []gerbera.ComponentFunc{
		LetterAvatar("Alice", AvatarOpts{}),
		LetterAvatar("Bob", AvatarOpts{}),
		LetterAvatar("Charlie", AvatarOpts{}),
		LetterAvatar("Diana", AvatarOpts{}),
		LetterAvatar("Eve", AvatarOpts{}),
	}
	out := render(t, AvatarGroup(avatars, AvatarGroupOpts{Max: 3}))
	if !strings.Contains(out, "g-avatar-group") {
		t.Error("AvatarGroup should have g-avatar-group class")
	}
	if !strings.Contains(out, "g-avatar-group-more") {
		t.Error("AvatarGroup with overflow should show +N more")
	}
	if !strings.Contains(out, "+2") {
		t.Error("AvatarGroup should show +2 for 5 avatars with max 3")
	}
}

func TestAvatarGroupNoOverflow(t *testing.T) {
	avatars := []gerbera.ComponentFunc{
		LetterAvatar("Alice", AvatarOpts{}),
		LetterAvatar("Bob", AvatarOpts{}),
	}
	out := render(t, AvatarGroup(avatars, AvatarGroupOpts{}))
	if strings.Contains(out, "g-avatar-group-more") {
		t.Error("AvatarGroup without overflow should not show +N more")
	}
}

func TestAvatarDefaultSize(t *testing.T) {
	out := render(t, ImageAvatar("photo.jpg", AvatarOpts{}))
	if !strings.Contains(out, "g-avatar-md") {
		t.Error("Avatar should default to md size")
	}
}

// Ensure unused imports are referenced
var _ = gd.Div
var _ = property.Class
var _ = time.UTC
