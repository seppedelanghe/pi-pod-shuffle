//go:build linux

package controller

import (
	"fmt"
	"pi-pod-shuffle/internal/player"
	"strings"

	"github.com/gvalkov/golang-evdev"
)

type AVRCPController struct {
	device *evdev.InputDevice
}

func NewAVRCPController() (*AVRCPController, error) {
	devices, err := evdev.ListInputDevices()
	if err != nil {
		return nil, err
	}

	var bestCandidate *evdev.InputDevice

	for _, dev := range devices {
		// Bluetooth media devices almost always have "AVRCP" in the name
		// e.g., "Sony WH-1000XM3 (AVRCP)" or "Pixel Buds (AVRCP)"
		if strings.Contains(dev.Name, "AVRCP") {
			fmt.Printf("Found Bluetooth Media Device: %s (%s)\n", dev.Name, dev.Fn)
			bestCandidate = dev
			break
		}
	}

	if bestCandidate == nil {
		return nil, fmt.Errorf("no Bluetooth headphones found. Is the device connected?")
	}

	return &AVRCPController{device: bestCandidate}, nil
}

func (b *AVRCPController) Run(p player.Player) error {
	b.device.Grab()
	defer b.device.Release()

	fmt.Printf("Listening for controls from: %s\n", b.device.Name)

	for {
		readEvents, err := b.device.Read()
		if err != nil {
			// If reading fails, the device probably disconnected.
			return fmt.Errorf("device disconnected: %v", err)
		}

		for _, event := range readEvents {
			// Only care about Key Presses (Value 1)
			if event.Type == evdev.EV_KEY && event.Value == 1 {
				fmt.Printf("Event: %s\n", event.String())

				switch event.Code {
				case evdev.KEY_PLAYPAUSE, evdev.KEY_PAUSECD, evdev.KEY_PLAY, evdev.KEY_PLAYCD:
					if p.State() == player.StatePlaying {
						p.Pause()
					} else {
						p.Play()
					}

				case evdev.KEY_NEXTSONG, evdev.KEY_NEXT:
					if err := p.Next(); err != nil {
						fmt.Println("Error skipping:", err)
					}

				case evdev.KEY_PREVIOUSSONG, evdev.KEY_PREVIOUS:
					if err := p.Previous(); err != nil {
						fmt.Println("Error rewinding:", err)
					}

				// Some headphones send volume keys, some don't (Absolute Volume)
				case evdev.KEY_VOLUMEUP:
					p.SetVolume(p.Volume() + 0.05)
				case evdev.KEY_VOLUMEDOWN:
					p.SetVolume(p.Volume() - 0.05)
				}
			}
		}
	}
}

func (b *AVRCPController) Stop() error {
	return nil
}
