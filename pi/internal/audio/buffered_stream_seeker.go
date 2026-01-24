package audio

import (
	"sync"
	"time"

	"github.com/faiface/beep"
)

// BufferedStreamSeeker wraps a StreamSeeker to provide buffering and lookahead.
// This prevents stuttering caused by disk I/O or heavy resampling.
type BufferedStreamSeeker struct {
	mu       sync.Mutex
	streamer beep.StreamSeeker

	// Buffer configuration
	bufferSize int // Total samples to hold
	chunkSize  int // Samples per read op

	// Data pipe
	chunks  chan [][2]float64
	current [][2]float64

	// Control
	quit     chan struct{}
	seekErr  error
	position int
}

func NewBufferedStreamSeeker(s beep.StreamSeeker, bufferDuration time.Duration, format beep.Format) *BufferedStreamSeeker {
	// Calculate buffer size based on duration (e.g., 2 seconds)
	bufferSize := format.SampleRate.N(bufferDuration)
	chunkSize := 1024 * 4 // 4K samples per chunk is a good balance

	b := &BufferedStreamSeeker{
		streamer:   s,
		bufferSize: bufferSize,
		chunkSize:  chunkSize,
		quit:       make(chan struct{}),
	}

	b.startWorker()
	return b
}

// startWorker starts the background goroutine that fills the buffer.
func (b *BufferedStreamSeeker) startWorker() {
	b.chunks = make(chan [][2]float64, b.bufferSize/b.chunkSize)

	go func() {
		defer close(b.chunks)

		for {
			// Check if we need to stop (e.g., during a Seek or Close)
			select {
			case <-b.quit:
				return
			default:
			}

			// Allocate a new chunk
			data := make([][2]float64, b.chunkSize)
			n, ok := b.streamer.Stream(data)

			if n > 0 {
				// Send data to the buffer.
				// If buffer is full, this blocks until the speaker consumes data.
				select {
				case b.chunks <- data[:n]:
				case <-b.quit:
					return
				}
			}

			if !ok {
				return // End of stream
			}
		}
	}()
}

func (b *BufferedStreamSeeker) Stream(samples [][2]float64) (n int, ok bool) {
	// We don't lock here to keep playback fast. Channel read is thread-safe.
	filled := 0
	for filled < len(samples) {
		if len(b.current) == 0 {
			newChunk, open := <-b.chunks
			if !open {
				return filled, filled > 0
			}
			b.current = newChunk
		}

		n := copy(samples[filled:], b.current)
		b.current = b.current[n:]
		filled += n
		b.position += n
	}
	return filled, true
}

func (b *BufferedStreamSeeker) Seek(p int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	close(b.quit)

	err := b.streamer.Seek(p)
	if err != nil {
		// Restart worker even if seek failed, to keep stream alive
		b.quit = make(chan struct{})
		b.startWorker()
		return err
	}

	b.current = nil
	b.position = p
	b.quit = make(chan struct{})

	b.startWorker()

	return nil
}

func (b *BufferedStreamSeeker) Err() error {
	return b.streamer.Err()
}

func (b *BufferedStreamSeeker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Signal worker to stop
	select {
	case <-b.quit:
	default:
		close(b.quit)
	}
}

// Helper to expose position (useful for UI)
func (b *BufferedStreamSeeker) Position() int {
	return b.position
}
