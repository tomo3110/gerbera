package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// ChatInputOpts configures a live ChatInput.
type ChatInputOpts struct {
	Name         string
	Value        string
	Placeholder  string
	SendEvent    string // event fired when send button is clicked
	InputEvent   string // event fired on input change
	KeydownEvent string // event fired on keydown (for Enter key handling)
}

// ChatInput renders a live chat input area with a text input and send button.
func ChatInput(opts ChatInputOpts) gerbera.ComponentFunc {
	ph := opts.Placeholder
	if ph == "" {
		ph = "Type a message..."
	}

	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-chat-input-field"),
		property.Name(opts.Name),
		property.Attr("value", opts.Value),
		property.Placeholder(ph),
		property.Attr("autocomplete", "off"),
	}
	if opts.InputEvent != "" {
		inputAttrs = append(inputAttrs, gl.Input(opts.InputEvent))
	}
	if opts.KeydownEvent != "" {
		inputAttrs = append(inputAttrs, gl.Keydown(opts.KeydownEvent))
	}

	sendBtn := []gerbera.ComponentFunc{
		property.Class("g-btn", "g-btn-primary", "g-chat-send"),
		property.Attr("type", "button"),
		property.AriaLabel("Send"),
		property.Value("Send"),
	}
	if opts.SendEvent != "" {
		sendBtn = append(sendBtn, gl.Click(opts.SendEvent))
	}

	return dom.Div(
		property.Class("g-chat-inputbar"),
		dom.Input(inputAttrs...),
		dom.Button(sendBtn...),
	)
}
