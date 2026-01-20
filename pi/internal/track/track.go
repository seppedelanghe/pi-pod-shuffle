package track

import "time"

type Track struct {
	Path      string
	Duration  time.Duration
	Embedding []float32
}
