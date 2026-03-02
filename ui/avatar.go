package ui

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// AvatarOpts configures an Avatar.
type AvatarOpts struct {
	Size  string // "xs"(24px), "sm"(32px), "md"(40px default), "lg"(48px), "xl"(64px)
	Shape string // "circle"(default), "rounded"
	Alt   string // ImageAvatar alt text
}

// AvatarGroupOpts configures an AvatarGroup.
type AvatarGroupOpts struct {
	Size string
	Max  int // max displayed avatars, 0 = show all
}

var avatarColors = []string{
	"#ef4444", "#f97316", "#f59e0b", "#84cc16",
	"#22c55e", "#14b8a6", "#06b6d4", "#3b82f6",
	"#6366f1", "#8b5cf6", "#a855f7", "#ec4899",
}

func avatarSize(size string) string {
	switch size {
	case "xs", "sm", "md", "lg", "xl":
		return size
	default:
		return "md"
	}
}

func avatarShape(shape string) string {
	if shape == "rounded" {
		return "rounded"
	}
	return "circle"
}

func avatarColorForName(name string) string {
	h := 0
	for _, r := range name {
		h = h*31 + int(r)
	}
	if h < 0 {
		h = -h
	}
	return avatarColors[h%len(avatarColors)]
}

func avatarInitial(name string) string {
	for _, r := range name {
		return string(r)
	}
	return "?"
}

// ImageAvatar renders an avatar with an image.
func ImageAvatar(src string, opts AvatarOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	size := avatarSize(opts.Size)
	shape := avatarShape(opts.Shape)
	attrs := []gerbera.ComponentFunc{
		property.Class("g-avatar", "g-avatar-"+shape, "g-avatar-"+size),
	}
	attrs = append(attrs, extra...)
	var imgOpts []gerbera.ComponentFunc
	if opts.Alt != "" {
		imgOpts = append(imgOpts, property.Attr("alt", opts.Alt))
	}
	attrs = append(attrs, dom.Img(src, imgOpts...))
	return dom.Div(attrs...)
}

// LetterAvatar renders an avatar with a letter initial and deterministic background color.
func LetterAvatar(name string, opts AvatarOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	size := avatarSize(opts.Size)
	shape := avatarShape(opts.Shape)
	color := avatarColorForName(name)
	attrs := []gerbera.ComponentFunc{
		property.Class("g-avatar", "g-avatar-"+shape, "g-avatar-"+size),
		property.Attr("style", fmt.Sprintf("background-color:%s", color)),
		property.Value(avatarInitial(name)),
	}
	attrs = append(attrs, extra...)
	return dom.Div(attrs...)
}

// AvatarGroup renders a group of overlapping avatars with an optional "+N" overflow indicator.
func AvatarGroup(avatars []gerbera.ComponentFunc, opts AvatarGroupOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Class("g-avatar-group"),
	}
	attrs = append(attrs, extra...)

	display := avatars
	overflow := 0
	if opts.Max > 0 && len(avatars) > opts.Max {
		display = avatars[:opts.Max]
		overflow = len(avatars) - opts.Max
	}

	attrs = append(attrs, display...)

	if overflow > 0 {
		size := avatarSize(opts.Size)
		attrs = append(attrs, dom.Div(
			property.Class("g-avatar", "g-avatar-circle", "g-avatar-"+size, "g-avatar-group-more"),
			property.Value(fmt.Sprintf("+%d", overflow)),
		))
	}

	return dom.Div(attrs...)
}
