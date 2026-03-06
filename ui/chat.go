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

// ChatInputOpts configures a ChatInput component.
type ChatInputOpts struct {
	Placeholder  string
	SendEvent    string // gerbera-click on send button
	InputEvent   string // gerbera-input on text field
	KeydownEvent string // gerbera-keydown on text field
}

// ChatContainer renders a scrollable chat message list.
func ChatContainer(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{
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

	var children gerbera.Components
	if avatarEl != nil && !msg.Sent {
		children = append(children, avatarEl)
	}

	bubble := gerbera.Components{
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

	rowAttrs := gerbera.Components{
		property.Class("g-chat-row"),
		property.ClassIf(msg.Sent, "g-chat-row-sent"),
		property.ClassIf(!msg.Sent, "g-chat-row-received"),
	}
	rowAttrs = append(rowAttrs, children...)
	return dom.Div(rowAttrs...)
}

// ChatInput renders a chat input area with a text input and send button.
func ChatInput(name, value string, opts ChatInputOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	ph := opts.Placeholder
	if ph == "" {
		ph = "Type a message..."
	}

	inputAttrs := gerbera.Components{
		property.Class("g-chat-input-field"),
		property.Name(name),
		property.Attr("value", value),
		property.Placeholder(ph),
		property.Attr("autocomplete", "off"),
	}
	inputAttrs = append(inputAttrs, extra...)
	if opts.InputEvent != "" {
		inputAttrs = append(inputAttrs, property.Attr("gerbera-input", opts.InputEvent))
	}
	if opts.KeydownEvent != "" {
		inputAttrs = append(inputAttrs, property.Attr("gerbera-keydown", opts.KeydownEvent))
	}

	btnType := "submit"
	if opts.SendEvent != "" {
		btnType = "button"
	}
	sendBtn := gerbera.Components{
		property.Class("g-btn", "g-btn-primary", "g-chat-send"),
		property.Attr("type", btnType),
		property.AriaLabel("Send"),
		property.Value("Send"),
	}
	if opts.SendEvent != "" {
		sendBtn = append(sendBtn, property.Attr("gerbera-click", opts.SendEvent))
	}

	return dom.Div(
		property.Class("g-chat-inputbar"),
		dom.Input(inputAttrs...),
		dom.Button(sendBtn...),
	)
}
