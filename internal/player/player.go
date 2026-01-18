package player

import (
	"fmt"
	"pi-pod-shuffle/internal/audio"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

type player struct {
	mu sync.Mutex

	state State

	queue   []Track
	history []Track
	current *Track

	mixer  *beep.Mixer
	ctrl   *beep.Ctrl
	volume *effects.Volume

	sampleRate beep.SampleRate
}

func newPlayer(sampleRate int) (*player, error) {
	sr := beep.SampleRate(sampleRate)

	if err := speaker.Init(sr, sr.N(time.Second/4)); err != nil {
		return nil, err
	}

	mixer := &beep.Mixer{}
	vol := &effects.Volume{
		Streamer: mixer,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}

	speaker.Play(vol)

	return &player{
		state:      StateStopped,
		mixer:      mixer,
		ctrl:       nil,
		volume:     vol,
		sampleRate: sr,
	}, nil
}

func (p *player) Enqueue(t Track) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = append(p.queue, t)
	return nil
}

func (p *player) EnqueueNext(t Track) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = append([]Track{t}, p.queue...)
	return nil
}

func (p *player) ClearQueue() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = nil
}

func (p *player) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.state {
	case StatePaused:
		speaker.Lock()
		if p.ctrl != nil {
			p.ctrl.Paused = false
		}
		speaker.Unlock()
		p.state = StatePlaying
	case StateStopped:
		return p.playNextLocked()
	}

	return nil
}

func (p *player) Pause() {
	speaker.Lock()
	defer speaker.Unlock()

	if p.state == StatePlaying {
		p.state = StatePaused
		p.ctrl.Paused = true
	}
}

func (p *player) Stop() {
	speaker.Lock()
	defer speaker.Unlock()

	p.state = StateStopped
	p.volume.Silent = true
	p.mixer.Clear()
	p.current = nil
}

func (p *player) Next() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.playNextLocked()
}

func (p *player) Previous() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.history) == 0 {
		return nil
	}

	last := p.history[len(p.history)-1]
	p.history = p.history[:len(p.history)-1]
	p.queue = append([]Track{last}, p.queue...)

	return p.playNextLocked()
}

func (p *player) playNextLocked() error {
	if len(p.queue) == 0 {
		p.state = StateStopped
		return fmt.Errorf("Player: Queue empty")
	}

	next := p.queue[0]
	p.queue = p.queue[1:]

	streamer, format, err := audio.Decode(next.Path)
	if err != nil {
		return err
	}

	var s beep.Streamer = streamer
	if format.SampleRate != p.sampleRate {
		s = beep.Resample(4, format.SampleRate, p.sampleRate, s)
	}

	p.current = &next
	p.history = append(p.history, next)
	p.state = StatePlaying

	speaker.Lock()
	p.volume.Silent = false

	p.ctrl = &beep.Ctrl{
		Paused:   false,
		Streamer: s,
	}
	p.ctrl.Streamer = s
	seq := beep.Seq(
		p.ctrl,
		beep.Callback(func() {
			streamer.Close()
			p.mu.Lock()
			defer p.mu.Unlock()
			p.playNextLocked()
		}),
	)

	p.mixer.Clear()
	p.mixer.Add(seq)
	speaker.Unlock()
	return nil
}

func (p *player) SetVolume(v Volume) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	speaker.Lock()
	defer speaker.Unlock()

	if v == 0 {
		p.volume.Silent = true
		return
	}

	p.volume.Silent = false
	p.volume.Volume = float64(v*2) - 1 // maps nicely to log scale
}

func (p *player) Volume() Volume {
	return Volume((p.volume.Volume + 1) / 2)
}

func (p *player) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *player) Current() *Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

func (p *player) Queue() []Track {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := make([]Track, len(p.queue))
	copy(out, p.queue)
	return out
}
