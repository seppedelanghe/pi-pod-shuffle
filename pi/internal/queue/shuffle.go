package queue

import (
	"fmt"
	"math/rand/v2"
	"pi-pod-shuffle/internal/track"
)

// True shuffle -> nothing smart
type ShuffledQueue struct {
	stack []track.Track

	current *track.Track
	queue   []*track.Track
	history []*track.Track
}

func NewShuffledQueue(tracks []track.Track) ShuffledQueue {
	rand.Shuffle(len(tracks), func(i, j int) {
		tracks[i], tracks[j] = tracks[j], tracks[i]
	})

	current := &tracks[0]
	queue := make([]*track.Track, len(tracks)-1)
	history := make([]*track.Track, 0)

	fmt.Printf("len(tracks): %v\n", len(tracks))
	for i := range tracks[1:] {
		queue[i] = &tracks[i+1]
	}

	return ShuffledQueue{
		stack:   tracks,
		current: current,
		queue:   queue,
		history: history,
	}
}

func (q *ShuffledQueue) Empty() bool {
	return len(q.queue) == 0
}

func (q *ShuffledQueue) Clear() {
	q.queue = nil
}

func (q *ShuffledQueue) Next() *track.Track {
	if len(q.queue) == 0 {
		return nil
	}

	q.history = append(q.history, q.current)
	q.current = q.queue[0]
	q.queue = q.queue[1:]

	fmt.Printf("q.current.Path: %v\n", q.current.Path)
	return q.current
}

func (q *ShuffledQueue) Current() *track.Track {
	fmt.Printf("q.current.Path: %v\n", q.current.Path)
	return q.current
}

func (q *ShuffledQueue) Previous() *track.Track {
	if len(q.history) == 0 {
		return nil
	}

	q.queue = append([]*track.Track{q.current}, q.queue...)
	lastIndex := len(q.history) - 1

	q.current = q.history[lastIndex]
	q.history = q.history[:lastIndex]
	fmt.Printf("q.current.Path: %v\n", q.current.Path)

	return q.current
}
