package base

import (
	"github.com/visualfc/atk/tk"
	"strconv"
)

type Base struct {
	Window *tk.Window
	Text *tk.Text
	Content string
	Id int
}

func New(id int) Base {
	content := "Voici le fichier initial"

	win := tk.RootWindow()

	btn := tk.NewButton(win, "Quit")
	btn.OnCommand(func() {
		tk.Quit()
	})

	txt := tk.NewText(win)
	txt.InsertText(1.0, content)
	txt.SetReadOnly(true)

	lineNumber := tk.NewEntry(win)
	labelLine := tk.NewLabel(win, "Numéro de ligne")

	numberLayout = tk.NewVPackLayout(win)
	numberLayout.AddWidgets(labelLine, lineNumber)
	
	tk.NewHPackLayout(win).AddWidgets(txt, btn)

	tk.NewVPackLayout(win).AddWidgets(txt, btn)

	win.ResizeN(700,700)
	win.SetTitle("Utilisateur n°" + strconv.Itoa(id))
	win.Center(nil)

    base := Base{
		Window: win,
		Text: txt,
		Content: content,
		Id: id,
	}
	return base
}