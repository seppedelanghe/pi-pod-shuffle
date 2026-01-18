package player

type State int

const (
	StateStopped State = iota
	StatePlaying
	StatePaused
)
