package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)

// FileInfo holds metadata about an uploaded file.
type FileInfo struct {
	Name     string
	Size     int64
	MIMEType string
}

func (f FileInfo) ToMap() g.Map {
	return g.Map{
		"name":     f.Name,
		"size":     f.Size,
		"mimetype": f.MIMEType,
	}
}

// UploadView demonstrates file upload with the UploadHandler interface.
type UploadView struct {
	Files []FileInfo
}

func (v *UploadView) Mount(_ gl.Params) error {
	return nil
}

func (v *UploadView) HandleUpload(_ string, files []gl.UploadedFile) error {
	for _, f := range files {
		v.Files = append(v.Files, FileInfo{
			Name:     f.Name,
			Size:     f.Size,
			MIMEType: f.MIMEType,
		})
	}
	return nil
}

func (v *UploadView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "gerbera:upload_complete":
		// UI will re-render automatically; files were already added in HandleUpload.
	case "clear":
		v.Files = nil
	}
	return nil
}

func (v *UploadView) Render() []g.ComponentFunc {
	fileItems := make([]g.ConvertToMap, len(v.Files))
	for i, f := range v.Files {
		fileItems[i] = f
	}

	return []g.ComponentFunc{
		gd.Head(
			gd.Title("File Upload — Gerbera Demo"),
			gs.CSS(`
				body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f0f2f5; margin: 0; padding: 24px; color: #333; }
				h1 { color: #2c3e50; }
				.card { background: #fff; border-radius: 8px; padding: 20px 24px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
				.upload-area { border: 2px dashed #bdc3c7; border-radius: 8px; padding: 32px; text-align: center; margin-bottom: 16px; }
				.upload-area p { color: #7f8c8d; margin: 8px 0 16px; }
				input[type="file"] { font-size: 1rem; }
				table { width: 100%; border-collapse: collapse; margin-top: 12px; }
				th { background: #34495e; color: #fff; padding: 10px 12px; text-align: left; font-weight: 500; }
				th:first-child { border-radius: 4px 0 0 0; }
				th:last-child { border-radius: 0 4px 0 0; }
				td { padding: 10px 12px; border-bottom: 1px solid #ecf0f1; }
				tr:hover td { background: #f8f9fa; }
				.empty { color: #95a5a6; font-style: italic; padding: 20px; text-align: center; }
				button { padding: 8px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 0.9rem; font-weight: 500; background: #e74c3c; color: #fff; margin-top: 12px; }
				button:hover { opacity: 0.85; }
				.count { color: #7f8c8d; font-size: 0.9rem; margin-top: 8px; }
			`),
		),
		gd.Body(
			gd.H1(gp.Value("File Upload Demo")),

			gd.Div(
				gp.Class("card"),
				gd.Div(
					gp.Class("upload-area"),
					gd.P(gp.Value("Select files to upload")),
					gd.Input(
						gp.Type("file"),
						gl.Upload("on_upload"),
						gl.UploadMultiple(),
					),
				),
			),

			gd.Div(
				gp.Class("card"),
				gd.H3(gp.Value("Uploaded Files")),
				ge.If(len(v.Files) > 0,
					gd.Div(
						gd.Table(
							gd.Thead(
								gd.Tr(
									gd.Th(gp.Value("Name")),
									gd.Th(gp.Value("Size")),
									gd.Th(gp.Value("MIME Type")),
								),
							),
							gd.Tbody(
								ge.Each(fileItems, func(item g.ConvertToMap) g.ComponentFunc {
									m := item.ToMap()
									name := m.Get("name").(string)
									size := m.Get("size").(int64)
									mime := m.Get("mimetype").(string)
									return gd.Tr(
										gd.Td(gp.Value(name)),
										gd.Td(gp.Value(formatSize(size))),
										gd.Td(gp.Value(mime)),
									)
								}),
							),
						),
						gd.P(gp.Class("count"), gp.Value(fmt.Sprintf("%d file(s) uploaded", len(v.Files)))),
						gd.Button(gl.Click("clear"), gp.Value("Clear")),
					),
					gd.P(gp.Class("empty"), gp.Value("No files uploaded yet.")),
				),
			),
		),
	}
}

// formatSize returns a human-readable file size string.
func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func main() {
	addr := flag.String("addr", ":8885", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func() gl.View { return &UploadView{} }, opts...))
	log.Printf("upload running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
