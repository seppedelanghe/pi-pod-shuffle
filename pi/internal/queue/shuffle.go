package queue

import (
	"log"
	"math/rand/v2"
	"pi-pod-shuffle/internal/track"
)

// True shuffle -> nothing smart
type ShuffledQueue struct {
	stack []string

	current *track.Track
	queue   []string
	history []string
}

func NewShuffledQueue(files []string) ShuffledQueue {
	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	current, err := track.LoadTrack(files[0])
	if err != nil {
		log.Fatal(err)
	}

	return ShuffledQueue{
		stack:   files,
		current: current,
		queue:   files[1:],
		history: make([]string, 0),
	}
}

func (q *ShuffledQueue) Empty() bool {
	return len(q.queue) == 0
}

func (q *ShuffledQueue) Clear() {
	q.queue = nil
}

func (q *ShuffledQueue) Next(playtime float32) *track.Track {
	if len(q.queue) == 0 {
		return nil
	}

	_ = playtime

	q.history = append(q.history, q.current.Path)
	newTrack, err := track.LoadTrack(q.queue[0])
	if err != nil {
		log.Fatal(err)
	}
	q.queue = q.queue[1:]
	q.current = newTrack

	return q.current
}

func (q *ShuffledQueue) Current() *track.Track {
	return q.current
}

func (q *ShuffledQueue) Previous() *track.Track {
	if len(q.history) == 0 {
		return nil
	}

	q.queue = append([]string{q.current.Path}, q.queue...)
	lastIndex := len(q.history) - 1

	prevTrack, err := track.LoadTrack(q.history[lastIndex])
	if err != nil {
		log.Fatal(err)
	}
	q.current = prevTrack
	q.history = q.history[:lastIndex]

	return q.current
}
