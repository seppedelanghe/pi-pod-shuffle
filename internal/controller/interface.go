package controller

import "pi-pod-shuffle/internal/player"

type Controller interface {
	Run(p player.Player) error
	Stop() error
}
