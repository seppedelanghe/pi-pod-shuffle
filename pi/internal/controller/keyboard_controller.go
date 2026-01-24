//go:build !linux

package controller

import (
	"fmt"
	"os"
	"pi-pod-shuffle/internal/player"

	"golang.org/x/term"
)

type KeyboardController struct {
	oldState *term.State
}

func NewKeyboardController() *KeyboardController {
	return &KeyboardController{}
}

func (k *KeyboardController) Run(p player.Player) error {
	fmt.Print("\033[H\033[2J")
	fmt.Print("Pi Pod Shuffle Player\r\n")
	fmt.Printf("Controls:\tspace=play/pause\tn=next\tp=prev\t+=vol up\t-=vol down\tq=quit")

	fd := int(os.Stdin.Fd())

	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	k.oldState = state
	defer term.Restore(fd, state)

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
			err := p.Next()
			if err != nil {
				fmt.Println(err)
				p.Stop()
			}

		case 'p':
			err := p.Previous()
			if err != nil {
				fmt.Println(err)
				p.Stop()
			}

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
	}
}

func (k *KeyboardController) Stop() error {
	if k.oldState != nil {
		return term.Restore(int(os.Stdin.Fd()), k.oldState)
	}
	return nil
}
