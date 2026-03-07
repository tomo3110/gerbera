package main

import (
	"fmt"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
)

// iconHeart returns an inline SVG heart icon wrapped in a span.
// Using gd.Span wrapper ensures the button's Element tree children match the browser DOM
// (g.Literal sets parent.Value, which creates phantom DOM elements not tracked by the diff engine).
func iconHeart(filled bool) g.ComponentFunc {
	if filled {
		return gd.Span(g.Literal(`<svg class="icon-heart" viewBox="0 0 24 24" fill="#ef4444" stroke="#ef4444" stroke-width="2"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>`))
	}
	return gd.Span(g.Literal(`<svg class="icon-heart" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>`))
}

// iconRetweet returns an inline SVG retweet icon wrapped in a span.
func iconRetweet() g.ComponentFunc {
	return gd.Span(g.Literal(`<svg class="icon-retweet" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 1l4 4-4 4"/><path d="M3 11V9a4 4 0 0 1 4-4h14"/><path d="M7 23l-4-4 4-4"/><path d="M21 13v2a4 4 0 0 1-4 4H3"/></svg>`))
}

// iconComment returns an inline SVG comment/chat icon wrapped in a span.
func iconComment() g.ComponentFunc {
	return gd.Span(g.Literal(`<svg class="icon-comment" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`))
}

// iconShare returns an inline SVG share icon wrapped in a span.
func iconShare() g.ComponentFunc {
	return gd.Span(g.Literal(`<svg class="icon-share" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>`))
}

// formatTimeAgo returns a relative time string.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	default:
		return t.Format("Jan 2")
	}
}

// postCard renders a single post card with actions.
func postCard(p PostWithMeta) g.ComponentFunc {
	postIDStr := fmt.Sprintf("%d", p.Post.ID)
	authorIDStr := fmt.Sprintf("%d", p.Author.ID)

	likeClass := "post-action-btn"
	if p.Liked {
		likeClass = "post-action-btn liked"
	}
	retweetClass := "post-action-btn"
	if p.Retweeted {
		retweetClass = "post-action-btn retweeted"
	}

	avatar := avatarComponent(p.Author)

	return gd.Div(
		gp.Class("post-card"),
		gd.Div(
			gp.Class("post-header"),
			avatar,
			gd.Span(gp.Class("post-author"),
				gl.Click("viewProfile"), gl.ClickValue(authorIDStr),
				gp.Value(p.Author.DisplayName),
			),
			gd.Span(gp.Class("post-username"), gp.Value("@"+p.Author.Username)),
			gd.Span(gp.Class("post-time"), gp.Value(formatTimeAgo(p.Post.CreatedAt))),
		),
		gd.Div(
			gp.Class("post-content"),
			gl.Click("viewPost"), gl.ClickValue(postIDStr),
			gp.Value(p.Post.Content),
		),
		expr.If(p.Post.ImagePath != "",
			gd.Img(p.Post.ImagePath, gp.Class("post-image"), gp.Attr("alt", "Post image")),
		),
		gd.Div(
			gp.Class("post-actions"),
			gd.Button(
				gp.Class(likeClass),
				gl.Click("toggleLike"), gl.ClickValue(postIDStr),
				iconHeart(p.Liked),
				gd.Span(gp.Value(fmt.Sprintf("%d", p.LikeCount))),
			),
			gd.Button(
				gp.Class(retweetClass),
				gl.Click("toggleRetweet"), gl.ClickValue(postIDStr),
				iconRetweet(),
				gd.Span(gp.Value(fmt.Sprintf("%d", p.RetweetCount))),
			),
			gd.Button(
				gp.Class("post-action-btn"),
				gl.Click("viewPost"), gl.ClickValue(postIDStr),
				iconComment(),
				gd.Span(gp.Value(fmt.Sprintf("%d", p.CommentCount))),
			),
			gd.Button(
				gp.Class("post-action-btn"),
				gl.Click("sharePost"), gl.ClickValue(postIDStr),
				iconShare(),
			),
		),
	)
}

// avatarComponent renders the appropriate avatar for a user.
func avatarComponent(u User) g.ComponentFunc {
	if u.AvatarPath != "" {
		return gu.ImageAvatar(u.AvatarPath, gu.AvatarOpts{Size: "sm"})
	}
	name := u.DisplayName
	if name == "" {
		name = u.Username
	}
	return gu.LetterAvatar(name, gu.AvatarOpts{Size: "sm"})
}

// navLink renders a sidebar navigation link with optional badge.
func navLink(icon, label, page string, active bool, badgeCount int) g.ComponentFunc {
	cls := "sns-nav-link"
	if active {
		cls = "sns-nav-link active"
	}
	var children g.Components
	children = append(children,
		gp.Class(cls),
		gl.Click("nav"), gl.ClickValue(page),
		gu.Icon(icon, ""),
		gd.Span(gp.Value(label)),
	)
	if badgeCount > 0 {
		children = append(children,
			gd.Span(gp.Class("sns-nav-badge"), gp.Value(fmt.Sprintf("%d", badgeCount))),
		)
	}
	return gd.Button(children...)
}

// composeBox renders the post compose form.
func composeBox(draft string, charCount int) g.ComponentFunc {
	charClass := "char-count"
	if charCount > 250 {
		charClass = "char-count over"
	}

	return gd.Div(
		gp.Class("compose-box"),
		gu.FormTextarea("compose", gp.Attr("value", draft),
			gp.Placeholder("What's happening?"),
			gp.Attr("maxlength", "250"),
			gl.Input("composeInput"),
		),
		gd.Div(
			gp.Class("compose-footer"),
			gd.Span(gp.Class(charClass), gp.Value(fmt.Sprintf("%d/250", charCount))),
			gd.Div(
				gp.Class("compose-actions"),
				gd.Label(
					gp.Attr("style", "cursor:pointer"),
					gd.Input(gp.Type("file"), gl.Upload("postPhoto"), gl.UploadAccept("image/*"),
						gp.Attr("style", "display:none")),
					gu.Icon("upload", "sm"),
				),
				gu.Button("Post", gu.ButtonPrimary, gu.ButtonSmall,
					gl.Click("submitPost"),
				),
			),
		),
	)
}
