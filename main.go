package main

import (
	_ "embed"
	"log"

	"github.com/hirschmann-koxha-gbr/hcd/window"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	mw := window.NewMainWindow()
	mw.Window.ShowAndRun()
}
