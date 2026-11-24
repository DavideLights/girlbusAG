package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

func decodeLeftChannel(pcm []byte) []int16 {
	frames := len(pcm) / 4
	out := make([]int16, frames)
	for i := 0; i < frames; i++ {
		out[i] = int16(binary.LittleEndian.Uint16(pcm[i*4 : i*4+2]))
	}
	return out
}

func TestSynthReaderProducesValidPCM(t *testing.T) {
	r := newSynthReader()
	buf := make([]byte, 4096)

	modes := []struct {
		name    string
		playing bool
		tier    int
		boss    bool
		stars   int
		chronos bool
		paused  bool
	}{
		{"menu", false, 1, false, 0, false, false},
		{"tier2_play", true, 2, false, 2, false, false},
		{"tier4_boss_chronos", true, 4, true, 5, true, false},
		{"paused", true, 3, false, 1, false, true},
	}

	for _, m := range modes {
		r.setMusicState(m.playing, m.tier, m.boss, m.stars, m.chronos, m.paused)
		nonZero := false
		for block := 0; block < 8; block++ {
			n, err := r.Read(buf)
			if err != nil {
				t.Fatalf("%s: unexpected Read error: %v", m.name, err)
			}
			if n != len(buf) {
				t.Fatalf("%s: expected %d bytes, got %d", m.name, len(buf), n)
			}
			for _, b := range buf {
				if b != 0 {
					nonZero = true
					break
				}
			}
		}
		if !nonZero {
			t.Fatalf("%s: produced all-zero PCM output", m.name)
		}
	}

	// Verify unaligned Read buffer sizes never fail or lose byte continuity.
	for _, sz := range []int{0, 1, 3, 7, 13} {
		small := make([]byte, sz)
		n, err := r.Read(small)
		if err != nil {
			t.Fatalf("small Read(%d) error: %v", sz, err)
		}
		if n != sz {
			t.Fatalf("small Read(%d) got %d", sz, n)
		}
	}

	// Verify NaN/Inf sanitization helper.
	if got := floatToInt16(math.NaN()); got != 0 {
		t.Fatalf("floatToInt16(NaN) = %d, want 0", got)
	}
	if got := floatToInt16(math.Inf(1)); got != 32767 {
		t.Fatalf("floatToInt16(+Inf) = %d, want 32767", got)
	}
	if got := floatToInt16(math.Inf(-1)); got != -32768 {
		t.Fatalf("floatToInt16(-Inf) = %d, want -32768", got)
	}
}

func TestSynthReaderZeroAllocations(t *testing.T) {
	r := newSynthReader()
	r.setMusicState(true, 4, true, 5, false, false)
	buf := make([]byte, 4096)

	allocs := testing.AllocsPerRun(20, func() {
		n, err := r.Read(buf)
		if err != nil || n != len(buf) {
			t.Fatalf("unexpected Read result: n=%d err=%v", n, err)
		}
	})
	if allocs != 0 {
		t.Fatalf("synthReader.Read allocated %v times per call, want 0", allocs)
	}
}

func TestSynthReaderLowPassChronosAttenuation(t *testing.T) {
	rNormal := newSynthReader()
	rChronos := newSynthReader()

	rNormal.setMusicState(true, 3, false, 2, false, false)
	rChronos.setMusicState(true, 3, false, 2, true, false)

	// Settle cutoff smoothing and step state for 0.5 seconds.
	settle := make([]byte, SampleRate*2) // 0.5s * 4 bytes/frame
	if _, err := rNormal.Read(settle); err != nil {
		t.Fatal(err)
	}
	if _, err := rChronos.Read(settle); err != nil {
		t.Fatal(err)
	}

	// Measure 0.5 seconds of PCM output.
	normalBuf := make([]byte, SampleRate*2)
	chronosBuf := make([]byte, SampleRate*2)
	if _, err := rNormal.Read(normalBuf); err != nil {
		t.Fatal(err)
	}
	if _, err := rChronos.Read(chronosBuf); err != nil {
		t.Fatal(err)
	}

	hfRatio := func(samples []int16) float64 {
		var diffSum, ampSum float64
		for i := 1; i < len(samples); i++ {
			diffSum += math.Abs(float64(samples[i]) - float64(samples[i-1]))
			ampSum += math.Abs(float64(samples[i]))
		}
		if ampSum == 0 {
			return 0
		}
		return diffSum / ampSum
	}

	normalHF := hfRatio(decodeLeftChannel(normalBuf))
	chronosHF := hfRatio(decodeLeftChannel(chronosBuf))

	if normalHF <= 0 || chronosHF <= 0 {
		t.Fatalf("expected non-zero HF ratios, got normal=%v chronos=%v", normalHF, chronosHF)
	}
	if chronosHF >= normalHF*0.65 {
		t.Fatalf("expected Chronos 800Hz low-pass to significantly smooth HF energy: normal=%v chronos=%v", normalHF, chronosHF)
	}
}

func TestSynthReaderDynamicBPM(t *testing.T) {
	r := newSynthReader()

	cases := []struct {
		playing bool
		tier    int
		wantBPM float64
	}{
		{false, 4, 108.0},
		{true, 1, 108.0},
		{true, 2, 118.0},
		{true, 3, 128.0},
		{true, 4, 138.0},
		{true, 7, 138.0},
	}

	for _, tc := range cases {
		r.setMusicState(tc.playing, tc.tier, false, 0, false, false)
		if r.targetBPM != tc.wantBPM {
			t.Fatalf("playing=%v tier=%d: targetBPM=%v, want %v", tc.playing, tc.tier, r.targetBPM, tc.wantBPM)
		}
	}
}

func TestSFXBankBuffers(t *testing.T) {
	if sfxCount != 25 {
		t.Fatalf("expected 25 SFXID values, got %d", sfxCount)
	}

	bank := buildSFXBank()
	for id := SFXID(0); id < sfxCount; id++ {
		buf := bank[id]
		if len(buf) == 0 {
			t.Fatalf("SFXID %d produced empty buffer", id)
		}
		if len(buf)%4 != 0 {
			t.Fatalf("SFXID %d buffer length %d is not 4-byte stereo frame aligned", id, len(buf))
		}

		hasSignal := false
		for _, b := range buf {
			if b != 0 {
				hasSignal = true
				break
			}
		}
		if !hasSignal {
			t.Fatalf("SFXID %d buffer is completely silent", id)
		}
	}
}

func TestEngineVolumesAndStateHeadlessSafe(t *testing.T) {
	e := NewEngine()
	if e == nil {
		t.Fatal("NewEngine returned nil")
	}

	// Volume clamping checks.
	e.SetVolumes(-0.5, 1.8, math.NaN())
	if e.masterVol != 0.0 || e.musicVol != 1.0 || e.sfxVol != 0.0 {
		t.Fatalf("unexpected clamped volumes: master=%v music=%v sfx=%v", e.masterVol, e.musicVol, e.sfxVol)
	}

	e.SetVolumes(0.75, 0.5, 0.9)
	if e.masterVol != 0.75 || e.musicVol != 0.5 || e.sfxVol != 0.9 {
		t.Fatalf("unexpected volumes: master=%v music=%v sfx=%v", e.masterVol, e.musicVol, e.sfxVol)
	}

	// Headless UpdateMusicState & PlaySFX before Unlock must never panic.
	e.UpdateMusicState(false, 1, false, 0, false, false)
	e.UpdateMusicState(true, 3, true, 5, true, false)
	e.UpdateMusicState(true, 4, false, 2, false, true)

	if e.synth.targetCutoff != chronosCutoffHz {
		t.Fatalf("expected paused state to engage low-pass cutoff %v, got %v", chronosCutoffHz, e.synth.targetCutoff)
	}

	e.PlaySFX(SFXThrow)
	e.PlaySFX(SFXID(-1))
	e.PlaySFX(sfxCount)
}

func TestSynthReaderFeverZeroAllocationsAndBPM(t *testing.T) {
	e := NewEngine()
	e.UpdateMusicState(true, 3, false, 2, false, false)
	e.SetFever(true)
	buf := make([]byte, 4096)

	// Settle fever gain so shimmering path executes inside Read.
	for i := 0; i < 8; i++ {
		if _, err := e.synth.Read(buf); err != nil {
			t.Fatalf("settle Read failed: %v", err)
		}
	}

	allocs := testing.AllocsPerRun(20, func() {
		n, err := e.synth.Read(buf)
		if err != nil || n != len(buf) {
			t.Fatalf("unexpected Read result: n=%d err=%v", n, err)
		}
	})
	if allocs != 0 {
		t.Fatalf("fever synthReader.Read allocated %v times per call, want 0", allocs)
	}

	bpm := e.CurrentBPM()
	if bpm < 108.0 || bpm > 138.0 {
		t.Fatalf("unexpected CurrentBPM = %v, want within [108, 138]", bpm)
	}
}
