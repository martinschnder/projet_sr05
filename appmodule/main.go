package main

import (
	"github.com/visualfc/atk/tk"
	"appmodule/base"
)

func main() {
	tk.MainLoop(func() {
		base := base.New()
		base.Window.ShowNormal()
	})
}