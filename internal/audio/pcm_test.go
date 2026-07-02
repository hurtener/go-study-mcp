package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestMonoToInterleavedStereo(t *testing.T) {
	mono := []int16{1, 2, 3, -4}
	st := monoToInterleavedStereo(mono)

	// Length must be a whole number of encoder passes (2×framePass), or
	// shine-mp3 drops/over-reads the trailing block.
	if len(st)%(framePass*2) != 0 {
		t.Fatalf("stereo length %d is not aligned to a pass (%d)", len(st), framePass*2)
	}
	// Each mono sample is duplicated into L and R.
	for i, v := range mono {
		if st[2*i] != v || st[2*i+1] != v {
			t.Errorf("sample %d not duplicated: L=%d R=%d want %d", i, st[2*i], st[2*i+1], v)
		}
	}
	// Padding is silence.
	for i := len(mono) * 2; i < len(st); i++ {
		if st[i] != 0 {
			t.Errorf("pad sample %d = %d, want 0", i, st[i])
		}
	}
}

// TestPCMToMP3FullLength guards against the shine-mp3 mono regression, where
// Encoder.Write dropped every other block and produced a ~half-length,
// glitchy MP3. A constant-bitrate MP3's size is proportional to its duration,
// so a half-length encode shows up as ~half the expected bytes.
func TestPCMToMP3FullLength(t *testing.T) {
	const rate = 24000
	const n = rate // 1 second of audio

	pcm := make([]byte, n*2)
	for i := 0; i < n; i++ {
		v := int16(6000 * math.Sin(2*math.Pi*220*float64(i)/float64(rate)))
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(v))
	}

	mp3, err := PCMToMP3(pcm, rate)
	if err != nil {
		t.Fatalf("PCMToMP3: %v", err)
	}
	if len(mp3) < 4 || mp3[0] != 0xFF || mp3[1]&0xE0 != 0xE0 {
		t.Fatalf("output does not start with an MPEG frame sync: % x", mp3[:min(4, len(mp3))])
	}

	padded := n
	if r := padded % framePass; r != 0 {
		padded += framePass - r
	}
	// 128 kbps CBR ≈ 16000 bytes/sec.
	want := 16000.0 * float64(padded) / float64(rate)
	got := float64(len(mp3))
	if got < want*0.85 || got > want*1.15 {
		t.Errorf("mp3 size %.0f bytes, want ~%.0f (±15%%); a half-length regression roughly halves this", got, want)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
