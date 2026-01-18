package main

import (
	"log"
	"os"
	"pi-pod-shuffle/internal/player"
	"time"
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

	p.SetVolume(0.8)
	err = p.Play()
	if err != nil {
		panic(err)
	}

	// crude but effective: keep process alive
	for p.State() != player.StateStopped {
		time.Sleep(200 * time.Millisecond)
	}
}
