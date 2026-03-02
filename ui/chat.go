package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// ChatMessage represents a single chat message.
type ChatMessage struct {
	Author    string
	Content   string
	Timestamp string
	Sent      bool   // true = sent by current user, false = received
	Avatar    string // optional avatar text/emoji
}

// ChatContainer renders a scrollable chat message list.
func ChatContainer(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Class("g-chat"),
		property.Role("log"),
		property.AriaLabel("Chat messages"),
	}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// ChatMessageView renders a single chat message bubble.
func ChatMessageView(msg ChatMessage) gerbera.ComponentFunc {
	bubbleClass := "g-chat-bubble-received"
	if msg.Sent {
		bubbleClass = "g-chat-bubble-sent"
	}

	var avatarEl gerbera.ComponentFunc
	if msg.Avatar != "" {
		avatarEl = dom.Div(
			property.Class("g-chat-avatar"),
			property.Value(msg.Avatar),
		)
	}

	var children []gerbera.ComponentFunc
	if avatarEl != nil && !msg.Sent {
		children = append(children, avatarEl)
	}

	bubble := []gerbera.ComponentFunc{
		property.Class("g-chat-bubble", bubbleClass),
	}
	if msg.Author != "" {
		bubble = append(bubble, dom.Div(
			property.Class("g-chat-author"),
			property.Value(msg.Author),
		))
	}
	bubble = append(bubble, dom.Div(
		property.Class("g-chat-content"),
		property.Value(msg.Content),
	))
	if msg.Timestamp != "" {
		bubble = append(bubble,
			gerbera.Tag("time",
				property.Class("g-chat-time"),
				property.Value(msg.Timestamp),
			),
		)
	}

	children = append(children, dom.Div(bubble...))

	if avatarEl != nil && msg.Sent {
		children = append(children, avatarEl)
	}

	rowAttrs := []gerbera.ComponentFunc{
		property.Class("g-chat-row"),
		property.ClassIf(msg.Sent, "g-chat-row-sent"),
		property.ClassIf(!msg.Sent, "g-chat-row-received"),
	}
	rowAttrs = append(rowAttrs, children...)
	return dom.Div(rowAttrs...)
}

// ChatInput renders a chat input area with a text input and send button.
func ChatInput(name, value string, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-chat-input-field"),
		property.Name(name),
		property.Attr("value", value),
		property.Placeholder("Type a message..."),
		property.Attr("autocomplete", "off"),
	}
	inputAttrs = append(inputAttrs, opts...)

	return dom.Div(
		property.Class("g-chat-inputbar"),
		dom.Input(inputAttrs...),
		dom.Button(
			property.Class("g-btn", "g-btn-primary", "g-chat-send"),
			property.Attr("type", "submit"),
			property.AriaLabel("Send"),
			property.Value("Send"),
		),
	)
}
