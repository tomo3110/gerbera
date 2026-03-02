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
		Bg:        "#fafafa",
		BgSurface: "#ffffff",
		BgOverlay: "#f5f5f5",
		BgInset:   "#f0f0f1",

		Text:          "#1a1a1a",
		TextSecondary: "#737373",
		TextTertiary:  "#a3a3a3",
		TextInverse:   "#ffffff",

		Border:       "#e5e5e5",
		BorderStrong: "#d4d4d4",

		ShadowSm: "0 1px 2px rgba(0,0,0,0.03), 0 1px 3px rgba(0,0,0,0.02)",
		Shadow:   "0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.03)",
		ShadowLg: "0 8px 30px rgba(0,0,0,0.06), 0 2px 8px rgba(0,0,0,0.03)",

		Accent:      "#171717",
		AccentLight: "#f5f5f5",
		AccentHover: "#262626",

		Danger:       "#dc2626",
		DangerBg:     "#fef2f2",
		DangerBorder: "#fecaca",
		DangerHover:  "#b91c1c",

		Success:       "#16a34a",
		SuccessBg:     "#f0fdf4",
		SuccessBorder: "#bbf7d0",

		Warning:       "#ca8a04",
		WarningBg:     "#fefce8",
		WarningBorder: "#fef08a",

		Info:       "#2563eb",
		InfoBg:     "#eff6ff",
		InfoBorder: "#bfdbfe",

		SpaceXs: "4px",
		SpaceSm: "8px",
		SpaceMd: "16px",
		SpaceLg: "28px",
		SpaceXl: "44px",

		Font:     `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`,
		FontMono: `"SF Mono", "Cascadia Code", Consolas, monospace`,
		Radius:   "10px",
	}
}

// DarkTheme returns a dark theme configuration.
func DarkTheme() ThemeConfig {
	return ThemeConfig{
		Bg:        "#0a0a0a",
		BgSurface: "#171717",
		BgOverlay: "#262626",
		BgInset:   "#1c1c1e",

		Text:          "#fafafa",
		TextSecondary: "#a3a3a3",
		TextTertiary:  "#737373",
		TextInverse:   "#0a0a0a",

		Border:       "#262626",
		BorderStrong: "#404040",

		ShadowSm: "0 1px 2px rgba(0,0,0,0.15), 0 1px 3px rgba(0,0,0,0.12)",
		Shadow:   "0 2px 8px rgba(0,0,0,0.20), 0 1px 3px rgba(0,0,0,0.15)",
		ShadowLg: "0 8px 30px rgba(0,0,0,0.25), 0 2px 8px rgba(0,0,0,0.15)",

		Accent:      "#e5e5e5",
		AccentLight: "#262626",
		AccentHover: "#ffffff",

		Danger:       "#ef4444",
		DangerBg:     "#3d1f1f",
		DangerBorder: "#6b3333",
		DangerHover:  "#dc2626",

		Success:       "#4ade80",
		SuccessBg:     "#1f3d1f",
		SuccessBorder: "#2b6b2b",

		Warning:       "#facc15",
		WarningBg:     "#3d331f",
		WarningBorder: "#6b5c2b",

		Info:       "#60a5fa",
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
  border: none;
  border-radius: 12px;
  box-shadow: var(--g-shadow);
}

.g-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--g-space-md) var(--g-space-lg);
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
  background: var(--g-bg);
  color: var(--g-text-secondary);
  font-size: 13px;
  border-radius: 0 0 12px 12px;
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
  font-weight: 500;
  color: var(--g-text-secondary);
  border-bottom: 1px solid var(--g-border);
  font-size: 12px;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.g-table td {
  padding: var(--g-space-sm) var(--g-space-md);
  color: var(--g-text);
}

.g-table tbody tr:hover {
  background: var(--g-bg-overlay);
}

.g-table tbody tr:nth-child(even) {
  background: var(--g-bg);
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
  border-radius: 999px;
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
  border: none;
}

.g-alert-info { background: var(--g-info-bg); color: var(--g-info); }
.g-alert-success { background: var(--g-success-bg); color: var(--g-success); }
.g-alert-warning { background: var(--g-warning-bg); color: var(--g-warning); }
.g-alert-danger { background: var(--g-danger-bg); color: var(--g-danger); }

/* StatCard */
.g-stat {
  background: var(--g-bg-surface);
  border: none;
  border-radius: 12px;
  padding: var(--g-space-lg);
  box-shadow: var(--g-shadow);
}

.g-stat-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--g-text-tertiary);
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
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.g-sidebar-header {
  padding: var(--g-space-lg);
  font-size: 16px;
  font-weight: 700;
  color: var(--g-text);
  letter-spacing: -0.01em;
}

.g-sidebar-link {
  display: block;
  padding: var(--g-space-sm) var(--g-space-lg);
  color: var(--g-text-secondary);
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
}

.g-sidebar-link:hover { background: var(--g-bg-overlay); color: var(--g-text); }
.g-sidebar-link-active { background: var(--g-accent-light); color: var(--g-accent); }

.g-sidebar-divider {
  height: 1px;
  background: transparent;
  margin: var(--g-space-md) 0;
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
  padding: 8px 20px;
  font-size: 13px;
  font-weight: 500;
  font-family: var(--g-font);
  border-radius: 8px;
  border: 1px solid transparent;
  background: var(--g-bg-surface);
  color: var(--g-text);
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, box-shadow 0.15s;
  line-height: 1.4;
  white-space: nowrap;
  box-shadow: var(--g-shadow-sm);
}

.g-btn:hover { background: var(--g-bg-overlay); border-color: transparent; }
.g-btn:active { box-shadow: inset 0 1px 2px rgba(0,0,0,0.06); }

.g-btn-primary { background: var(--g-accent); color: var(--g-text-inverse); border-color: transparent; }
.g-btn-primary:hover { background: var(--g-accent-hover); }

.g-btn-outline { background: transparent; box-shadow: none; border-color: var(--g-border); }
.g-btn-outline:hover { background: var(--g-bg-overlay); }

.g-btn-danger { background: var(--g-danger); color: var(--g-text-inverse); border-color: transparent; }
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
  padding: 10px 14px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-inset);
  border: 1px solid transparent;
  border-radius: 8px;
  transition: border-color 0.15s, box-shadow 0.15s;
  outline: none;
}

.g-form-input:focus, .g-form-select:focus {
  border-color: transparent;
  box-shadow: 0 0 0 3px var(--g-accent-light);
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
  height: 8px;
  background: var(--g-bg-inset);
  border-radius: 4px;
  overflow: hidden;
}

.g-progress-bar {
  height: 100%;
  background: var(--g-accent);
  border-radius: 4px;
  transition: width 0.3s ease;
}

/* Modal */
.g-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.g-modal-panel {
  background: var(--g-bg-surface);
  border-radius: 16px;
  box-shadow: var(--g-shadow-lg);
  width: 90%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
}

.g-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--g-space-md) var(--g-space-lg);
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
  border: none;
  animation: g-toast-in 0.2s ease;
}

@keyframes g-toast-in {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

.g-toast-info { background: var(--g-info-bg); color: var(--g-info); }
.g-toast-success { background: var(--g-success-bg); color: var(--g-success); }
.g-toast-warning { background: var(--g-warning-bg); color: var(--g-warning); }
.g-toast-danger { background: var(--g-danger-bg); color: var(--g-danger); }

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
  border: none;
  border-radius: var(--g-radius);
  box-shadow: var(--g-shadow-lg);
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
  transition: color 0.15s, border-color 0.15s, background 0.15s;
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
  padding: 5px var(--g-space-sm);
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
  padding: 10px 14px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-inset);
  border: 1px solid transparent;
  border-radius: 8px;
  transition: border-color 0.15s, box-shadow 0.15s;
  outline: none;
  resize: vertical;
  min-height: 80px;
  line-height: 1.5;
}

.g-form-textarea:focus {
  border-color: transparent;
  box-shadow: 0 0 0 3px var(--g-accent-light);
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
  box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.10);
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
  padding: 10px 14px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-inset);
  border: 1px solid transparent;
  border-radius: 8px;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.g-searchselect-input:focus {
  border-color: transparent;
  box-shadow: 0 0 0 3px var(--g-accent-light);
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
  border: none;
  border-radius: var(--g-radius);
  box-shadow: var(--g-shadow-lg);
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
  background: rgba(0,0,0,0.2);
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
}

.g-hamburger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
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

/* ============ Spinner ============ */

@keyframes g-spin {
  to { transform: rotate(360deg); }
}

.g-spinner {
  display: flex;
  align-items: center;
  justify-content: center;
}

.g-spinner-inline { display: inline-flex; }

.g-spinner-arc {
  display: block;
  border-radius: 50%;
  border: 2px solid var(--g-border);
  border-top-color: var(--g-accent);
  animation: g-spin 0.6s linear infinite;
}

.g-spinner-sm .g-spinner-arc { width: 16px; height: 16px; }
.g-spinner-md .g-spinner-arc { width: 24px; height: 24px; }
.g-spinner-lg .g-spinner-arc { width: 40px; height: 40px; border-width: 3px; }

/* ============ NumberInput ============ */

.g-numberinput {
  display: inline-flex;
  align-items: center;
  background: var(--g-bg-inset);
  border-radius: 8px;
  overflow: hidden;
}

.g-numberinput-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  color: var(--g-text);
  font-size: 16px;
  font-family: var(--g-font);
  cursor: pointer;
  transition: background 0.15s;
  flex-shrink: 0;
}

.g-numberinput-btn:hover:not(:disabled) { background: var(--g-bg-overlay); }
.g-numberinput-btn:disabled { color: var(--g-text-tertiary); cursor: not-allowed; }

.g-numberinput-field {
  width: 64px;
  text-align: center;
  border: none;
  background: transparent;
  font-size: 14px;
  font-family: var(--g-font);
  color: var(--g-text);
  outline: none;
  padding: var(--g-space-xs) 0;
  -moz-appearance: textfield;
}

.g-numberinput-field::-webkit-inner-spin-button,
.g-numberinput-field::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* ============ TimePicker ============ */

.g-timepicker {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.g-timepicker-unit {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.g-timepicker-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--g-text-secondary);
  font-size: 10px;
  font-family: var(--g-font);
  cursor: pointer;
  transition: background 0.15s;
  flex-shrink: 0;
  border-radius: 4px;
}

.g-timepicker-btn:hover:not(:disabled) { background: var(--g-bg-inset); }
.g-timepicker-btn:disabled { color: var(--g-text-tertiary); cursor: not-allowed; }

.g-timepicker-field {
  width: 40px;
  text-align: center;
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  background: var(--g-bg-surface);
  font-size: 16px;
  font-family: monospace;
  color: var(--g-text);
  outline: none;
  padding: 4px 0;
}

.g-timepicker-field:focus {
  border-color: var(--g-accent);
  box-shadow: 0 0 0 3px var(--g-accent-light);
}

.g-timepicker-sep {
  font-size: 18px;
  font-weight: 700;
  color: var(--g-text-secondary);
  line-height: 1;
  padding: 0 2px;
}

.g-timepicker-ampm {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-left: 8px;
}

.g-timepicker-ampm-btn {
  padding: 4px 8px;
  font-size: 11px;
  font-weight: 600;
  font-family: var(--g-font);
  color: var(--g-text-secondary);
  background: var(--g-bg-inset);
  border: 1px solid var(--g-border);
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.g-timepicker-ampm-btn:hover:not(:disabled) { background: var(--g-bg-overlay); }

.g-timepicker-ampm-btn[aria-pressed="true"] {
  background: var(--g-accent);
  color: var(--g-text-inverse);
  border-color: var(--g-accent);
}

.g-timepicker-ampm-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* ============ Slider ============ */

.g-slider { width: 100%; }

.g-slider-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: var(--g-space-xs);
}

.g-slider-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--g-text);
}

.g-slider-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--g-text);
  font-family: var(--g-font-mono);
}

.g-slider-input {
  -webkit-appearance: none;
  appearance: none;
  width: 100%;
  height: 6px;
  border-radius: 3px;
  background: var(--g-bg-inset);
  outline: none;
  cursor: pointer;
}

.g-slider-input::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--g-bg-surface);
  box-shadow: var(--g-shadow), 0 0 0 1px var(--g-border);
  cursor: pointer;
  transition: box-shadow 0.15s;
}

.g-slider-input::-webkit-slider-thumb:hover {
  box-shadow: var(--g-shadow-lg), 0 0 0 1px var(--g-border-strong);
}

.g-slider-input::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--g-bg-surface);
  box-shadow: var(--g-shadow), 0 0 0 1px var(--g-border);
  border: none;
  cursor: pointer;
  transition: box-shadow 0.15s;
}

.g-slider-input::-moz-range-thumb:hover {
  box-shadow: var(--g-shadow-lg), 0 0 0 1px var(--g-border-strong);
}

.g-slider-input::-moz-range-track {
  height: 6px;
  border-radius: 3px;
  background: var(--g-bg-inset);
}

/* ============ Calendar ============ */

.g-calendar {
  background: var(--g-bg-surface);
  border-radius: 12px;
  box-shadow: var(--g-shadow);
  padding: var(--g-space-md);
  width: 100%;
  max-width: 320px;
}

.g-calendar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--g-space-sm);
}

.g-calendar-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--g-text);
}

.g-calendar-selectors {
  display: flex;
  align-items: center;
  gap: var(--g-space-xs);
}

.g-calendar-select {
  appearance: none;
  -webkit-appearance: none;
  background: var(--g-bg-inset);
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  color: var(--g-text);
  font-size: 13px;
  font-weight: 600;
  font-family: var(--g-font);
  padding: 4px 24px 4px 8px;
  cursor: pointer;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6'%3E%3Cpath d='M0 0l5 6 5-6z' fill='%23888'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 6px center;
  transition: border-color 0.15s;
}

.g-calendar-select:hover {
  border-color: var(--g-border-strong);
}

.g-calendar-select:focus {
  outline: none;
  border-color: var(--g-accent);
  box-shadow: 0 0 0 2px var(--g-accent-light);
}

.g-calendar-nav {
  min-width: 28px;
  padding: 2px 6px;
}

.g-calendar-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  text-align: center;
  margin-bottom: var(--g-space-xs);
}

.g-calendar-dayname {
  font-size: 11px;
  font-weight: 500;
  color: var(--g-text-tertiary);
  padding: var(--g-space-xs) 0;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.g-calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}

.g-calendar-day {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 36px;
  font-size: 13px;
  border-radius: 8px;
  cursor: pointer;
  color: var(--g-text);
  transition: background 0.15s;
}

.g-calendar-day:hover { background: var(--g-bg-overlay); }

.g-calendar-day-outside {
  color: var(--g-text-tertiary);
  cursor: default;
}

.g-calendar-day-outside:hover { background: transparent; }

.g-calendar-day-today {
  font-weight: 700;
  box-shadow: inset 0 0 0 1px var(--g-border-strong);
}

.g-calendar-day-selected {
  background: var(--g-accent);
  color: var(--g-text-inverse);
  font-weight: 600;
}

.g-calendar-day-selected:hover { background: var(--g-accent-hover); }

/* ============ Chat ============ */

.g-chat {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-sm);
  padding: var(--g-space-md);
  overflow-y: auto;
}

.g-chat-row {
  display: flex;
  align-items: flex-end;
  gap: var(--g-space-sm);
}

.g-chat-row-received { justify-content: flex-start; }
.g-chat-row-sent { justify-content: flex-end; }

.g-chat-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--g-bg-inset);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  flex-shrink: 0;
  color: var(--g-text-secondary);
}

.g-chat-bubble {
  max-width: 75%;
  padding: var(--g-space-sm) var(--g-space-md);
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
}

.g-chat-bubble-received {
  background: var(--g-bg-surface);
  box-shadow: var(--g-shadow-sm);
  border-radius: 12px 12px 12px 4px;
  color: var(--g-text);
}

.g-chat-bubble-sent {
  background: var(--g-bg-inset);
  color: var(--g-text);
  border-radius: 12px 12px 4px 12px;
}

.g-chat-author {
  font-size: 11px;
  font-weight: 600;
  margin-bottom: 2px;
  color: var(--g-text-secondary);
}

.g-chat-bubble-sent .g-chat-author {
  color: var(--g-text-secondary);
}

.g-chat-content { white-space: pre-wrap; }

.g-chat-time {
  display: block;
  font-size: 10px;
  margin-top: 4px;
  opacity: 0.6;
}

.g-chat-inputbar {
  display: flex;
  gap: var(--g-space-sm);
  padding: var(--g-space-sm) var(--g-space-md);
  background: var(--g-bg-surface);
  border-top: 1px solid var(--g-border);
  align-items: center;
}

.g-chat-input-field {
  flex: 1;
  padding: 10px 14px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-inset);
  border: 1px solid transparent;
  border-radius: 8px;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.g-chat-input-field:focus {
  border-color: transparent;
  box-shadow: 0 0 0 3px var(--g-accent-light);
}

.g-chat-send { flex-shrink: 0; }

/* ── Pagination ── */

.g-pagination {
  display: flex;
  align-items: center;
  gap: var(--g-space-md);
  flex-wrap: wrap;
}

.g-pagination-info {
  font-size: 13px;
  color: var(--g-text-secondary);
  white-space: nowrap;
}

.g-pagination-pages {
  display: flex;
  align-items: center;
  gap: 2px;
}

.g-pagination-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  padding: 0 8px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: transparent;
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
}

.g-pagination-btn:hover:not(:disabled) {
  background: var(--g-bg-inset);
  border-color: var(--g-border-strong);
}

.g-pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.g-pagination-active {
  background: var(--g-accent);
  color: var(--g-text-inverse);
  border-color: var(--g-accent);
  font-weight: 600;
}

.g-pagination-active:hover:not(:disabled) {
  background: var(--g-accent-hover);
  border-color: var(--g-accent-hover);
}

.g-pagination-ellipsis {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  font-size: 13px;
  color: var(--g-text-tertiary);
}

.g-pagination-prev,
.g-pagination-next {
  font-size: 18px;
  font-weight: 600;
}

/* ── ButtonGroup ── */

.g-btngroup {
  display: inline-flex;
  border-radius: var(--g-radius);
  overflow: hidden;
  border: 1px solid var(--g-border);
}

.g-btngroup-btn {
  padding: 8px 16px;
  font-size: 13px;
  font-family: var(--g-font);
  color: var(--g-text);
  background: var(--g-bg-surface);
  border: none;
  border-right: 1px solid var(--g-border);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
}

.g-btngroup-btn:last-child { border-right: none; }

.g-btngroup-btn:hover:not(.g-btngroup-active) {
  background: var(--g-bg-inset);
}

.g-btngroup-active {
  background: var(--g-accent);
  color: var(--g-text-inverse);
  font-weight: 600;
}

.g-btngroup-sm .g-btngroup-btn {
  padding: 4px 10px;
  font-size: 12px;
}

/* ── Accordion ── */

.g-accordion {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.g-accordion-item {
  border-radius: var(--g-radius);
  background: var(--g-bg-surface);
  box-shadow: var(--g-shadow-sm);
  overflow: hidden;
}

.g-accordion-header {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 14px 16px;
  font-size: 14px;
  font-weight: 500;
  font-family: var(--g-font);
  color: var(--g-text);
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s, color 0.15s;
  border-radius: var(--g-radius);
}

.g-accordion-header:hover {
  background: var(--g-bg-inset);
}

.g-accordion-header::before {
  content: "\25B8";
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-right: 10px;
  border-radius: 4px;
  background: var(--g-bg-inset);
  font-size: 11px;
  color: var(--g-text-secondary);
  transition: transform 0.2s, background 0.2s, color 0.2s;
  flex-shrink: 0;
}

details[open] > .g-accordion-header::before,
.g-accordion-open > .g-accordion-header::before {
  transform: rotate(90deg);
  background: var(--g-accent-light);
  color: var(--g-accent);
}

.g-accordion-body {
  padding: 0 16px 14px 46px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--g-text-secondary);
}

/* ── Stepper ── */

.g-stepper {
  display: flex;
  align-items: flex-start;
}

.g-stepper-step {
  display: flex;
  align-items: center;
  position: relative;
  flex: 1;
  min-width: 0;
}

.g-stepper-step:last-child { flex: 0 0 auto; }

.g-stepper-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
  transition: background 0.2s, box-shadow 0.2s;
}

.g-stepper-completed .g-stepper-indicator {
  background: var(--g-accent);
  color: var(--g-text-inverse);
}

.g-stepper-active .g-stepper-indicator {
  background: var(--g-bg-surface);
  color: var(--g-accent);
  border: 2px solid var(--g-accent);
  box-shadow: 0 0 0 3px var(--g-accent-light);
}

.g-stepper-upcoming .g-stepper-indicator {
  background: var(--g-bg-inset);
  color: var(--g-text-tertiary);
  border: 1px solid var(--g-border);
}

.g-stepper-content {
  display: flex;
  flex-direction: column;
  margin-left: 8px;
  min-width: 0;
}

.g-stepper-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--g-text);
  white-space: nowrap;
}

.g-stepper-active .g-stepper-label { color: var(--g-accent); font-weight: 600; }

.g-stepper-upcoming .g-stepper-label { color: var(--g-text-tertiary); }

.g-stepper-desc {
  font-size: 11px;
  color: var(--g-text-tertiary);
  margin-top: 2px;
}

.g-stepper-connector {
  flex: 1;
  height: 2px;
  background: var(--g-border);
  margin: 0 12px;
  align-self: center;
  min-width: 24px;
}

.g-stepper-completed + .g-stepper-step .g-stepper-connector,
.g-stepper-completed .g-stepper-connector {
  background: var(--g-accent);
}

/* Vertical stepper */

.g-stepper-vertical {
  flex-direction: column;
}

.g-stepper-vertical .g-stepper-step {
  flex-direction: row;
  align-items: flex-start;
  flex: 0 0 auto;
  padding-bottom: 24px;
  position: relative;
}

.g-stepper-vertical .g-stepper-step:last-child { padding-bottom: 0; }

.g-stepper-vertical .g-stepper-connector {
  position: absolute;
  left: 15px;
  top: 36px;
  bottom: 0;
  width: 2px;
  height: auto;
  margin: 0;
  min-width: 0;
}

@media (max-width: 768px) {
  .g-stepper:not(.g-stepper-vertical) {
    flex-direction: column;
  }
  .g-stepper:not(.g-stepper-vertical) .g-stepper-step {
    flex-direction: row;
    flex: 0 0 auto;
    padding-bottom: 24px;
  }
  .g-stepper:not(.g-stepper-vertical) .g-stepper-step:last-child { padding-bottom: 0; }
  .g-stepper:not(.g-stepper-vertical) .g-stepper-connector {
    position: absolute;
    left: 15px;
    top: 36px;
    bottom: 0;
    width: 2px;
    height: auto;
    margin: 0;
    min-width: 0;
  }
}

/* ── InfiniteScroll ── */

.g-infinitescroll {
  display: flex;
  flex-direction: column;
}

.g-infinitescroll-toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  padding: var(--g-space-sm) 0;
}

.g-infinitescroll-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  font-size: 16px;
  background: transparent;
  border: 1px solid var(--g-border);
  border-radius: var(--g-radius);
  color: var(--g-text-secondary);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.g-infinitescroll-toggle:hover { background: var(--g-bg-inset); }

.g-infinitescroll-toggle-active {
  background: var(--g-accent);
  color: var(--g-text-inverse);
  border-color: var(--g-accent);
}

.g-infinitescroll-toggle-active:hover {
  background: var(--g-accent-hover);
}

.g-infinitescroll-content {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-sm);
  max-height: 400px;
  overflow-y: auto;
  padding: var(--g-space-xs);
}

.g-infinitescroll-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  flex-direction: unset;
}

.g-infinitescroll-loader {
  display: flex;
  justify-content: center;
  padding: var(--g-space-md);
}
`
