package player

type Player interface {
	// Queue management
	Enqueue(track Track) error
	EnqueueNext(track Track) error
	ClearQueue()

	// Transport
	Play() error
	Pause()
	Stop()

	Next() error
	Previous() error

	// Volume
	SetVolume(v Volume)
	Volume() Volume

	// Introspection
	State() State
	Current() *Track
	Queue() []Track
}

func New(sampleRate int) (Player, error) {
	return newPlayer(sampleRate)
}
