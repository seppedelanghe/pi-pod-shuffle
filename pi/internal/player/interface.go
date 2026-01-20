package player

import "pi-pod-shuffle/internal/queue"

type Player interface {
	// Transport
	Play() error
	Pause()
	Stop()

	Next() error
	Previous() error

	// Volume
	SetVolume(v Volume)
	Volume() Volume

	// Introspection
	State() State
}

func New(sampleRate int, musicQueue queue.MusicQueue) (Player, error) {
	return newPlayer(sampleRate, musicQueue)
}
