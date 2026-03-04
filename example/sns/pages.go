package main

import (
	"fmt"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

func (v *SNSView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("SNS"),
			gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
			gu.Theme(),
			gs.CSS(snsCSS),
		),
		gd.Body(
			// Mobile header
			gd.Div(
				gp.Class("sns-mobile-header"),
				gd.Button(
					gp.Attr("style", "background:none;border:none;cursor:pointer"),
					gl.Click("toggleDrawer"),
					gu.Icon("menu", ""),
				),
				gd.H1(gp.Value("SNS")),
			),
			// Main layout (always at body.Children[1] for stable diff paths)
			gd.Div(
				gp.Class("sns-layout"),
				// Desktop sidebar
				gd.Nav(
					gp.Class("sns-sidebar"),
					gd.Div(
						gp.Attr("style", "padding: 0 var(--g-space-md) var(--g-space-md); font-size: 1.3rem; font-weight: 700"),
						gp.Value("SNS"),
					),
					gd.Div(v.renderNavLinks()...),
				),
				// Content area
				gd.Main(
					gp.Class("sns-main"),
					v.renderContent(),
				),
			),
			// Overlays (after layout to avoid shifting layout index)
			gul.Drawer(v.drawerOpen, "closeDrawer", "left",
				gul.DrawerHeader("SNS", "closeDrawer"),
				gul.DrawerBody(v.renderDrawerNavLinks()...),
			),
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
			gul.Confirm(v.confirmOpen, "Delete Post",
				"Are you sure you want to delete this post?",
				"doDelete", "cancelDelete"),
		),
	}
}

func (v *SNSView) renderNavLinks() []g.ComponentFunc {
	return []g.ComponentFunc{
		navLink("home", "Home", "home", v.page == "home", 0),
		navLink("search", "Search", "search", v.page == "search", 0),
		navLink("bell", "Notifications", "home", false, v.unreadNotifications),
		navLink("mail", "Messages", "messages", v.page == "messages", v.unreadMessages),
		navLink("user", "Profile", "profile", v.page == "profile", 0),
		navLink("settings", "Settings", "settings", v.page == "settings", 0),
		gd.Div(gp.Attr("style", "padding: var(--g-space-sm) var(--g-space-md)"),
			gd.A(gp.Attr("href", "/logout"),
				gu.Button("Logout", gu.ButtonOutline, gu.ButtonSmall, gp.Attr("style", "width:100%")),
			),
		),
	}
}

// renderDrawerNavLinks returns nav links for the mobile drawer (separate closures from sidebar).
func (v *SNSView) renderDrawerNavLinks() []g.ComponentFunc {
	return []g.ComponentFunc{
		navLink("home", "Home", "home", v.page == "home", 0),
		navLink("search", "Search", "search", v.page == "search", 0),
		navLink("bell", "Notifications", "home", false, v.unreadNotifications),
		navLink("mail", "Messages", "messages", v.page == "messages", v.unreadMessages),
		navLink("user", "Profile", "profile", v.page == "profile", 0),
		navLink("settings", "Settings", "settings", v.page == "settings", 0),
		gd.Div(gp.Attr("style", "padding: var(--g-space-sm) var(--g-space-md)"),
			gd.A(gp.Attr("href", "/logout"),
				gu.Button("Logout", gu.ButtonOutline, gu.ButtonSmall, gp.Attr("style", "width:100%")),
			),
		),
	}
}

func (v *SNSView) renderContent() g.ComponentFunc {
	switch v.page {
	case "profile":
		if v.profileUser == nil {
			return g.Tag("page-profile", gu.EmptyState("User not found"))
		}
		return g.Tag("page-profile", v.renderProfile()...)
	case "post":
		if v.detailPost == nil {
			return g.Tag("page-post", gu.EmptyState("Post not found"))
		}
		return g.Tag("page-post", v.renderPostDetail()...)
	case "messages":
		if v.chatPartner != nil {
			return g.Tag("page-messages", append([]g.ComponentFunc{gp.Key("chat")}, v.renderChat()...)...)
		}
		return g.Tag("page-messages", append([]g.ComponentFunc{gp.Key("list")}, v.renderConversationList()...)...)
	case "settings":
		return g.Tag("page-settings", v.renderSettings()...)
	case "search":
		return g.Tag("page-search", v.renderSearch()...)
	default:
		return g.Tag("page-home", v.renderTimeline()...)
	}
}

// --- Home Timeline ---

func (v *SNSView) renderTimeline() []g.ComponentFunc {
	var postCards []g.ComponentFunc
	for _, p := range v.posts {
		postCards = append(postCards, postCard(p))
	}

	return []g.ComponentFunc{
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
	}
}

// --- Profile ---

func (v *SNSView) renderProfile() []g.ComponentFunc {
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

	var postCards []g.ComponentFunc
	for _, p := range v.profilePosts {
		postCards = append(postCards, postCard(p))
	}

	return []g.ComponentFunc{
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
					gu.Button("Message", gu.ButtonOutline, gu.ButtonSmall,
						gp.Attr("style", "margin-top: var(--g-space-sm)"),
						gl.Click("openChat"), gl.ClickValue(fmt.Sprintf("%d", u.ID)),
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
	}
}

// --- Post Detail ---

func (v *SNSView) renderPostDetail() []g.ComponentFunc {
	p := v.detailPost
	isOwner := p.Post.UserID == v.userID

	var commentItems []g.ComponentFunc
	for _, c := range v.detailComments {
		commentItems = append(commentItems, renderComment(c))
	}

	return []g.ComponentFunc{
		gd.Button(
			gp.Class("chat-back-btn"),
			gl.Click("nav"), gl.ClickValue("home"),
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
	}
}

func renderComment(c CommentWithUser) g.ComponentFunc {
	avatar := avatarComponent(c.Author)
	return gd.Div(
		gp.Class("comment-item"),
		avatar,
		gd.Div(
			gp.Class("comment-body"),
			gd.Span(gp.Class("comment-author"), gp.Value(c.Author.DisplayName)),
			gd.P(gp.Class("comment-text"), gp.Value(c.Comment.Content)),
			gd.Span(gp.Class("comment-time"), gp.Value(formatTimeAgo(c.Comment.CreatedAt))),
		),
	)
}

// --- Messages ---

func (v *SNSView) renderMessages() []g.ComponentFunc {
	if v.chatPartner != nil {
		return v.renderChat()
	}
	return v.renderConversationList()
}

func (v *SNSView) renderConversationList() []g.ComponentFunc {
	var items []g.ComponentFunc
	for _, c := range v.conversations {
		items = append(items, renderConvoItem(c))
	}

	return []g.ComponentFunc{
		gd.Div(gp.Attr("style", "padding: var(--g-space-md); font-size: 1.1rem; font-weight: 700"),
			gp.Value("Messages"),
		),
		gu.Card(
			expr.If(len(v.conversations) == 0,
				gu.EmptyState("No conversations yet"),
			),
			expr.If(len(v.conversations) > 0,
				gd.Div(items...),
			),
		),
	}
}

func renderConvoItem(c ConversationSummary) g.ComponentFunc {
	avatar := avatarComponent(c.Partner)
	return gd.Div(
		gp.Class("convo-item"),
		gl.Click("openChat"), gl.ClickValue(fmt.Sprintf("%d", c.Partner.ID)),
		avatar,
		gd.Div(
			gp.Class("convo-info"),
			gd.Div(gp.Class("convo-name"), gp.Value(c.Partner.DisplayName)),
			gd.Div(gp.Class("convo-preview"), gp.Value(truncate(c.LastMessage, 50))),
		),
		gd.Div(
			gp.Class("convo-meta"),
			gd.Div(gp.Class("convo-time"), gp.Value(formatTimeAgo(c.LastMessageAt))),
			expr.If(c.UnreadCount > 0,
				gd.Span(gp.Class("sns-nav-badge"), gp.Value(fmt.Sprintf("%d", c.UnreadCount))),
			),
		),
	)
}

func (v *SNSView) renderChat() []g.ComponentFunc {
	var msgViews []g.ComponentFunc
	msgViews = append(msgViews, gp.ID("chat-messages"))
	for _, m := range v.chatMessages {
		msgViews = append(msgViews, gu.ChatMessageView(gu.ChatMessage{
			Author:    "",
			Content:   m.Content,
			Timestamp: m.CreatedAt.Format("15:04"),
			Sent:      m.SenderID == v.userID,
		}))
	}

	return []g.ComponentFunc{
		gd.Div(
			gp.Attr("style", "display:flex; align-items:center; gap:var(--g-space-sm); padding:var(--g-space-sm) 0"),
			gd.Button(
				gp.Class("chat-back-btn"),
				gl.Click("backToConversations"),
				gp.Value("< Back"),
			),
			avatarComponent(*v.chatPartner),
			gd.Strong(gp.Value(v.chatPartner.DisplayName)),
		),
		gu.Card(
			gu.CardBody(
				gd.Div(gp.Attr("style", "height:400px;display:flex;flex-direction:column"),
					gd.Div(gp.Attr("style", "flex:1;overflow-y:auto"),
						gu.ChatContainer(msgViews...),
					),
					gu.ChatInput("chatMsg", v.chatDraft, gu.ChatInputOpts{
						Placeholder:  "Type a message...",
						SendEvent:    "chatSend",
						InputEvent:   "chatInput",
						KeydownEvent: "chatKeydown",
					}),
				),
			),
		),
	}
}

// --- Settings ---

func (v *SNSView) renderSettings() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Div(gp.Attr("style", "padding: var(--g-space-md) 0; font-size: 1.1rem; font-weight: 700"),
			gp.Value("Settings"),
		),
		gu.Card(
			gd.Div(gp.Class("settings-section"),
				gd.H3(gp.Value("Profile")),
				gu.Stack(
					gu.FormGroup(
						gu.FormLabel("Display Name", "display_name"),
						gu.FormInput("display_name",
							gp.ID("display_name"),
							gp.Attr("value", v.settingsDisplayName),
							gl.Input("settingsDisplayNameInput"),
						),
					),
					gu.FormGroup(
						gu.FormLabel("Email", "email"),
						gu.FormInput("email",
							gp.ID("email"),
							gp.Attr("value", v.settingsEmail),
							gp.Attr("type", "email"),
							gl.Input("settingsEmailInput"),
						),
					),
					gu.FormGroup(
						gu.FormLabel("Bio", "bio"),
						gu.FormTextarea("bio",
							gp.Attr("value", v.settingsBio),
							gp.Placeholder("Tell us about yourself"),
							gp.Attr("maxlength", "160"),
							gl.Input("settingsBioInput"),
						),
					),
					gu.Button("Save Profile", gu.ButtonPrimary, gl.Click("saveProfile")),
				),
			),
			gd.Div(gp.Class("settings-section"),
				gd.H3(gp.Value("Change Password")),
				gd.Form(
					gl.Submit("savePassword"),
					gu.Stack(
						gu.FormGroup(
							gu.FormLabel("Current Password", "old_password"),
							gu.FormInput("old_password",
								gp.ID("old_password"),
								gp.Attr("type", "password"),
							),
						),
						gu.FormGroup(
							gu.FormLabel("New Password", "new_password"),
							gu.FormInput("new_password",
								gp.ID("new_password"),
								gp.Attr("type", "password"),
							),
						),
						gu.Button("Update Password", gu.ButtonPrimary),
					),
				),
			),
			gd.Div(gp.Class("settings-section"),
				gd.H3(gp.Value("Avatar")),
				gd.Div(
					gp.Attr("style", "display:flex; align-items:center; gap:var(--g-space-md)"),
					func() g.ComponentFunc {
						if v.user.AvatarPath != "" {
							return gu.ImageAvatar(v.user.AvatarPath, gu.AvatarOpts{Size: "xl"})
						}
						return gu.LetterAvatar(v.user.DisplayName, gu.AvatarOpts{Size: "xl"})
					}(),
					gd.Label(
						gp.Attr("style", "cursor:pointer"),
						gd.Input(gp.Type("file"), gl.Upload("avatarUpload"), gl.UploadAccept("image/*"),
							gp.Attr("style", "display:none")),
						gd.Span(gp.Class("g-btn", "g-btn-outline", "g-btn-sm"), gp.Value("Upload Avatar")),
					),
				),
			),
		),
	}
}

// --- Search ---

func (v *SNSView) renderSearch() []g.ComponentFunc {
	var results []g.ComponentFunc

	if v.searchQuery != "" {
		// User results
		if len(v.searchUsers) > 0 {
			results = append(results,
				gd.Div(gp.Class("search-section-label"), gp.Value("People")),
			)
			for _, u := range v.searchUsers {
				results = append(results, renderSearchUserItem(u))
			}
		}

		// Post results
		if len(v.searchPosts) > 0 {
			results = append(results,
				gd.Div(gp.Class("search-section-label"), gp.Value("Posts")),
			)
			for _, p := range v.searchPosts {
				results = append(results, postCard(p))
			}
		}

		if len(v.searchUsers) == 0 && len(v.searchPosts) == 0 {
			results = append(results, gu.EmptyState("No results found"))
		}
	}

	return []g.ComponentFunc{
		gd.Div(gp.Attr("style", "padding: var(--g-space-md) 0; font-size: 1.1rem; font-weight: 700"),
			gp.Value("Search"),
		),
		gu.Card(
			gd.Div(gp.Class("search-input-wrap"),
				gu.FormInput("search",
					gp.Attr("value", v.searchQuery),
					gp.Placeholder("Search people and posts..."),
					gl.Input("searchInput"),
					gl.Debounce(300),
				),
			),
			gd.Div(results...),
		),
	}
}

func renderSearchUserItem(u User) g.ComponentFunc {
	avatar := avatarComponent(u)
	return gd.Div(
		gp.Class("search-user-item"),
		gl.Click("viewProfile"), gl.ClickValue(fmt.Sprintf("%d", u.ID)),
		avatar,
		gd.Div(
			gp.Class("search-user-info"),
			gd.Strong(gp.Value(u.DisplayName)),
			gd.Span(gp.Value("@"+u.Username)),
		),
	)
}

// --- Helpers ---

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// Suppress unused import warnings by referencing time.
var _ = time.Now
