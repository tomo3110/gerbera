package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
	gl "github.com/tomo3110/gerbera/live"
)

// MarkdownView implements gl.View for the Markdown viewer/editor.
type MarkdownView struct {
	Source     string    `json:"source"`
	HTML       string    `json:"html"`
	FilePath   string    `json:"filePath"`
	Preview    bool      `json:"preview"`
	ModTime    time.Time `json:"modTime"`
	FileName   string    `json:"fileName"`
	SaveStatus string    `json:"saveStatus"`
}

func (v *MarkdownView) Mount(_ gl.Params) error {
	if v.FilePath != "" {
		data, err := os.ReadFile(v.FilePath)
		if err != nil {
			return err
		}
		v.Source = string(data)
		info, err := os.Stat(v.FilePath)
		if err != nil {
			return err
		}
		v.ModTime = info.ModTime()
		v.FileName = filepath.Base(v.FilePath)
	}
	if v.Source == "" && !v.Preview {
		v.Source = "# Hello Markdown\n\nStart typing here..."
	}
	v.HTML = renderMarkdown(v.Source)
	return nil
}

func (v *MarkdownView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "edit":
		v.Source = payload["value"]
		v.HTML = renderMarkdown(v.Source)
		v.SaveStatus = ""
	case "save":
		if v.FilePath == "" {
			return nil
		}
		if err := os.WriteFile(v.FilePath, []byte(v.Source), 0644); err != nil {
			v.SaveStatus = "error"
			return nil
		}
		info, err := os.Stat(v.FilePath)
		if err == nil {
			v.ModTime = info.ModTime()
		}
		v.SaveStatus = "saved"
	case "check-file":
		if v.FilePath == "" {
			return nil
		}
		info, err := os.Stat(v.FilePath)
		if err != nil {
			return nil
		}
		if info.ModTime().After(v.ModTime) {
			data, err := os.ReadFile(v.FilePath)
			if err != nil {
				return nil
			}
			v.Source = string(data)
			v.HTML = renderMarkdown(v.Source)
			v.ModTime = info.ModTime()
		}
	}
	return nil
}

func (v *MarkdownView) Render() []g.ComponentFunc {
	title := "Markdown Viewer"
	if v.FileName != "" {
		title = v.FileName + " — Markdown Viewer"
	}

	return []g.ComponentFunc{
		gd.Head(
			gd.Title(title),
			gd.Meta(gp.Attr("charset", "utf-8")),
			gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
			g.Tag("style",gp.Value(cssStyles)),
		),
		gd.Body(
			gs.Style(g.StyleMap{"margin": "0", "padding": "0", "font-family": "-apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif"}),
			v.renderHeader(),
			v.renderMain(),
			v.renderScripts(),
		),
	}
}

func (v *MarkdownView) renderHeader() g.ComponentFunc {
	return gd.Header(
		gs.Style(g.StyleMap{
			"background":    "#24292e",
			"color":         "#fff",
			"padding":       "8px 16px",
			"display":       "flex",
			"align-items":   "center",
			"gap":           "12px",
			"height":        "40px",
			"box-sizing":    "border-box",
		}),
		g.Tag("span",
			gs.Style(g.StyleMap{"font-weight": "bold", "font-size": "14px"}),
			ge.If(v.FileName != "",
				gp.Value(v.FileName),
				gp.Value("Markdown Viewer"),
			),
		),
		ge.If(v.FilePath != "" && !v.Preview,
			gd.Button(
				gl.Click("save"),
				gs.Style(g.StyleMap{
					"background":    "#2ea44f",
					"color":         "#fff",
					"border":        "none",
					"padding":       "4px 12px",
					"border-radius": "4px",
					"cursor":        "pointer",
					"font-size":     "12px",
				}),
				gp.Value("Save"),
			),
			g.Skip(),
		),
		ge.If(v.SaveStatus == "saved",
			g.Tag("span",gs.Style(g.StyleMap{"color": "#85e89d", "font-size": "12px"}), gp.Value("Saved")),
			g.Skip(),
		),
		ge.If(v.SaveStatus == "error",
			g.Tag("span",gs.Style(g.StyleMap{"color": "#f97583", "font-size": "12px"}), gp.Value("Save failed")),
			g.Skip(),
		),
		ge.If(v.FilePath == "" && !v.Preview,
			v.renderDownloadLink(),
			g.Skip(),
		),
	)
}

func (v *MarkdownView) renderDownloadLink() g.ComponentFunc {
	encoded := url.PathEscape(v.Source)
	raw := fmt.Sprintf(
		`<a download="document.md" href="data:text/markdown;charset=utf-8,%s" style="color:#79b8ff;font-size:12px;text-decoration:none;">Download .md</a>`,
		encoded,
	)
	return g.Tag("span",g.Literal(raw))
}

func (v *MarkdownView) renderMain() g.ComponentFunc {
	return gd.Main(
		gp.Class("md-container"),
		ge.Unless(v.Preview,
			gd.Div(
				gp.Class("md-editor-pane"),
				gd.Textarea(
					gp.ID("md-editor"),
					gp.Class("md-textarea"),
					gl.Input("edit"),
					gp.Value(v.Source),
				),
			),
		),
		gd.Div(
			gp.Class("md-preview-pane"),
			gd.Div(
				gp.ID("md-preview"),
				gp.Class("md-preview"),
				g.Literal(v.HTML),
			),
		),
	)
}

func (v *MarkdownView) renderScripts() g.ComponentFunc {
	var scripts string
	if !v.Preview {
		scripts += `<script>
(function() {
  var ta = document.getElementById('md-editor');
  var pv = document.querySelector('.md-preview-pane');
  if (ta && pv) {
    ta.addEventListener('scroll', function() {
      var max = ta.scrollHeight - ta.clientHeight;
      if (max <= 0) return;
      var pct = ta.scrollTop / max;
      pv.scrollTop = pct * (pv.scrollHeight - pv.clientHeight);
    });
  }
})();
</script>`
	}
	if v.FilePath != "" {
		scripts += `<script>
(function() {
  var btn = document.getElementById('file-check-trigger');
  if (btn) setInterval(function() { btn.click(); }, 2000);
})();
</script>`
	}
	return gd.Div(
		ge.If(v.FilePath != "",
			gd.Button(
				gp.ID("file-check-trigger"),
				gl.Click("check-file"),
				gs.Style(g.StyleMap{"display": "none"}),
			),
			g.Skip(),
		),
		g.Literal(scripts),
	)
}

const cssStyles = `
* { box-sizing: border-box; }
.md-container {
  display: flex;
  height: calc(100vh - 40px);
}
.md-editor-pane {
  flex: 1;
  display: flex;
  border-right: 1px solid #e1e4e8;
}
.md-textarea {
  width: 100%;
  height: 100%;
  border: none;
  outline: none;
  resize: none;
  padding: 16px;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 14px;
  line-height: 1.6;
  background: #fafbfc;
}
.md-preview-pane {
  flex: 1;
  overflow-y: auto;
}
.md-preview {
  padding: 16px 32px;
  font-size: 14px;
  line-height: 1.6;
  color: #24292e;
}
.md-preview h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; margin-top: 24px; margin-bottom: 16px; }
.md-preview h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; margin-top: 24px; margin-bottom: 16px; }
.md-preview h3 { font-size: 1.25em; margin-top: 24px; margin-bottom: 16px; }
.md-preview p { margin-top: 0; margin-bottom: 16px; }
.md-preview code {
  background: #f6f8fa;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-size: 85%;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}
.md-preview pre {
  background: #f6f8fa;
  padding: 16px;
  border-radius: 6px;
  overflow-x: auto;
  line-height: 1.45;
}
.md-preview pre code { background: none; padding: 0; font-size: 100%; }
.md-preview blockquote {
  margin: 0 0 16px 0;
  padding: 0 1em;
  color: #6a737d;
  border-left: 0.25em solid #dfe2e5;
}
.md-preview table { border-collapse: collapse; margin-bottom: 16px; }
.md-preview th, .md-preview td { border: 1px solid #dfe2e5; padding: 6px 13px; }
.md-preview th { background: #f6f8fa; font-weight: 600; }
.md-preview img { max-width: 100%; }
.md-preview a { color: #0366d6; text-decoration: none; }
.md-preview a:hover { text-decoration: underline; }
.md-preview ul, .md-preview ol { padding-left: 2em; margin-bottom: 16px; }
.md-preview li + li { margin-top: 0.25em; }
.md-preview hr { height: 0.25em; padding: 0; margin: 24px 0; background-color: #e1e4e8; border: 0; }
.md-preview input[type="checkbox"] { margin-right: 0.5em; }
`
