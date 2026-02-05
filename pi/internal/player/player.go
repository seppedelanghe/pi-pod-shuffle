package player

import (
	"fmt"
	"pi-pod-shuffle/internal/audio"
	"pi-pod-shuffle/internal/queue"
	"pi-pod-shuffle/internal/track"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

type player struct {
	mu sync.Mutex

	state State
	queue queue.MusicQueue

	mixer  *beep.Mixer
	ctrl   *beep.Ctrl
	volume *effects.Volume

	sampleRate beep.SampleRate
}

func newPlayer(sampleRate int, musicQueue queue.MusicQueue) (*player, error) {
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
		queue:      musicQueue,
		mixer:      mixer,
		ctrl:       nil,
		volume:     vol,
		sampleRate: sr,
	}, nil
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
		return p.playNextLocked(p.queue.Current())
	}

	return nil
}

func (p *player) Pause() {
	speaker.Lock()
	defer speaker.Unlock()

	if p.state == StatePlaying {
		p.state = StatePaused
		if p.ctrl != nil {
			p.ctrl.Paused = true
		}
	}
}

func (p *player) Stop() {
	speaker.Lock()
	defer speaker.Unlock()

	p.state = StateStopped
	p.volume.Silent = true
	p.mixer.Clear()

	if curr := p.queue.Current(); curr != nil && curr.Streamer != nil {
		curr.Close()
	}

	p.queue.Clear()
}

func (p *player) Next() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	playtime := p.calculatePlaytime()
	if curr := p.queue.Current(); curr != nil && curr.Streamer != nil {
		curr.Close()
	}

	next := p.queue.Next(playtime)
	return p.playNextLocked(next)
}

func (p *player) Previous() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if curr := p.queue.Current(); curr != nil && curr.Streamer != nil {
		curr.Close()
	}

	track := p.queue.Previous()
	return p.playNextLocked(track)
}

func (p *player) playNextLocked(track *track.Track) error {
	buffSeeker := audio.NewBufferedStreamSeeker(track.Streamer, time.Second*2, track.Format)

	var finalStream beep.Streamer = buffSeeker
	if track.Format.SampleRate != p.sampleRate {
		finalStream = beep.Resample(4, track.Format.SampleRate, p.sampleRate, buffSeeker)
	}

	p.state = StatePlaying
	speaker.Lock()
	p.volume.Silent = false

	p.ctrl = &beep.Ctrl{
		Paused:   false,
		Streamer: finalStream,
	}

	seq := beep.Seq(
		p.ctrl,
		beep.Callback(func() {
			go p.handleTrackFinished()
		}),
	)

	p.mixer.Clear()
	p.mixer.Add(seq)
	speaker.Unlock()

	fmt.Printf("\r\nPlaying track: '%s'", track.Path)
	return nil
}

func (p *player) handleTrackFinished() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StateStopped {
		return
	}

	if curr := p.queue.Current(); curr != nil && curr.Streamer != nil {
		curr.Close()
	}

	next := p.queue.Next(1.0)
	if err := p.playNextLocked(next); err != nil {
		fmt.Println("Error playing next track:", err)
		p.state = StateStopped
	}
}

func (p *player) calculatePlaytime() float32 {
	curr := p.queue.Current()
	if curr == nil || curr.Streamer == nil || curr.Streamer.Len() == 0 {
		return 0.0
	}

	pos := curr.Streamer.Position()
	trackLen := curr.Streamer.Len()

	percent := float32(pos) / float32(trackLen)

	if percent > 1.0 {
		return 1.0
	}
	return percent
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
