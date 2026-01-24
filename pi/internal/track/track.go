package track

import (
	"pi-pod-shuffle/internal/audio"
	"time"

	"github.com/faiface/beep"
)

type Track struct {
	Path         string
	Streamer     beep.StreamSeekCloser
	Format       beep.Format
	TotalSamples int
}

func LoadTrack(path string) (*Track, error) {
	streamer, format, err := audio.Decode(path)
	if err != nil {
		return nil, err
	}

	return &Track{
		Path:         path,
		Streamer:     streamer,
		Format:       format,
		TotalSamples: streamer.Len(),
	}, nil
}

func (t *Track) GetTotalDuration() time.Duration {
	seconds := float64(t.TotalSamples) / float64(t.Format.SampleRate)
	return time.Duration(seconds * float64(time.Second))
}

func (t *Track) GetPlaytimePercentage() float32 {
	if t.TotalSamples == 0 {
		return 0.0
	}

	currentPos := t.Streamer.Position()
	ratio := float32(currentPos) / float32(t.TotalSamples)

	if ratio > 1.0 {
		return 1.0
	}
	return ratio
}

func (t *Track) Close() {
	t.Streamer.Close()
}
