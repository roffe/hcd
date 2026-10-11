package main

import (
	_ "embed"
	"log"

	"github.com/roffe/hcd/window"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	window.NewMainWindow().Window.ShowAndRun()
}
