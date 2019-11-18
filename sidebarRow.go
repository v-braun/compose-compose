package main

import (
	"fmt"

	"github.com/marcusolsson/tui-go"
)

type sidebarRow struct {
	*tui.Box
	muteLbl  *tui.Label
	playLbl  *tui.Label
	titleLbl *tui.Label
	selected bool

	style     string
	actionIdx int

	stack *stack
}

func (s *sidebarRow) update() {
	s.style = "normal"
	if s.selected {
		s.style = "selected"
	}

	mute := "🔈"
	if s.stack.muted {
		mute = "🔇"
	}

	stat := s.stack.GetStatus()

	start := "🔴"
	if stat.Status == runningStatus {
		start = "✅"
	} else if stat.Status == stoppedStatus {
		start = "🔴"
	} else if stat.Status == unknownStatus {
		start = "⏳"
	} else if stat.Status == warningStatus {
		start = "🔶"
	}

	if !s.selected {
		s.muteLbl.SetText(fmt.Sprintf(" %s ", mute))
		s.playLbl.SetText(fmt.Sprintf(" %s ", start))
	} else if s.actionIdx == 0 {
		s.muteLbl.SetText(fmt.Sprintf("(%s)", mute))
		s.playLbl.SetText(fmt.Sprintf(" %s ", start))
	} else if s.actionIdx == 1 {
		s.muteLbl.SetText(fmt.Sprintf(" %s ", mute))
		s.playLbl.SetText(fmt.Sprintf("(%s)", start))
	}
}

func (s *sidebarRow) Draw(p *tui.Painter) {
	p.WithStyle(s.style, func(p *tui.Painter) {
		s.Box.Draw(p)
	})
}

func newSidebarRow(stack *stack) *sidebarRow {
	status := stack.GetStatus()
	result := new(sidebarRow)
	result.muteLbl = tui.NewLabel("")
	result.playLbl = tui.NewLabel("")
	result.titleLbl = tui.NewLabel(status.Title)
	result.stack = stack

	actions := tui.NewHBox(
		result.muteLbl,
		result.playLbl,
	)
	actions.SetSizePolicy(tui.Maximum, tui.Maximum)

	row := tui.NewHBox(
		tui.NewPadder(1, 0,
			actions,
		),
		tui.NewPadder(1, 0,
			result.titleLbl,
		),
	)

	result.selected = false
	result.Box = row
	return result
}
