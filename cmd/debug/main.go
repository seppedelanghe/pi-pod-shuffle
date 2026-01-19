package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/speaker"
)

func main() {
	f, err := os.Open("track.flac")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := flac.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	sr := format.SampleRate
	speaker.Init(sr, sr.N(time.Second/4))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
