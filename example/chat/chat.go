package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
	gu "github.com/tomo3110/gerbera/ui"
)

// shared Hub instance
var hub = NewHub()

// ChatView implements the LiveView for a multi-user chat room.
type ChatView struct {
	gl.CommandQueue

	Username string
	Messages []ChatMessage
	Draft    string

	hub     *Hub
	session *gl.Session
}

func (v *ChatView) Mount(params gl.Params) error {
	v.hub = hub
	v.session = params.Conn.LiveSession
	return nil
}

func (v *ChatView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "join":
		name := payload["username"]
		if name == "" {
			return nil
		}
		v.Username = name
		v.hub.Join(v.session)
		v.hub.Broadcast(ChatMessage{
			Author:    name,
			Content:   fmt.Sprintf("%s joined the room", name),
			Timestamp: time.Now().Format("15:04"),
			System:    true,
		}, v.session.ID)

	case "input":
		v.Draft = payload["value"]

	case "send":
		if v.Draft == "" || v.Username == "" {
			return nil
		}
		msg := ChatMessage{
			Author:    v.Username,
			Content:   v.Draft,
			Timestamp: time.Now().Format("15:04"),
		}
		v.Messages = append(v.Messages, msg)
		v.Draft = ""
		v.hub.Broadcast(msg, v.session.ID)
		v.ScrollIntoPct("#chat-messages", "1.0")

	case "keydown":
		if payload["key"] != "Enter" {
			return nil
		}
		if v.Draft == "" || v.Username == "" {
			return nil
		}
		msg := ChatMessage{
			Author:    v.Username,
			Content:   v.Draft,
			Timestamp: time.Now().Format("15:04"),
		}
		v.Messages = append(v.Messages, msg)
		v.Draft = ""
		v.hub.Broadcast(msg, v.session.ID)
		v.ScrollIntoPct("#chat-messages", "1.0")
	}
	return nil
}

// HandleInfo receives ChatMessages broadcast from other users via the Hub.
func (v *ChatView) HandleInfo(msg any) error {
	if cm, ok := msg.(ChatMessage); ok {
		v.Messages = append(v.Messages, cm)
		v.ScrollIntoPct("#chat-messages", "1.0")
	}
	return nil
}

// Unmount is called when the WebSocket connection closes.
func (v *ChatView) Unmount() {
	if v.Username != "" && v.session != nil {
		v.hub.Leave(v.session.ID)
		v.hub.Broadcast(ChatMessage{
			Author:    v.Username,
			Content:   fmt.Sprintf("%s left the room", v.Username),
			Timestamp: time.Now().Format("15:04"),
			System:    true,
		}, v.session.ID)
	}
}

func (v *ChatView) Render() g.Components {
	return g.Components{
		gd.Head(
			gd.Title("Chat — Gerbera LiveView"),
			gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
			gu.Theme(),
			gs.CSS(pageCSS),
		),
		gd.Body(
			gu.Container(
				gu.Stack(
					// Header
					gd.Div(
						gp.Class("chat-header"),
						gd.H1(gp.Value("Chat Room")),
						gu.Badge(fmt.Sprintf("Online: %d", v.hub.OnlineCount())),
					),

					// Main content
					expr.If(v.Username == "", v.renderJoinForm()),
					expr.If(v.Username != "", v.renderChat()),
				),
			),
		),
	}
}

func (v *ChatView) renderJoinForm() g.ComponentFunc {
	return gu.Card(
		gu.CardHeader("Join Chat"),
		gu.CardBody(
			gd.Form(
				gl.Submit("join"),
				gu.Stack(
					gu.FormGroup(
						gu.FormLabel("Username", "username"),
						gu.FormInput("username",
							gp.ID("username"),
							gp.Placeholder("Enter your name"),
							gp.Attr("required", "required"),
						),
					),
					gu.Button("Join", gu.ButtonPrimary),
				),
			),
		),
	)
}

func (v *ChatView) renderChat() g.ComponentFunc {
	var msgViews g.Components
	msgViews = append(msgViews, gp.ID("chat-messages"))
	for _, m := range v.Messages {
		msg := m // capture
		if msg.System {
			msgViews = append(msgViews, gd.Div(
				gp.Class("chat-system-msg"),
				gp.Value(msg.Content),
			))
		} else {
			msgViews = append(msgViews, gu.ChatMessageView(gu.ChatMessage{
				Author:    msg.Author,
				Content:   msg.Content,
				Timestamp: msg.Timestamp,
				Sent:      msg.Author == v.Username,
				Avatar:    string([]rune(msg.Author)[0]),
			}))
		}
	}

	return gu.Card(
		gu.CardBody(
			gu.ChatContainer(msgViews...),
		),
		gu.CardFooter(
			gu.ChatInput("draft", v.Draft, gu.ChatInputOpts{
				Placeholder:  "Type a message...",
				SendEvent:    "send",
				InputEvent:   "input",
				KeydownEvent: "keydown",
			}),
		),
	)
}

const pageCSS = `
body {
	background: var(--g-bg);
	color: var(--g-text);
	font-family: var(--g-font);
	margin: 0;
	padding: var(--g-space-md) 0;
}
.chat-header {
	display: flex;
	align-items: center;
	gap: var(--g-space-sm);
}
.chat-header h1 {
	margin: 0;
	font-size: 1.4rem;
	flex: 1;
}
.g-chat {
	min-height: 300px;
	max-height: 60vh;
}
.chat-system-msg {
	text-align: center;
	color: var(--g-text-tertiary);
	font-size: 0.85rem;
	padding: var(--g-space-xs) 0;
}
`

func main() {
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ context.Context) gl.View {
		return &ChatView{}
	}, opts...))

	log.Println("Chat running on http://localhost:8920")
	log.Fatal(http.ListenAndServe(":8920", g.Serve(http.DefaultServeMux)))
}
