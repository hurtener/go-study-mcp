// Package audio provides PCM-to-MP3 conversion using shine-mp3.
package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/braheezy/shine-mp3/pkg/mp3"
)

const (
	// GeminiTTS sample rate
	defaultSampleRate = 24000
	// 16-bit signed LE
	bytesPerSample = 2
	// Mono
	channels = 1

	// framePass is the number of samples per encoder pass we align to. The
	// shine-mp3 Write loop consumes 2×samplesPerPass interleaved values per
	// pass; samplesPerPass is GRANULE_SIZE (576) for MPEG-2 (≤24 kHz) and
	// 2×576 for MPEG-1. Padding the mono length to a multiple of 1152 makes
	// the interleaved (×2) buffer a multiple of every possible pass width, so
	// the encoder never drops a trailing block or reads past the buffer.
	framePass = 1152
)

// monoToInterleavedStereo converts mono int16 samples into an interleaved
// dual-mono stereo buffer, zero-padded to a whole number of encoder passes.
//
// This is required because shine-mp3 v0.1.0's Encoder.Write is implemented for
// interleaved stereo only: its pass loop advances 2×samplesPerPass and
// encodeBufferInterleaved reads consecutive samples as L/R pairs. Feeding it a
// mono buffer makes it skip every other block and mis-interpret samples as two
// channels — producing half-length, glitchy audio. Duplicating each sample
// into both channels yields correct full-length output.
func monoToInterleavedStereo(mono []int16) []int16 {
	padded := len(mono)
	if r := padded % framePass; r != 0 {
		padded += framePass - r
	}
	stereo := make([]int16, padded*2)
	for i, s := range mono {
		stereo[2*i] = s
		stereo[2*i+1] = s
	}
	// Samples in [len(mono), padded) stay zero — trailing silence.
	return stereo
}

func decodeMonoPCM(pcmData []byte) ([]int16, error) {
	if len(pcmData) == 0 {
		return nil, fmt.Errorf("empty PCM data")
	}
	if len(pcmData)%bytesPerSample != 0 {
		return nil, fmt.Errorf("PCM data length %d is not aligned to sample size %d", len(pcmData), bytesPerSample)
	}
	numSamples := len(pcmData) / bytesPerSample
	samples := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*bytesPerSample : (i+1)*bytesPerSample]))
	}
	return samples, nil
}

// PCMToMP3 converts raw 16-bit signed LE mono PCM audio to MP3.
// pcmData is the raw PCM bytes, sampleRate is the audio sample rate (e.g. 24000).
func PCMToMP3(pcmData []byte, sampleRate int) ([]byte, error) {
	if sampleRate == 0 {
		sampleRate = defaultSampleRate
	}

	mono, err := decodeMonoPCM(pcmData)
	if err != nil {
		return nil, err
	}
	stereo := monoToInterleavedStereo(mono)

	// Encode as 2-channel: shine-mp3's Write is correct only for interleaved
	// stereo (see monoToInterleavedStereo). The result is dual-mono, which
	// plays identically to mono on every player.
	encoder := mp3.NewEncoder(sampleRate, 2)

	var buf bytes.Buffer
	if err := encoder.Write(&buf, stereo); err != nil {
		return nil, fmt.Errorf("MP3 encoding failed: %w", err)
	}

	return buf.Bytes(), nil
}

// PCMToMP3Writer writes MP3-encoded audio directly to a writer.
func PCMToMP3Writer(w io.Writer, pcmData []byte, sampleRate int) error {
	if sampleRate == 0 {
		sampleRate = defaultSampleRate
	}

	mono, err := decodeMonoPCM(pcmData)
	if err != nil {
		return err
	}
	stereo := monoToInterleavedStereo(mono)

	encoder := mp3.NewEncoder(sampleRate, 2)
	return encoder.Write(w, stereo)
}
