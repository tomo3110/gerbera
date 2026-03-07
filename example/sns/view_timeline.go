package main

import (
	"database/sql"
	"fmt"
	"strings"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

type TimelineView struct {
	baseView

	posts        []PostWithMeta
	composeDraft string
	composeChars int
	confirmOpen  bool
	confirmAction string
}

func NewTimelineView(db *sql.DB, hub *Hub) *TimelineView {
	return &TimelineView{baseView: baseView{db: db, hub: hub}}
}

func (v *TimelineView) Mount(params gl.Params) error {
	if err := v.mountBase(params); err != nil {
		return err
	}
	return v.loadTimeline()
}

func (v *TimelineView) Unmount() {
	v.unmountBase()
}

func (v *TimelineView) loadTimeline() error {
	posts, err := dbGetTimeline(v.db, v.userID, 50, 0)
	if err != nil {
		return err
	}
	v.posts = posts
	return nil
}

func (v *TimelineView) HandleEvent(event string, payload gl.Payload) error {
	if handled, err := v.handlePostAction(event, payload); handled {
		if err != nil {
			return err
		}
		v.loadTimeline()
		return nil
	}

	switch event {
	case "composeInput":
		v.composeDraft = payload["value"]
		v.composeChars = len([]rune(v.composeDraft))
	case "submitPost":
		content := strings.TrimSpace(v.composeDraft)
		if content == "" || len([]rune(content)) > 250 {
			return nil
		}
		_, err := dbCreatePost(v.db, v.userID, content, "")
		if err != nil {
			return err
		}
		v.composeDraft = ""
		v.composeChars = 0
		v.loadTimeline()
		v.hub.Broadcast(NewPostNotif{AuthorName: v.user.DisplayName}, v.sess.ID)
	case "confirmDeletePost":
		v.confirmOpen = true
		v.confirmAction = payload["value"]
	case "doDelete":
		if v.confirmAction != "" {
			postID := parseID(v.confirmAction)
			if postID > 0 {
				dbDeletePost(v.db, postID, v.userID)
			}
		}
		v.confirmOpen = false
		v.confirmAction = ""
		v.loadTimeline()
		v.showToast("Post deleted", "success")
	case "cancelDelete":
		v.confirmOpen = false
		v.confirmAction = ""
	case "gerbera:upload_complete":
		// handled in HandleUpload
	}
	return nil
}

func (v *TimelineView) HandleUpload(event string, files []gl.UploadedFile) error {
	if event == "postPhoto" && len(files) > 0 {
		fmt.Printf("received upload: %s (%d bytes)\n", files[0].Name, files[0].Size)
	}
	return nil
}

func (v *TimelineView) HandleInfo(msg any) error {
	if v.handleBaseInfo(msg) {
		v.loadTimeline()
		return nil
	}
	switch msg.(type) {
	case NewPostNotif:
		v.loadTimeline()
	case NewMessageNotif:
		v.showToast("New message received", "info")
	}
	return nil
}

func (v *TimelineView) Render() g.Components {
	var postCards g.Components
	for _, p := range v.posts {
		postCards = append(postCards, postCard(p))
	}

	return g.Components{
		gd.Body(
			gu.Card(
				composeBox(v.composeDraft, v.composeChars),
			),
			gu.Card(
				expr.If(len(v.posts) == 0,
					gu.EmptyState("No posts yet. Follow someone or create a post!"),
				),
				expr.If(len(v.posts) > 0,
					gd.Div(postCards...),
				),
			),
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
			gul.Confirm(v.confirmOpen, "Delete Post",
				"Are you sure you want to delete this post?",
				"doDelete", "cancelDelete"),
		),
	}
}
