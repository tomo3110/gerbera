package main

import (
	"database/sql"
	"fmt"

	gl "github.com/tomo3110/gerbera/live"
)

// baseView holds fields and helpers shared by all SNS views.
type baseView struct {
	gl.CommandQueue
	db   *sql.DB
	hub  *Hub
	sess *gl.Session

	userID int64
	user   *User

	toastMessage string
	toastVariant string
	toastVisible bool
}

func (v *baseView) mountBase(params gl.Params) error {
	v.sess = params.Conn.LiveSession

	if params.Conn.Session != nil {
		if uid, ok := params.Conn.Session.Get("user_id").(int64); ok {
			v.userID = uid
		}
	}

	if v.userID == 0 {
		return fmt.Errorf("not authenticated")
	}

	user, err := dbGetUserByID(v.db, v.userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	v.user = user

	v.hub.Join(v.sess, v.userID)
	return nil
}

func (v *baseView) unmountBase() {
	if v.sess != nil {
		v.hub.Leave(v.sess.ID, v.userID)
	}
}

func (v *baseView) showToast(msg, variant string) {
	v.toastMessage = msg
	v.toastVariant = variant
	v.toastVisible = true
}

// handlePostAction handles common post-interaction events (toggleLike, toggleRetweet, sharePost).
// Returns true if the event was handled.
func (v *baseView) handlePostAction(event string, payload gl.Payload) (bool, error) {
	switch event {
	case "toggleLike":
		return true, v.doToggleLike(payload)
	case "toggleRetweet":
		return true, v.doToggleRetweet(payload)
	case "sharePost":
		v.showToast("Share link: /post/"+payload["value"], "info")
		return true, nil
	case "dismissToast":
		v.toastVisible = false
		return true, nil
	}
	return false, nil
}

func (v *baseView) doToggleLike(payload gl.Payload) error {
	postID := parseID(payload["value"])
	if postID == 0 {
		return nil
	}
	liked, err := dbToggleLike(v.db, v.userID, postID)
	if err != nil {
		return err
	}
	if liked {
		post, _ := dbGetPost(v.db, postID, v.userID)
		if post != nil && post.Post.UserID != v.userID {
			dbCreateNotification(v.db, post.Post.UserID, v.userID, "like", &postID)
			v.hub.Notify(post.Post.UserID, NewLikeNotif{PostID: postID, ActorName: v.user.DisplayName})
		}
	}
	return nil
}

func (v *baseView) doToggleRetweet(payload gl.Payload) error {
	postID := parseID(payload["value"])
	if postID == 0 {
		return nil
	}
	retweeted, err := dbToggleRetweet(v.db, v.userID, postID)
	if err != nil {
		return err
	}
	if retweeted {
		post, _ := dbGetPost(v.db, postID, v.userID)
		if post != nil && post.Post.UserID != v.userID {
			dbCreateNotification(v.db, post.Post.UserID, v.userID, "retweet", &postID)
			v.hub.Notify(post.Post.UserID, NewRetweetNotif{PostID: postID, ActorName: v.user.DisplayName})
		}
	}
	return nil
}

// handleBaseInfo handles common notification types. Returns true if handled.
func (v *baseView) handleBaseInfo(msg any) bool {
	switch m := msg.(type) {
	case NewLikeNotif:
		v.showToast(fmt.Sprintf("%s liked your post", m.ActorName), "info")
		return true
	case NewRetweetNotif:
		v.showToast(fmt.Sprintf("%s retweeted your post", m.ActorName), "info")
		return true
	case NewFollowNotif:
		v.showToast(fmt.Sprintf("%s followed you", m.ActorName), "info")
		return true
	case NewCommentNotif:
		v.showToast(fmt.Sprintf("%s commented on your post", m.ActorName), "info")
		return true
	}
	return false
}
