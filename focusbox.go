package main

import "github.com/marcusolsson/tui-go"

type focusbox struct {
	*tui.Box
	focused bool
}

func wrapFocusbox(b *tui.Box) *focusbox {
	return &focusbox{
		Box: b,
	}
}

func (w *focusbox) Draw(p *tui.Painter) {
	style := "not-focused"
	if w.focused {
		style = "focused"
	}

	p.WithStyle(style, func(p *tui.Painter) {
		w.Box.Draw(p)
	})
}

func (w *focusbox) SetFocused(f bool) {
	// logger.Printf("set focused %v", f)
	w.focused = f
	// w.Box.SetFocused(f)
}
func (w *focusbox) IsFocused() bool {
	return w.focused
}
