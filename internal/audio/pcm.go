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
)

// PCMToMP3 converts raw 16-bit signed LE PCM audio to MP3.
// pcmData is the raw PCM bytes, sampleRate is the audio sample rate (e.g. 24000).
func PCMToMP3(pcmData []byte, sampleRate int) ([]byte, error) {
	if sampleRate == 0 {
		sampleRate = defaultSampleRate
	}

	if len(pcmData) == 0 {
		return nil, fmt.Errorf("empty PCM data")
	}
	if len(pcmData)%bytesPerSample != 0 {
		return nil, fmt.Errorf("PCM data length %d is not aligned to sample size %d", len(pcmData), bytesPerSample)
	}

	// Convert bytes to int16 samples
	numSamples := len(pcmData) / bytesPerSample
	samples := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*bytesPerSample : (i+1)*bytesPerSample]))
	}

	// Create MP3 encoder
	encoder := mp3.NewEncoder(sampleRate, channels)

	// Encode to MP3
	var buf bytes.Buffer
	if err := encoder.Write(&buf, samples); err != nil {
		return nil, fmt.Errorf("MP3 encoding failed: %w", err)
	}

	return buf.Bytes(), nil
}

// PCMToMP3Writer writes MP3-encoded audio directly to a writer.
func PCMToMP3Writer(w io.Writer, pcmData []byte, sampleRate int) error {
	if sampleRate == 0 {
		sampleRate = defaultSampleRate
	}

	if len(pcmData) == 0 {
		return fmt.Errorf("empty PCM data")
	}

	numSamples := len(pcmData) / bytesPerSample
	samples := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*bytesPerSample : (i+1)*bytesPerSample]))
	}

	encoder := mp3.NewEncoder(sampleRate, channels)
	return encoder.Write(w, samples)
}
