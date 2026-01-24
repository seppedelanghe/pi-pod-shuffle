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

var extensions = []string{".flac"}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: player <library>")
	}

	dirname := os.Args[1]
	files, err := io.FindFiles(dirname, extensions)
	if err != nil {
		log.Fatal(err)
	}

	musicQueue := queue.NewShuffledQueue(files)
	fmt.Printf("Found %d songs\n", len(files))

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
