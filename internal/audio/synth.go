package audio

import (
	"math"
	"sync"
)

const (
	normalCutoffHz  = 16000.0
	chronosCutoffHz = 800.0
	stepsPerBar     = 16
	barsPerPhrase   = 4
	totalSteps      = stepsPerBar * barsPerPhrase
)

// Am - F - C - G chord roots (bass octave) and 4-note arpeggio voicings.
var chordBassHz = [barsPerPhrase]float64{
	55.00, // A1 (Am)
	43.65, // F1 (F)
	65.41, // C2 (C)
	49.00, // G1 (G)
}

var chordArpHz = [barsPerPhrase][4]float64{
	{220.00, 261.63, 329.63, 440.00}, // Am: A3 C4 E4 A4
	{174.61, 220.00, 261.63, 349.23}, // F:  F3 A3 C4 F4
	{261.63, 329.63, 392.00, 523.25}, // C:  C4 E4 G4 C5
	{196.00, 246.94, 293.66, 392.00}, // G:  G3 B3 D4 G4
}

var chordPadHz = [barsPerPhrase][3]float64{
	{220.00, 329.63, 440.00}, // Am5 + octave
	{174.61, 261.63, 349.23}, // F5 + octave
	{261.63, 392.00, 523.25}, // C5 + octave
	{196.00, 293.66, 392.00}, // G5 + octave
}

var arpPattern = [stepsPerBar]int{
	0, 1, 2, 3, 2, 1, 3, 2,
	0, 2, 1, 3, 2, 3, 1, 2,
}

// synthReader streams procedural 44.1kHz 16-bit LE stereo synthwave PCM with zero
// heap allocations in Read.
type synthReader struct {
	mu sync.Mutex

	// Target music parameters updated from game loop.
	playing       bool
	tier          int
	bossActive    bool
	wantedStars   int
	chronosActive bool
	paused        bool
	targetBPM     float64
	targetCutoff  float64

	// Smoothed synthesis parameters.
	currentBPM    float64
	currentCutoff float64
	arpGain       float64
	padGain       float64
	feverActive   bool
	feverGain     float64
	feverPhase    float64

	// Step sequencer state.
	stepIdx      int
	sampleInStep float64

	// Voice phases (0..1).
	kickPhase   float64
	snarePhase  float64
	bassPhase   float64
	bassSubPh   float64
	arpPhaseL   float64
	arpPhaseR   float64
	padPhaseA   float64
	padPhaseB   float64
	padPhaseHi  float64
	lfoPhase    float64

	// Voice envelopes / elapsed step time (seconds).
	kickEnvT   float64
	snareEnvT  float64
	hatEnvT    float64
	hatDecay   float64
	hatLevel   float64
	bassEnvT   float64
	bassFreq   float64
	bassLP     float64
	arpEnvT    float64
	arpFreq    float64
	sidechain  float64

	// Noise generator state.
	rng       uint32
	prevNoise float64

	// Master one-pole low-pass filter state (stereo).
	lpL float64
	lpR float64

	// Remainder buffer for unaligned io.Reader callers.
	remBuf [4]byte
	remIdx int
	remLen int
}

func newSynthReader() *synthReader {
	r := &synthReader{
		targetBPM:     108.0,
		currentBPM:    108.0,
		targetCutoff:  normalCutoffHz,
		currentCutoff: normalCutoffHz,
		arpGain:       0.12,
		rng:           0x9E3779B9,
		bassFreq:      chordBassHz[0] * 2.0,
		arpFreq:       chordArpHz[0][0],
		sidechain:     1.0,
		kickEnvT:      1.0,
		snareEnvT:     1.0,
		hatEnvT:       1.0,
		hatDecay:      0.03,
	}
	r.triggerStep(0)
	return r
}

func (r *synthReader) setMusicState(playing bool, tier int, bossActive bool, wantedStars int, chronosActive bool, paused bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.playing = playing
	r.tier = tier
	r.bossActive = bossActive
	r.wantedStars = wantedStars
	r.chronosActive = chronosActive
	r.paused = paused

	if !playing || tier <= 1 {
		r.targetBPM = 108.0
	} else if tier == 2 {
		r.targetBPM = 118.0
	} else if tier == 3 {
		r.targetBPM = 128.0
	} else {
		r.targetBPM = 138.0
	}

	if chronosActive || paused {
		r.targetCutoff = chronosCutoffHz
	} else {
		r.targetCutoff = normalCutoffHz
	}
}

func (r *synthReader) setFever(active bool) {
	r.mu.Lock()
	r.feverActive = active
	r.mu.Unlock()
}

func (r *synthReader) bpm() float64 {
	r.mu.Lock()
	b := r.currentBPM
	r.mu.Unlock()
	return b
}

func (r *synthReader) nextNoise() float64 {
	x := r.rng
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.rng = x
	// Map uint32 to [-1, 1].
	return (float64(x) / 2147483647.5) - 1.0
}

func (r *synthReader) triggerStep(step int) {
	bar := (step / stepsPerBar) % barsPerPhrase
	s16 := step % stepsPerBar

	// Kick: beats 1 & 3, plus syncopated pickup when intense.
	if s16 == 0 || s16 == 8 || (r.playing && r.tier >= 2 && s16 == 14) {
		r.kickEnvT = 0.0
		r.kickPhase = 0.0
		r.sidechain = 0.18
	}

	// Snare: beats 2 & 4.
	if s16 == 4 || s16 == 12 {
		r.snareEnvT = 0.0
		r.snarePhase = 0.0
	}

	// Hi-hat: 16th notes in gameplay, gentle 8ths on menu.
	if r.playing || s16%2 == 0 {
		r.hatEnvT = 0.0
		if s16%4 == 2 {
			r.hatLevel = 1.0
			r.hatDecay = 0.045
		} else if s16 == 10 || s16 == 14 {
			r.hatLevel = 0.78
			r.hatDecay = 0.065
		} else {
			r.hatLevel = 0.45
			r.hatDecay = 0.025
		}
	}

	// Bassline: warm 8ths on menu, driving octave 16ths in gameplay.
	root := chordBassHz[bar]
	if !r.playing {
		if s16%2 == 0 {
			r.bassEnvT = 0.0
			r.bassFreq = root * 2.0
		}
	} else {
		r.bassEnvT = 0.0
		if s16%2 == 1 {
			r.bassFreq = root * 2.0
		} else {
			r.bassFreq = root
		}
	}

	// Arpeggio lead.
	arpNote := chordArpHz[bar][arpPattern[s16]]
	if r.playing || s16%2 == 0 {
		r.arpEnvT = 0.0
		r.arpFreq = arpNote
	}
}

func (r *synthReader) nextFrame() (int16, int16) {
	const dt = 1.0 / float64(SampleRate)

	// Smooth tempo, cutoff, and dynamic layer gains.
	r.currentBPM += 0.0008 * (r.targetBPM - r.currentBPM)
	r.currentCutoff += 0.0025 * (r.targetCutoff - r.currentCutoff)

	targetArpGain := 0.10
	if r.playing {
		t := r.tier
		if t < 1 {
			t = 1
		}
		if t > 4 {
			t = 4
		}
		targetArpGain = 0.22 + 0.035*float64(t)
	}
	r.arpGain += 0.0006 * (targetArpGain - r.arpGain)

	targetPadGain := 0.0
	if r.bossActive || r.wantedStars >= 4 {
		targetPadGain = 0.20
	}
	r.padGain += 0.0002 * (targetPadGain - r.padGain)

	targetFeverGain := 0.0
	if r.playing && r.feverActive {
		targetFeverGain = 0.14
	}
	r.feverGain += 0.0005 * (targetFeverGain - r.feverGain)

	// Advance 16th-note step sequencer.
	samplesPerStep := float64(SampleRate) * 60.0 / (r.currentBPM * 4.0)
	r.sampleInStep += 1.0
	if r.sampleInStep >= samplesPerStep {
		r.sampleInStep -= samplesPerStep
		r.stepIdx = (r.stepIdx + 1) % totalSteps
		r.triggerStep(r.stepIdx)
	}

	bar := (r.stepIdx / stepsPerBar) % barsPerPhrase

	// 1. Kick drum: pitch-swept sine + click transient.
	var kick float64
	if r.kickEnvT < 0.28 {
		freq := 46.0 + 115.0*math.Exp(-r.kickEnvT*38.0)
		r.kickPhase += freq * dt
		if r.kickPhase >= 1.0 {
			r.kickPhase -= 1.0
		}
		amp := math.Exp(-r.kickEnvT * 12.5)
		click := 0.0
		if r.kickEnvT < 0.004 {
			click = 0.45 * (1.0 - r.kickEnvT/0.004)
		}
		kick = (math.Sin(2.0*math.Pi*r.kickPhase)*0.9 + click) * amp * 0.62
		r.kickEnvT += dt
	}

	// 2. Snare drum: punchy body + gated noise tail.
	noise := r.nextNoise()
	hpNoise := noise - r.prevNoise
	r.prevNoise = noise

	var snare float64
	if r.snareEnvT < 0.22 {
		bodyFreq := 145.0 + 75.0*math.Exp(-r.snareEnvT*45.0)
		r.snarePhase += bodyFreq * dt
		if r.snarePhase >= 1.0 {
			r.snarePhase -= 1.0
		}
		body := math.Sin(2.0*math.Pi*r.snarePhase) * math.Exp(-r.snareEnvT*28.0) * 0.45
		// Gated reverb feel: broad plateau followed by sharp gate drop.
		noiseGate := math.Exp(-r.snareEnvT * 11.0)
		if r.snareEnvT > 0.14 {
			noiseGate *= math.Exp(-(r.snareEnvT - 0.14) * 55.0)
		}
		snareScale := 0.24
		if r.playing {
			snareScale = 0.44
		}
		snare = (body + (noise*0.65+hpNoise*0.35)*noiseGate*0.55) * snareScale
		r.snareEnvT += dt
	}

	// 3. Hi-hat: filtered high-frequency crisp noise.
	var hat float64
	if r.hatEnvT < 0.12 {
		amp := math.Exp(-r.hatEnvT / r.hatDecay)
		hatLevel := r.hatLevel * 0.16
		if !r.playing {
			hatLevel *= 0.45
		}
		hat = hpNoise * amp * hatLevel
		r.hatEnvT += dt
	}

	// 4. Driving synthwave bassline (saw + sub-square, sidechain pumped).
	r.sidechain += 0.00055 * (1.0 - r.sidechain)
	r.bassPhase += r.bassFreq * dt
	if r.bassPhase >= 1.0 {
		r.bassPhase -= 1.0
	}
	r.bassSubPh += (r.bassFreq * 0.5) * dt
	if r.bassSubPh >= 1.0 {
		r.bassSubPh -= 1.0
	}
	sawBass := 2.0*r.bassPhase - 1.0
	subBass := -1.0
	if r.bassSubPh < 0.5 {
		subBass = 1.0
	}
	bassEnv := math.Exp(-r.bassEnvT * 7.5)
	if !r.playing {
		bassEnv = math.Exp(-r.bassEnvT * 5.0)
	}
	r.bassEnvT += dt
	rawBass := (sawBass*0.72 + subBass*0.28) * bassEnv
	// Subtle analog bass low-pass.
	r.bassLP += 0.16 * (rawBass - r.bassLP)
	bass := r.bassLP * 0.42 * r.sidechain

	// 5. Bright stereo detuned arpeggio lead.
	r.arpPhaseL += (r.arpFreq * 1.0025) * dt
	if r.arpPhaseL >= 1.0 {
		r.arpPhaseL -= 1.0
	}
	r.arpPhaseR += (r.arpFreq * 0.9975) * dt
	if r.arpPhaseR >= 1.0 {
		r.arpPhaseR -= 1.0
	}
	arpSawL := 2.0*r.arpPhaseL - 1.0
	arpSawR := 2.0*r.arpPhaseR - 1.0
	arpTriL := 1.0 - 4.0*math.Abs(r.arpPhaseL-0.5)
	arpTriR := 1.0 - 4.0*math.Abs(r.arpPhaseR-0.5)
	arpEnv := math.Exp(-r.arpEnvT * 11.0)
	r.arpEnvT += dt
	arpL := (arpSawL*0.55 + arpTriL*0.45) * arpEnv * r.arpGain
	arpR := (arpSawR*0.55 + arpTriR*0.45) * arpEnv * r.arpGain

	// 6. Extra tension pad / octave shimmer for boss or high heat (4-5 stars).
	var padL, padR float64
	if r.padGain > 0.001 {
		r.lfoPhase += 2.2 * dt
		if r.lfoPhase >= 1.0 {
			r.lfoPhase -= 1.0
		}
		vib := 1.0 + 0.003*math.Sin(2.0*math.Pi*r.lfoPhase)

		r.padPhaseA += chordPadHz[bar][0] * vib * dt
		if r.padPhaseA >= 1.0 {
			r.padPhaseA -= 1.0
		}
		r.padPhaseB += chordPadHz[bar][1] * 1.003 * dt
		if r.padPhaseB >= 1.0 {
			r.padPhaseB -= 1.0
		}
		r.padPhaseHi += (chordPadHz[bar][2] * 2.0) * 0.997 * dt
		if r.padPhaseHi >= 1.0 {
			r.padPhaseHi -= 1.0
		}
		vA := 2.0*r.padPhaseA - 1.0
		vB := 2.0*r.padPhaseB - 1.0
		vHi := math.Sin(2.0 * math.Pi * r.padPhaseHi)
		padL = (vA*0.45 + vB*0.30 + vHi*0.25) * r.padGain
		padR = (vA*0.30 + vB*0.45 + vHi*0.25) * r.padGain
	}

	// 7. Rush Hour (Streak x5+) octave-up sparkling arpeggio shimmer.
	var feverL, feverR float64
	if r.feverGain > 0.001 {
		r.feverPhase += (r.arpFreq * 2.0) * dt
		if r.feverPhase >= 1.0 {
			r.feverPhase -= 1.0
		}
		tri := 1.0 - 4.0*math.Abs(r.feverPhase-0.5)
		shim := tri * arpEnv * r.feverGain
		feverL = shim * 0.90
		feverR = shim * 1.10
	}

	// Mix dry stereo bus.
	dryL := kick + snare + hat*0.92 + bass + arpL + padL + feverL
	dryR := kick + snare + hat*1.08 + bass + arpR + padR + feverR

	// Master real one-pole low-pass filter: y += alpha * (x - y).
	cutoff := r.currentCutoff
	if cutoff < 80.0 {
		cutoff = 80.0
	}
	rc := 1.0 / (2.0 * math.Pi * cutoff)
	alpha := dt / (rc + dt)
	if alpha > 1.0 {
		alpha = 1.0
	}
	r.lpL += alpha * (dryL - r.lpL)
	r.lpR += alpha * (dryR - r.lpR)

	// Gentle soft-clipping to prevent int16 wraparound.
	outL := math.Tanh(r.lpL * 0.95)
	outR := math.Tanh(r.lpR * 0.95)

	return floatToInt16(outL), floatToInt16(outR)
}

func floatToInt16(v float64) int16 {
	if math.IsNaN(v) {
		return 0
	}
	scaled := v * 32767.0
	if scaled > 32767.0 {
		return 32767
	}
	if scaled < -32768.0 {
		return -32768
	}
	return int16(scaled)
}

// Read fills p with 44.1kHz 16-bit signed little-endian stereo PCM frames.
func (r *synthReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	written := 0

	// Drain any leftover bytes from a previous unaligned Read.
	for r.remIdx < r.remLen && written < len(p) {
		p[written] = r.remBuf[r.remIdx]
		r.remIdx++
		written++
	}
	if r.remIdx >= r.remLen {
		r.remIdx = 0
		r.remLen = 0
	}

	// Write whole 4-byte stereo frames directly into p.
	for written+4 <= len(p) {
		sL, sR := r.nextFrame()
		uL := uint16(sL)
		uR := uint16(sR)
		p[written+0] = byte(uL)
		p[written+1] = byte(uL >> 8)
		p[written+2] = byte(uR)
		p[written+3] = byte(uR >> 8)
		written += 4
	}

	// If caller provided a partial trailing byte count (< 4 bytes), generate 1 frame into remBuf.
	if written < len(p) {
		sL, sR := r.nextFrame()
		uL := uint16(sL)
		uR := uint16(sR)
		r.remBuf[0] = byte(uL)
		r.remBuf[1] = byte(uL >> 8)
		r.remBuf[2] = byte(uR)
		r.remBuf[3] = byte(uR >> 8)
		r.remLen = 4
		r.remIdx = 0
		for written < len(p) {
			p[written] = r.remBuf[r.remIdx]
			r.remIdx++
			written++
		}
	}

	return written, nil
}
