package main

import (
	"fmt"
	"log"
	"os"
	"pi-pod-shuffle/internal/controller"
	"pi-pod-shuffle/internal/io"
	"pi-pod-shuffle/internal/player"
	"pi-pod-shuffle/internal/queue"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: player <library>")
	}

	libraryPath := os.Args[1]
	library, err := io.LoadMusicLibary(libraryPath)
	if err != nil {
		log.Fatal(err)
	}

	musicQueue := queue.NewSmartShuffledQueue(library)
	fmt.Printf("Found %d songs\n", len(library.Files))

	p, err := player.New(44100, &musicQueue)
	if err != nil {
		log.Fatal(err)
	}

	p.SetVolume(0.0)
	ctrl := controller.NewKeyboardController()
	if err := ctrl.Run(p); err != nil {
		log.Fatal(err)
	}
}
