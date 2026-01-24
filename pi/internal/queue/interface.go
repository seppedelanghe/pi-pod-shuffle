package queue

import "pi-pod-shuffle/internal/track"

type MusicQueue interface {
	Empty() bool
	Clear()

	Next(playtime float32) *track.Track
	Current() *track.Track
	Previous() *track.Track
}
