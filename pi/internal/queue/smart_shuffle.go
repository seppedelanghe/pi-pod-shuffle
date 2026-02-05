package queue

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"pi-pod-shuffle/internal/io"
	"pi-pod-shuffle/internal/track"
)

const (
	MIN_LR    = 0.3
	MAX_LR    = 2.0
	DECAY     = 5.0
	STEEPNESS = 4.0
)

type SmartShuffledQueue struct {
	stack           []string
	current         *track.Track
	queue           []string
	history         []string
	embeddings      map[string][]float32
	scoredEmbedding []float32
	inertia         float32
}

func NewSmartShuffledQueue(library *io.MusicLibrary) SmartShuffledQueue {
	files := library.Filenames()
	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	normalizedEmbeddings := make(map[string][]float32, len(library.Files))
	for k, v := range library.Files {
		vec := make([]float32, len(v))
		copy(vec, v)
		normalize(vec)
		normalizedEmbeddings[k] = vec
	}

	current, err := track.LoadTrack(files[0])
	if err != nil {
		log.Fatal(err)
	}

	startEmbedding := make([]float32, 512)
	limit := 3
	if len(files) < 3 {
		limit = len(files)
	}

	for i := 0; i < limit; i++ {
		vec := normalizedEmbeddings[files[i]]
		for j := range 512 {
			startEmbedding[j] += vec[j]
		}
	}

	for i := range 512 {
		startEmbedding[i] = (startEmbedding[i] / float32(limit)) + (rand.Float32() - 0.5)
	}
	normalize(startEmbedding)

	return SmartShuffledQueue{
		stack:           files,
		current:         current,
		queue:           make([]string, 0),
		history:         make([]string, 0),
		embeddings:      normalizedEmbeddings,
		scoredEmbedding: startEmbedding,
		inertia:         1.0,
	}
}

func (q *SmartShuffledQueue) Empty() bool {
	return len(q.queue) == 0
}

func (q *SmartShuffledQueue) Clear() {
	q.queue = nil
}

func (q *SmartShuffledQueue) Next(playtime float32) *track.Track {
	if len(q.queue) == 0 {
		nextSong := q.findBestNextSong()
		if nextSong == "" {
			return nil
		}
		q.queue = append(q.queue, nextSong)
	}

	q.inertia += playtime

	currentEmbedding := q.trackEmbedding(q.current.Path)
	if currentEmbedding != nil {
		rawLR := float64(DECAY / q.inertia)
		learningRate := float32(math.Min(MAX_LR, math.Max(MIN_LR, rawLR)))

		centered := float64(playtime) - 0.5
		curve := math.Tanh(centered * STEEPNESS)
		feedbackScore := float32(curve * 1.5)

		fmt.Printf("Time: %.2f | Score: %.2f | LR: %.2f | Inertia: %.2f\n", playtime, feedbackScore, learningRate, q.inertia)

		for i, currentVal := range currentEmbedding {
			scoredVal := q.scoredEmbedding[i]
			delta := currentVal - scoredVal
			q.scoredEmbedding[i] = scoredVal + (delta * learningRate * feedbackScore)
		}

		normalize(q.scoredEmbedding)
	}

	q.history = append(q.history, q.current.Path)

	nextPath := q.queue[0]
	q.queue = q.queue[1:]

	newTrack, err := track.LoadTrack(nextPath)
	if err != nil {
		log.Printf("Error loading track %s: %v", nextPath, err)
		if len(q.stack) > 0 {
			return q.Next(0.0)
		}
		return nil
	}
	q.current = newTrack

	return q.current
}

func (q *SmartShuffledQueue) Current() *track.Track {
	return q.current
}

func (q *SmartShuffledQueue) Previous() *track.Track {
	if len(q.history) == 0 {
		return nil
	}

	q.queue = append([]string{q.current.Path}, q.queue...)

	lastIndex := len(q.history) - 1
	prevPath := q.history[lastIndex]
	q.history = q.history[:lastIndex]

	prevTrack, err := track.LoadTrack(prevPath)
	if err != nil {
		log.Printf("Error loading prev track: %v", err)
		return q.current
	}
	q.current = prevTrack

	return q.current
}

func (q *SmartShuffledQueue) trackEmbedding(path string) []float32 {
	return q.embeddings[path]
}

func (q *SmartShuffledQueue) findBestNextSong() string {
	seen := make(map[string]bool, len(q.history)+len(q.queue)+1)
	for _, path := range q.history {
		seen[path] = true
	}
	if q.current != nil {
		seen[q.current.Path] = true
	}
	for _, path := range q.queue {
		seen[path] = true
	}

	var bestScore float32 = -2.0
	var bestMatch string
	var fallbackMatch string

	for _, path := range q.stack {
		if seen[path] {
			continue
		}

		if fallbackMatch == "" {
			fallbackMatch = path
		}

		embedding := q.embeddings[path]
		var score float32
		for i, v := range q.scoredEmbedding {
			score += v * embedding[i]
		}

		if score > bestScore {
			bestScore = score
			bestMatch = path
		}
	}

	if bestMatch == "" {
		return fallbackMatch
	}

	return bestMatch
}

func normalize(vec []float32) {
	var sum float32
	for _, v := range vec {
		sum += v * v
	}
	magnitude := float32(math.Sqrt(float64(sum)))

	if magnitude > 1e-9 {
		for i := range vec {
			vec[i] /= magnitude
		}
	}
}
