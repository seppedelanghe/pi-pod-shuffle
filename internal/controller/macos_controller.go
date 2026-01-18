//go:build darwin

package controller

import (
	"fmt"
	"os"
	"pi-pod-shuffle/internal/player"
	"time"

	"golang.org/x/term"
)

type MacOSKeyboard struct {
	oldState *term.State
}

func NewMacOSKeyboard() *MacOSKeyboard {
	return &MacOSKeyboard{}
}

func (k *MacOSKeyboard) Run(p player.Player) error {
	fd := int(os.Stdin.Fd())

	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	k.oldState = state
	defer term.Restore(fd, state)

	fmt.Println("Controls: space=play/pause  n=next  p=prev  +=vol up  -=vol down  q=quit")

	buf := make([]byte, 1)

	for {
		if _, err := os.Stdin.Read(buf); err != nil {
			return err
		}

		switch buf[0] {
		case ' ':
			if p.State() == player.StatePlaying {
				p.Pause()
			} else {
				p.Play()
			}

		case 'n':
			p.Next()

		case 'p':
			p.Previous()

		case '+':
			v := p.Volume() + 0.05
			p.SetVolume(v)

		case '-':
			v := p.Volume() - 0.05
			p.SetVolume(v)

		case 'q':
			p.Stop()
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}
}

func (k *MacOSKeyboard) Stop() error {
	if k.oldState != nil {
		return term.Restore(int(os.Stdin.Fd()), k.oldState)
	}
	return nil
}
