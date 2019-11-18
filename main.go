package main

import (
	"log"
	"os"

	"github.com/marcusolsson/tui-go"
	"github.com/v-braun/go-must"
)

var logger tui.Logger

func main() {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println(fmt.Sprintf("%v", err))
	// 	}
	// }()

	f, _ := os.Create("debug.log")
	defer f.Close()
	logger = log.New(f, "", log.LstdFlags)

	cfg := loadOrCreateConf()
	stacks := createStacks(cfg)
	sidebar := newSidebar(stacks)
	sidebar.SetSizePolicy(tui.Maximum, tui.Maximum)

	logsInner := tui.NewList()
	logsScroll := tui.NewScrollArea(logsInner)
	logsScroll.SetAutoscrollToBottom(true)
	logs := wrapFocusbox(tui.NewVBox(logsScroll))
	logs.SetBorder(true)
	logs.SetTitle("logs")
	logs.SetSizePolicy(tui.Expanding, tui.Expanding)
	// logsInner.AddItems("hello world")

	statusLbl := tui.NewLabel("")
	statusLbl.SetWordWrap(true)

	status := wrapFocusbox(tui.NewHBox(
		statusLbl,
	))
	status.SetBorder(true)
	status.SetTitle("status")
	status.SetSizePolicy(tui.Maximum, tui.Maximum)

	sidebar.onUpdate = func(w *stacksWidget) {
		selected := w.getSelected()
		stat := selected.stack.GetStatus()
		statusLbl.SetText(stat.StatusMessage)
		status.SetTitle(stat.Title)
	}

	main := tui.NewVBox(
		logs,
		status,
	)

	root := tui.NewHBox(sidebar, main)
	ui, err := tui.New(root)
	must.NoError(err, "could not build GUI")

	tui.DefaultFocusChain.Set(sidebar, logs, status)

	t := tui.NewTheme()
	normal := tui.Style{Bg: tui.ColorBlack, Fg: tui.ColorWhite}
	t.SetStyle("normal", normal)
	t.SetStyle("selected", tui.Style{Bg: tui.ColorCyan, Fg: tui.ColorWhite})
	t.SetStyle("not-focused", tui.Style{Fg: tui.ColorDefault, Bg: tui.ColorBlack})
	t.SetStyle("focused", tui.Style{Fg: tui.ColorWhite, Bg: tui.ColorBlack})

	// force update before start
	sidebar.update()

	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetTheme(t)

	// tui.SetLogger(logger)

	for _, s := range stacks {
		s.onLog = func(msg string) {
			if msg != "" {
				logsInner.AddItems(msg)
			}

			sidebar.update()
			ui.Repaint()
		}
		s.onUpdate = func() {
			sidebar.update()
			ui.Repaint()
		}
	}

	err = ui.Run()
	must.NoError(err, "app failed")
}
