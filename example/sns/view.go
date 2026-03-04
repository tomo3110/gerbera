package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gl "github.com/tomo3110/gerbera/live"
)

// SNSView implements the single LiveView for the entire SNS app.
type SNSView struct {
	gl.CommandQueue
	db      *sql.DB
	hub     *Hub
	session *gl.Session

	// Current user
	userID int64
	user   *User

	// Navigation
	page string

	// Timeline
	posts        []PostWithMeta
	composeDraft string
	composeChars int

	// Profile
	profileUser      *User
	profilePosts     []PostWithMeta
	profileFollowers int
	profileFollowing int
	profilePostCount int
	isFollowing      bool

	// Post detail
	detailPost     *PostWithMeta
	detailComments []CommentWithUser
	commentDraft   string

	// Messages
	conversations []ConversationSummary
	chatPartner   *User
	chatMessages  []Message
	chatDraft     string
	chatPartnerID int64

	// Settings
	settingsEmail       string
	settingsDisplayName string
	settingsBio         string

	// Search
	searchQuery string
	searchUsers []User
	searchPosts []PostWithMeta

	// UI state
	toastMessage string
	toastVariant string
	toastVisible bool
	drawerOpen   bool
	confirmOpen  bool
	confirmAction string

	// Notification counts
	unreadNotifications int
	unreadMessages      int
}

func (v *SNSView) Mount(params gl.Params) error {
	v.session = params.Conn.LiveSession

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

	// Join hub for real-time notifications
	v.hub.Join(v.session, v.userID)

	// Load notification counts
	v.unreadNotifications, _ = dbUnreadNotificationCount(v.db, v.userID)
	v.unreadMessages, _ = dbUnreadMessageCount(v.db, v.userID)

	// Set initial page
	v.page = "home"
	if p := params.Get("page"); p != "" {
		v.page = p
	}

	// Load initial page data
	return v.loadPageData(params)
}

func (v *SNSView) loadPageData(params gl.Params) error {
	switch v.page {
	case "home":
		return v.loadTimeline()
	case "profile":
		idStr := params.Get("id")
		if idStr == "" {
			return v.loadProfile(v.userID)
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return v.loadProfile(v.userID)
		}
		return v.loadProfile(id)
	case "post":
		idStr := params.Get("id")
		if idStr == "" {
			return nil
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil
		}
		return v.loadPostDetail(id)
	case "messages":
		return v.loadConversations()
	case "settings":
		v.settingsDisplayName = v.user.DisplayName
		v.settingsEmail = v.user.Email
		v.settingsBio = v.user.Bio
	case "search":
		// empty initially
	}
	return nil
}

func (v *SNSView) loadTimeline() error {
	posts, err := dbGetTimeline(v.db, v.userID, 50, 0)
	if err != nil {
		return err
	}
	v.posts = posts
	return nil
}

func (v *SNSView) loadProfile(uid int64) error {
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

func (v *SNSView) loadPostDetail(postID int64) error {
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

func (v *SNSView) loadConversations() error {
	convos, err := dbGetConversations(v.db, v.userID)
	if err != nil {
		return err
	}
	v.conversations = convos
	v.chatPartner = nil
	v.chatMessages = nil
	v.chatDraft = ""
	v.chatPartnerID = 0
	return nil
}

func (v *SNSView) loadChat(partnerID int64) error {
	partner, err := dbGetUserByID(v.db, partnerID)
	if err != nil {
		return err
	}
	v.chatPartner = partner
	v.chatPartnerID = partnerID
	msgs, err := dbGetMessages(v.db, v.userID, partnerID, 100)
	if err != nil {
		return err
	}
	v.chatMessages = msgs
	v.chatDraft = ""
	_ = dbMarkMessagesRead(v.db, v.userID, partnerID)
	v.unreadMessages, _ = dbUnreadMessageCount(v.db, v.userID)
	return nil
}

func (v *SNSView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	// Navigation
	case "nav":
		v.page = payload["value"]
		v.drawerOpen = false
		v.toastVisible = false
		v.confirmOpen = false
		switch v.page {
		case "home":
			v.loadTimeline()
		case "profile":
			v.loadProfile(v.userID)
		case "messages":
			v.loadConversations()
		case "settings":
			v.settingsDisplayName = v.user.DisplayName
			v.settingsEmail = v.user.Email
			v.settingsBio = v.user.Bio
		case "search":
			v.searchQuery = ""
			v.searchUsers = nil
			v.searchPosts = nil
		}
		v.PushPatch(v.buildPath())

	case "toggleDrawer":
		v.drawerOpen = !v.drawerOpen
	case "closeDrawer":
		v.drawerOpen = false
	case "dismissToast":
		v.toastVisible = false

	// Compose
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
		v.hub.Broadcast(NewPostNotif{AuthorName: v.user.DisplayName}, v.session.ID)

	// Post actions
	case "toggleLike":
		postID, _ := strconv.ParseInt(payload["value"], 10, 64)
		if postID == 0 {
			return nil
		}
		liked, err := dbToggleLike(v.db, v.userID, postID)
		if err != nil {
			return err
		}
		// Notify post author
		if liked {
			post, _ := dbGetPost(v.db, postID, v.userID)
			if post != nil && post.Post.UserID != v.userID {
				dbCreateNotification(v.db, post.Post.UserID, v.userID, "like", &postID)
				v.hub.Notify(post.Post.UserID, NewLikeNotif{PostID: postID, ActorName: v.user.DisplayName})
			}
		}
		v.refreshCurrentPosts()

	case "toggleRetweet":
		postID, _ := strconv.ParseInt(payload["value"], 10, 64)
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
		v.refreshCurrentPosts()

	case "viewPost":
		postID, _ := strconv.ParseInt(payload["value"], 10, 64)
		if postID == 0 {
			return nil
		}
		v.page = "post"
		v.loadPostDetail(postID)
		v.PushPatch(v.buildPath())

	case "viewProfile":
		uid, _ := strconv.ParseInt(payload["value"], 10, 64)
		if uid == 0 {
			return nil
		}
		v.page = "profile"
		v.loadProfile(uid)
		v.PushPatch(v.buildPath())

	// Follow
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

	// Comments
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
		}, v.session.ID)
		v.loadPostDetail(postID)

	// Messages
	case "openChat":
		partnerID, _ := strconv.ParseInt(payload["value"], 10, 64)
		if partnerID == 0 {
			return nil
		}
		v.page = "messages"
		v.loadChat(partnerID)
		v.PushPatch(v.buildPath())
	case "backToConversations":
		v.page = "messages"
		v.loadConversations()
		v.PushPatch(v.buildPath())
	case "chatInput":
		v.chatDraft = payload["value"]
	case "chatSend":
		if strings.TrimSpace(v.chatDraft) == "" || v.chatPartnerID == 0 {
			return nil
		}
		msgID, err := dbSendMessage(v.db, v.userID, v.chatPartnerID, strings.TrimSpace(v.chatDraft))
		if err != nil {
			return err
		}
		_ = msgID
		v.hub.Notify(v.chatPartnerID, NewMessageNotif{SenderID: v.userID, Content: v.chatDraft})
		v.chatDraft = ""
		v.loadChat(v.chatPartnerID)
		v.ScrollIntoPct("#chat-messages", "1.0")
	case "chatKeydown":
		if payload["key"] != "Enter" {
			return nil
		}
		if strings.TrimSpace(v.chatDraft) == "" || v.chatPartnerID == 0 {
			return nil
		}
		dbSendMessage(v.db, v.userID, v.chatPartnerID, strings.TrimSpace(v.chatDraft))
		v.hub.Notify(v.chatPartnerID, NewMessageNotif{SenderID: v.userID, Content: v.chatDraft})
		v.chatDraft = ""
		v.loadChat(v.chatPartnerID)
		v.ScrollIntoPct("#chat-messages", "1.0")

	// Settings
	case "settingsDisplayNameInput":
		v.settingsDisplayName = payload["value"]
	case "settingsEmailInput":
		v.settingsEmail = payload["value"]
	case "settingsBioInput":
		v.settingsBio = payload["value"]
	case "saveProfile":
		err := dbUpdateUserProfile(v.db, v.userID, v.settingsDisplayName, v.settingsEmail, v.settingsBio)
		if err != nil {
			v.showToast("Failed to save profile", "danger")
			return nil
		}
		v.user.DisplayName = v.settingsDisplayName
		v.user.Email = v.settingsEmail
		v.user.Bio = v.settingsBio
		v.showToast("Profile saved", "success")

	case "savePassword":
		oldPass := payload["old_password"]
		newPass := payload["new_password"]
		if oldPass == "" || newPass == "" {
			v.showToast("Both fields are required", "warning")
			return nil
		}
		if !verifyPassword(v.user.PasswordHash, oldPass) {
			v.showToast("Current password is incorrect", "danger")
			return nil
		}
		hash := hashPassword(newPass)
		if err := dbUpdateUserPassword(v.db, v.userID, hash); err != nil {
			v.showToast("Failed to update password", "danger")
			return nil
		}
		v.user.PasswordHash = hash
		v.showToast("Password updated", "success")

	// Search
	case "searchInput":
		v.searchQuery = payload["value"]
		if strings.TrimSpace(v.searchQuery) == "" {
			v.searchUsers = nil
			v.searchPosts = nil
			return nil
		}
		v.searchUsers, _ = dbSearchUsers(v.db, v.searchQuery, 10)
		v.searchPosts, _ = dbSearchPosts(v.db, v.searchQuery, v.userID, 20)

	// Share
	case "sharePost":
		v.showToast("Share link: /?page=post&id="+payload["value"], "info")

	// Delete post confirmation
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
		v.refreshCurrentPosts()
		v.showToast("Post deleted", "success")
	case "cancelDelete":
		v.confirmOpen = false
		v.confirmAction = ""

	// Upload
	case "gerbera:upload_complete":
		// handled in HandleUpload
	}
	return nil
}

func (v *SNSView) HandleUpload(event string, files []gl.UploadedFile) error {
	if event == "postPhoto" && len(files) > 0 {
		// In a real app, save to disk/S3. For this demo, we skip storing the image.
		log.Printf("received upload: %s (%d bytes)", files[0].Name, files[0].Size)
	}
	if event == "avatarUpload" && len(files) > 0 {
		f := files[0]
		if !strings.HasPrefix(f.MIMEType, "image/") {
			return nil
		}
		ext := ".jpg"
		switch f.MIMEType {
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		}
		filename := fmt.Sprintf("%d%s", v.userID, ext)
		if err := os.WriteFile(filepath.Join("uploads", "avatars", filename), f.Data, 0644); err != nil {
			return err
		}
		avatarPath := "/avatars/" + filename
		if err := dbUpdateUserAvatar(v.db, v.userID, avatarPath); err != nil {
			return err
		}
		v.user.AvatarPath = avatarPath
		v.showToast("Avatar updated", "success")
	}
	return nil
}

// HandleInfo receives real-time notifications from the Hub.
func (v *SNSView) HandleInfo(msg any) error {
	switch m := msg.(type) {
	case NewLikeNotif:
		v.showToast(fmt.Sprintf("%s liked your post", m.ActorName), "info")
		v.unreadNotifications++
		v.refreshCurrentPosts()
	case NewRetweetNotif:
		v.showToast(fmt.Sprintf("%s retweeted your post", m.ActorName), "info")
		v.unreadNotifications++
		v.refreshCurrentPosts()
	case NewFollowNotif:
		v.showToast(fmt.Sprintf("%s followed you", m.ActorName), "info")
		v.unreadNotifications++
	case NewCommentNotif:
		v.showToast(fmt.Sprintf("%s commented on your post", m.ActorName), "info")
		v.unreadNotifications++
	case NewMessageNotif:
		v.unreadMessages++
		if v.page == "messages" && v.chatPartnerID == m.SenderID {
			v.loadChat(m.SenderID)
			v.ScrollIntoPct("#chat-messages", "1.0")
		} else {
			v.showToast("New message received", "info")
		}
	case NewPostNotif:
		if v.page == "home" {
			v.loadTimeline()
		}
	case NewCommentOnViewedPostNotif:
		if v.page == "post" && v.detailPost != nil && v.detailPost.Post.ID == m.PostID {
			v.loadPostDetail(m.PostID)
		}
	}
	return nil
}

// Unmount is called when the WebSocket connection closes.
func (v *SNSView) Unmount() {
	if v.session != nil {
		v.hub.Leave(v.session.ID, v.userID)
	}
}

func (v *SNSView) showToast(msg, variant string) {
	v.toastMessage = msg
	v.toastVariant = variant
	v.toastVisible = true
}

// HandleParams restores view state from URL query parameters (browser back/forward).
func (v *SNSView) HandleParams(params url.Values) error {
	page := params.Get("page")
	if page == "" {
		page = "home"
	}
	v.page = page
	v.drawerOpen = false
	v.toastVisible = false
	v.confirmOpen = false
	switch page {
	case "home":
		return v.loadTimeline()
	case "profile":
		if idStr := params.Get("id"); idStr != "" {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			return v.loadProfile(id)
		}
		return v.loadProfile(v.userID)
	case "post":
		if idStr := params.Get("id"); idStr != "" {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			return v.loadPostDetail(id)
		}
	case "messages":
		if chatStr := params.Get("chat"); chatStr != "" {
			id, _ := strconv.ParseInt(chatStr, 10, 64)
			return v.loadChat(id)
		}
		return v.loadConversations()
	case "settings":
		v.settingsDisplayName = v.user.DisplayName
		v.settingsEmail = v.user.Email
		v.settingsBio = v.user.Bio
	case "search":
		v.searchQuery = ""
		v.searchUsers = nil
		v.searchPosts = nil
	}
	return nil
}

// buildPath returns the URL path corresponding to the current view state.
func (v *SNSView) buildPath() string {
	q := url.Values{}
	if v.page != "" && v.page != "home" {
		q.Set("page", v.page)
	}
	if v.page == "profile" && v.profileUser != nil && v.profileUser.ID != v.userID {
		q.Set("id", strconv.FormatInt(v.profileUser.ID, 10))
	}
	if v.page == "post" && v.detailPost != nil {
		q.Set("id", strconv.FormatInt(v.detailPost.Post.ID, 10))
	}
	if v.page == "messages" && v.chatPartnerID > 0 {
		q.Set("chat", strconv.FormatInt(v.chatPartnerID, 10))
	}
	if len(q) == 0 {
		return "/"
	}
	return "/?" + q.Encode()
}

func (v *SNSView) refreshCurrentPosts() {
	switch v.page {
	case "home":
		v.loadTimeline()
	case "profile":
		if v.profileUser != nil {
			v.loadProfile(v.profileUser.ID)
		}
	case "post":
		if v.detailPost != nil {
			v.loadPostDetail(v.detailPost.Post.ID)
		}
	case "search":
		if v.searchQuery != "" {
			v.searchPosts, _ = dbSearchPosts(v.db, v.searchQuery, v.userID, 20)
		}
	}
}
