package ui

import (
	"fmt"
	"strings"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/styles"
)

// ThemeConfig holds CSS custom property values for the Gerbera UI theme.
// Zero-value fields are filled with defaults from DefaultTheme().
type ThemeConfig struct {
	// Background
	Bg        string // --g-bg
	BgSurface string // --g-bg-surface
	BgOverlay string // --g-bg-overlay
	BgInset   string // --g-bg-inset

	// Text
	Text          string // --g-text
	TextSecondary string // --g-text-secondary
	TextTertiary  string // --g-text-tertiary
	TextInverse   string // --g-text-inverse

	// Border
	Border       string // --g-border
	BorderStrong string // --g-border-strong

	// Shadow
	ShadowSm string // --g-shadow-sm
	Shadow   string // --g-shadow
	ShadowLg string // --g-shadow-lg

	// Accent (primary)
	Accent      string // --g-accent
	AccentLight string // --g-accent-light
	AccentHover string // --g-accent-hover

	// Semantic: Danger
	Danger       string // --g-danger
	DangerBg     string // --g-danger-bg
	DangerBorder string // --g-danger-border
	DangerHover  string // --g-danger-hover

	// Semantic: Success
	Success       string // --g-success
	SuccessBg     string // --g-success-bg
	SuccessBorder string // --g-success-border

	// Semantic: Warning
	Warning       string // --g-warning
	WarningBg     string // --g-warning-bg
	WarningBorder string // --g-warning-border

	// Semantic: Info
	Info       string // --g-info
	InfoBg     string // --g-info-bg
	InfoBorder string // --g-info-border

	// Spacing
	SpaceXs string // --g-space-xs
	SpaceSm string // --g-space-sm
	SpaceMd string // --g-space-md
	SpaceLg string // --g-space-lg
	SpaceXl string // --g-space-xl

	// Typography
	Font     string // --g-font
	FontMono string // --g-font-mono
	Radius   string // --g-radius
}

// DefaultTheme returns the default light theme configuration.
func DefaultTheme() ThemeConfig {
	return ThemeConfig{
		Bg:        "#f5f6f8",
		BgSurface: "#ffffff",
		BgOverlay: "#f0f1f4",
		BgInset:   "#e8eaed",

		Text:          "#1a1d21",
		TextSecondary: "#5a6069",
		TextTertiary:  "#8b949e",
		TextInverse:   "#ffffff",

		Border:       "#d8dce0",
		BorderStrong: "#b0b8c1",

		ShadowSm: "0 1px 2px rgba(0,0,0,0.06)",
		Shadow:   "0 2px 8px rgba(0,0,0,0.08)",
		ShadowLg: "0 4px 16px rgba(0,0,0,0.10)",

		Accent:      "#3d4450",
		AccentLight: "#e8eaef",
		AccentHover: "#2d333b",

		Danger:       "#8b3a3a",
		DangerBg:     "#fdf2f2",
		DangerBorder: "#e5c7c7",
		DangerHover:  "#6b2a2a",

		Success:       "#3a6b3a",
		SuccessBg:     "#f2fdf2",
		SuccessBorder: "#c7e5c7",

		Warning:       "#7a6a2a",
		WarningBg:     "#fdfaf2",
		WarningBorder: "#e5dfc7",

		Info:       "#3a5a8b",
		InfoBg:     "#f2f6fd",
		InfoBorder: "#c7d5e5",

		SpaceXs: "4px",
		SpaceSm: "8px",
		SpaceMd: "16px",
		SpaceLg: "24px",
		SpaceXl: "32px",

		Font:     `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`,
		FontMono: `"SF Mono", "Cascadia Code", Consolas, monospace`,
		Radius:   "6px",
	}
}

// DarkTheme returns a dark theme configuration.
func DarkTheme() ThemeConfig {
	return ThemeConfig{
		Bg:        "#1a1d21",
		BgSurface: "#24282e",
		BgOverlay: "#2d333b",
		BgInset:   "#161a1e",

		Text:          "#e6edf3",
		TextSecondary: "#8b949e",
		TextTertiary:  "#6e7681",
		TextInverse:   "#1a1d21",

		Border:       "#373e47",
		BorderStrong: "#525a65",

		ShadowSm: "0 1px 2px rgba(0,0,0,0.3)",
		Shadow:   "0 2px 8px rgba(0,0,0,0.4)",
		ShadowLg: "0 4px 16px rgba(0,0,0,0.5)",

		Accent:      "#6e7ee0",
		AccentLight: "#282d45",
		AccentHover: "#8a97e8",

		Danger:       "#e5534b",
		DangerBg:     "#3d1f1f",
		DangerBorder: "#6b3333",
		DangerHover:  "#f47067",

		Success:       "#57ab5a",
		SuccessBg:     "#1f3d1f",
		SuccessBorder: "#2b6b2b",

		Warning:       "#c69026",
		WarningBg:     "#3d331f",
		WarningBorder: "#6b5c2b",

		Info:       "#539bf5",
		InfoBg:     "#1f2d3d",
		InfoBorder: "#2b4a6b",
	}
}

// Theme returns a <style> element containing the full Gerbera UI design system CSS
// with the default light theme. Include this in <head> to apply the theme.
func Theme() gerbera.ComponentFunc {
	return ThemeWith(ThemeConfig{})
}

// ThemeWith returns a <style> element with the specified theme configuration.
// Zero-value fields in cfg are filled with DefaultTheme() values.
func ThemeWith(cfg ThemeConfig) gerbera.ComponentFunc {
	cfg = cfg.withDefaults(DefaultTheme())
	css := generateVarsCSS(":root", cfg) + "\n" + themeRulesCSS
	return styles.CSS(css)
}

// ThemeAuto returns a <style> element that switches between light and dark themes
// based on the user's OS preference (prefers-color-scheme).
// Zero-value fields in light are filled with DefaultTheme(), dark with DarkTheme().
func ThemeAuto(light, dark ThemeConfig) gerbera.ComponentFunc {
	light = light.withDefaults(DefaultTheme())
	dark = dark.withDefaults(DarkTheme()).withDefaults(DefaultTheme())
	var b strings.Builder
	b.WriteString(generateVarsCSS(":root", light))
	b.WriteString("\n@media (prefers-color-scheme: dark) {\n")
	b.WriteString(generateVarsCSS("  :root", dark))
	b.WriteString("}\n")
	b.WriteString(themeRulesCSS)
	return styles.CSS(b.String())
}

// withDefaults fills zero-value fields with values from defaults.
func (c ThemeConfig) withDefaults(defaults ThemeConfig) ThemeConfig {
	if c.Bg == "" {
		c.Bg = defaults.Bg
	}
	if c.BgSurface == "" {
		c.BgSurface = defaults.BgSurface
	}
	if c.BgOverlay == "" {
		c.BgOverlay = defaults.BgOverlay
	}
	if c.BgInset == "" {
		c.BgInset = defaults.BgInset
	}
	if c.Text == "" {
		c.Text = defaults.Text
	}
	if c.TextSecondary == "" {
		c.TextSecondary = defaults.TextSecondary
	}
	if c.TextTertiary == "" {
		c.TextTertiary = defaults.TextTertiary
	}
	if c.TextInverse == "" {
		c.TextInverse = defaults.TextInverse
	}
	if c.Border == "" {
		c.Border = defaults.Border
	}
	if c.BorderStrong == "" {
		c.BorderStrong = defaults.BorderStrong
	}
	if c.ShadowSm == "" {
		c.ShadowSm = defaults.ShadowSm
	}
	if c.Shadow == "" {
		c.Shadow = defaults.Shadow
	}
	if c.ShadowLg == "" {
		c.ShadowLg = defaults.ShadowLg
	}
	if c.Accent == "" {
		c.Accent = defaults.Accent
	}
	if c.AccentLight == "" {
		c.AccentLight = defaults.AccentLight
	}
	if c.AccentHover == "" {
		c.AccentHover = defaults.AccentHover
	}
	if c.Danger == "" {
		c.Danger = defaults.Danger
	}
	if c.DangerBg == "" {
		c.DangerBg = defaults.DangerBg
	}
	if c.DangerBorder == "" {
		c.DangerBorder = defaults.DangerBorder
	}
	if c.DangerHover == "" {
		c.DangerHover = defaults.DangerHover
	}
	if c.Success == "" {
		c.Success = defaults.Success
	}
	if c.SuccessBg == "" {
		c.SuccessBg = defaults.SuccessBg
	}
	if c.SuccessBorder == "" {
		c.SuccessBorder = defaults.SuccessBorder
	}
	if c.Warning == "" {
		c.Warning = defaults.Warning
	}
	if c.WarningBg == "" {
		c.WarningBg = defaults.WarningBg
	}
	if c.WarningBorder == "" {
		c.WarningBorder = defaults.WarningBorder
	}
	if c.Info == "" {
		c.Info = defaults.Info
	}
	if c.InfoBg == "" {
		c.InfoBg = defaults.InfoBg
	}
	if c.InfoBorder == "" {
		c.InfoBorder = defaults.InfoBorder
	}
	if c.SpaceXs == "" {
		c.SpaceXs = defaults.SpaceXs
	}
	if c.SpaceSm == "" {
		c.SpaceSm = defaults.SpaceSm
	}
	if c.SpaceMd == "" {
		c.SpaceMd = defaults.SpaceMd
	}
	if c.SpaceLg == "" {
		c.SpaceLg = defaults.SpaceLg
	}
	if c.SpaceXl == "" {
		c.SpaceXl = defaults.SpaceXl
	}
	if c.Font == "" {
		c.Font = defaults.Font
	}
	if c.FontMono == "" {
		c.FontMono = defaults.FontMono
	}
	if c.Radius == "" {
		c.Radius = defaults.Radius
	}
	return c
}

// generateVarsCSS generates a CSS rule block with custom properties.
func generateVarsCSS(selector string, cfg ThemeConfig) string {
	return fmt.Sprintf(`%s {
  --g-bg: %s;
  --g-bg-surface: %s;
  --g-bg-overlay: %s;
  --g-bg-inset: %s;

  --g-text: %s;
  --g-text-secondary: %s;
  --g-text-tertiary: %s;
  --g-text-inverse: %s;

  --g-border: %s;
  --g-border-strong: %s;

  --g-shadow-sm: %s;
  --g-shadow: %s;
  --g-shadow-lg: %s;

  --g-accent: %s;
  --g-accent-light: %s;
  --g-accent-hover: %s;

  --g-space-xs: %s;
  --g-space-sm: %s;
  --g-space-md: %s;
  --g-space-lg: %s;
  --g-space-xl: %s;

  --g-font: %s;
  --g-font-mono: %s;
  --g-radius: %s;

  --g-danger: %s;
  --g-danger-bg: %s;
  --g-danger-border: %s;
  --g-danger-hover: %s;
  --g-success: %s;
  --g-success-bg: %s;
  --g-success-border: %s;
  --g-warning: %s;
  --g-warning-bg: %s;
  --g-warning-border: %s;
  --g-info: %s;
  --g-info-bg: %s;
  --g-info-border: %s;
}
`,
		selector,
		cfg.Bg, cfg.BgSurface, cfg.BgOverlay, cfg.BgInset,
		cfg.Text, cfg.TextSecondary, cfg.TextTertiary, cfg.TextInverse,
		cfg.Border, cfg.BorderStrong,
		cfg.ShadowSm, cfg.Shadow, cfg.ShadowLg,
		cfg.Accent, cfg.AccentLight, cfg.AccentHover,
		cfg.SpaceXs, cfg.SpaceSm, cfg.SpaceMd, cfg.SpaceLg, cfg.SpaceXl,
		cfg.Font, cfg.FontMono, cfg.Radius,
		cfg.Danger, cfg.DangerBg, cfg.DangerBorder, cfg.DangerHover,
		cfg.Success, cfg.SuccessBg, cfg.SuccessBorder,
		cfg.Warning, cfg.WarningBg, cfg.WarningBorder,
		cfg.Info, cfg.InfoBg, cfg.InfoBorder,
	)
}

// themeRulesCSS contains the component CSS rules that reference CSS custom properties.
// This is the immutable part that does not change between themes.
const themeRulesCSS = `
*, *::before, *::after { box-sizing: border-box; }

body {
  margin: 0;
  font-family: var(--g-font);
  font-size: 14px;
  line-height: 1.5;
  color: var(--g-text);
  background: var(--g-bg);
  -webkit-font-smoothing: antialiased;
}

/* Card */
.g-card {
  background: var(--g-bg-surface);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  box-shadow: var(--g-shadow-sm);
}

.g-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--g-space-md) var(--g-space-lg);
  border-bottom: 1px solid var(--g-border);
}

.g-card-header-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--g-text);
  margin: 0;
}

.g-card-header-actions {
  display: flex;
  gap: var(--g-space-sm);
  align-items: center;
}

.g-card-footer {
  padding: var(--g-space-sm) var(--g-space-lg);
  border-top: 1px solid var(--g-border);
  color: var(--g-text-secondary);
  font-size: 13px;
}

/* Table */
.g-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.g-table th {
  text-align: left;
  padding: var(--g-space-sm) var(--g-space-md);
  font-weight: 600;
  color: var(--g-text-secondary);
  border-bottom: 2px solid var(--g-border);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.g-table td {
  padding: var(--g-space-sm) var(--g-space-md);
  border-bottom: 1px solid var(--g-border);
  color: var(--g-text);
}

.g-table tbody tr:hover {
  background: var(--g-bg-overlay);
}

.g-table th.g-sortable {
  cursor: pointer;
  user-select: none;
}

.g-table th.g-sortable:hover {
  color: var(--g-text);
}

.g-sort-indicator { margin-left: 4px; opacity: 0.5; }
.g-sort-active .g-sort-indicator { opacity: 1; }

/* Badge */
.g-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 12px;
  line-height: 1.5;
}

.g-badge-default { background: var(--g-bg-inset); color: var(--g-text-secondary); }
.g-badge-dark { background: var(--g-accent); color: var(--g-text-inverse); }
.g-badge-outline { background: transparent; color: var(--g-text-secondary); border: 1px solid var(--g-border-strong); }
.g-badge-light { background: var(--g-accent-light); color: var(--g-accent); }

/* Alert */
.g-alert {
  padding: var(--g-space-sm) var(--g-space-md);
  border-radius: var(--g-radius);
  font-size: 13px;
  line-height: 1.5;
  border: 1px solid;
}

.g-alert-info { background: var(--g-info-bg); color: var(--g-info); border-color: var(--g-info-border); }
.g-alert-success { background: var(--g-success-bg); color: var(--g-success); border-color: var(--g-success-border); }
.g-alert-warning { background: var(--g-warning-bg); color: var(--g-warning); border-color: var(--g-warning-border); }
.g-alert-danger { background: var(--g-danger-bg); color: var(--g-danger); border-color: var(--g-danger-border); }

/* StatCard */
.g-stat {
  background: var(--g-bg-surface);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  padding: var(--g-space-lg);
  box-shadow: var(--g-shadow-sm);
}

.g-stat-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--g-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin: 0 0 var(--g-space-xs) 0;
}

.g-stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--g-text);
  line-height: 1.2;
  margin: 0;
}

/* Sidebar */
.g-sidebar {
  width: 240px;
  min-height: 100vh;
  background: var(--g-bg-surface);
  border-right: 1px solid var(--g-border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.g-sidebar-header {
  padding: var(--g-space-lg);
  font-size: 16px;
  font-weight: 700;
  color: var(--g-text);
  border-bottom: 1px solid var(--g-border);
  letter-spacing: -0.01em;
}

.g-sidebar-link {
  display: block;
  padding: var(--g-space-sm) var(--g-space-lg);
  color: var(--g-text-secondary);
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: background 0.1s, color 0.1s;
}

.g-sidebar-link:hover { background: var(--g-bg-overlay); color: var(--g-text); }
.g-sidebar-link-active { background: var(--g-accent-light); color: var(--g-accent); }

.g-sidebar-divider {
  height: 1px;
  background: var(--g-border);
  margin: var(--g-space-sm) 0;
}

/* Breadcrumb */
.g-breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--g-space-xs);
  font-size: 13px;
  color: var(--g-text-tertiary);
  list-style: none;
  padding: 0;
  margin: 0;
}

.g-breadcrumb-sep { color: var(--g-text-tertiary); }
.g-breadcrumb a { color: var(--g-text-secondary); text-decoration: none; }
.g-breadcrumb a:hover { color: var(--g-text); text-decoration: underline; }
.g-breadcrumb-current { color: var(--g-text); font-weight: 500; }

/* Layout */
.g-admin-shell {
  display: flex;
  min-height: 100vh;
}

.g-admin-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.g-page-header {
  padding: var(--g-space-lg) var(--g-space-xl);
  border-bottom: 1px solid var(--g-border);
  background: var(--g-bg-surface);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.g-page-header-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--g-text);
  margin: 0;
}

.g-page-body {
  flex: 1;
  padding: var(--g-space-xl);
}

/* Button */
.g-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--g-space-xs);
  padding: 7px 14px;
  font-size: 13px;
  font-weight: 500;
  font-family: var(--g-font);
  border-radius: var(--g-radius);
  border: 1px solid var(--g-border);
  background: var(--g-bg-surface);
  color: var(--g-text);
  cursor: pointer;
  transition: background 0.1s, border-color 0.1s, box-shadow 0.1s;
  line-height: 1.4;
  white-space: nowrap;
}

.g-btn:hover { background: var(--g-bg-overlay); border-color: var(--g-border-strong); }
.g-btn:active { box-shadow: inset 0 1px 2px rgba(0,0,0,0.08); }

.g-btn-primary { background: var(--g-accent); color: var(--g-text-inverse); border-color: var(--g-accent); }
.g-btn-primary:hover { background: var(--g-accent-hover); }

.g-btn-outline { background: transparent; }
.g-btn-outline:hover { background: var(--g-bg-overlay); }

.g-btn-danger { background: var(--g-danger); color: var(--g-text-inverse); border-color: var(--g-danger); }
.g-btn-danger:hover { background: var(--g-danger-hover); }

.g-btn-sm { padding: 4px 10px; font-size: 12px; }

/* Form */
.g-form-group { margin-bottom: var(--g-space-md); }

.g-form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--g-text);
  margin-bottom: var(--g-space-xs);
}

.g-form-input, .g-form-select {
  display: block;
  width: 100%;
  padding: 7px 10px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-inset);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  transition: border-color 0.1s, box-shadow 0.1s;
  outline: none;
}

.g-form-input:focus, .g-form-select:focus {
  border-color: var(--g-border-strong);
  box-shadow: 0 0 0 2px var(--g-accent-light);
}

.g-form-select { appearance: auto; }

/* Misc */
.g-divider {
  height: 1px;
  background: var(--g-border);
  margin: var(--g-space-lg) 0;
  border: none;
}

.g-empty-state {
  text-align: center;
  padding: var(--g-space-xl) var(--g-space-lg);
  color: var(--g-text-tertiary);
}

.g-empty-state-msg {
  font-size: 15px;
  margin: 0 0 var(--g-space-md) 0;
}

.g-progress-track {
  width: 100%;
  height: 6px;
  background: var(--g-bg-inset);
  border-radius: 3px;
  overflow: hidden;
}

.g-progress-bar {
  height: 100%;
  background: var(--g-accent);
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* Modal */
.g-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.g-modal-panel {
  background: var(--g-bg-surface);
  border-radius: var(--g-radius);
  box-shadow: var(--g-shadow-lg);
  width: 90%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  border: 1px solid var(--g-border);
}

.g-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--g-space-md) var(--g-space-lg);
  border-bottom: 1px solid var(--g-border);
}

.g-modal-title { font-size: 15px; font-weight: 600; margin: 0; }

.g-modal-close {
  background: none;
  border: none;
  font-size: 18px;
  color: var(--g-text-tertiary);
  cursor: pointer;
  padding: 4px;
  line-height: 1;
}

.g-modal-close:hover { color: var(--g-text); }

.g-modal-body { padding: var(--g-space-lg); }

.g-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--g-space-sm);
  padding: var(--g-space-md) var(--g-space-lg);
  border-top: 1px solid var(--g-border);
}

/* Toast */
.g-toast {
  position: fixed;
  top: var(--g-space-lg);
  right: var(--g-space-lg);
  z-index: 1100;
  min-width: 280px;
  max-width: 400px;
  display: flex;
  align-items: flex-start;
  gap: var(--g-space-sm);
  padding: var(--g-space-sm) var(--g-space-md);
  border-radius: var(--g-radius);
  box-shadow: var(--g-shadow-lg);
  font-size: 13px;
  border: 1px solid;
  animation: g-toast-in 0.2s ease;
}

@keyframes g-toast-in {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

.g-toast-info { background: var(--g-info-bg); color: var(--g-info); border-color: var(--g-info-border); }
.g-toast-success { background: var(--g-success-bg); color: var(--g-success); border-color: var(--g-success-border); }
.g-toast-warning { background: var(--g-warning-bg); color: var(--g-warning); border-color: var(--g-warning-border); }
.g-toast-danger { background: var(--g-danger-bg); color: var(--g-danger); border-color: var(--g-danger-border); }

.g-toast-msg { flex: 1; }

.g-toast-close {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: inherit;
  opacity: 0.6;
  padding: 0;
  line-height: 1;
}

.g-toast-close:hover { opacity: 1; }

/* Dropdown */
.g-dropdown { position: relative; display: inline-block; }

.g-dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 900;
  min-width: 160px;
  margin-top: var(--g-space-xs);
  background: var(--g-bg-surface);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  box-shadow: var(--g-shadow);
  padding: var(--g-space-xs) 0;
}

.g-dropdown-item {
  display: block;
  width: 100%;
  padding: var(--g-space-sm) var(--g-space-md);
  font-size: 13px;
  color: var(--g-text);
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  text-decoration: none;
  font-family: var(--g-font);
}

.g-dropdown-item:hover { background: var(--g-bg-overlay); }

/* DataTable pagination */
.g-datatable-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--g-space-sm) var(--g-space-md);
  border-top: 1px solid var(--g-border);
  font-size: 12px;
  color: var(--g-text-secondary);
}

.g-datatable-pages {
  display: flex;
  gap: 2px;
}

.g-datatable-page {
  padding: 4px 10px;
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  background: var(--g-bg-surface);
  cursor: pointer;
  font-size: 12px;
  font-family: var(--g-font);
  color: var(--g-text-secondary);
}

.g-datatable-page:hover { background: var(--g-bg-overlay); }
.g-datatable-page-active { background: var(--g-accent); color: var(--g-text-inverse); border-color: var(--g-accent); }

/* Tabs */
.g-tabs { width: 100%; }

.g-tablist {
  display: flex;
  gap: 0;
  border-bottom: 2px solid var(--g-border);
}

.g-tab {
  padding: var(--g-space-sm) var(--g-space-md);
  border: none;
  background: transparent;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  font-size: 13px;
  font-weight: 500;
  font-family: var(--g-font);
  color: var(--g-text-secondary);
  transition: color 0.1s, border-color 0.1s, background 0.1s;
}

.g-tab:hover { background: var(--g-bg-overlay); color: var(--g-text); }
.g-tab-active { border-bottom-color: var(--g-accent); color: var(--g-text); font-weight: 600; }
.g-tabpanel { padding: var(--g-space-md) 0; }

/* Icon */
.g-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1em;
  height: 1em;
  vertical-align: -0.125em;
  fill: currentColor;
  flex-shrink: 0;
}

.g-icon-sm { width: 14px; height: 14px; }
.g-icon-md { width: 18px; height: 18px; }
.g-icon-lg { width: 24px; height: 24px; }

/* TreeView */
.g-tree { list-style: none; padding: 0; margin: 0; font-size: 13px; }
.g-tree .g-tree { padding-left: var(--g-space-lg); }

.g-tree-item { padding: 0; }

.g-tree-node {
  display: flex;
  align-items: center;
  gap: var(--g-space-xs);
  padding: 3px var(--g-space-sm);
  border-radius: var(--g-radius);
  cursor: default;
  color: var(--g-text);
  text-decoration: none;
}

.g-tree-node:hover { background: var(--g-bg-overlay); }
.g-tree-node-active { background: var(--g-accent-light); color: var(--g-accent); }

.g-tree-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border: none;
  background: none;
  cursor: pointer;
  padding: 0;
  color: var(--g-text-tertiary);
  font-size: 10px;
  transition: transform 0.15s;
  flex-shrink: 0;
}

.g-tree-toggle-open { transform: rotate(90deg); }
.g-tree-spacer { width: 18px; flex-shrink: 0; }
.g-tree-label { flex: 1; min-width: 0; }

/* Textarea */
.g-form-textarea {
  display: block;
  width: 100%;
  padding: 7px 10px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-inset);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  transition: border-color 0.1s, box-shadow 0.1s;
  outline: none;
  resize: vertical;
  min-height: 80px;
  line-height: 1.5;
}

.g-form-textarea:focus {
  border-color: var(--g-border-strong);
  box-shadow: 0 0 0 2px var(--g-accent-light);
}

/* Form Error */
.g-form-error {
  font-size: 12px;
  color: var(--g-danger);
  margin-top: var(--g-space-xs);
}

.g-form-input-error, .g-form-textarea-error, .g-form-select-error {
  border-color: var(--g-danger-border);
}

.g-form-input-error:focus, .g-form-textarea-error:focus, .g-form-select-error:focus {
  box-shadow: 0 0 0 2px rgba(139, 58, 58, 0.12);
}

/* Checkbox & Radio */
.g-form-check {
  display: flex;
  align-items: flex-start;
  gap: var(--g-space-sm);
  padding: 3px 0;
  cursor: pointer;
  font-size: 13px;
}

.g-form-check input[type="checkbox"],
.g-form-check input[type="radio"] {
  width: 16px;
  height: 16px;
  margin: 1px 0 0 0;
  accent-color: var(--g-accent);
  cursor: pointer;
  flex-shrink: 0;
}

.g-form-check-label { color: var(--g-text); cursor: pointer; user-select: none; }
.g-form-check-disabled { opacity: 0.5; cursor: not-allowed; }
.g-form-check-disabled input { cursor: not-allowed; }

/* SearchSelect */
.g-searchselect { position: relative; }

.g-searchselect-input {
  display: block;
  width: 100%;
  padding: 7px 10px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-inset);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  outline: none;
  transition: border-color 0.1s, box-shadow 0.1s;
}

.g-searchselect-input:focus {
  border-color: var(--g-border-strong);
  box-shadow: 0 0 0 2px var(--g-accent-light);
}

.g-searchselect-list {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  z-index: 910;
  max-height: 200px;
  overflow-y: auto;
  margin-top: 2px;
  background: var(--g-bg-surface);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  box-shadow: var(--g-shadow);
  list-style: none;
  padding: var(--g-space-xs) 0;
}

.g-searchselect-option {
  display: block;
  width: 100%;
  padding: var(--g-space-sm) var(--g-space-md);
  font-size: 13px;
  color: var(--g-text);
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  font-family: var(--g-font);
}

.g-searchselect-option:hover { background: var(--g-bg-overlay); }
.g-searchselect-option-highlight { background: var(--g-bg-overlay); }
.g-searchselect-option-active { background: var(--g-accent-light); }
.g-searchselect-empty {
  padding: var(--g-space-sm) var(--g-space-md);
  font-size: 13px;
  color: var(--g-text-tertiary);
}

/* Drawer */
.g-drawer-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.35);
  z-index: 1000;
  animation: g-fade-in 0.15s ease;
}

@keyframes g-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.g-drawer {
  position: fixed;
  top: 0;
  bottom: 0;
  z-index: 1001;
  width: 320px;
  max-width: 85vw;
  background: var(--g-bg-surface);
  box-shadow: var(--g-shadow-lg);
  display: flex;
  flex-direction: column;
  animation: g-drawer-slide 0.2s ease;
}

.g-drawer-left { left: 0; }
.g-drawer-right { right: 0; }

@keyframes g-drawer-slide {
  from { transform: translateX(-100%); }
  to { transform: translateX(0); }
}

.g-drawer-right.g-drawer {
  animation-name: g-drawer-slide-right;
}

@keyframes g-drawer-slide-right {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.g-drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--g-space-md) var(--g-space-lg);
  border-bottom: 1px solid var(--g-border);
  flex-shrink: 0;
}

.g-drawer-title { font-size: 15px; font-weight: 600; margin: 0; }

.g-drawer-close {
  background: none;
  border: none;
  font-size: 18px;
  color: var(--g-text-tertiary);
  cursor: pointer;
  padding: 4px;
  line-height: 1;
}

.g-drawer-close:hover { color: var(--g-text); }

.g-drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--g-space-lg);
}

/* Mobile hamburger for sidebar */
.g-mobile-header {
  display: none;
  align-items: center;
  gap: var(--g-space-sm);
  padding: var(--g-space-sm) var(--g-space-md);
  background: var(--g-bg-surface);
  border-bottom: 1px solid var(--g-border);
}

.g-hamburger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  background: var(--g-bg-surface);
  cursor: pointer;
  font-size: 18px;
  color: var(--g-text);
}

.g-hamburger:hover { background: var(--g-bg-overlay); }

.g-mobile-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--g-text);
}

/* Flexbox layout */
.g-row    { display:flex; flex-direction:row; gap:var(--g-space-sm); }
.g-col    { display:flex; flex-direction:column; gap:var(--g-space-sm); }
.g-stack  { display:flex; flex-direction:column; gap:var(--g-space-md); }
.g-hstack { display:flex; flex-direction:row; flex-wrap:wrap; align-items:center; gap:var(--g-space-sm); }
.g-vstack { display:flex; flex-direction:column; align-items:center; gap:var(--g-space-sm); }
.g-center { display:flex; align-items:center; justify-content:center; }

/* Container */
.g-container { width:100%; max-width:960px; margin:0 auto; padding:0 var(--g-space-lg); }
.g-container-narrow { max-width:640px; }
.g-container-wide   { max-width:1280px; }

/* Spacer */
.g-spacer { flex:1; }
.g-space-y-xs { height:var(--g-space-xs); }
.g-space-y-sm { height:var(--g-space-sm); }
.g-space-y-md { height:var(--g-space-md); }
.g-space-y-lg { height:var(--g-space-lg); }
.g-space-y-xl { height:var(--g-space-xl); }

/* Gap modifiers */
.g-gap-none { gap:0; }
.g-gap-xs { gap:var(--g-space-xs); }
.g-gap-sm { gap:var(--g-space-sm); }
.g-gap-md { gap:var(--g-space-md); }
.g-gap-lg { gap:var(--g-space-lg); }
.g-gap-xl { gap:var(--g-space-xl); }

/* Flex wrap */
.g-wrap { flex-wrap:wrap; }

/* justify-content */
.g-justify-start { justify-content:flex-start; }
.g-justify-center { justify-content:center; }
.g-justify-end { justify-content:flex-end; }
.g-justify-between { justify-content:space-between; }
.g-justify-around { justify-content:space-around; }

/* align-items */
.g-align-start { align-items:flex-start; }
.g-align-center { align-items:center; }
.g-align-end { align-items:flex-end; }
.g-align-stretch { align-items:stretch; }
.g-align-baseline { align-items:baseline; }

/* Flex child */
.g-grow { flex-grow:1; }
.g-shrink-0 { flex-shrink:0; }

/* Grid helper */
.g-grid { display: grid; gap: var(--g-space-lg); }
.g-grid-2 { grid-template-columns: repeat(2, 1fr); }
.g-grid-3 { grid-template-columns: repeat(3, 1fr); }
.g-grid-4 { grid-template-columns: repeat(4, 1fr); }
.g-grid-5 { grid-template-columns: repeat(5, 1fr); }
.g-grid-6 { grid-template-columns: repeat(6, 1fr); }

/* ============ Responsive ============ */

/* Tablet: <= 1024px */
@media (max-width: 1024px) {
  .g-grid-4 { grid-template-columns: repeat(2, 1fr); }
  .g-grid-3 { grid-template-columns: repeat(2, 1fr); }
  .g-grid-5, .g-grid-6 { grid-template-columns: repeat(3, 1fr); }

  .g-container { padding: 0 var(--g-space-md); }

  .g-page-header {
    padding: var(--g-space-md) var(--g-space-lg);
    flex-wrap: wrap;
    gap: var(--g-space-sm);
  }

  .g-page-body { padding: var(--g-space-lg); }

  .g-stat-value { font-size: 22px; }
}

/* Mobile: <= 768px */
@media (max-width: 768px) {
  .g-sidebar { display: none; }

  .g-mobile-header { display: flex; }

  .g-admin-shell { flex-direction: column; }

  .g-grid-2, .g-grid-3, .g-grid-4, .g-grid-5, .g-grid-6 { grid-template-columns: 1fr; }

  .g-stack { gap: var(--g-space-sm); }
  .g-container { padding: 0 var(--g-space-sm); }

  .g-page-header {
    padding: var(--g-space-sm) var(--g-space-md);
  }

  .g-page-header-title { font-size: 16px; }

  .g-page-body { padding: var(--g-space-md); }

  .g-card-header { padding: var(--g-space-sm) var(--g-space-md); }

  .g-table { font-size: 12px; }
  .g-table th, .g-table td { padding: var(--g-space-xs) var(--g-space-sm); }

  .g-card { overflow-x: auto; }

  .g-modal-panel { width: 95%; max-width: none; }

  .g-toast { left: var(--g-space-md); right: var(--g-space-md); min-width: unset; max-width: none; }

  .g-datatable-footer { flex-direction: column; gap: var(--g-space-sm); align-items: flex-start; }

  .g-breadcrumb { flex-wrap: wrap; }

  .g-drawer { width: 280px; }
}

/* Small mobile: <= 480px */
@media (max-width: 480px) {
  body { font-size: 13px; }

  .g-btn { padding: 6px 10px; font-size: 12px; }

  .g-stat { padding: var(--g-space-md); }
  .g-stat-value { font-size: 20px; }
  .g-stat-label { font-size: 11px; }

  .g-form-input, .g-form-select, .g-form-textarea, .g-searchselect-input {
    font-size: 16px; /* prevent iOS zoom */
  }
}
`
