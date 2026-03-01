# File Upload Tutorial

## Overview

Build a file upload demo that lets users select files, upload them to the server, and display the results. This tutorial covers the following features:

- `gl.UploadHandler` interface — Handling file uploads on the server
- `gl.Upload()` — Binding an upload event to a file input
- `gl.UploadMultiple()` — Allowing multiple file selection
- `gl.UploadedFile` — Accessing uploaded file metadata and data
- `ge.Each` / `ge.If` — Conditional and iterative rendering

This example demonstrates the **Gerbera upload flow**: the client JS automatically POSTs files when selected, the server processes them via `HandleUpload`, then the client sends a `gerbera:upload_complete` event via WebSocket to trigger a UI re-render.

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Importing Packages

```go
import (
	"flag"
	"fmt"
	"log"
	"net/http"

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)
```

## Step 2: Defining the FileInfo Type

```go
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
```

`FileInfo` stores metadata about each uploaded file. It implements `g.ConvertToMap` so it can be used with `ge.Each` for rendering the file table.

## Step 3: Defining the View Struct

```go
type UploadView struct {
	Files []FileInfo
}
```

The view simply holds a list of uploaded file metadata. Unlike JSCommandView, there is no need to embed `CommandQueue` — upload handling is interface-based.

## Step 4: Implementing HandleUpload

```go
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
```

This is the core of the upload feature. By implementing the `UploadHandler` interface, your View receives uploaded files with:
- `Name` — Original filename
- `Size` — File size in bytes
- `MIMEType` — Content type (e.g., `"image/png"`, `"application/pdf"`)
- `Data` — Raw file bytes (not used in this demo, but available for processing)

The first parameter `event` matches the event name from `gl.Upload("on_upload")`. You can use this to distinguish between different upload inputs on the same page.

## Step 5: Implementing HandleEvent

```go
func (v *UploadView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "gerbera:upload_complete":
		// UI will re-render automatically; files were already added in HandleUpload.
	case "clear":
		v.Files = nil
	}
	return nil
}
```

Two events are handled:
- `gerbera:upload_complete` — Sent automatically by the client JS after a successful upload. The files are already in `v.Files` from `HandleUpload`, so we just need the re-render that happens after `HandleEvent` returns.
- `clear` — Resets the file list.

## Step 6: Binding the Upload Input

```go
gd.Input(
	gp.Type("file"),
	gl.Upload("on_upload"),
	gl.UploadMultiple(),
)
```

Key attributes:
- `gp.Type("file")` — Standard HTML file input
- `gl.Upload("on_upload")` — Sets `gerbera-upload="on_upload"`, telling gerbera.js to auto-POST when files are selected
- `gl.UploadMultiple()` — Sets `multiple="multiple"` to allow selecting multiple files

Optional: use `gl.UploadAccept("image/*,.pdf")` to restrict file types, and `gl.UploadMaxSize(5 << 20)` to set a client-side size limit.

## Step 7: Rendering the File Table

```go
fileItems := make([]g.ConvertToMap, len(v.Files))
for i, f := range v.Files {
	fileItems[i] = f
}

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
		gd.Button(gl.Click("clear"), gp.Value("Clear")),
	),
	gd.P(gp.Class("empty"), gp.Value("No files uploaded yet.")),
)
```

The file list is rendered as an HTML table using `ge.Each`. A helper function `formatSize` converts byte counts to human-readable strings (B, KB, MB).

## Upload Flow

```
Browser                                   Go Server
  |                                          |
  |  1. User selects files in <input>        |
  |                                          |
  |  2. gerbera.js detects change event      |
  |     on [gerbera-upload] input            |
  |                                          |
  |  3. POST /?gerbera-upload=1              |
  |     &session=xxx&event=on_upload         |
  |     Content-Type: multipart/form-data    |
  |     [file data]                          |
  |                                          |
  |      Handler routes to handleUpload()    |
  |      -> ParseMultipartForm              |
  |      -> UploadHandler.HandleUpload()     |
  |      <- 200 OK {"ok":true}              |
  |                                          |
  |  4. gerbera.js receives 200 OK           |
  |  --> WebSocket: {"e":                    |
  |       "gerbera:upload_complete",         |
  |       "p":{"event":"on_upload"}}         |
  |                                          |
  |      HandleEvent -> Render -> Diff       |
  |                                          |
  |  <-- DOM patches (file table updated)    |
  |                                          |
  |  5. Browser shows updated file list      |
```

The upload is a two-phase process:
1. **HTTP POST** — File data is sent via multipart form, processed by `HandleUpload`
2. **WebSocket event** — Client notifies the server that upload is complete, triggering a re-render

## Running

```bash
go run example/upload/upload.go
```

Open http://localhost:8885 in your browser. Select one or more files — they will be uploaded immediately and their metadata displayed in the table. Click **Clear** to reset.

### Debug Mode

```bash
go run example/upload/upload.go -debug
```

In the debug panel, observe the **Events** tab to see the `gerbera:upload_complete` event arrive after each upload.

## Exercises

1. Add `gl.UploadAccept("image/*")` to restrict uploads to images only
2. Add `gl.UploadMaxSize(1 << 20)` to limit file size to 1MB
3. Store the uploaded `Data` bytes and display image previews using a data URL
4. Add a per-file delete button that removes individual entries from the list

## API Reference

| Function | Description |
|----------|-------------|
| `gl.UploadHandler` | Interface: extends `View` with `HandleUpload(event, files)` |
| `gl.UploadedFile` | Struct: `Name`, `Size`, `MIMEType`, `Data` |
| `gl.Upload(event)` | Binds upload event to a file input |
| `gl.UploadMultiple()` | Allows multiple file selection |
| `gl.UploadAccept(types)` | Restricts accepted file types |
| `gl.UploadMaxSize(bytes)` | Sets client-side max file size |
| `ge.Each(items, cb)` | Iterates over a list to render elements |
| `ge.If(cond, t, f)` | Conditional rendering |
