package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

type PostDetailView struct {
	baseView

	detailPost     *PostWithMeta
	detailComments []CommentWithUser
	commentDraft   string
	confirmOpen    bool
	confirmAction  string
}

func NewPostDetailView(db *sql.DB, hub *Hub) *PostDetailView {
	return &PostDetailView{baseView: baseView{db: db, hub: hub}}
}

func (v *PostDetailView) Mount(params gl.Params) error {
	if err := v.mountBase(params); err != nil {
		return err
	}
	idStr := params.Query.Get("id")
	if idStr == "" {
		return nil
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil
	}
	return v.loadPostDetail(id)
}

func (v *PostDetailView) Unmount() {
	v.unmountBase()
}

func (v *PostDetailView) loadPostDetail(postID int64) error {
	post, err := dbGetPost(v.db, postID, v.userID)
	if err != nil {
		return err
	}
	v.detailPost = post
	comments, err := dbGetComments(v.db, postID)
	if err != nil {
		return err
	}
	v.detailComments = comments
	v.commentDraft = ""
	return nil
}

func (v *PostDetailView) HandleEvent(event string, payload gl.Payload) error {
	if handled, err := v.handlePostAction(event, payload); handled {
		if err != nil {
			return err
		}
		if v.detailPost != nil {
			v.loadPostDetail(v.detailPost.Post.ID)
		}
		return nil
	}

	switch event {
	case "commentInput":
		v.commentDraft = payload["value"]
	case "submitComment":
		if v.detailPost == nil || strings.TrimSpace(v.commentDraft) == "" {
			return nil
		}
		postID := v.detailPost.Post.ID
		_, err := dbCreateComment(v.db, v.userID, postID, strings.TrimSpace(v.commentDraft))
		if err != nil {
			return err
		}
		if v.detailPost.Post.UserID != v.userID {
			dbCreateNotification(v.db, v.detailPost.Post.UserID, v.userID, "comment", &postID)
			v.hub.Notify(v.detailPost.Post.UserID, NewCommentNotif{PostID: postID, ActorName: v.user.DisplayName})
		}
		v.hub.Broadcast(NewCommentOnViewedPostNotif{
			PostID:    postID,
			ActorName: v.user.DisplayName,
		}, v.sess.ID)
		v.loadPostDetail(postID)
	case "confirmDeletePost":
		v.confirmOpen = true
		v.confirmAction = payload["value"]
	case "doDelete":
		if v.confirmAction != "" {
			postID, _ := strconv.ParseInt(v.confirmAction, 10, 64)
			if postID > 0 {
				dbDeletePost(v.db, postID, v.userID)
			}
		}
		v.confirmOpen = false
		v.confirmAction = ""
		v.showToast("Post deleted", "success")
		v.Navigate("/")
	case "cancelDelete":
		v.confirmOpen = false
		v.confirmAction = ""
	}
	return nil
}

func (v *PostDetailView) HandleInfo(msg any) error {
	if v.handleBaseInfo(msg) {
		return nil
	}
	switch m := msg.(type) {
	case NewCommentOnViewedPostNotif:
		if v.detailPost != nil && v.detailPost.Post.ID == m.PostID {
			v.loadPostDetail(m.PostID)
		}
	case NewMessageNotif:
		v.showToast("New message received", "info")
	}
	return nil
}

func (v *PostDetailView) Render() g.Components {
	if v.detailPost == nil {
		return g.Components{
			gd.Body(gu.EmptyState("Post not found")),
		}
	}

	p := v.detailPost
	isOwner := p.Post.UserID == v.userID

	var commentItems g.Components
	for _, c := range v.detailComments {
		commentItems = append(commentItems, renderComment(c))
	}

	return g.Components{
		gd.Body(
			gd.A(
				gp.Class("chat-back-btn"),
				gp.Attr("href", "/"),
				gp.Value("< Back"),
			),
			gu.Card(
				postCard(*p),
				expr.If(isOwner,
					gd.Div(gp.Attr("style", "padding: 0 var(--g-space-md) var(--g-space-sm)"),
						gu.Button("Delete", gu.ButtonDanger, gu.ButtonSmall,
							gl.Click("confirmDeletePost"), gl.ClickValue(fmt.Sprintf("%d", p.Post.ID)),
						),
					),
				),
			),
			gu.Card(
				gd.Div(gp.Attr("style", "padding: var(--g-space-sm) var(--g-space-md); font-weight: 600"),
					gp.Value(fmt.Sprintf("Comments (%d)", len(v.detailComments))),
				),
				expr.If(len(v.detailComments) == 0,
					gu.EmptyState("No comments yet"),
				),
				expr.If(len(v.detailComments) > 0,
					gd.Div(commentItems...),
				),
				gd.Div(
					gp.Class("comment-form"),
					gu.FormTextarea("comment",
						gp.Attr("value", v.commentDraft),
						gp.Placeholder("Write a comment..."),
						gp.Attr("maxlength", "250"),
						gl.Input("commentInput"),
					),
					gu.Button("Reply", gu.ButtonPrimary, gu.ButtonSmall,
						gl.Click("submitComment"),
					),
				),
			),
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
			gul.Confirm(v.confirmOpen, "Delete Post",
				"Are you sure you want to delete this post?",
				"doDelete", "cancelDelete"),
		),
	}
}

func renderComment(c CommentWithUser) g.ComponentFunc {
	avatar := avatarComponent(c.Author)
	return gd.Div(
		gp.Class("comment-item"),
		gp.Key(fmt.Sprintf("comment-%d", c.Comment.ID)),
		avatar,
		gd.Div(
			gp.Class("comment-body"),
			gd.Span(gp.Class("comment-author"), gp.Value(c.Author.DisplayName)),
			gd.P(gp.Class("comment-text"), gp.Value(c.Comment.Content)),
			gd.Span(gp.Class("comment-time"), gp.Value(formatTimeAgo(c.Comment.CreatedAt))),
		),
	)
}
