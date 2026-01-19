package controller

type ControlCommand int

const (
	CmdPlayPause ControlCommand = iota
	CmdNext
	CmdPrevious
	CmdVolumeUp
	CmdVolumeDown
	CmdQuit
)
