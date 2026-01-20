package main

import (
	"log"
	"os"
	"path/filepath"
	"pi-pod-shuffle/internal/controller"
	"pi-pod-shuffle/internal/player"
	"pi-pod-shuffle/internal/queue"
	"pi-pod-shuffle/internal/track"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: player <dir>")
	}

	dirname := os.Args[1]
	files, err := os.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}

	tracks := make([]track.Track, 0)
	for _, file := range files {
		if !file.IsDir() {
			relPath := filepath.Join(dirname, file.Name())
			absPath, err := filepath.Abs(relPath)
			if err != nil {
				log.Fatal(err)
			}
			tracks = append(tracks, track.Track{Path: absPath})
		}
	}
	musicQueue := queue.NewShuffledQueue(tracks)

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
