package live

import (
	"fmt"
	"io"
	"net/http"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/property"
)

// UploadConfig configures file upload behavior.
type UploadConfig struct {
	// MaxFileSize is the maximum size per file in bytes (default 10MB).
	MaxFileSize int64
	// MaxFiles is the maximum number of files allowed (default 1).
	MaxFiles int
	// Accept is a comma-separated list of MIME types or extensions (e.g., "image/*,.pdf").
	Accept string
}

// UploadedFile represents a file uploaded by the client.
type UploadedFile struct {
	Name     string
	Size     int64
	MIMEType string
	Data     []byte
}

// UploadHandler is an optional interface for Views that handle file uploads.
type UploadHandler interface {
	View
	// HandleUpload is called when files are uploaded.
	HandleUpload(event string, files []UploadedFile) error
}

// Upload binds a file upload event to the element (typically an <input type="file">).
func Upload(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-upload", event)
}

// UploadAccept sets the accepted file types for an upload input.
func UploadAccept(accept string) gerbera.ComponentFunc {
	return property.Attr("accept", accept)
}

// UploadMultiple allows multiple file selection.
func UploadMultiple() gerbera.ComponentFunc {
	return property.Attr("multiple", "multiple")
}

// UploadMaxSize sets the maximum upload size attribute (used for client-side validation).
func UploadMaxSize(bytes int64) gerbera.ComponentFunc {
	return property.Attr("gerbera-upload-max", fmt.Sprintf("%d", bytes))
}

// handleUpload processes multipart file upload HTTP requests.
func handleUpload(w http.ResponseWriter, r *http.Request, store *sessionStore, dlog *debugLogger) {
	sessionID := r.URL.Query().Get("session")
	event := r.URL.Query().Get("event")
	sess := store.get(sessionID)
	if sess == nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	uh, ok := sess.View.(UploadHandler)
	if !ok {
		http.Error(w, "view does not support uploads", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	var files []UploadedFile
	for _, fHeaders := range r.MultipartForm.File {
		for _, fh := range fHeaders {
			f, err := fh.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				continue
			}
			files = append(files, UploadedFile{
				Name:     fh.Filename,
				Size:     fh.Size,
				MIMEType: fh.Header.Get("Content-Type"),
				Data:     data,
			})
		}
	}

	sess.mu.Lock()
	err := uh.HandleUpload(event, files)
	sess.mu.Unlock()

	if err != nil {
		dlog.handleError(sessionID, "HandleUpload", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
