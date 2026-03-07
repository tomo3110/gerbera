package main

import (
	"database/sql"
	"fmt"
	"net/url"
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

type MessagesView struct {
	baseView

	conversations []ConversationSummary
	chatPartner   *User
	chatMessages  []Message
	chatDraft     string
	chatPartnerID int64
}

func NewMessagesView(db *sql.DB, hub *Hub) *MessagesView {
	return &MessagesView{baseView: baseView{db: db, hub: hub}}
}

func (v *MessagesView) Mount(params gl.Params) error {
	if err := v.mountBase(params); err != nil {
		return err
	}
	idStr := params.Query.Get("id")
	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil && id > 0 {
			return v.loadChat(id)
		}
	}
	return v.loadConversations()
}

func (v *MessagesView) Unmount() {
	v.unmountBase()
}

func (v *MessagesView) loadConversations() error {
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

func (v *MessagesView) loadChat(partnerID int64) error {
	if partnerID == v.userID {
		return v.loadConversations()
	}
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
	return nil
}

func (v *MessagesView) buildMessagesPath() string {
	if v.chatPartnerID > 0 {
		return "/messages/" + strconv.FormatInt(v.chatPartnerID, 10)
	}
	return "/messages"
}

func (v *MessagesView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "openChat":
		partnerID, _ := strconv.ParseInt(payload["value"], 10, 64)
		if partnerID == 0 {
			return nil
		}
		v.loadChat(partnerID)
		v.PushPatch(v.buildMessagesPath())
	case "backToConversations":
		v.loadConversations()
		v.PushPatch(v.buildMessagesPath())
	case "chatInput":
		v.chatDraft = payload["value"]
	case "chatSend":
		if strings.TrimSpace(v.chatDraft) == "" || v.chatPartnerID == 0 {
			return nil
		}
		dbSendMessage(v.db, v.userID, v.chatPartnerID, strings.TrimSpace(v.chatDraft))
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
	case "dismissToast":
		v.toastVisible = false
	}
	return nil
}

func (v *MessagesView) HandleParams(path string, params url.Values) error {
	v.toastVisible = false
	// Extract partner ID from path (/messages/{id}) or query (?id=...)
	var id int64
	if idStr := params.Get("id"); idStr != "" {
		id, _ = strconv.ParseInt(idStr, 10, 64)
	} else if strings.HasPrefix(path, "/messages/") {
		idStr := strings.TrimPrefix(path, "/messages/")
		id, _ = strconv.ParseInt(idStr, 10, 64)
	}
	if id > 0 {
		return v.loadChat(id)
	}
	return v.loadConversations()
}

func (v *MessagesView) HandleInfo(msg any) error {
	if v.handleBaseInfo(msg) {
		return nil
	}
	switch m := msg.(type) {
	case NewMessageNotif:
		if v.chatPartnerID == m.SenderID {
			v.loadChat(m.SenderID)
			v.ScrollIntoPct("#chat-messages", "1.0")
		} else {
			v.showToast("New message received", "info")
		}
	}
	return nil
}

func (v *MessagesView) Render() g.Components {
	if v.chatPartner != nil {
		return v.renderChat()
	}
	return v.renderConversationList()
}

func (v *MessagesView) renderConversationList() g.Components {
	var items g.Components
	for _, c := range v.conversations {
		items = append(items, renderConvoItem(c))
	}

	return g.Components{
		gd.Body(
			gp.Key("list"),
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
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
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

func (v *MessagesView) renderChat() g.Components {
	var msgViews g.Components
	msgViews = append(msgViews, gp.ID("chat-messages"))
	for _, m := range v.chatMessages {
		msgViews = append(msgViews, gu.ChatMessageView(gu.ChatMessage{
			Author:    "",
			Content:   m.Content,
			Timestamp: m.CreatedAt.Format("15:04"),
			Sent:      m.SenderID == v.userID,
		}))
	}

	return g.Components{
		gd.Body(
			gp.Key("chat"),
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
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
		),
	}
}
