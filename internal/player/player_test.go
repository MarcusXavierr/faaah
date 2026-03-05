package player

import (
	"bytes"
	"testing"

	"github.com/MarcusXavierr/faaah/assets"
	"github.com/hajimehoshi/go-mp3"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectError bool
	}{
		{
			name:        "Decode valid mp3",
			input:       assets.SoundFile,
			expectError: false,
		},
		{
			name:        "Decode invalid data",
			input:       []byte("not an mp3"),
			expectError: true,
		},
		{
			name:        "Decode empty data",
			input:       []byte{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder, err := mp3.NewDecoder(bytes.NewReader(tt.input))
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if decoder.SampleRate() <= 0 {
				t.Errorf("expected positive SampleRate, got %d", decoder.SampleRate())
			}

			if decoder.Length() <= 0 {
				t.Errorf("expected positive Length, got %d", decoder.Length())
			}
		})
	}
}
