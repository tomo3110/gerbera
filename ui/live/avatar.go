package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/ui"
)

// LiveAvatarOpts extends AvatarOpts with a click event.
type LiveAvatarOpts struct {
	ui.AvatarOpts
	ClickEvent string
}

// LiveAvatarGroupOpts extends AvatarGroupOpts with a click event for "+N more".
type LiveAvatarGroupOpts struct {
	ui.AvatarGroupOpts
	ClickEvent string // "+N more" click event
}

// ImageAvatar renders a live image avatar with optional click event.
func ImageAvatar(src string, opts LiveAvatarOpts) gerbera.ComponentFunc {
	var extra []gerbera.ComponentFunc
	if opts.ClickEvent != "" {
		extra = append(extra,
			gl.Click(opts.ClickEvent),
			gl.ClickValue(src),
			property.Attr("style", "cursor:pointer"),
		)
	}
	return ui.ImageAvatar(src, opts.AvatarOpts, extra...)
}

// LetterAvatar renders a live letter avatar with optional click event.
func LetterAvatar(name string, opts LiveAvatarOpts) gerbera.ComponentFunc {
	var extra []gerbera.ComponentFunc
	if opts.ClickEvent != "" {
		extra = append(extra,
			gl.Click(opts.ClickEvent),
			gl.ClickValue(name),
			property.Attr("style", "cursor:pointer"),
		)
	}
	return ui.LetterAvatar(name, opts.AvatarOpts, extra...)
}

// AvatarGroup renders a live avatar group with optional click event on "+N more".
func AvatarGroup(avatars []gerbera.ComponentFunc, opts LiveAvatarGroupOpts) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Class("g-avatar-group"),
	}

	display := avatars
	overflow := 0
	max := opts.Max
	if max > 0 && len(avatars) > max {
		display = avatars[:max]
		overflow = len(avatars) - max
	}

	attrs = append(attrs, display...)

	if overflow > 0 {
		size := opts.Size
		if size == "" {
			size = "md"
		}
		moreAttrs := []gerbera.ComponentFunc{
			property.Class("g-avatar", "g-avatar-circle", "g-avatar-"+size, "g-avatar-group-more"),
			property.Value(fmt.Sprintf("+%d", overflow)),
		}
		if opts.ClickEvent != "" {
			moreAttrs = append(moreAttrs,
				gl.Click(opts.ClickEvent),
				gl.ClickValue("more"),
				property.Attr("style", "cursor:pointer"),
			)
		}
		attrs = append(attrs, dom.Div(moreAttrs...))
	}

	return dom.Div(attrs...)
}
