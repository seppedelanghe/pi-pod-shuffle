package queue

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"pi-pod-shuffle/internal/track"
	"pi-pod-shuffle/internal/utils"
)

type SmartShuffledQueue struct {
	stack []string

	current    *track.Track
	queue      []string
	history    []string
	embeddings map[string][]float32

	scoredEmbedding []float32
}

func NewSmartShuffledQueue(files []string, embeddings [][]float32) SmartShuffledQueue {
	embeddingsMap := make(map[string][]float32)
	for i, file := range files {
		embeddingsMap[file] = embeddings[i]
	}

	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	current, err := track.LoadTrack(files[0])
	if err != nil {
		log.Fatal(err)
	}

	startEmbedding := make([]float32, 512)
	for i := range 512 {
		startEmbedding[i] = rand.Float32()
	}

	return SmartShuffledQueue{
		stack:           files,
		current:         current,
		queue:           make([]string, 0),
		history:         make([]string, 0),
		embeddings:      embeddingsMap,
		scoredEmbedding: startEmbedding,
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

	currentEmbedding := q.trackEmbedding(q.current.Path)

	inertia := float32(len(q.history) + 1)
	learningRate := 1.0 / inertia
	feedbackScore := (playtime * 2) - 1

	fmt.Printf("learningRate: %v\n", learningRate)
	fmt.Printf("feedbackScore: %v\n", feedbackScore)

	for i, currentVal := range currentEmbedding {
		scoredVal := q.scoredEmbedding[i]
		delta := currentVal - scoredVal
		q.scoredEmbedding[i] = scoredVal + (delta * learningRate * feedbackScore)
	}

	q.history = append(q.history, q.current.Path)

	nextPath := q.queue[0]
	q.queue = q.queue[1:]

	newTrack, err := track.LoadTrack(nextPath)
	if err != nil {
		log.Fatal(err) // TODO: handle gracefully
	}
	q.current = newTrack

	nextBest := q.findBestNextSong()
	if nextBest != "" {
		q.queue = append(q.queue, nextBest)
	}

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
		log.Fatal(err)
	}
	q.current = prevTrack

	return q.current
}

func (q *SmartShuffledQueue) trackEmbedding(path string) []float32 {
	return q.embeddings[path]
}

func (q *SmartShuffledQueue) findBestNextSong() string {
	seen := make(map[string]bool)

	for _, path := range q.history {
		seen[path] = true
	}
	if q.current != nil {
		seen[q.current.Path] = true
	}
	for _, path := range q.queue {
		seen[path] = true
	}

	var bestScore float32 = math.MaxFloat32
	var bestMatch string

	for path, embedding := range q.embeddings {
		if seen[path] {
			continue
		}

		score := utils.CosineSimilarity(q.scoredEmbedding, embedding)

		if score < bestScore {
			bestScore = score
			bestMatch = path
		}
	}

	return bestMatch
}
