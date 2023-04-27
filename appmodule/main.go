package main

import (
	"github.com/visualfc/atk/tk"
	"appmodule/base"
)

func main() {
	tk.MainLoop(func() {
		base := base.New(1)
		base.Window.ShowNormal()
	})
}