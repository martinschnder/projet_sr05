package base

import (
	"github.com/visualfc/atk/tk"
)

type Base struct {
	Window *tk.Window
	Content string
}

func New() Base {
    base := Base{
		Window: NewWindow(),
		Content: "aaaaa",
	}
	return base
}

func NewWindow() *tk.Window {
	win := tk.RootWindow()
	lbl := tk.NewLabel(win, "Hello ATK")
	btn := tk.NewButton(win, "Quit")
	btn.OnCommand(func() {
		tk.Quit()
	})
	tk.NewVPackLayout(win).AddWidgets(lbl, tk.NewLayoutSpacer(win, 0, true), btn)
	win.ResizeN(300, 200)
	win.SetTitle("ATK Sample")
	win.Center(nil)
	return win
}