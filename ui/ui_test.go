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

// Ensure unused imports are referenced
var _ = gd.Div
var _ = property.Class
