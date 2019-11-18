package main

import (
	"github.com/marcusolsson/tui-go"
)

type stacksWidget struct {
	*focusbox
	stacks      []*stack
	selectedIdx int
	rows        []*sidebarRow
	actionIdx   int

	onUpdate func(w *stacksWidget)
	focused  bool
}

func (w *stacksWidget) toggleAction() {
	row := w.rows[w.selectedIdx]
	if row.actionIdx == 0 {
		row.stack.toggleMute()
	} else if row.actionIdx == 1 {
		row.stack.toggleRunning()
	}
}

func (w *stacksWidget) update() {
	if w.actionIdx > 1 {
		w.actionIdx = 0
	} else if w.actionIdx < 0 {
		w.actionIdx = 1
	}

	if w.selectedIdx < 0 {
		w.selectedIdx = len(w.rows) - 1
	} else if w.selectedIdx >= len(w.rows) {
		w.selectedIdx = 0
	}

	for i, r := range w.rows {
		r.selected = i == w.selectedIdx
		r.actionIdx = w.actionIdx

		r.update()
	}

	w.onUpdate(w)
}

func (w *stacksWidget) getSelected() *sidebarRow {
	return w.rows[w.selectedIdx]
}

func (w *stacksWidget) OnKeyEvent(ev tui.KeyEvent) {
	if !w.IsFocused() {
		return
	}

	switch ev.Key {
	case tui.KeyUp:
		w.selectedIdx--
	case tui.KeyDown:
		w.selectedIdx++
	case tui.KeyLeft, tui.KeyRight:
		w.actionIdx++
	case tui.KeyRune:
		if ev.Name() == " " {
			w.toggleAction()
		}
	}

	w.update()

	w.focusbox.OnKeyEvent(ev)
}

func newSidebar(stacks []*stack) *stacksWidget {
	result := new(stacksWidget)
	result.focusbox = wrapFocusbox(tui.NewVBox())
	result.SetBorder(true)
	result.SetTitle("stacks")
	result.rows = []*sidebarRow{}
	result.stacks = stacks
	result.onUpdate = func(w *stacksWidget) {

	}

	for _, s := range result.stacks {

		row := newSidebarRow(s)
		result.rows = append(result.rows, row)

		padder := tui.NewPadder(1, 0, row)
		result.Append(padder)
	}
	result.Append(tui.NewSpacer())
	result.update()

	return result
}
