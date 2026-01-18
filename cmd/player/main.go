package main

import (
	"log"
	"os"
	"pi-pod-shuffle/internal/controller"
	"pi-pod-shuffle/internal/player"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: player <audiofile1> [audiofile2...]")
	}

	p, err := player.New(44100)
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range os.Args[1:] {
		if err := p.Enqueue(player.Track{Path: path}); err != nil {
			log.Fatal(err)
		}
	}

	p.SetVolume(1.0)
	ctrl := controller.NewMacOSKeyboard()
	if err := ctrl.Run(p); err != nil {
		log.Fatal(err)
	}
}
