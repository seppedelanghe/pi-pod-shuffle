package audio

import "github.com/faiface/beep"

type AudioBackend interface {
	Init(sampleRate int) error
	Play(stream beep.Streamer)
	Lock()
	Unlock()
}
