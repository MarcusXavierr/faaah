package player

import (
	"bytes"
	"time"

	"github.com/hajimehoshi/oto/v2"
	"github.com/hajimehoshi/go-mp3"
)

// Play decodes the given mp3 data and plays it through the default audio output.
// Blocks until playback is complete.
// Returns an error if the data cannot be decoded or if audio output fails.
func Play(soundData []byte) error {
	reader := bytes.NewReader(soundData)
	decoder, err := mp3.NewDecoder(reader)
	if err != nil {
		return err
	}

	// Create an oto context
	ctx, ready, err := oto.NewContext(decoder.SampleRate(), 2, 2)
	if err != nil {
		return err
	}

	// Wait for context to be ready
	<-ready

	// Create a new player passing the decoder as the io.Reader
	p := ctx.NewPlayer(decoder)

	// Play the audio
	p.Play()

	// Wait for playback to finish
	for p.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	// Close the player
	if err := p.Close(); err != nil {
		return err
	}

	return nil
}
