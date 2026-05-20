package main

import (
	_ "embed"
	"log"

	"github.com/roffe/hcd/store"
	"github.com/roffe/hcd/window"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	db, err := store.New()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	window.NewMainWindow(db).Window.ShowAndRun()
}
