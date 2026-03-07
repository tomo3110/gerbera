package main

import (
	"database/sql"
	"fmt"
	"strconv"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

type ProfileView struct {
	baseView

	profileUser      *User
	profilePosts     []PostWithMeta
	profileFollowers int
	profileFollowing int
	profilePostCount int
	isFollowing      bool
}

func NewProfileView(db *sql.DB, hub *Hub) *ProfileView {
	return &ProfileView{baseView: baseView{db: db, hub: hub}}
}

func (v *ProfileView) Mount(params gl.Params) error {
	if err := v.mountBase(params); err != nil {
		return err
	}
	idStr := params.Query.Get("id")
	if idStr == "" {
		return v.loadProfile(v.userID)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return v.loadProfile(v.userID)
	}
	return v.loadProfile(id)
}

func (v *ProfileView) Unmount() {
	v.unmountBase()
}

func (v *ProfileView) loadProfile(uid int64) error {
	user, err := dbGetUserByID(v.db, uid)
	if err != nil {
		return err
	}
	v.profileUser = user
	posts, err := dbGetUserPosts(v.db, uid, v.userID, 50, 0)
	if err != nil {
		return err
	}
	v.profilePosts = posts
	v.profileFollowers, _ = dbFollowerCount(v.db, uid)
	v.profileFollowing, _ = dbFollowingCount(v.db, uid)
	v.profilePostCount, _ = dbPostCount(v.db, uid)
	if uid != v.userID {
		v.isFollowing, _ = dbIsFollowing(v.db, v.userID, uid)
	}
	return nil
}

func (v *ProfileView) HandleEvent(event string, payload gl.Payload) error {
	if handled, err := v.handlePostAction(event, payload); handled {
		if err != nil {
			return err
		}
		if v.profileUser != nil {
			v.loadProfile(v.profileUser.ID)
		}
		return nil
	}

	switch event {
	case "toggleFollow":
		uid, _ := strconv.ParseInt(payload["value"], 10, 64)
		if uid == 0 || uid == v.userID {
			return nil
		}
		following, err := dbToggleFollow(v.db, v.userID, uid)
		if err != nil {
			return err
		}
		if following {
			dbCreateNotification(v.db, uid, v.userID, "follow", nil)
			v.hub.Notify(uid, NewFollowNotif{ActorName: v.user.DisplayName})
		}
		if v.profileUser != nil && v.profileUser.ID == uid {
			v.isFollowing = following
			v.profileFollowers, _ = dbFollowerCount(v.db, uid)
		}
	}
	return nil
}

func (v *ProfileView) HandleInfo(msg any) error {
	if v.handleBaseInfo(msg) {
		return nil
	}
	if _, ok := msg.(NewMessageNotif); ok {
		v.showToast("New message received", "info")
	}
	return nil
}

func (v *ProfileView) Render() g.Components {
	if v.profileUser == nil {
		return g.Components{
			gd.Body(gu.EmptyState("User not found")),
		}
	}

	u := v.profileUser
	isOwnProfile := u.ID == v.userID

	avatar := gu.LetterAvatar(u.DisplayName, gu.AvatarOpts{Size: "xl"})
	if u.AvatarPath != "" {
		avatar = gu.ImageAvatar(u.AvatarPath, gu.AvatarOpts{Size: "xl"})
	}

	var followBtn g.ComponentFunc
	if !isOwnProfile {
		if v.isFollowing {
			followBtn = gu.Button("Unfollow", gu.ButtonOutline, gu.ButtonSmall,
				gl.Click("toggleFollow"), gl.ClickValue(fmt.Sprintf("%d", u.ID)))
		} else {
			followBtn = gu.Button("Follow", gu.ButtonPrimary, gu.ButtonSmall,
				gl.Click("toggleFollow"), gl.ClickValue(fmt.Sprintf("%d", u.ID)))
		}
	}

	var postCards g.Components
	for _, p := range v.profilePosts {
		postCards = append(postCards, postCard(p))
	}

	return g.Components{
		gd.Body(
			gu.Card(
				gd.Div(
					gp.Class("profile-header"),
					gd.Div(
						gp.Class("profile-info"),
						avatar,
						gd.Div(
							gp.Class("profile-names"),
							gd.H2(gp.Value(u.DisplayName)),
							gd.Span(gp.Class("username"), gp.Value("@"+u.Username)),
						),
						expr.If(!isOwnProfile, followBtn),
					),
					expr.If(u.Bio != "",
						gd.P(gp.Class("profile-bio"), gp.Value(u.Bio)),
					),
					gd.Div(
						gp.Class("profile-stats"),
						gd.Div(gp.Class("profile-stat"),
							gd.Strong(gp.Value(fmt.Sprintf("%d", v.profilePostCount))),
							gd.Span(gp.Value("Posts")),
						),
						gd.Div(gp.Class("profile-stat"),
							gd.Strong(gp.Value(fmt.Sprintf("%d", v.profileFollowers))),
							gd.Span(gp.Value("Followers")),
						),
						gd.Div(gp.Class("profile-stat"),
							gd.Strong(gp.Value(fmt.Sprintf("%d", v.profileFollowing))),
							gd.Span(gp.Value("Following")),
						),
					),
					expr.If(!isOwnProfile,
						gd.A(
							gp.Attr("href", fmt.Sprintf("/messages/%d", u.ID)),
							gu.Button("Message", gu.ButtonOutline, gu.ButtonSmall,
								gp.Attr("style", "margin-top: var(--g-space-sm)"),
							),
						),
					),
				),
			),
			gu.Card(
				expr.If(len(v.profilePosts) == 0,
					gu.EmptyState("No posts yet"),
				),
				expr.If(len(v.profilePosts) > 0,
					gd.Div(postCards...),
				),
			),
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
		),
	}
}
